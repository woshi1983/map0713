# MAP Runtime Configuration

## Setup

|Environment Variable|Command-Line Flag|Default Value|
|:-|:-|:-|
|`MAPD_FQDN`|`-D`, `--fqdn`|`$(hostname)`|
|`MAPD_LISTEN_ON`|`-l`, `--listen-on`|`[::]:9800`|
|`MAPD_TLS_CERT`|`--tls-cert`||
|`MAPD_TLS_KEY`|`--tls-key`||
|`MAPD_IDP_ISSUER`|`-I`, `--idp-issuer`|`http://localhost:5556/dex`|
|`MAPD_IDP_CALLBACK_URI`|`--idp-callback-uri`||
|`MAPD_IDP_CA_CERT`|`--idp-ca-cert`||
|`MAPD_IDP_DEX_API`|`--idp-dex-api`|`localhost:5557`|
|`MAPD_IDP_DEX_API_CERT`|`--idp-dex-api-cert`||
|`MAPD_IDP_DEX_API_KEY`|`--idp-dex-api-key`||
|`MAPD_IDP_CROSS_CLIENT_TOKEN_SHARING`|`--idp-cross-client-token-sharing`|`false`|
|`MAPD_LOG_LEVEL`|`-L`, `--log-level`|`info`|
|`MAPD_DB_HOST`|`--db-host`|`[::1]:3306`|
|`MAPD_DB_NAME`|`--db-name`||
|`MAPD_DB_USER`|`--db-user`||
|`MAPD_DB_PASSWORD`|`--db-password`||
|`MAPD_DB_MAX_IDLE_CONNS`||`8`|
|`MAPD_DB_MAX_OPEN_CONNS`||`64`|
|`MAPD_DB_DEBUG`||`false`|
|`MAPD_DAEMON_SHUTDOWN_TIMEOUT`||`3s`|
|`MAPD_DAEMON_CLEANUP_TIMEOUT`||`3s`|
|`MAPD_AUTH_ADMIN_INITIAL_PASSWORD`||`vmware`|
|`MAPD_AUTH_DISABLE_LOCAL_USER`||`false`|
|`MAPD_AUTH_MAX_REQUESTS_PER_IP`||`1000`|
|`MAPD_AUTH_MAX_REQUESTS_DURATION`||`1h`|
|`MAPD_CACHE_DIR`||`${TMPDIR}/vmware-system-map/cache`|
|`MAPD_CLUSTER_IMAGE_PROJECT`||`map`|
|`MAPD_CLUSTER_INITIATOR_IMAGE`||`cluster-initiator`|
|`MAPD_EMBEDDED_BUILDER_IMAGE`||`kaniko-builder`|
|`MAPD_EMBEDDED_BUILD_HISTORY_PRESERVED`||`10`|
|`MAPD_HTTP_CLIENT_CA_CERT`||(system default)|
|`MAPD_HTTP_CLIENT_TIMEOUT`||`10s`|
|`MAPD_CONDUCTOR_SYNCER_INTERVAL`||`10s`|