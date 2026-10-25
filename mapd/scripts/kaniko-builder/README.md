# Binary build execute permission fix

This directory contains the `/kaniko/build` entrypoint extracted from
`harbor.huaruicloud.cn/map/kaniko-builder:v0.12.0` (manifest digest pinned in
the Dockerfile). For binary builds, the entrypoint sets the downloaded file's
execute bit in the build context before Kaniko processes `ADD`. Kaniko did not
capture the prior `RUN chmod +x` change in the resulting image layer.

Build the patched image with:

```sh
docker build -t kaniko-builder:v0.13.0-execfix.2 mapd/scripts/kaniko-builder
```

MAPD v0.13.0 pulls `kaniko-builder:v0.13.0` with `Always` for build Jobs. A
release of this fix must preserve the previous tag under a backup tag before
updating `v0.13.0` in the registry. Existing binary images need a new build to
acquire execute permission.
