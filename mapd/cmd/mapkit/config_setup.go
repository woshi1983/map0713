// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gopkg.in/yaml.v2"
)

// SetupConfigCmd represents the config setup command.
var SetupConfigCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup docker-compose and env files from YAML config file.",
	Long:  "Setup docker-compose and env files from YAML config file.",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			cfg      Config
			exitCode int
		)
		b, err := os.ReadFile(configPath)
		if err != nil {
			exitCode = 1
			goto EXIT
		}
		if err = yaml.Unmarshal(b, &cfg); err != nil {
			exitCode = 1
			goto EXIT
		}
		if err = setup(outputPath, &cfg); err != nil {
			exitCode = 1
			goto EXIT
		}
		return

	EXIT:
		if err != nil {
			fmt.Fprintf(os.Stderr, "FATAL(%d): %v.\n", exitCode, err)
		}
		os.Exit(exitCode)
	},
}

var (
	imageMappingFile string
	outputPath       string
)

func init() {
	ConfigCmd.AddCommand(SetupConfigCmd)

	SetupConfigCmd.Flags().StringVarP(&outputPath, "output", "o", outputPath, "Path of the output directory.")
	SetupConfigCmd.Flags().StringVarP(&imageMappingFile, "image-mapping-file", "I", imageMappingFile, "Path of the image mapping file.")
}

type Config struct {
	Hostname string      `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	HTTP     HTTPConfig  `json:"http,omitempty" yaml:"http,omitempty"`
	HTTPS    HTTPSConfig `json:"https,omitempty" yaml:"https,omitempty"`

	MAPAdminPassword              string         `json:"map_admin_password,omitempty" yaml:"map_admin_password,omitempty"`
	MAPDisableLocalUser           bool           `json:"map_disable_local_user,omitempty" yaml:"map_disable_local_user,omitempty"`
	EnableCrossClientTokenSharing bool           `json:"enable_cross_client_token_sharing,omitempty" yaml:"enable_cross_client_token_sharing,omitempty"`
	Timezone                      string         `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	Database                      DatabaseConfig `json:"database,omitempty" yaml:"database,omitempty"`
	DataVolume                    string         `json:"data_volume,omitempty" yaml:"data_volume,omitempty"`
	Log                           LogConfig      `json:"log,omitempty" yaml:"log,omitempty"`
	Version                       string         `json:"_version,omitempty" yaml:"_version,omitempty"`
}

type HTTPConfig struct {
	Port          int    `json:"port,omitempty" yaml:"port,omitempty"`
	ClientTimeout string `json:"client_timeout,omitempty" yaml:"client_timeout,omitempty"`
}

type HTTPSConfig struct {
	Port          int    `json:"port,omitempty" yaml:"port,omitempty"`
	Certificate   string `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	PrivateKey    string `json:"private_key,omitempty" yaml:"private_key,omitempty"`
	CACertificate string `json:"ca_certificate,omitempty" yaml:"ca_certificate,omitempty"`
}

type DatabaseConfig struct {
	Password     string `json:"password,omitempty" yaml:"password,omitempty"`
	MaxIdleConns string `json:"max_idle_conns,omitempty" yaml:"max_idle_conns,omitempty"`
	MaxOpenConns string `json:"max_open_conns,omitempty" yaml:"max_open_conns,omitempty"`
}

type LogConfig struct {
	Level       string `json:"level,omitempty" yaml:"level,omitempty"`
	RotateCount int64  `json:"rotate_count,omitempty" yaml:"rotate_count,omitempty"`
	RotateSize  string `json:"rotate_size,omitempty" yaml:"rotate_size,omitempty"`
	Location    string `json:"location,omitempty" yaml:"location,omitempty"`
}

const (
	dockerComposeTemplate = `version: '{{.MAPVersion}}'
services:
  log:
    image: {{if ne .PublicImageRegistry ""}}{{.PublicImageRegistry}}/{{end}}map/rsyslog:8.2106.0
    container_name: map-log
    restart: always
    dns_search: .
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - DAC_OVERRIDE
      - SETGID
      - SETUID
    volumes:
      - {{.Config.Log.Location}}:/var/log/docker:z
      - type: bind
        source: ./common/config/log/logrotate.conf
        target: /etc/logrotate.d/logrotate.conf
      - type: bind
        source: ./common/config/log/rsyslog_docker.conf
        target: /etc/rsyslog.d/rsyslog_docker.conf
    ports:
      - 127.0.0.1:4514:10514
    networks:
      - map
  mariadb:
    image: library/mariadb:10.6.7
    container_name: map-mariadb
    restart: unless-stopped
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - DAC_OVERRIDE
      - SETGID
      - SETUID
    ports:
      - 127.0.0.1:3306:3306
    volumes:
      - {{.Config.DataVolume}}/mariadb:/var/lib/mysql:z
      - ./common/config/mariadb/init.sh:/docker-entrypoint-initdb.d/init.sh:z
    depends_on:
      log:
        condition: service_healthy
    networks:
      - map
    dns_search: .
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://localhost:4514"
        tag: "mariadb"
    env_file:
      - ./common/config/mariadb/env
    healthcheck:
      test:
        - CMD
        - healthcheck.sh
        - --su-mysql
        - --connect
        - --innodb_initialized
      interval: 2s
      timeout: 1s
      retries: 8
  dex:
    image: {{if ne .PublicImageRegistry ""}}{{.PublicImageRegistry}}/{{end}}map/dex:v2.32.0
    container_name: map-dex
    restart: unless-stopped
    dns_search: .
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - SETGID
      - SETUID
    ports:
      - 127.0.0.1:5556:5556
      - 127.0.0.1:5557:5557
      - 127.0.0.1:5558:5558
    volumes:
      - ./common/config/dex/config-template.yaml:/opt/dex/conf/dex-config-template.yaml:z
    depends_on:
      mariadb:
        condition: service_healthy
      log:
        condition: service_healthy
    networks:
      - map
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://localhost:4514"
        tag: "dex"
    env_file:
      - ./common/config/dex/env
    healthcheck:
      test:
        - CMD
        - curl
        - -f
        - http://localhost:5556/dex/healthz
      interval: 2s
      timeout: 1s
      retries: 8
  map-console:
    image: {{if ne .PublicImageRegistry ""}}{{.PublicImageRegistry}}/{{end}}map/map-console:{{.MAPVersion}}
    container_name: map-console
    restart: unless-stopped
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - SETGID
      - SETUID
      - NET_BIND_SERVICE
    ports:
      - {{.Config.HTTP.Port}}:{{.Config.HTTP.Port}}
      - {{.Config.HTTPS.Port}}:{{.Config.HTTPS.Port}}
    networks:
      - map
    dns_search: .
    volumes:
      - {{.Config.HTTPS.Certificate}}:/etc/pki/tls/certs/server.crt:z
      - {{.Config.HTTPS.PrivateKey}}:/etc/pki/tls/private/server.key:z
      - ./common/config/map-console/nginx.conf.template:/etc/nginx/templates/nginx.conf.template:z
    depends_on:
      mapd:
        condition: service_started
      dex:
        condition: service_started
      log:
        condition: service_healthy
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://localhost:4514"
        tag: "console"
    env_file:
      - ./common/config/map-console/env
    healthcheck:
      test:
        - CMD
        - curl
        - -k
        - -f
        - https://localhost{{if ne .Config.HTTPS.Port 443}}:{{.Config.HTTPS.Port}}{{end}}/healthz
      interval: 2s
      timeout: 1s
      retries: 8
  mapd:
    image: {{if ne .PublicImageRegistry ""}}{{.PublicImageRegistry}}/{{end}}map/mapd:{{.MAPVersion}}
    container_name: map-mapd
    restart: unless-stopped
    dns_search: .
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - SETGID
      - SETUID
    ports:
      - 127.0.0.1:9800:9800
    volumes:
      {{if ne .Config.HTTPS.CACertificate ""}}- {{.Config.HTTPS.CACertificate}}:/etc/pki/ca-trust/source/anchors/rootca.crt:z{{end}}
    depends_on:
      mariadb:
        condition: service_healthy
      dex:
        condition: service_healthy
      log:
        condition: service_healthy
    networks:
      - map
    env_file:
      - ./common/config/mapd/env
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://localhost:4514"
        tag: "mapd"
    healthcheck:
      test:
        - CMD
        - curl
        - -f
        - http://localhost:9800/api/healthz
      interval: 2s
      timeout: 1s
      retries: 8
networks:
  map:
    external: false`
	nginxConfigTemplate = `server {
  listen {{.HTTP.Port}} default_server;
  server_name _;
  return 301 https://$host{{if ne .HTTPS.Port 443}}:{{.HTTPS.Port}}{{end}}$request_uri;
}
server {
  listen       {{.HTTPS.Port}};
  server_name  {{.Hostname}};

  ssl                 on;
  ssl_certificate     /etc/pki/tls/certs/server.crt;
  ssl_certificate_key /etc/pki/tls/private/server.key;

  include /etc/nginx/mime.types;
  access_log off;

  gzip on;
  gzip_comp_level  2;
  gzip_types text/plain text/css text/html application/json application/javascript image/png application/xml;
  gzip_proxied no-cache no-store private expired auth;
  gzip_buffers 16 8k;
  gzip_min_length 1024;
  gzip_disable "MSIE [1-6]\.";

  proxy_buffer_size          128k;
  proxy_buffers              8 256k;
  proxy_busy_buffers_size    256k;

  location = / {
    return 301 /portal;
  }

  location = /healthz {
    return 200 'OK';
  }

  location / {
    root   /usr/share/nginx/html;
    index  index.html;
    try_files $uri $uri/ /portal/index.html;
  }

  location ^~ /api/ {
    rewrite ^/js/?(.*)$ /$1 break;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-NginX-Proxy true;
    proxy_pass ${MAP_API_ENDPOINT};
  }

  location ^~ /devtool/ {
    rewrite  ^/js/?(.*)$ /$1 break;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-NginX-Proxy true;
    proxy_pass ${MAP_API_ENDPOINT};
  }

  location ^~ /dex/ {
    rewrite ^/js/?(.*)$ /$1 break;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-NginX-Proxy true;
    proxy_pass ${DEX_IDP_ENDPOINT};
  }

  location ^~ /stream/ {
    rewrite  ^/stream/(.*)$ /api/graphql break;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-NginX-Proxy true;
    proxy_pass ${MAP_API_ENDPOINT};
  }

  location ~* /assets/monaco-editor/.* {
    root   /usr/share/nginx/html/portal;
  }

  error_page 500 502 503 504 /50x.html;
  location = /50x.html {
    root /usr/share/nginx/html;
  }
}`
)

type dockerComposeRenderConfig struct {
	MAPVersion          string
	Config              *Config
	PublicImageRegistry string
}

func setup(outputDir string, cfg *Config) error {
	err := setupMapd(outputDir, cfg)
	if err != nil {
		return err
	}
	err = setupMapConsole(outputDir, cfg)
	if err != nil {
		return err
	}
	err = setupDex(outputDir, cfg)
	if err != nil {
		return err
	}
	err = setupMariadb(outputDir, cfg)
	if err != nil {
		return err
	}
	err = setupLog(outputDir, cfg)
	if err != nil {
		return err
	}

	fpath := filepath.Join(outputDir, "docker-compose.yml")
	f, err := os.Create(fpath)
	if err != nil {
		return errors.Wrapf(err, "failed to write docker-compose file to %s", fpath)
	}
	defer f.Close()

	tpl, err := template.New("main").Parse(dockerComposeTemplate)
	if err != nil {
		panic(err)
	}
	if err = tpl.Execute(f, &dockerComposeRenderConfig{
		MAPVersion:          mapd.Version,
		Config:              cfg,
		PublicImageRegistry: mapd.PublicImageRegistry,
	}); err != nil {
		return errors.Wrap(err, "unable to generate docker-compose file")
	}
	return nil
}

func setupMapd(outputDir string, cfg *Config) error {
	portStr := ""
	if cfg.HTTPS.Port != 443 {
		portStr = fmt.Sprintf(":%d", cfg.HTTPS.Port)
	}

	fpath := filepath.Join(outputDir, "common/config/mapd/env")
	val := map[string]string{
		"MAPD_AUTH_ADMIN_INITIAL_PASSWORD":    cfg.MAPAdminPassword,
		"MAPD_AUTH_DISABLE_LOCAL_USER":        strconv.FormatBool(cfg.MAPDisableLocalUser),
		"MAPD_CLUSTER_IMAGE_PROJECT":          "map",
		"MAPD_DB_HOST":                        "mariadb:3306",
		"MAPD_DB_NAME":                        "map",
		"MAPD_DB_USER":                        "mapd",
		"MAPD_DB_PASSWORD":                    cfg.Database.Password,
		"MAPD_DB_MAX_IDLE_CONNS":              cfg.Database.MaxIdleConns,
		"MAPD_DB_MAX_OPEN_CONNS":              cfg.Database.MaxOpenConns,
		"MAPD_FQDN":                           cfg.Hostname,
		"MAPD_HTTP_CLIENT_TIMEOUT":            cfg.HTTP.ClientTimeout,
		"MAPD_IDP_ISSUER":                     fmt.Sprintf("https://%s%s/dex", cfg.Hostname, portStr),
		"MAPD_IDP_CALLBACK_URI":               fmt.Sprintf("https://%s%s", cfg.Hostname, portStr),
		"MAPD_IDP_CROSS_CLIENT_TOKEN_SHARING": strconv.FormatBool(cfg.EnableCrossClientTokenSharing),
		"MAPD_IDP_DEX_API":                    "dex:5557",
		"MAPD_IDP_DEX_API_INSECURE":           "true",
		"MAPD_LOG_LEVEL":                      cfg.Log.Level,
		"TZ":                                  cfg.Timezone,
	}
	if cfg.HTTPS.CACertificate != "" {
		val["MAPD_HTTP_CLIENT_CA_CERT"] = "/etc/pki/ca-trust/source/anchors/rootca.crt"
	}
	if err := writeEnvFile(fpath, val); err != nil {
		return err
	}
	return nil
}

func setupMapConsole(outputDir string, cfg *Config) error {
	fpath := filepath.Join(outputDir, "common/config/map-console/env")
	if err := writeEnvFile(fpath, map[string]string{
		"DEX_IDP_ENDPOINT": "http://dex:5556",
		"MAP_API_ENDPOINT": "http://mapd:9800",
		"TZ":               cfg.Timezone,
	}); err != nil {
		return err
	}

	fpath = filepath.Join(outputDir, "common/config/map-console/nginx.conf.template")
	f, err := os.Create(fpath)
	if err != nil {
		return errors.Wrapf(err, "failed to write nginx config template file to %s", fpath)
	}
	defer f.Close()

	tpl, err := template.New("main").Parse(nginxConfigTemplate)
	if err != nil {
		panic(err)
	}
	if err = tpl.Execute(f, cfg); err != nil {
		return errors.Wrap(err, "unable to generate nginx config template file")
	}
	return nil
}

func setupDex(outputDir string, cfg *Config) error {
	portStr := ""
	if cfg.HTTPS.Port != 443 {
		portStr = fmt.Sprintf(":%d", cfg.HTTPS.Port)
	}

	fpath := filepath.Join(outputDir, "common/config/dex/env")
	if err := writeEnvFile(fpath, map[string]string{
		"DEX_FQDN":                    cfg.Hostname,
		"DEX_ISSUER":                  fmt.Sprintf("https://%s%s/dex", cfg.Hostname, portStr),
		"DEX_EXPIRY_DEVICE_REQUESTS":  "5m",
		"DEX_EXPIRY_SIGNING_KEYS":     "6h",
		"DEX_EXPIRY_ID_TOKENS":        "24h",
		"DEX_EXPIRY_AUTH_REQUESTS":    "24h",
		"DEX_LOCAL_CONNECTOR_ENABLED": strconv.FormatBool(!cfg.MAPDisableLocalUser),
		"DEX_LOG_LEVEL":               "info",
		"DEX_LOG_FORMAT":              "text",
		"MYSQL_HOST":                  "mariadb",
		"MYSQL_PORT":                  "3306",
		"MYSQL_DEX_DATABASE":          "dex",
		"MYSQL_USER":                  "dex",
		"MYSQL_PASSWORD":              cfg.Database.Password,
		"TZ":                          cfg.Timezone,
	}); err != nil {
		return err
	}

	fpath = filepath.Join(outputDir, "common/config/dex/config-template.yaml")
	if _, err := os.Stat(fpath); err != nil && os.IsNotExist(err) {
		if err = writeFile(fpath, []byte(`issuer: ${DEX_ISSUER}
storage:
  type: mysql
  config:
    host: ${MYSQL_HOST}
    port: ${MYSQL_PORT}
    database: ${MYSQL_DEX_DATABASE}
    user: ${MYSQL_USER}
    password: ${MYSQL_PASSWORD}
    ssl:
      mode: "false"
web:
  http: 0.0.0.0:5556
  allowedOrigins: ['*']
telemetry:
  http: 0.0.0.0:5558
grpc:
  addr: 0.0.0.0:5557
frontend:
  issuer: ${MAP_ISSUER}
  dir: /opt/dex/web-${DEX_PAGE_LANG}
  theme: map
oauth2:
  skipApprovalScreen: true
enablePasswordDB: ${DEX_LOCAL_CONNECTOR_ENABLED}
expiry:
  deviceRequests: ${DEX_EXPIRY_DEVICE_REQUESTS}
  signingKeys: ${DEX_EXPIRY_SIGNING_KEYS}
  idTokens: ${DEX_EXPIRY_ID_TOKENS}
  authRequests: ${DEX_EXPIRY_AUTH_REQUESTS}
logger:
  level: ${DEX_LOG_LEVEL}
  format: ${DEX_LOG_FORMAT}
staticClients:
connectors:
# Example of LDAP connector. Full documentation can be found here: https://dexidp.io/docs/connectors.
#- type: ldap
#  id: ldap
#  name: Microsoft AD
#  config:
#    host: 10.1.2.20:636
#    # The following field is required if using port 389.
#    #insecureNoSSL: true
#    #rootCA: tmp/rootca.pem
#    insecureSkipVerify: true
#    bindDN: CN=Service Account,OU=dev,DC=example,DC=org
#    bindPW: P@ssw0rd
#    usernamePrompt: Email Address
#    userSearch:
#      baseDN: OU=dev,DC=example,DC=org
#      filter: "(objectClass=person)"
#      username: userPrincipalName
#      idAttr: DN
#      emailAttr: userPrincipalName
#      nameAttr: name
#      preferredUsernameAttr: sAMAccountName
#    groupSearch:
#      baseDN: OU=dev,DC=example,DC=org
#      filter: "(|(objectClass=posixGroup)(objectClass=group))"
#      userMatchers:
#      - userAttr: DN
#        groupAttr: member
#      nameAttr: cn
`)); err != nil {
			return err
		}
	}
	return nil
}

func setupMariadb(outputDir string, cfg *Config) error {
	fpath := filepath.Join(outputDir, "common/config/mariadb/env")
	if err := writeEnvFile(fpath, map[string]string{
		"MARIADB_ROOT_PASSWORD":        cfg.Database.Password,
		"MARIADB_DATABASE":             "map",
		"MARIADB_USER":                 "mapd",
		"MARIADB_MYSQL_LOCALHOST_USER": "true",
		"MARIADB_PASSWORD":             cfg.Database.Password,
		"MARIADB_DEX_DATABASE":         "dex",
		"MARIADB_DEX_USER":             "dex",
		"MARIADB_DEX_PASSWORD":         cfg.Database.Password,
		"TZ":                           cfg.Timezone,
	}); err != nil {
		return err
	}
	fpath = filepath.Join(outputDir, "common/config/mariadb/init.sh")
	if err := writeFile(fpath, []byte(`#!/bin/bash

mysql -hlocalhost -uroot -p${MARIADB_ROOT_PASSWORD} << SQL
CREATE DATABASE IF NOT EXISTS ${MARIADB_DEX_DATABASE};
CREATE USER IF NOT EXISTS '${MARIADB_DEX_USER}'@'%' IDENTIFIED BY '${MARIADB_DEX_PASSWORD}';
GRANT ALL ON ${MARIADB_DEX_DATABASE}.* TO ${MARIADB_DEX_USER};
SQL
`)); err != nil {
		return err
	}
	return nil
}

func setupLog(outputDir string, cfg *Config) error {
	fpath := filepath.Join(outputDir, "common/config/log/logrotate.conf")
	if err := writeFile(fpath, []byte(fmt.Sprintf(`/var/log/docker/*.log {
    rotate %d
    size %s
    copytruncate
    compress
    missingok
    nodateext
}`, cfg.Log.RotateCount, cfg.Log.RotateSize))); err != nil {
		return err
	}
	fpath = filepath.Join(outputDir, "common/config/log/rsyslog_docker.conf")
	if err := writeFile(fpath, []byte(`# Rsyslog configuration file for docker.

template(name="DynaFile" type="string" string="/var/log/docker/%programname%.log")

if $programname != "rsyslogd" then {
	action(type="omfile" dynaFile="DynaFile")
}`)); err != nil {
		return err
	}
	return nil
}

func writeEnvFile(name string, vals map[string]string) error {
	var buf bytes.Buffer
	for k, v := range vals {
		buf.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	return writeFile(name, buf.Bytes())
}

func writeFile(name string, p []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return errors.Wrapf(err, "failed to write file %s", name)
	}
	if err := os.WriteFile(name, p, 0644); err != nil {
		return errors.Wrapf(err, "failed to write file %s", name)
	}
	return nil
}
