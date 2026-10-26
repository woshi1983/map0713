# Cluster initiator without default Velero installation

MAPD v0.13.0 starts `cluster-initiator:v0.13.0` as a Job when a cluster is
created or attached. The original image installs Velero and Restic at the end
of `/usr/bin/entrypoint.sh`, even when no backup storage location exists.

The Dockerfile pins the original Harbor manifest. The patch script checks the
entrypoint checksum, removes the unused Velero functions and default install
block, and preserves the remaining cluster initialization steps. It fails
instead of changing an unexpected upstream script.

Build with:

```sh
docker build -t harbor.huaruicloud.cn/map/cluster-initiator:v0.13.0 mapd/scripts/cluster-initiator
```

Back up the current Harbor `v0.13.0` tag before replacing it. MAPD uses image
pull policy `Always`, so new cluster initialization Jobs will pull the updated
image. This change does not remove Velero from existing clusters.
