# MAP IP Mode Deployment

IP mode is the fixed deployment mode for test hosts where browser redirects must not contain Docker container names.

In this mode, every browser-facing URL uses the OS host IP:

- Dex issuer: `http://<host-ip>/dex`
- Dex login URL: `http://<host-ip>/dex/auth...`
- MAP callback: `http://<host-ip>:<mapd-port>/api/auth/.well-known/callback`
- Portal: `http://<host-ip>/portal/`

The deployment script is:

```bash
scripts/deploy-ip-mode.sh
```

By default it auto-detects the host IP. To pin an IP explicitly:

```bash
MAP_HOST_IP=10.121.160.226 scripts/deploy-ip-mode.sh
```

Default ports:

- Console: `80`
- Console compatibility alias: `18080`
- mapd: `9800`
- mapd alternate test port: `19800`
- Dex HTTP: `5556`
- Dex gRPC: `5557`

Default images:

```text
registry-stg.eng.vmware.com/map/map-console:v0.12.0
registry-stg.eng.vmware.com/map/mapd:v0.12.0
```

Override images when needed:

```bash
MAP_CONSOLE_IMAGE=registry-stg.eng.vmware.com/map/map-console:v0.12.0 \
MAPD_IMAGE=registry-stg.eng.vmware.com/map/mapd:v0.12.0 \
MAP_HOST_IP=<host-ip> \
scripts/deploy-ip-mode.sh
```

The script seeds a `default` project and admin binding by default so the initial admin login lands in the portal instead of `/portal/no-org`.

Default login:

```text
admin@tanzu.local / admin123
```

After deployment, verify that login redirects do not contain container DNS names:

```bash
curl -H "Host: <host-ip>" -D - \
  "http://127.0.0.1/api/auth/.well-known/login?redirectURI=http%3A%2F%2F<host-ip>%2Fportal%2F"
```

The `Location` header must start with:

```text
http://<host-ip>/dex/auth
```
