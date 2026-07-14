// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package mapd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/pflag"
	"github.com/universonic/turret/envutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/logging"
)

type config struct {
	DB   *dbConfig
	Auth *authConfig
	Log  *logConfig

	// FQDN is used by IDP to determine this app's callback entrypoint.
	FQDN string
	// ListenOn determines the port that the HTTP service will listen on.
	ListenOn string
	TLSCert  string
	TLSKey   string

	DaemonShutdownTimeout         time.Duration
	DaemonCleanupTimeout          time.Duration
	CacheDir                      string
	SupportFilesDirs              []string
	ClusterImageProject           string
	ClusterInitiatorImage         string
	EmbeddedBuilderImage          string
	EmbeddedBuildHistoryPreserved int64
	HTTPClientCACert              string
	HTTPClientTimeout             time.Duration
	ConductorSyncerInterval       time.Duration
}

func (cfg *config) urlWithPath(path string) *url.URL {
	_, listenPort, _ := net.SplitHostPort(cfg.ListenOn)

	u := &url.URL{Host: cfg.FQDN, Path: path}
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		u.Scheme = "https"
		if listenPort != "443" {
			u.Host += ":" + listenPort
		}
	} else {
		u.Scheme = "http"
		if listenPort != "80" {
			u.Host += ":" + listenPort
		}
	}
	return u
}

func (cfg *config) GetLoginURL() *url.URL {
	return cfg.urlWithPath("/api/auth/.well-known/login")
}

func (cfg *config) GetCallbackURI() *url.URL {
	defaultURI := cfg.urlWithPath("/api/auth/.well-known/callback")
	if cfg.Auth.IDPCallbackURI != "" {
		uri, err := url.Parse(cfg.Auth.IDPCallbackURI)
		if err == nil {
			if uri.Scheme == "" {
				uri.Scheme = defaultURI.Scheme
			}
			if uri.Host == "" {
				uri.Host = defaultURI.Host
			}
			uri.Path = "/api/auth/.well-known/callback"
			uri.RawQuery = url.Values{}.Encode()
			return uri
		}
	}
	return defaultURI
}

func (cfg *config) Apply(flags *pflag.FlagSet, envs *envutil.Namespace) {
	cfg.DB.Apply(flags, envs)
	cfg.Auth.Apply(flags, envs)
	cfg.Log.Apply(flags, envs)

	flags.StringVarP(
		&cfg.FQDN, "fqdn", "D", cfg.FQDN, "Determine the FQDN that will be used for communicating with IDP.",
	)
	envs.BindString("FQDN", &cfg.FQDN, cfg.FQDN)

	flags.StringVarP(
		&cfg.ListenOn, "listen-on", "l", cfg.ListenOn, "Determine the address that this application will listen on.",
	)
	envs.BindString("LISTEN_ON", &cfg.ListenOn, cfg.ListenOn)

	flags.StringVar(
		&cfg.TLSCert, "tls-cert", cfg.TLSCert, "Specify a TLS certificate file that will be used for securing the endpoint.",
	)
	envs.BindString("TLS_CERT", &cfg.TLSCert, cfg.TLSCert)

	flags.StringVar(
		&cfg.TLSKey, "tls-key", cfg.TLSKey, "Specify a TLS key file that will be used for securing the endpoint.",
	)
	envs.BindString("TLS_KEY", &cfg.TLSKey, cfg.TLSKey)

	envs.BindDuration("DAEMON_SHUTDOWN_TIMEOUT", &cfg.DaemonShutdownTimeout, cfg.DaemonShutdownTimeout)
	envs.BindDuration("DAEMON_CLEANUP_TIMEOUT", &cfg.DaemonCleanupTimeout, cfg.DaemonCleanupTimeout)
	envs.BindString("CACHE_DIR", &cfg.CacheDir, cfg.CacheDir)
	envs.BindString("CLUSTER_IMAGE_PROJECT", &cfg.ClusterImageProject, cfg.ClusterImageProject)
	envs.BindString("CLUSTER_INITIATOR_IMAGE", &cfg.ClusterInitiatorImage, cfg.ClusterInitiatorImage)
	envs.BindString("EMBEDDED_BUILDER_IMAGE", &cfg.EmbeddedBuilderImage, cfg.EmbeddedBuilderImage)
	envs.BindInt("EMBEDDED_BUILD_HISTORY_PRESERVED", &cfg.EmbeddedBuildHistoryPreserved, cfg.EmbeddedBuildHistoryPreserved)
	envs.BindString("HTTP_CLIENT_CA_CERT", &cfg.HTTPClientCACert, cfg.HTTPClientCACert)
	envs.BindDuration("HTTP_CLIENT_TIMEOUT", &cfg.HTTPClientTimeout, cfg.HTTPClientTimeout)
	envs.BindDuration("CONDUCTOR_SYNCER_INTERVAL", &cfg.ConductorSyncerInterval, cfg.ConductorSyncerInterval)
}

func newConfig() *config {
	var defaultSupportFilesDirs []string
	switch runtime.GOOS {
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err == nil {
			defaultSupportFilesDirs = append(defaultSupportFilesDirs, filepath.Join(homeDir, "Library", "Application Support", "com.vmware.tanzu.mapd"))
		}
		defaultSupportFilesDirs = append(defaultSupportFilesDirs, filepath.Join("/Library", "Application Support", "com.vmware.tanzu.mapd"))
	case "linux":
		defaultSupportFilesDirs = append(defaultSupportFilesDirs, "/var/lib/mapd/support-files")
	default:
		panic(fmt.Errorf("unsupported platform: %s", runtime.GOOS))
	}

	return &config{
		DB:   newDBConfig(),
		Auth: newAuthConfig(),
		Log:  newLogConfig(),

		FQDN:                          "localhost",
		ListenOn:                      "[::]:9800",
		DaemonShutdownTimeout:         time.Second * 3,
		DaemonCleanupTimeout:          time.Second * 3,
		CacheDir:                      filepath.Join(os.TempDir(), "vmware-system-map/cache"),
		SupportFilesDirs:              defaultSupportFilesDirs,
		ClusterImageProject:           "map",
		ClusterInitiatorImage:         "cluster-initiator",
		EmbeddedBuilderImage:          "kaniko-builder",
		EmbeddedBuildHistoryPreserved: 10,
		HTTPClientTimeout:             time.Second * 10,
		ConductorSyncerInterval:       time.Second * 10,
	}
}

type dbConfig struct {
	Host         string
	Name         string
	User         string
	Password     string
	MaxOpenConns int64
	MaxIdleConns int64
	Debug        bool
}

func (cfg *dbConfig) Apply(flags *pflag.FlagSet, envs *envutil.Namespace) {
	flags.StringVar(
		&cfg.Host, "db-host", cfg.Host, "Specify the host of MySQL database.",
	)
	envs.BindString("DB_HOST", &cfg.Host, cfg.Host)

	flags.StringVar(
		&cfg.Name, "db-name", cfg.Name, "Specify the name of MySQL database.",
	)
	envs.BindString("DB_NAME", &cfg.Name, cfg.Name)

	flags.StringVar(
		&cfg.User, "db-user", cfg.User, "Specify the login user of MySQL database.",
	)
	envs.BindString("DB_USER", &cfg.User, cfg.User)

	flags.StringVar(
		&cfg.Password, "db-password", cfg.Password, "Specify the login password of MySQL database.",
	)
	envs.BindString("DB_PASSWORD", &cfg.Password, cfg.Password)

	envs.BindInt("DB_MAX_OPEN_CONNS", &cfg.MaxOpenConns, cfg.MaxOpenConns)
	envs.BindInt("DB_MAX_IDLE_CONNS", &cfg.MaxIdleConns, cfg.MaxIdleConns)
	envs.BindBool("DB_DEBUG", &cfg.Debug, cfg.Debug)
}

func newDBConfig() *dbConfig {
	return &dbConfig{
		Host:         "127.0.0.1:3306",
		Name:         "mapd",
		User:         "mapd",
		Password:     "mapd123",
		MaxOpenConns: 64,
		MaxIdleConns: 8,
	}
}

type authConfig struct {
	IDPIssuer                  string
	IDPCallbackURI             string
	IDPCACert                  string
	IDPDexAPI                  string
	IDPDexAPICert              string
	IDPDexAPIKey               string
	IDPDexAPIInsecure          bool
	IDPCrossClientTokenSharing bool

	AdminInitialPassword string
	DisableLocalUser     bool
	MaxRequestsPerIP     int64
	MaxRequestsDuration  time.Duration
}

func (cfg *authConfig) Apply(flags *pflag.FlagSet, envs *envutil.Namespace) {
	flags.StringVarP(
		&cfg.IDPIssuer, "idp-issuer", "I", cfg.IDPIssuer, "Specify the IDP issuer that will be used for OpenID Connect.",
	)
	envs.BindString("IDP_ISSUER", &cfg.IDPIssuer, cfg.IDPIssuer)

	flags.StringVar(
		&cfg.IDPCallbackURI, "idp-callback-uri", cfg.IDPCallbackURI, "Specify a custom callback URI for the client of IDP issuer.",
	)
	envs.BindString("IDP_CALLBACK_URI", &cfg.IDPCallbackURI, cfg.IDPCallbackURI)

	flags.StringVar(
		&cfg.IDPCACert, "idp-ca-cert", cfg.IDPCACert, "Specify the certificate file of root CA that will be used for communicating with IDP.",
	)
	envs.BindString("IDP_CA_CERT", &cfg.IDPCACert, cfg.IDPCACert)

	flags.StringVar(
		&cfg.IDPDexAPI, "idp-dex-api", cfg.IDPDexAPI, "Specify the gRPC API endpoint of Dex IDP.",
	)
	envs.BindString("IDP_DEX_API", &cfg.IDPDexAPI, cfg.IDPDexAPI)

	flags.StringVar(
		&cfg.IDPDexAPICert, "idp-dex-api-cert", cfg.IDPDexAPICert, "Specify the certificate file that will be used for communicating with Dex API.",
	)
	envs.BindString("IDP_DEX_API_CERT", &cfg.IDPDexAPICert, cfg.IDPDexAPICert)

	flags.StringVar(
		&cfg.IDPDexAPIKey, "idp-dex-api-key", cfg.IDPDexAPIKey, "Specify the key file that will be used for communicating with Dex API.",
	)
	envs.BindString("IDP_DEX_API_KEY", &cfg.IDPDexAPIKey, cfg.IDPDexAPIKey)

	flags.BoolVar(
		&cfg.IDPDexAPIInsecure, "idp-dex-api-insecure", cfg.IDPDexAPIInsecure, "Do not validate the TLS connection while communicating with Dex API.",
	)
	envs.BindBool("IDP_DEX_API_INSECURE", &cfg.IDPDexAPIInsecure, cfg.IDPDexAPIInsecure)

	flags.BoolVar(
		&cfg.IDPCrossClientTokenSharing, "idp-cross-client-token-sharing", cfg.IDPCrossClientTokenSharing, "Whether allow to use tokens that are issued to other client",
	)
	envs.BindBool("IDP_CROSS_CLIENT_TOKEN_SHARING", &cfg.IDPCrossClientTokenSharing, cfg.IDPCrossClientTokenSharing)

	envs.BindString("AUTH_ADMIN_INITIAL_PASSWORD", &cfg.AdminInitialPassword, cfg.AdminInitialPassword)
	envs.BindBool("AUTH_DISABLE_LOCAL_USER", &cfg.DisableLocalUser, cfg.DisableLocalUser)
	envs.BindInt("AUTH_MAX_REQUESTS_PER_IP", &cfg.MaxRequestsPerIP, cfg.MaxRequestsPerIP)
	envs.BindDuration("AUTH_MAX_REQUESTS_DURATION", &cfg.MaxRequestsDuration, cfg.MaxRequestsDuration)
}

func newAuthConfig() *authConfig {
	return &authConfig{
		IDPIssuer: "http://localhost:5556/dex",
		IDPDexAPI: "localhost:5557",

		MaxRequestsPerIP:    1000,
		MaxRequestsDuration: time.Hour,
	}
}

type logConfig struct {
	Level         logging.LogLevel
	RotateSizeMiB int64
	File          string
}

func (cfg *logConfig) Apply(flags *pflag.FlagSet, envs *envutil.Namespace) {
	flags.VarP(
		logging.DefaultLogLevelValue(&cfg.Level), "log-level", "L", "Specify the log level (valid: \"debug\", \"info\", \"warn\", \"error\").",
	)
	envs.BindFunc("LOG_LEVEL", func(value string, exists bool) string {
		if exists {
			cfg.Level.Set(value)
		}
		return cfg.Level.String()
	})

	flags.StringVar(
		&cfg.File, "log-file", cfg.File, "Specify a location to write logging message to file.",
	)
	envs.BindString("LOG_FILE", &cfg.File, cfg.File)

	envs.BindInt("LOG_ROTATE_SIZE_MIB", &cfg.RotateSizeMiB, cfg.RotateSizeMiB)
}

func newLogConfig() *logConfig {
	return &logConfig{
		RotateSizeMiB: 64,
	}
}

// GlobalConfig is the global configuration of mapd.
var GlobalConfig = newConfig()

var defaultSupportFilesDirs []string
