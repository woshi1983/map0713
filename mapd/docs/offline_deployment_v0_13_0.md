# MAP v0.13.0 离线部署手册

本文适用于当前 `vks` 分支生成的 MAP v0.13.0 主安装包，以及 `map-offline-installer-dependencies-part1-v0_13_0.tgz`。主程序安装与集群依赖镜像导入是两个独立步骤。

## 1. 制品和用途

| 制品 | 当前位置 | 用途 |
| --- | --- | --- |
| `map-offline-installer-v0_13_0.tgz` | 本仓库 `mapd/dist/` | 安装 MAP 主程序；安装脚本直接 `docker load`，无需镜像仓库 |
| `map-offline-installer-dependencies-part1-v0_13_0.tgz` | 226 主机 `/data/0.13/` | 向内部镜像仓库导入集群创建及扩展所需镜像 |

本次主包的 SHA-256 为 `ecf479a6d4c0f2d72308d3f2d5951ea2fd943279b8671d07be40a5431e87b4a8`。重新构建后应以新文件的校验值为准。主包内有六个本地镜像标签：`mariadb:10.6.7`、`map/dex:v2.32.0`、`map/rsyslog:8.2106.0`、`map/map-console:v0.13.0`、`map/map-prepare:v0.13.0` 和 `map/mapd:v0.13.0`。

part1 压缩包约 9.8 GiB，解包后的 `dependencies-part1-v0_13_0.tar` 为 10,687,830,528 字节。其清单中虽然保留了原始来源镜像名，**上传时会把仓库主机名改写为指定的内部仓库地址**；上传过程从 tar 读取镜像数据，不需要访问原始仓库。不要运行 part1 包自带的旧版 `upload_dependency.sh`，应使用主包中的新版脚本。

## 2. 安装前检查

1. 目标 Linux 主机已安装 Docker 和独立命令 `docker-compose`；用 `docker --version`、`docker-compose --version` 检查。安装脚本需要 root 权限。
2. 准备供用户和 MAP 容器访问的主机名或 IP，以及对应的 HTTPS 证书和私钥。证书的 SAN 必须包含所选主机名或 IP。首次安装保持 `map_disable_local_user: false`。
3. 检查 HTTP/HTTPS 端口未被其他服务使用。211 主机同时运行 Kubernetes，其 CNI 使用 443；该主机实测使用 18080/18443。使用其他端口时，要在 `config.yml` 中同时修改 HTTP 和 HTTPS 端口。
4. 主包只需要目标主机上的 Docker。要使用 part1 镜像创建集群或安装扩展，还需一个**目标集群节点也能访问**的内部 Harbor 2.0 仓库，并为镜像与临时解包文件预留磁盘空间。本文不使用 `harbor.huaruicloud.cn` 作为上传目标。

## 3. 安装 MAP 主程序

把主包复制到目标主机的工作目录。本例以 211 测试主机为例；其他主机应替换地址和路径。

```bash
install -d -m 700 /data/map-offline-test-v0_13_0
cd /data/map-offline-test-v0_13_0
# 从打包机复制 mapd/dist/map-offline-installer-v0_13_0.tgz 至当前目录
sha256sum map-offline-installer-v0_13_0.tgz
tar -xzf map-offline-installer-v0_13_0.tgz
cd map-offline-installer
cp config.yml.tmpl config.yml
chmod 600 config.yml
```

编辑 `config.yml`，至少设置以下字段：

| 字段 | 211 测试值或设置要求 |
| --- | --- |
| `hostname` | `10.121.166.211`；不能用 `localhost` 或 `127.0.0.1` |
| `http.port` / `https.port` | 211 使用 `18080` / `18443`；其他主机按实际空闲端口设置 |
| `https.certificate` / `https.private_key` | 目标主机上的证书、私钥绝对路径 |
| `https.ca_certificate` | 签发证书的 CA 路径；测试用自签名证书时可填该证书路径 |
| `map_admin_password` | 设置新的强密码；不要沿用模板密码 |
| `database.password` | 设置新的强密码；不要沿用模板密码 |
| `data_volume` / `log.location` | 指向目标主机上持久化数据及日志的位置 |

自签名证书仅用于测试。若用 IP 访问，可在目标主机生成包含 IP SAN 的测试证书，例如：

```bash
install -d -m 700 /data/map-offline-test-v0_13_0/cert
openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 30 \
  -keyout /data/map-offline-test-v0_13_0/cert/server.key \
  -out /data/map-offline-test-v0_13_0/cert/server.crt \
  -subj '/CN=10.121.166.211' \
  -addext 'subjectAltName=IP:10.121.166.211'
chmod 600 /data/map-offline-test-v0_13_0/cert/server.key
```

配置完成后运行：

```bash
cd /data/map-offline-test-v0_13_0/map-offline-installer
TERM=xterm ./installer.sh
```

脚本加载包内镜像、生成 Compose 配置、启动五个服务，并等待其健康检查通过。重跑脚本会执行 `docker-compose down`，**不会删除数据库持久化目录**。首次管理员账号是 `admin@tanzu.local`，密码为 `config.yml` 中首次安装时设置的 `map_admin_password`；更改已有实例的配置文件不会重置该账号密码。

验证命令：

```bash
docker-compose ps
curl -kfsS https://10.121.166.211:18443/portal/ >/dev/null
curl -kfsS https://10.121.166.211:18443/dex/.well-known/openid-configuration >/dev/null
curl -fsS http://127.0.0.1:9800/api/healthz
```

`docker-compose ps` 应显示 `map-log`、`map-mariadb`、`map-dex`、`map-mapd`、`map-console` 全部为 `healthy`。`curl -k` 只用于测试自签名证书；正式证书应按正常链路验证。

## 4. 导入 part1 集群依赖镜像

主程序可先独立运行。要创建集群或安装运维扩展，再执行本节。part1 镜像要由集群节点拉取，**只在 MAP 主机上 `docker load` 不够**。

先把 226 上的 part1 压缩包复制到目标主机并解包。包较大，按现场可用磁盘与网络选择传输方式：

```bash
cd /data/map-offline-test-v0_13_0
scp root@10.121.160.226:/data/0.13/map-offline-installer-dependencies-part1-v0_13_0.tgz .
tar -xzf map-offline-installer-dependencies-part1-v0_13_0.tgz
test -f map-offline-installer-dependencies-part1/dependencies-part1-v0_13_0.tar
```

确认内部 Harbor 2.0 已运行，MAP 主机及目标集群节点均能解析并访问其 HTTPS 地址。上传参数 `-r` 填**仓库主机名及可选端口，不带 `https://`**。当前脚本会在目标 Harbor 自动创建项目，并在上传时跳过目标仓库证书校验；集群节点仍需信任该仓库证书。

```bash
cd /data/map-offline-test-v0_13_0/map-offline-installer
export REGISTRY_USERNAME='<内部 Harbor 用户名>'
read -rsp '内部 Harbor 密码: ' REGISTRY_PASSWORD; echo
export REGISTRY_PASSWORD

./upload_dependency.sh \
  -r '<内部仓库主机名[:端口]>' \
  -i /data/map-offline-test-v0_13_0/map-offline-installer-dependencies-part1/dependencies-part1-v0_13_0.tar

unset REGISTRY_PASSWORD
```

上传结束应输出 `Succeeded.`。在内部 Harbor 检查镜像，例如 `<内部仓库>/bitnami/elasticsearch:9.5.5`、`<内部仓库>/bitnami/kibana:9.5.5` 和 `<内部仓库>/tkg/packages/standard/repo:v2023.10.16`。不要把包内原始仓库名当作安装后的镜像地址。

## 5. 在 MAP 中关联内部仓库

上传镜像**不会自动修改 MAP 数据库中的仓库设置**。使用平台超级管理员登录后：

1. 在“系统设置 → 仓库管理”添加内部 Harbor，Endpoint 填 `https://<内部仓库主机名[:端口]>`，配置认证和证书。
2. 在集群供应商设置中关联该仓库，并按需要设为默认仓库。创建集群和安装扩展前，确认供应商使用的仓库地址与 part1 上传目标一致。
3. 如需使用 TKG Standard Package Repo，在“标准包管理”显式添加已上传的 `<内部仓库>/tkg/packages/standard/repo:v2023.10.16`，再按供应商配置选择。此 part1 实际携带的 bundle 标签是 `v2023.10.16`；纯本地安装不会自动创建标准包记录。

### 使用 VKS 时配置授权

VKS 供应商另需配置 MAP 的 Ed25519 签发公钥。将公钥写入安装目录中的 `docker-compose.override.yml`，以便重跑安装脚本时仍能保留；不要把签发私钥放在 MAP 主机上：

```yaml
version: '2.4'
services:
  mapd:
    environment:
      MAP_VKS_LICENSE_PUBLIC_KEY: '<签发端提供的 base64 公钥>'
```

保存后在安装目录执行 `TERM=xterm ./installer.sh` 重建容器。随后在 VKS 供应商详情的“授权”页签获取 Supervisor UID，由签发端生成 `.lic`，再上传授权文件。完整流程参见 [VKS Supervisor 授权](../../docs/vks-supervisor-license.md)。主程序健康并不代表 VKS 授权已经配置；211 的主程序安装测试没有覆盖 VKS 授权。

## 6. 常见故障

| 现象 | 检查与处理 |
| --- | --- |
| 安装脚本提示缺少本地镜像 | 检查 `core-images.txt` 与 `docker images`；使用完整主包重新运行 `installer.sh` |
| `map-mapd` 重启，日志报 Dex 证书缺少 IP SAN | 检查 `hostname`、证书 SAN 和 `https.ca_certificate`；确认 MAP 容器实际访问的是本次安装的 HTTPS 端口 |
| 目标主机的 Kubernetes CNI 抢占 443 | 在 `config.yml` 使用其他空闲端口，例如 211 的 18080/18443，然后重跑安装脚本 |
| docker-compose 重建时报 `ContainerConfig` | Docker 29 与 docker-compose 1.29.2 的原地重建存在兼容问题；在安装目录运行 `docker-compose down`，再运行 `TERM=xterm ./installer.sh`，不要使用 `down -v` |
| part1 上传成功但集群拉取失败 | 核对 MAP 仓库地址、供应商关联、节点到内部 Harbor 的连通性和证书信任 |
| 只安装主包后标准包列表为空 | 这是当前纯本地安装的预期行为；按第 5 节添加实际已上传的 bundle |

## 7. 211 主机实测记录

2026-11-04 在 `10.121.166.211` 验证：主包仅加载六个本地镜像，五个容器健康；从 226 访问门户与 Dex 发现文档均返回 HTTP 200，`mapd /api/healthz` 返回 HTTP 200。211 的安装目录为 `/data/map-offline-test-v0_13_0/map-offline-installer`，门户地址为 `https://10.121.166.211:18443/portal/`。该次验证**未上传 part1，也未执行浏览器交互登录**。
