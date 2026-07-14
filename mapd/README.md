# Mapd

Daemon Utility for MAP.

## Setup

**System requirements:**

**On a Linux host:** docker 1.10.0+ and docker-compose 1.6.0+ .

Download binaries of **[MAP release](https://gitlab.eng.vmware.com/cn-se-solution/mapd/-/tags)** and follow **[Installation & Configuration Guide](docs/installation_guide.md)** to install MAP.

Refer to **[User Guide](docs/user_guide.md)** for more details on how to use MAP.

## Build

```bash
docker build -t ${IMAGE_REGISTRY}/map/mapd:${TAG} .
```

If you are in China mainland, please use 3rd party goproxy. For example:

```bash
docker build --build-arg GOPROXY=https://goproxy.cn,direct -t ${IMAGE_REGISTRY}/map/mapd:${TAG} .
```

## Contribution
We welcome contributions from the community. If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a pull request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq). 

Refer to **[Contribution Guide](CONTRIBUTING.md)** for more details.

## License
MAP is available under the [Apache 2 license](LICENSE).

This project uses open source components which have additional licensing terms.  The official docker images and licensing terms for these open source components can be found at the following locations:

* Rocky Linux 8: [docker image](https://hub.docker.com/_/rockylinux), [license](https://rockylinux.org/licensing/)
* Dex 2.31: [license](https://github.com/dexidp/dex/blob/master/LICENSE)
* MariaDB 10.6: [docker image](https://hub.docker.com/_/mariadb), [license](https://github.com/MariaDB/mariadb-docker/blob/master/LICENSE)
