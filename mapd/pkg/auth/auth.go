// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dexidp/dex/api/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	echo "github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/identicon"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/randutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

const (
	signatureKey = "secret:oidc"
	clientID     = "b56dffd5-b1b0-47ef-a24c-421fdcd3fd53"
	adminID      = "1dda94c3-93b0-46b1-8164-73e621e8cb90"
	adminEmail   = "admin@tanzu.local"

	sessionKeyState       = "state"
	sessionKeyRedirectURI = "redirect_uri"

	sessionCookieSID  = "_sid"
	sessionCookieSSID = "_ssid"
)

// Auth is an OIDC authenticator.
type Auth struct {
	idp *oidc.Provider
	cli api.DexClient
	cfg *oauth2.Config
	lsm *httputil.SessionManager // login session
	lim *Limiter
	log *zap.SugaredLogger
}

func (o *Auth) DexClient() api.DexClient {
	return o.cli
}

// Login handles `GET /auth/.well-known/login` requests.
func (o *Auth) Login(ctx echo.Context) error {
	if err := o.lim.IsBanned(ctx); err != nil {
		return ctx.String(http.StatusForbidden, "error: access prohibited")
	}

	var thisHost string
	if !strings.ContainsRune(ctx.Request().Host, ':') {
		thisHost = ctx.Request().Host
	} else {
		var err error
		thisHost, _, err = net.SplitHostPort(ctx.Request().Host)
		if err != nil {
			panic(err)
		}
	}
	if thisHost != mapd.GlobalConfig.FQDN {
		return ctx.String(http.StatusUnprocessableEntity, "error: unrecognized host")
	}

	req := &portalRequest{}
	if err := ctx.Bind(req); err != nil {
		o.lim.Fail(ctx)
		return err
	}
	if req.RedirectURI != "" {
		u, err := url.Parse(req.RedirectURI)
		if err != nil {
			o.lim.Fail(ctx)
			return ctx.String(http.StatusUnprocessableEntity, "error: invalid redirect URI")
		}

		if u.Host != "" {
			switch u.Scheme {
			case "http", "https":
			default:
				o.lim.Fail(ctx)
				return ctx.String(http.StatusUnprocessableEntity, "error: invalid redirect URI")
			}

			var host string
			if !strings.ContainsRune(u.Host, ':') {
				host = u.Host
			} else {
				host, _, err = net.SplitHostPort(u.Host)
				if err != nil {
					o.lim.Fail(ctx)
					return ctx.String(http.StatusUnprocessableEntity, "error: invalid redirect URI")
				}
			}

			if mapd.GlobalConfig.FQDN != host {
				o.lim.Fail(ctx)
				return ctx.String(http.StatusUnprocessableEntity, "error: invalid redirect URI")
			}
		}
	}

	var lsess httputil.SessionStore

	rtoken, err := GetSecuredCookie(ctx, sessionCookieSID, true)
	if err == nil {
		if !req.NewSession {
			// authorized, renew session
			lsess, err = o.lsm.Start(context.TODO(), ctx.Response().Writer, ctx.Request())
			if err != nil {
				return err
			}
			defer lsess.Save()

			t := &oauth2.Token{
				RefreshToken: rtoken,
				Expiry:       time.Now().Add(-time.Hour),
			}
			token, err := o.cfg.TokenSource(ctx.Request().Context(), t).Token()
			if err == nil {
				return o.renderRedirectPage(ctx, lsess, req.RedirectURI, token)
			}

			// user has logged into another client, current session is not valid anymore.
		}
		// delete current session
		o.cleanupCookies(ctx)
	}

	if lsess == nil {
		// not authorized
		lsess, err = o.lsm.Start(context.TODO(), ctx.Response().Writer, ctx.Request())
		if err != nil {
			return err
		}
		defer lsess.Save()
	}

	state := uuid.NewString()
	lsess.Set(sessionKeyState, state)
	lsess.Set(sessionKeyRedirectURI, req.RedirectURI)

	return ctx.Redirect(http.StatusSeeOther, o.cfg.AuthCodeURL(state, oauth2.AccessTypeOffline))
}

// Logout handles `GET /auth/.well-known/logout` requests.
func (o *Auth) Logout(ctx echo.Context) error {
	lsess, err := o.lsm.Start(context.TODO(), ctx.Response().Writer, ctx.Request())
	if err != nil {
		return err
	}
	defer lsess.Save()

	lsess.Flush()

	o.cleanupCookies(ctx)
	return nil
}

func (o *Auth) cleanupCookies(ctx echo.Context) {
	past := time.Unix(0, 0)
	SetSecuredCookie(ctx, sessionCookieSID, "", past)
	SetBrowserReadableSecuredCookie(ctx, sessionCookieSSID, "", past)
}

// Callback handles `GET /auth/.well-known/callback` requests. This is used for
// redirecting identities from IDP to users (browsers).
func (o *Auth) Callback(ctx echo.Context) error {
	lsess, err := o.lsm.Start(context.TODO(), ctx.Response().Writer, ctx.Request())
	if err != nil {
		return err
	}
	defer lsess.Save()

	req := &oidcCallbackRequest{}
	if err := ctx.Bind(req); err != nil {
		return err
	}

	// security assurance, according to:
	//
	// - https://openid.net/specs/openid-connect-core-1_0.html#ImplicitAuthResponse
	prevState, ok := lsess.Get(sessionKeyState)
	if !ok {
		return echo.NewHTTPError(http.StatusForbidden)
	}
	prevStateStr, ok := prevState.(string)
	if !ok || prevStateStr != req.State {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// inject custom http client into context
	cli := httputil.NewClient()
	cTx := context.WithValue(ctx.Request().Context(), oauth2.HTTPClient, cli.Client)
	token, err := o.cfg.Exchange(cTx, req.Code)
	if err != nil {
		panic(err)
	}

	var redirectURI string
	redirectURIIntf, ok := lsess.Get(sessionKeyRedirectURI)
	if ok {
		redirectURIStr, _ := redirectURIIntf.(string)
		if redirectURIStr != "" {
			redirectURI = redirectURIStr
		}
	}

	return o.renderRedirectPage(ctx, lsess, redirectURI, token)
}

// renderRedirectPage renders a page for redirecting browser to a new location.
// This is called when user is successfully authenticated.
func (o *Auth) renderRedirectPage(ctx echo.Context, sessStore httputil.SessionStore, uri string, token *oauth2.Token) error {
	var redirectURI *url.URL
	if uri != "" {
		redirectURI, _ = url.Parse(uri)
	}
	if redirectURI == nil {
		redirectURI = &url.URL{}
	}
	if redirectURI.Scheme == "" {
		redirectURI.Scheme = ctx.Request().URL.Scheme
	}
	if redirectURI.Host == "" {
		redirectURI.Scheme = ctx.Request().URL.Host
	}
	if redirectURI.Path == "" {
		redirectURI.Path = "/"
	}

	return o.finalizeAuthorization(ctx, sessStore, token, redirectURI)
}

func (o *Auth) ValidateToken(ctx context.Context, token string) (*oidc.IDToken, error) {
	cfg := &oidc.Config{ClientID: o.cfg.ClientID}
	if mapd.GlobalConfig.Auth.IDPCrossClientTokenSharing {
		cfg.ClientID = ""
		cfg.SkipClientIDCheck = true
	}
	return o.idp.Verifier(cfg).Verify(ctx, token)
}

func (o *Auth) finalizeAuthorization(ctx echo.Context, sessStore httputil.SessionStore, token *oauth2.Token, redirectURI *url.URL) error {
	idToken, err := o.ValidateToken(ctx.Request().Context(), token.AccessToken)
	if err != nil {
		panic(err)
	}

	userInfo, err := o.BindUserForTokenOnDemand(sessStore, idToken, true)
	if err != nil {
		panic(err)
	}

	if mapd.GlobalConfig.Auth.DisableLocalUser && userInfo.Managed {
		return echo.NewHTTPError(http.StatusForbidden, "local-managed user has been disabled.")
	}

	// deprecated: this is only used for frontend to initialize views.
	if err := o.initializeUserInfoCookie(ctx, userInfo, idToken.Expiry); err != nil {
		panic(err)
	}

	RuntimeSessionCache.OnLogin(uuid.MustParse(userInfo.ID), userInfo.Name)
	// save identity to cookie
	SetSecuredCookie(ctx, sessionCookieSID, token.RefreshToken, time.Time{})
	SetBrowserReadableSecuredCookie(ctx, sessionCookieSSID, token.AccessToken, token.Expiry)

	tmpl, err := template.New("main").Parse(callbackPageTemplate)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, redirectURI.String()); err != nil {
		panic(err)
	}
	return ctx.HTML(http.StatusAccepted, buf.String())
}

func (o *Auth) initializeUserInfoCookie(ctx echo.Context, userInfo *UserInfo, tokenExpiry time.Time) error {
	info := &UserInfo{
		ID:       userInfo.ID,
		Name:     userInfo.Name,
		FullName: userInfo.FullName,
		Email:    userInfo.Email,
		Remark:   userInfo.Remark,
		Admin:    userInfo.Admin,
		Managed:  userInfo.Managed,
	}
	if len(userInfo.Organizations) > 0 {
		// only preserve the first one as the default organization
		info.DefaultOrgnization = userInfo.Organizations[0]
	}

	b, err := json.Marshal(info)
	if err != nil {
		return errors.Wrap(err, "failed to save user info to cookie")
	}
	SetBrowserReadableSecuredCookie(ctx, "__user_info", string(b), tokenExpiry)
	return nil
}

// BindUserForTokenOnDemand caches a user's identity into session store on demand.
func (o *Auth) BindUserForTokenOnDemand(sessStore httputil.SessionStore, token *oidc.IDToken, force bool) (*UserInfo, error) {
	if !force {
		info, err := RuntimeSessionCache.Get(token)
		if err != nil {
			panic(err)
		}

		if info != nil {
			var claims JWTClaims
			if err := token.Claims(&claims); err != nil {
				return nil, errors.Wrap(err, "failed to extract token claims")
			}
			if claims.Email == info.Email {
				return info, nil
			}
		}
	}

	info, err := o.BindUserForToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to bind user info")
	}
	RuntimeSessionCache.Set(token, info)
	return info, nil
}

func (o *Auth) BindUserForToken(token *oidc.IDToken) (*UserInfo, error) {
	if len(token.Audience) < 1 {
		return nil, fmt.Errorf("no audience is available in token")
	}

	var (
		info   UserInfo
		claims JWTClaims
	)
	if err := token.Claims(&claims); err != nil {
		return nil, errors.Wrap(err, "failed to extract token claims")
	}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		var binding model.UserBinding
		if err := tx.Where("`email` = ?", claims.Email).Take(&binding).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			binding.ID = uuid.New()
			binding.FullName = claims.Name
			binding.Email = claims.Email
			binding.Remark = claims.PreferredUsername

			if binding.Remark != "" {
				binding.Name = binding.Remark
			} else {
				ss := strings.Fields(binding.FullName)
				for i := range ss {
					if textutil.ContainsCJK([]byte(ss[i])) {
						return fmt.Errorf("username must not contains CJK characters: %s", ss[i])
					}
					ss[i] = strings.ToLower(ss[i])
				}
				binding.Name = strings.Join(ss, "-")
			}

			if err := tx.Create(&binding).Error; err != nil {
				return err
			}
			if err := tx.Take(&binding).Error; err != nil {
				return err
			}
		}

		info.ID = binding.ID.String()
		info.Name = binding.Name
		info.FullName = binding.FullName
		info.Email = binding.Email
		info.Remark = binding.Remark
		info.Admin = binding.Admin
		info.Managed = binding.Managed
		info.Role = make(map[string]model.Role)
		info.Perm = make(map[string]*SessionPerm)

		var (
			orgs         []model.Organization
			roleBindings []model.RoleBinding
		)

		db := tx
		if !info.Admin {
			db = db.Where("`id` IN (SELECT `organization_id` FROM `role_binding` WHERE `user_binding_id` = ?)", info.ID)
		}
		if err := db.Find(&orgs).Error; err != nil {
			return err
		}
		for _, each := range orgs {
			info.Organizations = append(info.Organizations, &Organization{
				ID:   each.ID.String(),
				Name: each.Name,
			})
		}
		sort.Slice(info.Organizations, func(i, j int) bool { return info.Organizations[i].Name < info.Organizations[j].Name })

		if err := tx.Where("`user_binding_id` = ?", binding.ID).Find(&roleBindings).Error; err != nil {
			return err
		}
		for _, each := range roleBindings {
			policies := model.GetRoleBindingPolicies(each.Role, binding.Admin)
			// user does not have any permission by default
			sort.Slice(policies, func(i, j int) bool {
				return policies[i].IsReversed() && !policies[j].IsReversed()
			})
			// 255 has higher priority than 0
			sort.SliceStable(policies, func(i, j int) bool {
				return policies[i].AbsPriority() < policies[j].AbsPriority()
			})

			perm := &SessionPerm{}
			for _, each := range policies {
				if each.Perm.Has(model.PermReadOnly) {
					if each.IsReversed() {
						perm.R.Remove(each.Object)
					} else {
						perm.R.Add(each.Object)
					}
				}
				if each.Perm.Has(model.PermWriteOnly) {
					if each.IsReversed() {
						perm.W.Remove(each.Object)
					} else {
						perm.W.Add(each.Object)
					}
				}
				if each.Perm.Has(model.PermExecOnly) {
					if each.IsReversed() {
						perm.X.Remove(each.Object)
					} else {
						perm.X.Add(each.Object)
					}
				}
				if each.Perm.Has(model.PermDeleteOnly) {
					if each.IsReversed() {
						perm.D.Remove(each.Object)
					} else {
						perm.D.Add(each.Object)
					}
				}
			}

			info.Role[each.OrganizationID.String()] = each.Role.Normalize()
			info.Perm[each.OrganizationID.String()] = perm
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &info, nil
}

// Middleware exports Auth as an API auth guard.
func (o *Auth) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			str := ctx.Request().Header.Get("Authorization")
			if str == "" {
				return echo.ErrUnauthorized
			}
			ss := strings.Fields(strings.TrimSpace(str))
			if len(ss) < 2 || ss[0] != "Bearer" {
				return echo.ErrUnauthorized
			}

			idToken, err := o.ValidateToken(ctx.Request().Context(), ss[1])
			if err != nil {
				return echo.ErrUnauthorized
			}
			var claims JWTClaims
			if err := idToken.Claims(&claims); err != nil {
				return echo.ErrUnauthorized
			}

			lsess, err := o.lsm.Start(context.TODO(), ctx.Response().Writer, ctx.Request())
			if err != nil {
				return err
			}
			defer lsess.Save()

			info, err := o.BindUserForTokenOnDemand(lsess, idToken, false)
			if err != nil {
				panic(err)
			}

			if mapd.GlobalConfig.Auth.DisableLocalUser && info.Managed {
				return echo.ErrForbidden
			}

			uid := uuid.MustParse(info.ID)
			ip, _, _ := net.SplitHostPort(ctx.Request().RemoteAddr)
			sess := &SessionInfo{
				UserID:   uid,
				UserName: info.Name,
				RemoteIP: ip,
				Admin:    info.Admin,
				Perm:     make(map[uuid.UUID]*SessionPerm),
			}
			for k, v := range info.Perm {
				org, err := uuid.Parse(k)
				if err != nil {
					continue
				}
				sess.Perm[org] = &SessionPerm{
					R: v.R,
					W: v.W,
					X: v.X,
					D: v.D,
				}
			}

			RuntimeSessionCache.OnMiddleware(uid, info.Name)

			ctx.SetRequest(ctx.Request().WithContext(Context(ctx.Request().Context(), sess)))
			if err := next(ctx); err != nil {
				ctx.Error(err)
			}
			return nil
		}
	}
}

func (o *Auth) GraphQLWebsocketHandler(srv *handler.Server) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().Header.Get("Upgrade") == "" {
			return echo.ErrBadRequest
		}

		srv.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	}
}

func (o *Auth) GraphQLWebsocketConfig() transport.Websocket {
	return transport.Websocket{
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			if initPayload == nil {
				return ctx, fmt.Errorf("unauthorized")
			}
			autho := initPayload["Authorization"]
			if autho == nil {
				return ctx, fmt.Errorf("unauthorized")
			}
			str, ok := autho.(string)
			if !ok {
				return ctx, fmt.Errorf("unauthorized")
			}

			ss := strings.Fields(strings.TrimSpace(str))
			if len(ss) < 2 || ss[0] != "Bearer" {
				return ctx, fmt.Errorf("unauthorized")
			}

			idToken, err := o.ValidateToken(ctx, ss[1])
			if err != nil {
				return ctx, fmt.Errorf("unauthorized")
			}
			var claims JWTClaims
			if err := idToken.Claims(&claims); err != nil {
				return ctx, fmt.Errorf("unauthorized")
			}

			userInfo, err := RuntimeSessionCache.Get(idToken)
			if err != nil {
				panic(err)
			}
			if userInfo == nil || userInfo.Email != claims.Email {
				userInfo, err = o.BindUserForToken(idToken)
				if err != nil {
					panic(err)
				}
				RuntimeSessionCache.Set(idToken, userInfo)
			}

			if mapd.GlobalConfig.Auth.DisableLocalUser && userInfo.Managed {
				return ctx, fmt.Errorf("prohibited")
			}

			uid := uuid.MustParse(userInfo.ID)
			sess := &SessionInfo{
				UserID:   uid,
				UserName: userInfo.Name,
				Admin:    userInfo.Admin,
				Perm:     make(map[uuid.UUID]*SessionPerm),
			}
			for k, v := range userInfo.Perm {
				org, err := uuid.Parse(k)
				if err != nil {
					continue
				}
				sess.Perm[org] = &SessionPerm{
					R: v.R,
					W: v.W,
					X: v.X,
					D: v.D,
				}
			}

			RuntimeSessionCache.OnMiddleware(uid, userInfo.Name)

			return Context(ctx, sess), nil
		},
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		KeepAlivePingInterval: 10 * time.Second,
	}
}

// NewAuth registers route into given route group.
func NewAuth(group *echo.Group, logger *zap.SugaredLogger) (*Auth, error) {
	ctx := oidc.ClientContext(context.Background(), httputil.NewClient().Client)
	provider, err := oidc.NewProvider(ctx, mapd.GlobalConfig.Auth.IDPIssuer)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query identity provider %q", mapd.GlobalConfig.Auth.IDPIssuer)
	}

	var clientCfg *tls.Config
	if !mapd.GlobalConfig.Auth.IDPDexAPIInsecure {
		caPool, _ := x509.SystemCertPool()
		if caPool == nil {
			caPool = x509.NewCertPool()
		}
		if len(mapd.GlobalConfig.Auth.IDPCACert) > 0 {
			b, err := os.ReadFile(mapd.GlobalConfig.Auth.IDPCACert)
			if err != nil {
				return nil, errors.Wrap(err, "failed to load CA certificate")
			}
			if !caPool.AppendCertsFromPEM(b) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
		}
		clientCfg = &tls.Config{
			RootCAs:    caPool,
			MinVersion: tls.VersionTLS12,
		}
		if mapd.GlobalConfig.Auth.IDPDexAPICert != "" && mapd.GlobalConfig.Auth.IDPDexAPIKey != "" {
			clientCert, err := tls.LoadX509KeyPair(mapd.GlobalConfig.Auth.IDPDexAPICert, mapd.GlobalConfig.Auth.IDPDexAPIKey)
			if err != nil {
				return nil, errors.Wrap(err, "error loading TLS client key pair")
			}

			clientCfg.Certificates = append(clientCfg.Certificates, clientCert)
		}
	}

	dex, err := newDexClient(mapd.GlobalConfig.Auth.IDPDexAPI, clientCfg)
	if err != nil {
		return nil, errors.Wrap(err, "error initialize dex client")
	}

	oauth2Cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: randutil.StringN(64),
		RedirectURL:  mapd.GlobalConfig.GetCallbackURI().String(),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups", "offline_access"},
	}
	if err := initDexData(dex, oauth2Cfg); err != nil {
		return nil, errors.Wrap(err, "error initialize dex data")
	}

	auth := &Auth{
		idp: provider,
		cli: dex,
		cfg: oauth2Cfg,
		lsm: httputil.NewManager(
			httputil.SetDomain(mapd.GlobalConfig.FQDN),
			httputil.SetCookieName("_lsid"),
		),
		lim: NewLimiter(),
	}
	group.GET("/login", auth.Login)
	group.GET("/logout", auth.Logout)
	group.GET("/callback", auth.Callback)
	return auth, nil
}

func newDexClient(hostAndPort string, clientCfg *tls.Config) (api.DexClient, error) {
	var creds credentials.TransportCredentials
	if clientCfg != nil {
		creds = credentials.NewTLS(clientCfg)
	} else {
		creds = insecure.NewCredentials()
	}
	conn, err := grpc.Dial(hostAndPort, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("error establishing gRPC connection with %s: %v", hostAndPort, err)
	}
	return api.NewDexClient(conn), nil
}

func initDexData(dex api.DexClient, cfg *oauth2.Config) error {
	if err := initDexClient(dex, cfg); err != nil {
		return errors.Wrap(err, "failed to initialize OpenID client for Dex")
	}
	if err := initAdminAndDefaultOrg(dex, mapd.GlobalConfig.Auth.AdminInitialPassword); err != nil {
		return errors.Wrap(err, "failed to initialize password for admin")
	}
	return nil
}

func initDexClient(dex api.DexClient, cfg *oauth2.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	client := &api.Client{
		Id:           cfg.ClientID,
		Secret:       cfg.ClientSecret,
		Name:         "vmware-tanzu-map",
		RedirectUris: []string{cfg.RedirectURL},
	}

	// Try to delete first, ignore error if client doesn't exist or is static (read-only)
	_, err := dex.DeleteClient(ctx, &api.DeleteClientReq{Id: cfg.ClientID})
	if err != nil {
		// Check if it's a read-only static client error, if so skip deletion
		errMsg := err.Error()
		if strings.Contains(errMsg, "read-only") || strings.Contains(errMsg, "static") {
			// Client is static/read-only, skip deletion and just try to create
		} else {
			return errors.Wrap(err, "failed to remove OIDC client")
		}
	}

	// Create the client, ignore "already exists" error for static clients
	_, err = dex.CreateClient(ctx, &api.CreateClientReq{Client: client})
	if err != nil {
		// If already exists, that's fine (for static clients)
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return errors.Wrap(err, "failed to create OIDC client")
	}
	return nil
}

func initAdminAndDefaultOrg(dex api.DexClient, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	newPassword := password
	if newPassword == "" {
		newPassword = "vmware"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		org := &model.Organization{}
		if err := tx.Where("`name` = ?", "default").Take(org).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.Wrap(err, "failed to get default organization")
			}

			org.ID = uuid.New()
			org.Name = "default"
			org.Description = "Default organization."
			org.Avatar = identicon.NewDataURL(org.ID[:]).String()
			if err := tx.Create(org).Error; err != nil {
				return errors.Wrap(err, "failed to create default organization")
			}

			quota := &model.OrganizationQuota{ID: uuid.New(), OrganizationID: model.NewNullUUID(org.ID)}
			if err := tx.Create(quota).Error; err != nil {
				return errors.Wrap(err, "failed to create quota for default organization")
			}
		}

		if !mapd.GlobalConfig.Auth.DisableLocalUser {
			binding := &model.UserBinding{ID: uuid.MustParse(adminID)}
			if err := tx.Take(binding).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				binding = &model.UserBinding{
					ID:       uuid.MustParse(adminID),
					Name:     "admin",
					FullName: "Admin User",
					Email:    adminEmail,
					Admin:    true,
					Managed:  true,
				}
				if err := tx.Create(binding).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(binding).Updates(map[string]interface{}{
					"name":      "admin",
					"full_name": "Admin User",
					"email":     adminEmail,
					"admin":     true,
					"managed":   true,
				}).Error; err != nil {
					return err
				}
			}

			pwd := &api.Password{
				Email:    adminEmail,
				Hash:     hash,
				Username: "admin",
				UserId:   adminID,
			}
			// this call automatically ignores the error if the user exists.
			_, err := dex.CreatePassword(ctx, &api.CreatePasswordReq{Password: pwd})
			if err != nil {
				return errors.Wrap(err, "failed to create admin user")
			}

			roleBinding := &model.RoleBinding{}
			if err := tx.Where("`user_binding_id` = ? AND `organization_id` = ?", binding.ID, org.ID).Take(roleBinding).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.Wrap(err, "failed to get default organization")
				}

				roleBinding.ID = uuid.New()
				roleBinding.Role = model.RoleAdmin
				roleBinding.UserBindingID = model.NewNullUUID(binding.ID)
				roleBinding.OrganizationID = model.NewNullUUID(org.ID)
				if err := tx.Create(roleBinding).Error; err != nil {
					return errors.Wrap(err, "failed to add admin to default organization as a member")
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

type portalRequest struct {
	// RedirectURI is used for frontend to redirect page back
	RedirectURI string `json:"redirectURI" form:"redirectURI" query:"redirectURI"`
	// NewSession enforces to create a new login session
	NewSession bool `json:"newSession" form:"newSession" query:"newSession"`
}

func (r portalRequest) String() string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(r)
	return buf.String()
}

type oidcCallbackRequest struct {
	Error string `json:"error" form:"error" query:"error"`
	Code  string `json:"code" form:"code" query:"code"`
	State string `json:"state" form:"state" query:"state"`
}

const callbackPageTemplate = `<!doctype html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="icon" type="image/x-icon" href="data:image/png;base64,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AAAAAAAAAAAAANqRAADakQAA2pEAJdqRAN7akQBNtm0MAIpCGoCKQhrGikIaDIpCGgCKQhoAikIaAIpCGoCKQhrIfzgdDN2UABfakQDX2pEAcdqRAADakQAA2pEAANqRAADZkwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA2JAAANqRAADakQAA2pEAANqRAADakQBb2pEA39uSACNLBC8DikIatIpCGpeKQhoAikIaAIpCGgCKQhoAikIaSYpCGuKKQhotAACGANqRAJPakQDn2pEAjNqRACrakgAA2pEAANqRAADakQAA3ZEAAAAAAAAAAAAA//8AANqRAADakQAA2pEAANqRAADakQAj2pEAgdqRAOHakQCj960ABYlBGhuKQhrZikIaYYpCGgCKQhoAAAAAAIpCGgCKQhofikIa1YlBGkqscw8A2pEAEdqRAGnakQDN2pEA39qRAIzakQAp2o8AANqRAADakQAA2pEAANqRAADakQAA2pEAANqRAADakQAl2pEAg9qRANvakQDS2pEActqRABavdw0AiD8bM4pCGtqKQhoyikIaAIhDGQAAAAAAikIaAIpCGgiJQRpKrn8OHtfEAB/biAAA2KYAANqRABbakQBq2pEAztqRAN/akQCL2pEAKduRAADakQAA2pEAANuRAADakQAm2pEAhdqRANzakQDR2pEAcdqRABrYpgAA16oAANfDABy7lQkciUEaS4pCGg+KQhoAAAAAAAAAAACKQhoArX0OAII0HQDWwgBu1cEA29XBAIzVwQAp2ZwAANimAADakQAW2pEAatqRAM7akQDf2pEAitqRACnakQAo2pEAiNqRAN3akQDQ2pEAb9qRABnYpgAA16kAANXBACLVwQB/1cEA2NXCAIf//wACtIoLAIpCGgAAAAAAAAAAAIpDGgDRugEA0bsCANXBABjVwQBq1cEAzdXBAN/VwQCL1cEAKdmeAADYpgAA2pEAFtqRAGvakQDO2pEA4dqRAODakQDP2pEAbdqRABjYpgAA2KgAANXBACPVwQCB1cEA2tXBANbVwQB41cEAINbCAAHSvAEAi0IZAAAAAAAAAAAAAAAAANXBAADVwQAA1cEAANXBAADVwQAW1cEAatXBAM7VwQDf1cEAi9XBACnZnwAA2KYAANqRABfakQBj2pEAYtqRABfYpgAA2KcAANXBACTVwQCD1cEA29XBANXVwQB21cEAHdXBAADVwQAA1cEAANXBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADWwQAA0cQAANXBAADVwQAA1cEAFtXBAGrVwQDO1cEA39XBAIvVwQAp2J8AANmfAADZnwAA2KUAANXBACXVwQCF1cEA3NXBANPVwQB01cEAHNXBAADVwQAA1cEAANXBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA1cEAANXDAADVwQAA1cEAANXBABbVwQBr1cEAztXBAN/VwQCK1cEAGdXBABjVwQCG1cEA3dXBANLVwQBy1cEAGtXBAADVwQAA1cEAANXBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANXBAADUwwAA1cEAANXBAADVwQAW1cEAatXBAMzVwQBV1cEAVdXBAM/VwQBv1cEAGdXBAADVwQAA1cEAANXBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVwQAA0MQAANXBAADVwQAA1cEAG9XBABrVwQAb1cEAHdXBAADVwQAA1cIAANXBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA+AAAP/gAAB/wAAAP4AAAB8AAAAOAAAADgAAAAQAH4AAAD/AAAB/4AAAf+AAAP/wAAH/+AAD//wAA//8AAP//AAB//gAAf/4AAH/+AAB//gAAH/wAAAfgAAABgACAAAAAgAAAAYAAAAGAAAABwAAAA/AAAA/8AAA//wAA///AA/8="><title>Modern Application Platform</title><meta name="apple-mobile-web-app-title" content="Modern Application Platform"><meta http-equiv="refresh" content="2; URL={{.}}"><style>html,body{height:100%}body{display:flex;align-items:center;justify-content:center}.spinner{-webkit-animation:rotator 1.4s linear infinite;animation:rotator 1.4s linear infinite}@-webkit-keyframes rotator{0%{transform:rotate(0deg)}100%{transform:rotate(270deg)}}@keyframes rotator{0%{transform:rotate(0deg)}100%{transform:rotate(270deg)}}.path{stroke-dasharray:187;stroke-dashoffset:0;transform-origin:center;-webkit-animation:dash 1.4s ease-in-out infinite,colors 5.6s ease-in-out infinite;animation:dash 1.4s ease-in-out infinite,colors 5.6s ease-in-out infinite}@-webkit-keyframes colors{0%{stroke:#4285F4}25%{stroke:#DE3E35}50%{stroke:#F7C223}75%{stroke:#1B9A59}100%{stroke:#4285F4}}@keyframes colors{0%{stroke:#4285F4}25%{stroke:#DE3E35}50%{stroke:#F7C223}75%{stroke:#1B9A59}100%{stroke:#4285F4}}@-webkit-keyframes dash{0%{stroke-dashoffset:187}50%{stroke-dashoffset:46.75;transform:rotate(135deg)}100%{stroke-dashoffset:187;transform:rotate(450deg)}}@keyframes dash{0%{stroke-dashoffset:187}50%{stroke-dashoffset:46.75;transform:rotate(135deg)}100%{stroke-dashoffset:187;transform:rotate(450deg)}}</style></head><body translate="no"><svg class="spinner" width="65px" height="65px" viewBox="0 0 66 66" xmlns="http://www.w3.org/2000/svg"><circle class="path" fill="none" stroke-width="6" stroke-linecap="round" cx="33" cy="33" r="30"></circle></svg></body></html>`
