// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package submodule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	gql "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/resolver"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"go.uber.org/zap"
)

var (
	APIRouter = Router.Group("/api")
)

type noOpWriter struct{}

func (*noOpWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// newNoOpWriter returns a writer that will never write.
func newNoOpWriter() io.Writer {
	return &noOpWriter{}
}

const (
	gqlUploadMaxSize   = 4 * util.Gi  // 4GiB
	gqlUploadMaxMemory = 16 * util.Gi // 16GiB
)

type API struct {
	status APIStatus
	log    *zap.SugaredLogger
}

func (m *API) PreRun() error {
	logger := m.log.Named("http")
	requestIDMiddleware := middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: uuid.New().String,
	})
	protoMiddleware := protoMiddlewareForLogger(logger)
	detailsMiddleware := middleware.BodyDump(func(ctx echo.Context, reqBody, respBody []byte) {
		log := logger
		if xrid := ctx.Request().Header.Get(echo.HeaderXRequestID); xrid != "" {
			log = log.Named(xrid)
		}
		defer log.Sync()

		var buf bytes.Buffer
		// request
		buf.WriteString(fmt.Sprintf("%s %s %s\n", ctx.Request().Method, ctx.Request().URL, ctx.Request().Proto))
		for header, value := range ctx.Request().Header {
			buf.WriteString("< " + header + ": " + strings.Join(value, " ") + "\n")
		}
		if l := len(reqBody); l < util.Mi {
			buf.Write(reqBody)
			if l == 0 || reqBody[l-1] != '\n' {
				buf.WriteRune('\n')
			}
		} else {
			buf.WriteString(">>>BODY OMITTED (>1Mi)<<<\n")
		}

		// response
		size := util.DataSize(ctx.Response().Size)
		buf.WriteString(fmt.Sprintf("> %s %d %sB\n", ctx.Request().Proto, ctx.Response().Status, size.HumanReadable('b', 3)))
		for header, value := range ctx.Response().Header() {
			buf.WriteString("> " + header + ": " + strings.Join(value, " ") + "\n")
		}
		if len(reqBody) < util.Mi {
			buf.Write(reqBody)
		} else {
			buf.WriteString(">>>BODY OMITTED (>1Mi)<<<\n")
		}

		log.Debug(buf.String())
	})
	recov := &recoverPanic{logger}

	Router.Pre(requestIDMiddleware, protoMiddleware, detailsMiddleware, recov.Middleware)

	authLogger := m.log.Named("auth")
	if mapd.GlobalConfig.Auth.DisableLocalUser {
		authLogger.Warnf("local-managed users are disabled.")
	} else {
		authLogger.Infof("local-managed users are enabled.")
	}

	guard, err := auth.NewAuth(APIRouter.Group("/auth/.well-known"), authLogger)
	if err != nil {
		return err
	}
	resolverRoot, err := resolver.NewResolver(guard.DexClient())
	if err != nil {
		return err
	}

	gqlPost := handler.New(graphql.NewExecutableSchema(graphql.Config{Resolvers: resolverRoot}))
	gqlPost.AddTransport(transport.POST{})
	gqlPost.AddTransport(transport.MultipartForm{
		MaxUploadSize: gqlUploadMaxSize,
		MaxMemory:     gqlUploadMaxMemory,
	})
	gqlGet := handler.New(graphql.NewExecutableSchema(graphql.Config{Resolvers: resolverRoot}))
	gqlGet.AddTransport(transport.GET{})
	gqlGet.AddTransport(guard.GraphQLWebsocketConfig())

	log := m.log.Named("graphql")
	for _, each := range []*handler.Server{gqlPost, gqlGet} {
		each.AddTransport(transport.Options{})

		each.SetQueryCache(lru.New(1000))
		each.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
			defer log.Sync()

			errmsg := fmt.Sprintf("handler has just recovered from a panic: %v. Please contact customer support to report a bug. Stacktrace:\n---\n%s\n", err, debug.Stack())
			log.Errorf(strings.ReplaceAll(errmsg, "\n", "\n> "))
			e, isErr := err.(error)
			if isErr && resolver.IsDBError(e) {
				return &gqlerror.Error{
					Message: "internal database error",
					Extensions: map[string]interface{}{
						"code": "0x" + fmt.Sprintf("%08X", resolver.CodeInternalDatabaseError),
					},
				}
			}
			return &gqlerror.Error{
				Message: "internal server error",
				Extensions: map[string]interface{}{
					"code": "0x" + fmt.Sprintf("%08X", resolver.CodeInternalServerError),
				},
			}
		})
		each.AroundResponses(func(ctx context.Context, next gql.ResponseHandler) *gql.Response {
			defer log.Sync()

			opCtx := gql.GetOperationContext(ctx)
			b, _ := json.MarshalIndent(opCtx.Variables, "", "  ")
			log.Debugf("request(%s) => %s", strings.ReplaceAll(string(b), "\n", "\n> "), strings.ReplaceAll(opCtx.RawQuery, "\n", "\n> "))

			resp := next(ctx)
			b, _ = json.MarshalIndent(resp, "", "  ")
			log.Debugf("response => %s", strings.ReplaceAll(string(b), "\n", "\n> "))
			return resp
		})
		each.Use(extension.FixedComplexityLimit(512))
		each.Use(extension.Introspection{})
		each.Use(extension.AutomaticPersistedQuery{Cache: lru.New(100)})
	}

	APIRouter.GET("/healthz", func(c echo.Context) error {
		m.status.mux.RLock()
		defer m.status.mux.RUnlock()

		code := http.StatusOK
		if !m.status.Ready {
			code = http.StatusServiceUnavailable
		}
		return c.JSON(code, &m.status)
	})
	APIRouter.POST("/graphql", echo.WrapHandler(gqlPost), guard.Middleware())
	APIRouter.GET("/graphql", guard.GraphQLWebsocketHandler(gqlGet))
	return nil
}

func (m *API) Run() <-chan error {
	defer m.log.Sync()

	errCh := make(chan error, 1)
	go func() {
		defer m.log.Sync()

		var err error
		if mapd.GlobalConfig.TLSCert != "" && mapd.GlobalConfig.TLSKey != "" {
			m.log.Infof("api server will listen on: https://%s.", mapd.GlobalConfig.ListenOn)
			err = Router.StartTLS(mapd.GlobalConfig.ListenOn, mapd.GlobalConfig.TLSCert, mapd.GlobalConfig.TLSKey)
		} else {
			m.log.Infof("api server will listen on: http://%s.", mapd.GlobalConfig.ListenOn)
			m.log.Warnf("api server could be vulnerable: unencrypted HTTP connections.")
			err = Router.Start(mapd.GlobalConfig.ListenOn)
		}
		if err == http.ErrServerClosed {
			err = nil
		}
		errCh <- err
		m.log.Info("api server is stopped.")
	}()
	m.log.Info("api server started to serve.")
	m.status.MarkReady()

	clozCh := make(chan error, 1)
	go func() {
		err := <-errCh
		if err != nil {
			m.log.Errorf("shutting down immediately due to a critical error: %v.", err)
			Router.Close()
		}
		clozCh <- err
	}()
	return clozCh
}

func (m *API) Close(ctx context.Context) error {
	var err error
	if e := Router.Shutdown(ctx); e != nil {
		err = e
	}
	if e := Router.Close(); e != nil && err == nil {
		err = e
	}
	return err
}

func NewAPI(logger *zap.Logger) *API {
	return &API{log: logger.Named("api").Sugar()}
}

type APIStatus struct {
	Ready bool `json:"ready,omitempty"`

	mux sync.RWMutex `json:"-"`
}

func (s *APIStatus) MarkReady() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.Ready = true
}

type contextKey struct {
	Name string
}

// ContextWithValue set value into key within underlay context.Context on
// http.Request.
func ContextWithValue(ctx echo.Context, key, value interface{}) echo.Context {
	c := context.WithValue(ctx.Request().Context(), key, value)
	ctx.SetRequest(ctx.Request().WithContext(c))
	return ctx
}

// ValueFromContext gets value with given key from context.Context.
func ValueFromContext(ctx context.Context, key interface{}) interface{} {
	return ctx.Value(key)
}

var (
	ctxLogger = &contextKey{"logger"}
)

// ContextWithLogger set logger into underlay context.Context on http.Request.
func ContextWithLogger(ctx echo.Context, logger *zap.SugaredLogger) echo.Context {
	return ContextWithValue(ctx, ctxLogger, logger)
}

// LoggerFromContext gets logger from context.Context.
func LoggerFromContext(ctx context.Context) *zap.SugaredLogger {
	raw, _ := ValueFromContext(ctx, ctxLogger).(*zap.SugaredLogger)
	return raw
}

type recoverPanic struct {
	l *zap.SugaredLogger
}

// Middleware is the middleware function.
func (m *recoverPanic) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				e, ok := r.(error)
				if !ok {
					e = fmt.Errorf("%v", r)
				}
				errmsg := fmt.Sprintf("handler has just recovered from a panic: %v. Please contact customer support to report a bug. Stacktrace:\n---\n%s\n", e, debug.Stack())
				m.l.Errorf(strings.ReplaceAll(errmsg, "\n", "\n> "))
				err = e
			}
		}()
		return next(ctx)
	}
}

func protoMiddlewareForLogger(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			log := logger
			if xrid := ctx.Request().Header.Get(echo.HeaderXRequestID); xrid != "" {
				log = log.Named(xrid)
			}
			ctx.Set("logger", log)
			ctx = ContextWithLogger(ctx, log)

			// ignore for /api/healthz
			if ctx.Request().URL.Path == "/api/healthz" {
				return next(ctx)
			}

			begin := time.Now()

			ctx.Response().Before(func() {
				defer logger.Sync()

				logger.Infof("%s %s %s <- %s %s %s", ctx.Request().Method, ctx.Request().URL, ctx.Request().Proto, ctx.Request().RemoteAddr, ctx.Request().UserAgent(), ctx.Request().Referer())
			})
			ctx.Response().After(func() {
				defer logger.Sync()

				size := util.DataSize(ctx.Response().Size)
				switch {
				case ctx.Response().Status < 400:
					logger.Infof("%s %d %s %sB %v", ctx.Request().Proto, ctx.Response().Status, ctx.Request().URL, size.HumanReadable('b', 3), time.Since(begin).Truncate(time.Millisecond))
				case ctx.Response().Status < 500:
					logger.Warnf("%s %d %s %sB %v", ctx.Request().Proto, ctx.Response().Status, ctx.Request().URL, size.HumanReadable('b', 3), time.Since(begin).Truncate(time.Millisecond))
				default:
					logger.Errorf("%s %d %s %sB %v", ctx.Request().Proto, ctx.Response().Status, ctx.Request().URL, size.HumanReadable('b', 3), time.Since(begin).Truncate(time.Millisecond))
				}
			})
			return next(ctx)
		}
	}
}

func init() {
	APIRouter.GET("/version", func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, &model.SoftwareVersion{Version: abstract.MustParseVersion(mapd.Version), Revision: mapd.Revision})
	})
}
