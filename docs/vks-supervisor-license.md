# VKS Supervisor 授权

MAP 只对 VKS 主管集群执行本授权检查。TKGM 和其他集群类型保持原有行为。同一主管集群下的多个 VKS 供应商与工作负载集群共用授权，绑定键为主管集群 `kube-system` Namespace 的 UID，而非地址、用户名或 MAP 供应商 ID。

## 签发和安装

签发端离线执行 `cd mapd && go run ./cmd/vks-license-issuer -generate-key`，安全保存输出的 `PRIVATE`。把 `PUBLIC` 作为 mapd 的 `MAP_VKS_LICENSE_PUBLIC_KEY` 环境变量配置，并重启 mapd。不要把私钥放到 MAP 服务器、仓库、镜像或日志中。

在 VKS 供应商详情的“授权”页签查看 Supervisor UID。签发端设置 `MAP_VKS_LICENSE_PRIVATE_KEY` 为私钥，并执行：

```sh
go run ./cmd/vks-license-issuer -supervisor-uid '<页面显示的 UID>' -expires-at '2027-11-01T00:00:00Z' > supervisor.lic
```

平台超级管理员在同一页签上传 `.lic` 文件。签发工具输出 `base64url(payload).base64url(Ed25519 signature)`；后端在入库前校验签名、UID、`vks` 功能和到期时间，并在每次 VKS 写操作时重新校验。公钥必须在启用授权前完成配置，否则 VKS 写操作会被拒绝。

## 行为与运维

- 授权状态查询可展示已保存的 UID 和授权状态；写操作必须实时连接主管集群并核对 UID。连接故障会阻止 VKS 写操作。
- 同地址的主管集群重建后 UID 变化，旧授权不能沿用；应新建供应商并签发新授权。地址、密码或证书修改但 UID 不变时授权继续有效。
- 授权过期或未配置时，VKS 创建、升级、扩缩容和正常远端删除被拒绝；只读查询、供应商连接测试及超级管理员的数据库强制清理仍可使用。
- 当前不自动发放试用授权。部署启用此功能前，先给现有 VKS 主管集群签发并上传授权文件。
