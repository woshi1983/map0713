# Binary build execute permission fix

This directory contains the `/kaniko/build` entrypoint extracted from
`harbor.huaruicloud.cn/map/kaniko-builder:v0.12.0` (manifest digest pinned in
the Dockerfile). The only behavior change is that a binary build's generated
Dockerfile runs `chmod +x` on the downloaded binary after setting its owner.

Build the patched image with:

```sh
docker build -t kaniko-builder:v0.12.0-execfix.1 mapd/scripts/kaniko-builder
```

MAPD v0.12.0 pulls `kaniko-builder:v0.12.0` with `Always` for build Jobs. A
release of this fix must preserve the previous tag under a backup tag before
updating `v0.12.0` in the registry. Existing binary images need a new build to
acquire execute permission.
