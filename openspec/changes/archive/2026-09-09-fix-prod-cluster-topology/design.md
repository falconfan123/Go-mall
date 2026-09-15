## Context

- 见 `proposal.md` — Why（跨节点 RPC 不可达 + Recreate 无零停机）。
- 当前节点分布（实证，2026-09-08）：
  - **master**（vm-4-3-ubuntu，Allocatable 内存 1720112Ki ≈ 1.64Gi）：gateway / users / auths / system / order / inventory / coupons / payment / dtm / **postgres / redis / etcd**（基础设施中 DB/缓存/注册中心在 master）。
  - **worker**（vm-0-5-ubuntu）：product / carts / checkout（待迁业务）+ elasticsearch / rabbitmq / minio（基础设施，留 worker）。
- master 当前内存 requests 合计 ≈ 1.2Gi（13 个 pod，多数 64–128Mi）；limits 合计 ≈ 2.5Gi（超 Allocatable，无流量下未触发 OOM）。
- dtm 回调依赖 hostNetwork + 单 `BusiHost`（master 节点 IP:10004/10007/10009），该架构保持不变（spec Requirement 2）。

## Goals / Non-Goals

- **Goals**：product/carts/checkout 收敛到 master，消除跨节点 hostNetwork 不可达；gateway 滚动更新零停机；其余服务 RollingUpdate 语义明确；迁移/更新以实证数字验收。
- **Non-Goals**：不改业务代码/逻辑；不引入多活/负载均衡/新服务；不迁移基础设施（postgres/redis 已随 master；ES/rabbitmq/minio 留 worker）。

## Decisions

### D1 节点收敛：nodeSelector 统一 master（已定，附备选）

- **选定**：给待迁 3 服务（product/carts/checkout）加 `nodeSelector: go-mall/node-role: node1`（master 标签实证存在），与其余业务一致。worker 不再调度业务 pod。
- **备选（已评估）**：
  - 网络层修复（修 flannel/kube-proxy 使 worker pod → master 节点 IP:hostNetwork 跨节点可达）：不确定性高，且治标不治本（跨节点仍存在）。
  - 改 dtm 回调架构（每服务独立回调地址、放弃单 BusiHost）：改动大、触碰已验收的 settlement-fallback-layers 链路，非本 change 目标。
- **依据**：pod→pod（overlay）跨节点通、pod→节点IP:hostNetwork 跨节点不通（实证），同节点后两者皆通且最简单。

### D2 基础设施访问：经 ClusterIP overlay（无需改动）

- postgres/redis/etcd 已在 master；ES/rabbitmq/minio 在 worker。业务服务一律经 Service ClusterIP 访问，kube-proxy 保证同/跨节点可达（迁移前 master 业务已如此访问，实证无异常）。迁移后无新增基础设施访问路径风险；验收补一条实证（master pod → worker 上 ES/rabbitmq/minio 的 ClusterIP 连通）。

### D3 滚动更新策略：网关双副本零停机，其余单副本滚动

- **gateway**：`replicas: 2` + `strategy: RollingUpdate{maxSurge: 1, maxUnavailable: 0}` → 更新期间新旧各 1 副本并存，入口零中断。**前置条件**：gateway deployment MUST 具备就绪探测（readinessProbe，TCP 探测 8888）——`maxUnavailable: 0` 仅在"新副本就绪后才切流、旧副本保持到新就绪"下成立；若无 readinessProbe，须先补充（属配置补全，非业务代码）。
- **其余业务服务**：`RollingUpdate{maxSurge: 0, maxUnavailable: 1}`（单副本滚动，先停旧再起新，存在秒级窗口）→ 与 spec"其余服务可含短暂窗口"的取舍一致；`maxSurge: 0` 避免更新期临时双副本挤占 master 内存。
- **dtm 回调目标**（order/inventory/coupons）单副本滚动时窗口期 saga 分支可能偶发不可达 → dtm 自带失败重试（cron 退避），可自愈，无需为此保双副本（trade-off 见 R2）。

### D4 资源可行性（实证支撑）

- master 内存 1.64Gi；迁入 product（req ~50–100Mi）/carts（~50–100Mi）/checkout（~100Mi）后 requests ≈ 1.4–1.5Gi，**贴边但可行**（实际使用远低于 requests，因无自然流量）。
- **双副本仅授予 gateway**（+64Mi req / +128Mi lim → requests ≈ 1.5Gi），且作为可选：若实测 OOM/内存紧张，回退 gateway 单副本 + `maxUnavailable: 1`（零停机承诺降级为网关短暂窗口，记录取舍）。
- **CPU（2C）边界**：master 为 2C，承载 ~13 pod + 网关双副本。压测回归（步骤 6）会制造流量、可能推高 CPU；本 change 的验收目标是**链路可达性与更新零停机**，CPU 饱和/吞吐上限不构成本 change 验收失败项（压测数字仅作链路恢复实证，注明非容量测试）。

### D5 迁移顺序：先改策略后迁移（最小中断）

- 先为待迁服务设置 `RollingUpdate`（避免以 Recreate 迁移造成不必要停机），再施加 `nodeSelector`（触发重建至 master）。两步可通过一次 `kubectl patch` 合并完成，但顺序上保证最终态为 RollingUpdate。

## Risks / Trade-offs

- **[master 内存贴边（1.5Gi/1.64Gi）]** → 迁移后监控节点 `kubectl top node`；gateway 双副本观察实际使用，超限即回退单副本（D4 预案）；不引入其余双副本。
- **[单副本滚动窗口期 dtm 回调偶发失败]** → dtm 分支失败自动退避重试（实证：TouchCronTime 指数退避），saga 终态可达，仅延迟增大。
- **[worker 仅剩基础设施的依赖链]** → ES/rabbitmq/minio 经 ClusterIP overlay 访问（实证后固化）；若这些服务自身异常不影响业务节点拓扑。
- **[迁移触发 pod 重建的瞬时中断]** → 测试环境可接受；滚动/重建期间由 gateway 与客户端重试吸收（无自然流量，影响面可控）。

## Migration Plan

1. **改策略**：product/carts/checkout 的 `strategy` → `RollingUpdate{maxSurge:0, maxUnavailable:1}`。
2. **迁节点**：三服务加 `nodeSelector` master，rollout 至 Running；**显式验证** `kubectl get node --show-labels`（确认 `go-mall/node-role=node1` 落在 vm-4-3-ubuntu，输出留证），确认 etcd 注册地址变为 master 侧可达（worker 上的业务 pod 消失）。
3. **链路验证**：`gateway→/api/v1/products`（product→inventory）、`checkout→order`、dtm 结算 saga（order/inventory/coupons 三分支）实证可达；master pod→worker 基础设施 ClusterIP 连通。
4. **策略改造**：gateway `replicas:2` + RollingUpdate{1,0}；其余业务 RollingUpdate{0,1}。
5. **零停机验证**：gateway 滚动更新期间持续探活（后台 `curl` 打点 `/api/v1/products` 或 gateway 端口，统计失败计数），连接中断计数 = 0。观测手段：滚动期间起一个长跑探活循环，记录时间戳序列，结束后统计非 200 个数。
6. **压测回归**：下单/商品链路压测（ab 或探针），记录 QPS/错误率。
7. **记录实证**：全部数字写入验证记录（附可复现命令）。

**回滚**：恢复各 deployment 原 `nodeSelector`/`strategy`/`replicas` 即回到迁移前拓扑；全程不触碰业务代码与业务数据。

## Open Questions

- **saga 终态判定窗口**：spec"下单链路"Scenario 中"dtm saga 到达终态"的验收时限（默认 30s 内，dtm cron 间隔下足够）——由 tasks 落地时在验证命令中固化，不改变 spec/方案。
- **gateway 双副本真实内存占用**：实测后定（D4 预案覆盖），无需改方案。