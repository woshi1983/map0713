# MAP v0.12 Manual Full UI Test Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Re-run the MAP v0.12 user manual end to end through the UI on the 226 environment, rebuild `test02`, validate all manual features, and fix product defects found during testing.

**Architecture:** Treat the Word manual as the source of truth for UI behavior. Execute destructive environment preparation only after explicit user approval, capture before/after state, then validate each feature area with UI actions plus backend/Kubernetes evidence where needed.

**Tech Stack:** MAP v0.12, Angular frontend in `map-portal`, Go backend in `mapd`, Playwright/browser UI testing, Kubernetes/Tanzu, Harbor, AVI, MinIO/Velero.

## Global Constraints

- Repository root: `/data/claude/map0713`.
- Backend: `/data/claude/map0713/mapd`.
- Frontend: `/data/claude/map0713/map-portal`.
- GitHub remote: `https://github.com/woshi1983/map0713.git`.
- Prefer this checkout for code changes, tests, commits, and pushes.
- Keep MAP product version at `v0.12.0` unless explicitly asked to change it.
- Keep standard package default repo version aligned with `v2.1.0` unless explicitly asked otherwise.
- Build backend binaries only with the existing `mapd-binary-builder` container.
- Do not rebuild Docker images for backend binary-only changes.
- Build static backend binaries with `CGO_ENABLED=0`.
- Deploy backend changes by replacing the `mapd` binary in the relevant container on `10.121.160.226`.
- Do not persist cleartext passwords in repo files, logs, screenshots, or test artifacts.
- Destructive remote operations require explicit approval immediately before execution.

---

## Environment Inputs

- MAP environment host: `10.121.160.226`.
- Supplier endpoint to inspect and recreate: `https://172.16.84.41:6443`.
- Management kubeconfig on 226: `/root/mgmt-kubeconfig.yaml`.
- Cluster under test: `test02`.
- Cluster sizing: 1 control-plane node, 3 worker nodes, `medium` class.
- Harbor endpoint: `https://harbor.huaruicloud.cn`.
- AVI endpoint: `https://172.16.82.100`.
- Standard package repository: `projects.packages.broadcom.com/tkg/packages/standard/repo:v2023.10.1`.
- MinIO endpoint: `http://172.16.84.22:9000`.
- Backup bucket: `velero`.
- DHCP server cleanup target: `172.16.82.120`, root login as provided by user.
- Manual source: `/root/MAP v0.12 使用手册 - 20231223.docx`.
- Extracted manual text: `/tmp/map-manual-extract/MAP v0.12 使用手册 - 20231223.txt`.

## Manual Coverage Matrix

- System settings: repository management, load balancer management, cluster supplier management, supplier CA certificates.
- Cluster management: create cluster, add existing cluster, upgrade where available, vertical resize, horizontal resize, delete cluster, namespace CRUD, backup, restore, operations services.
- Operations services: Octant/resource manager, Grafana/Prometheus monitoring, ElasticSearch/FluentBit logging, external service shortcut, edit/disable services.
- Application management: create/delete application, create BASIC component with ConfigMap/Secret/PVC, create RAW component, create HELM component through service catalog, edit/clone/delete components, configuration storage.
- Component deployment: deploy, inspect task details and service address, update deployed version, add deployment instance, delete instance.
- Application routing: L7 route create/edit/delete, multi-cluster route create/edit/delete when prerequisites exist.
- Build management: binary image create/edit/delete, Dockerfile image create/edit/delete, source-code image create/edit/delete if TBS is enabled, start build, inspect/delete build task.
- Project/user management: project create/edit/quota/member/delete, user create/edit/reset password/delete, normal user profile/password change.
- Troubleshooting: system logs for cluster initialization, initialization failure evidence, x509 external service failure path.

## Approval Gate

Before executing Task 2 or any later destructive step, ask the user to approve these exact actions:

- Delete supplier/cluster records and all current application data from MAP on 226.
- Delete cluster `test02` from MAP and, if necessary, the backing Tanzu/Kubernetes resources.
- Clean related DHCP/IP allocations on `172.16.82.120`.
- Recreate supplier `https://172.16.84.41:6443` and cluster `test02`.

Do not continue past this gate without an explicit user reply approving the cleanup.

### Task 1: Manual Extraction and Test Case Inventory

**Files:**
- Read: `/root/MAP v0.12 使用手册 - 20231223.docx`
- Read: `/tmp/map-manual-extract/MAP v0.12 使用手册 - 20231223.txt`
- Create: `docs/superpowers/plans/2026-07-14-map-v012-manual-full-ui-test-test02.md`

**Interfaces:**
- Consumes: Word manual and extracted text.
- Produces: Coverage matrix and execution checklist in this plan.

- [x] **Step 1: Convert the manual to readable text and HTML**

Run:

```bash
mkdir -p /tmp/map-manual-extract
libreoffice --headless --convert-to html --outdir /tmp/map-manual-extract '/root/MAP v0.12 使用手册 - 20231223.docx'
libreoffice --headless --convert-to txt --outdir /tmp/map-manual-extract '/root/MAP v0.12 使用手册 - 20231223.docx'
```

Expected: HTML, TXT, and screenshot image assets are generated under `/tmp/map-manual-extract`.

- [x] **Step 2: Extract functional coverage from the manual**

Run:

```bash
rg -n "集群|供应商|应用|组件|运维|备份|恢复|harbor|avi|标准|仓库|用户|角色|权限|监控|告警|日志|升级|删除|创建" '/tmp/map-manual-extract/MAP v0.12 使用手册 - 20231223.txt'
```

Expected: all manual sections listed in the coverage matrix above.

### Task 2: Baseline Capture Before Cleanup

**Files:**
- Create: `.remember/map-v012-test02-baseline-YYYYMMDD-HHMM.md`
- Create: `.gstack/artifacts/map-v012-test02/baseline/`

**Interfaces:**
- Consumes: MAP UI/API state, 226 kubeconfig, supplier endpoint, DHCP state.
- Produces: a timestamped baseline record used to restore or audit destructive cleanup.

- [ ] **Step 1: Confirm destructive approval**

Ask:

```text
请确认是否允许我在 226 上删除当前供应商/集群/应用数据、删除 test02、并清理 172.16.82.120 上相关 DHCP/IP 记录。确认后我才执行。
```

Expected: user explicitly approves.

- [ ] **Step 2: Record MAP current state through UI**

Use browser UI to log in to MAP on 226 and record screenshots/list exports for:

```text
系统设置 -> 仓库管理
系统设置 -> 负载均衡器管理
系统设置 -> 集群供应商管理
集群管理 -> 集群列表
应用组件 -> 应用、组件、部署实例
应用路由 -> 七层路由、多集群路由
构建管理 -> 镜像和构建任务
项目管理 -> 项目、成员
用户管理 -> 用户
```

Expected: each page has a screenshot and a short markdown inventory entry.

- [ ] **Step 3: Record supplier and cluster state from 226**

Run on 226 after connecting with the established project method:

```bash
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml config get-contexts
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml get ns
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml get clusters -A 2>/dev/null || true
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml get tkc -A 2>/dev/null || true
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml get cluster,tkc,machine,vspheremachine -A 2>/dev/null || true
```

Expected: baseline identifies any existing `test02` backing resources and supplier namespace resources.

- [ ] **Step 4: Record DHCP state on 172.16.82.120**

Run read-only commands first:

```bash
ssh root@172.16.82.120 'hostname; ip addr; systemctl status dhcpd --no-pager || systemctl status isc-dhcp-server --no-pager || true'
ssh root@172.16.82.120 'find /var/lib -maxdepth 3 -type f \( -name "*lease*" -o -name "dhcpd.leases" \) -print'
```

Expected: DHCP service type and lease file locations are known before cleanup.

### Task 3: Environment Cleanup

**Files:**
- Append: `.remember/map-v012-test02-baseline-YYYYMMDD-HHMM.md`
- Create: `.gstack/artifacts/map-v012-test02/cleanup/`

**Interfaces:**
- Consumes: baseline record from Task 2.
- Produces: clean MAP state, removed `test02`, cleared relevant DHCP/IP allocations.

- [ ] **Step 1: Delete application data through UI**

Delete in dependency order:

```text
应用路由 -> 删除多集群路由、七层路由
组件详情 -> 删除部署实例
应用组件 -> 删除 HELM、RAW、BASIC 组件
应用组件 -> 删除应用
构建管理 -> 删除构建任务，再删除镜像定义
集群详情 -> 删除备份任务/备份位置 only if created for this test scope
```

Expected: application, route, deployment, and build pages contain no stale test data.

- [ ] **Step 2: Delete `test02` through UI**

Use:

```text
集群管理 -> test02 -> 更多操作 -> 删除集群 -> 输入 test02 确认
```

Expected: `test02` disappears from MAP and backing Kubernetes resources begin deletion.

- [ ] **Step 3: Verify backing resource deletion**

Run on 226:

```bash
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml get cluster,tkc,machine,vspheremachine -A 2>/dev/null | grep -i test02 || true
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml get all -A 2>/dev/null | grep -i test02 || true
```

Expected: no `test02` resources remain. If resources remain blocked, collect events/logs before any manual cleanup.

- [ ] **Step 4: Delete cluster supplier and stale cluster-list entries through UI**

Use:

```text
系统设置 -> 集群供应商管理 -> 删除 target supplier
集群管理 -> 删除/强制删除 stale cluster records only after backing resource check
```

Expected: supplier and stale records are absent from MAP.

- [ ] **Step 5: Clean related DHCP/IP leases**

Only remove leases that baseline evidence ties to `test02` nodes or old supplier-created nodes:

```bash
ssh root@172.16.82.120 'cp -a /var/lib/dhcpd/dhcpd.leases /var/lib/dhcpd/dhcpd.leases.map-test02-backup-$(date +%Y%m%d%H%M%S) 2>/dev/null || true'
ssh root@172.16.82.120 'cp -a /var/lib/dhcp/dhcpd.leases /var/lib/dhcp/dhcpd.leases.map-test02-backup-$(date +%Y%m%d%H%M%S) 2>/dev/null || true'
```

Expected: lease file is backed up before edits. Perform the actual lease removal only after confirming exact lease records.

### Task 4: Recreate Foundation Services in MAP UI

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/recreate-foundation/`

**Interfaces:**
- Consumes: endpoints and credentials from the user message, without persisting cleartext passwords.
- Produces: working Harbor repository record, AVI load balancer record, TKGm supplier record.

- [ ] **Step 1: Add or verify Harbor repository**

Use UI:

```text
系统设置 -> 仓库管理 -> 添加仓库
地址: https://harbor.huaruicloud.cn
用户名: user-provided Harbor admin
密码: user-provided Harbor password
连接测试 -> 添加
```

Expected: connection test succeeds and repository appears in the list.

- [ ] **Step 2: Add or verify AVI load balancer**

Use UI:

```text
系统设置 -> 负载均衡器管理 -> 添加负载均衡器
类型: AVI
地址: https://172.16.82.100
用户名: user-provided AVI admin
密码: user-provided AVI password
连接测试 -> 添加
```

Expected: connection test succeeds and load balancer appears in the list.

- [ ] **Step 3: Add TKGm cluster supplier**

Use UI:

```text
系统设置 -> 集群供应商管理 -> 添加集群供应商
类型: TKGm
管理集群 KubeConfig: /root/mgmt-kubeconfig.yaml content from 226
Supervisor/API endpoint under test: https://172.16.84.41:6443
关联 Harbor repository
关联 AVI load balancer
连接测试 -> 下一步 -> 添加
```

Expected: supplier connection test passes and supplier is available for cluster creation.

- [ ] **Step 4: Add supplier CA certificates if needed**

Use UI:

```text
集群供应商详情 -> 证书管理 -> 添加 CA 证书
```

Expected: certificates needed for external services are present before creating `test02`.

### Task 5: Create Cluster `test02`

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/cluster-create/`

**Interfaces:**
- Consumes: supplier from Task 4.
- Produces: healthy cluster `test02`.

- [ ] **Step 1: Create the cluster from UI**

Use UI:

```text
集群管理 -> 创建集群
名称: test02
供应商: recreated TKGm supplier
Ingress 控制器: enabled unless environment blocks it
控制节点: 1
工作节点: 3
规格: medium
标准库仓库: projects.packages.broadcom.com/tkg/packages/standard/repo:v2023.10.1
```

Expected: creation request succeeds and cluster status starts as `INITIALIZING`.

- [ ] **Step 2: Watch cluster until healthy**

Use UI refresh plus 226 evidence:

```bash
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml get cluster,tkc,machine -A | grep -i test02 || true
kubectl --kubeconfig /root/mgmt-kubeconfig.yaml get events -A --sort-by=.lastTimestamp | tail -80
```

Expected: MAP status becomes `HEALTHY`; backing control-plane and worker nodes are ready.

- [ ] **Step 3: Capture initialization logs if delayed or failed**

Use UI:

```text
右上角用户菜单 -> 系统日志 -> 搜索 test02
```

Expected: if initialization is not healthy within the expected environment window, capture errors before fixing.

### Task 6: Cluster Management Feature Tests

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/cluster-management/`

**Interfaces:**
- Consumes: healthy `test02`.
- Produces: validated cluster lifecycle and namespace features.

- [ ] **Step 1: Namespace CRUD**

Use UI:

```text
test02 -> 命名空间 -> 创建命名空间 map-ui-ns
编辑标签: purpose=manual-test
删除命名空间 map-ui-ns
```

Expected: namespace appears, label updates, and namespace is removed.

- [ ] **Step 2: Horizontal resize**

Use UI:

```text
test02 -> 更多操作 -> 调整工作节点个数
临时调整到 4
等待健康后再调整回 3
```

Expected: MAP and Kubernetes both show worker count changes and final count returns to 3.

- [ ] **Step 3: Vertical resize feasibility check**

Use UI:

```text
test02 -> 更多操作 -> 调整节点规格
```

Expected: UI either offers valid size options or shows a clear disabled/unavailable state. Do not permanently leave the cluster on a different size unless explicitly needed.

- [ ] **Step 4: Upgrade feasibility check**

Use UI:

```text
test02 -> 详情 -> 升级
```

Expected: if upgrade versions are available, document options and perform only after confirming enough resources and backup. If no version is available, record the expected no-op state.

### Task 7: Backup and Restore With MinIO/Velero

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/backup-restore/`

**Interfaces:**
- Consumes: healthy `test02`, MinIO endpoint, `velero` bucket.
- Produces: available backup location, successful backup, successful restore.

- [ ] **Step 1: Create backup location**

Use UI:

```text
test02 -> 数据保护 -> 添加备份存储位置
类型: MinIO
地址: http://172.16.84.22:9000
桶: velero
访问模式: READ_WRITE
连接测试 -> 创建
```

Expected: backup location status becomes `AVAILABLE`.

- [ ] **Step 2: Create test workload for backup**

Use UI via RAW component or namespace page:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backup-smoke-config
  namespace: backup-smoke
data:
  marker: map-v012-manual-test
```

Expected: object exists before backup.

- [ ] **Step 3: Create immediate backup**

Use UI:

```text
备份恢复 -> 创建备份
模式: 单次立即执行
命名空间: backup-smoke
备份数据卷: default
保留天数: 7
```

Expected: backup status reaches success.

- [ ] **Step 4: Restore with namespace mapping**

Use UI:

```text
删除/alter source namespace marker if needed
备份恢复 -> 恢复备份
命名空间映射: backup-smoke -> backup-smoke-restore
```

Expected: restore status succeeds and restored namespace contains `backup-smoke-config`.

### Task 8: Operations Services Tests

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/ops-services/`

**Interfaces:**
- Consumes: healthy `test02`, Harbor, AVI, standard package repository.
- Produces: validated operations service install/edit/open/delete flows.

- [ ] **Step 1: Install all built-in operations service suites**

Use UI:

```text
test02 -> 运维服务 -> 添加运维服务
Install resource management / Octant
Install monitoring / Prometheus
Install Grafana and associate Prometheus if UI requests it
Install logging / ElasticSearch + FluentBit
Add external service shortcut
```

Expected: every service reaches ready/enabled state.

- [ ] **Step 2: Open each ready service**

Use UI cards:

```text
Click Octant/resource manager card
Click Grafana card
Click Prometheus card if exposed
Click logging service card
Click external shortcut
```

Expected: each opens in browser without MAP-side error; capture landing page screenshots with secrets redacted.

- [ ] **Step 3: Edit and disable services**

Use UI:

```text
Edit one service configuration with a reversible change
Disable one service
Re-enable it if required by later tests
```

Expected: status transitions complete within documented 5-15 minute window or errors are captured.

### Task 9: Application, Component, Deployment, and Routing Tests

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/application-flows/`

**Interfaces:**
- Consumes: healthy `test02`, operations services, Harbor.
- Produces: validated app/component/deployment/routing flows.

- [ ] **Step 1: Create application**

Use UI:

```text
应用组件 -> 新建应用
名称: manual-app
```

Expected: `manual-app` appears in the application selector.

- [ ] **Step 2: Create BASIC component**

Use UI:

```text
添加应用组件 -> BASIC
名称: basic-nginx
ConfigMap: basic-config with key message=hello-map
Secret: basic-secret GenericSecret
PVC: basic-pvc if storage class is available
Deployment container: nginx image reachable by test02
Service port: 80
```

Expected: component is created and version 1 is visible.

- [ ] **Step 3: Create RAW component**

Use UI with format validation:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: raw-nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: raw-nginx
  template:
    metadata:
      labels:
        app: raw-nginx
    spec:
      containers:
        - name: nginx
          image: nginx:1.25
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: raw-nginx
spec:
  selector:
    app: raw-nginx
  ports:
    - port: 80
      targetPort: 80
```

Expected: validation passes and RAW component is created.

- [ ] **Step 4: Create HELM component through service catalog**

Use UI:

```text
服务目录 -> choose available chart from Harbor ChartMuseum
添加到组件 -> manual-app -> validate values -> 添加
```

Expected: HELM component appears. If ChartMuseum is empty, record that prerequisite is missing and test mapkit upload path separately.

- [ ] **Step 5: Edit, clone, and delete component**

Use UI:

```text
Edit RAW component and change replicas from 1 to 2
Confirm new component version appears
Clone RAW component to raw-nginx-clone
Delete raw-nginx-clone
```

Expected: versioning, clone, and delete behaviors match manual.

- [ ] **Step 6: Deploy components**

Use UI:

```text
Component detail -> 部署组件
Cluster: test02
Namespace: manual-app-ns, create if missing
```

Expected: deployment status reaches `PROVISIONED`; per-instance task status is `SUCCESS`; service address is visible.

- [ ] **Step 7: Update and delete deployment instances**

Use UI:

```text
Update deployed RAW component to new version
Roll back to previous version
Delete deployment instance
```

Expected: update, rollback, and delete complete without stale resources.

- [ ] **Step 8: Test L7 route**

Use UI:

```text
应用路由 -> 七层路由 -> 创建
Target: deployed service on test02
Then edit route
Then delete route
```

Expected: create/edit/delete works. If Contour or ingress class is unavailable, record prerequisite failure.

- [ ] **Step 9: Test multi-cluster route when prerequisites exist**

Use UI:

```text
应用路由 -> 多集群路由 -> 创建
Target services: available LoadBalancer service endpoints, max 2
Adjust traffic ratio if two services exist
Then edit route
Then delete route
```

Expected: route reaches `PROVISIONED` and exposes VIP. If only one cluster exists or LoadBalancer prerequisite is absent, record scoped limitation rather than forcing unsupported topology.

### Task 10: Build Management Tests

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/build-management/`

**Interfaces:**
- Consumes: healthy `test02`, Harbor.
- Produces: validated build definitions and build task behavior.

- [ ] **Step 1: Create/edit/delete binary image definition**

Use UI:

```text
构建管理 -> 创建二进制镜像
Edit definition
Delete definition after task checks
```

Expected: CRUD works.

- [ ] **Step 2: Create/edit/delete Dockerfile image definition**

Use UI:

```text
构建管理 -> 创建 Dockerfile 镜像
Use a small known-good Dockerfile package or Git source reachable from the environment
Edit definition
Delete definition after task checks
```

Expected: CRUD works.

- [ ] **Step 3: Create source-code image definition if TBS is enabled**

Use UI:

```text
构建管理 -> 创建源码镜像
```

Expected: if TBS is enabled on `test02`, build definition works; otherwise record feature unavailable due to missing extension.

- [ ] **Step 4: Start build and inspect task**

Use UI:

```text
镜像详情 -> 启动构建
Wait for PENDING/RUNNING/SUCCESS or capture FAILURE logs
Delete build task
```

Expected: build task states and logs match manual.

### Task 11: Project and User Management Tests

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/project-user-management/`

**Interfaces:**
- Consumes: super-admin account and existing MAP projects.
- Produces: validated project, membership, role, user, reset password, and normal-user profile flows.

- [ ] **Step 1: Create and edit project**

Use UI:

```text
项目管理 -> 新建项目 manual-project
Set finite CPU/memory/storage quota if UI accepts it
Edit description/quota
```

Expected: project appears and quota changes persist.

- [ ] **Step 2: Create users and assign roles**

Use UI:

```text
用户管理 -> 新建用户 manual-admin-user
用户管理 -> 新建用户 manual-normal-user
项目详情 -> 添加成员 -> assign ADMIN/USER/GUEST roles in separate checks
Edit role
Remove member
```

Expected: role permissions match manual.

- [ ] **Step 3: Reset password and normal profile change**

Use UI:

```text
超级管理员 -> 重设密码
普通用户登录 -> 用户信息 -> 修改密码
```

Expected: password reset and self-service change work.

- [ ] **Step 4: Delete test users and project**

Use UI:

```text
Remove project members
Delete manual-project
Delete manual-admin-user and manual-normal-user
```

Expected: no test users/projects remain.

### Task 12: Defect Fix Loop

**Files:**
- Modify only files identified by root-cause investigation in `map-portal/` or `mapd/`.
- Create focused tests next to the affected feature.
- Create: `.gstack/artifacts/map-v012-test02/defects/<defect-id>/`

**Interfaces:**
- Consumes: failing UI action, screenshots, console/network logs, backend logs, Kubernetes evidence.
- Produces: minimal fix, regression test, local verification, deployed binary/frontend update when needed.

- [ ] **Step 1: Capture reproducible defect evidence**

For each defect, record:

```text
Manual section
Exact UI steps
Expected result
Actual result
Screenshot
Browser console errors
Network request/response
Backend logs
Kubernetes events/logs if relevant
```

Expected: the defect is reproducible or clearly tied to environment prerequisites.

- [ ] **Step 2: Write a failing regression test**

Choose the narrowest relevant test:

```text
Frontend: component/service/unit test or Playwright test under map-portal
Backend: Go unit/integration test under mapd
E2E: Playwright flow only when the bug is cross-layer UI behavior
```

Expected: test fails before the fix for the observed reason.

- [ ] **Step 3: Implement the minimal fix**

Follow local patterns:

```text
Frontend changes stay in map-portal feature files/services.
Backend changes stay in mapd handlers/services/repositories touched by the failing behavior.
Do not change product version or default package repo version.
Do not rebuild backend Docker images for binary-only fixes.
```

Expected: failing test passes and no unrelated behavior changes.

- [ ] **Step 4: Verify and deploy fix to 226 when needed**

Run relevant tests first, then deploy according to change type:

```text
Frontend fix: build/test map-portal, deploy according to existing 226 frontend workflow.
Backend fix: build static mapd with CGO_ENABLED=0 inside mapd-binary-builder, replace mapd binary in target 226 container.
```

Expected: fixed UI flow passes on 226.

### Task 13: Final Retest and Report

**Files:**
- Create: `.gstack/artifacts/map-v012-test02/final-report.md`

**Interfaces:**
- Consumes: artifacts from all tasks.
- Produces: final pass/fail report and unresolved issue list.

- [ ] **Step 1: Re-run failed and high-risk manual flows**

Re-run:

```text
Supplier connection
Cluster test02 health
Operations service launch
Backup/restore
Application deploy/update/delete
Routes
Build task
Project/user permission checks
```

Expected: all previously failing flows pass or are documented as environment-blocked.

- [ ] **Step 2: Produce final report**

Report format:

```markdown
# MAP v0.12 Manual UI Test Report

Date:
Environment:
Cluster:

## Summary
- Passed:
- Failed:
- Blocked:
- Fixed:

## Manual Coverage
...

## Defects Fixed
...

## Remaining Risks
...

## Artifacts
...
```

Expected: report gives enough evidence for another engineer to audit the run without rerunning everything.
