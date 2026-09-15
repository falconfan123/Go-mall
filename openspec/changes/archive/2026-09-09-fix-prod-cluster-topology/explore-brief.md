# Explore Brief · fix-prod-cluster-topology

> 探索会话（2026-09-08，线上 k3s 集群实证采集）结论落盘，作为提案审核的需求基线 checklist。

## 背景（实证）

- 线上为腾讯云轻量 k3s 双节点：master `vm-4-3-ubuntu`（2C2G）、worker `vm-0-5-ubuntu`（2C2G），命名空间 `go-mall`。
- settlement-fallback-layers 线上验证时，为满足 dtm 回调的单 `BusiHost` 约束，将 `payment/order/inventory/coupons/dtm` 改为 **hostNetwork + nodeSelector 固定 master**（dtm 回调走 master 节点 IP:10004/10007/10009）。
- **遗留副作用（实证）**：`product-rpc`/`carts-rpc`/`checkout-rpc` 仍在 worker 节点。worker pod 调 master 节点 IP:hostNetwork 端口**跨节点超时**（`product→inventory` GetInventory 报 `context canceled while waiting for connections`；gateway 调 `/api/v1/products` 返回 504）。pod→pod（flannel overlay）跨节点正常，pod→节点IP:hostNetwork端口跨节点不通。
- **部署策略**：全部业务 deployment 为 `Recreate` + 单副本，无滚动更新零停机能力。
- 线上无自然业务流量（orders 仅探针单、users 0 用户），压测下单链路被跨节点问题阻塞。

## 目标 / 非目标

- **目标**：恢复全业务链路跨服务 RPC 可达（下单/商品/购物车链路可通、可压测）；部署策略改为 RollingUpdate，网关实现零停机更新。
- **非目标**：不改业务代码/业务逻辑；不做多活/负载均衡改造；不引入新服务。

## 待决策点（已由用户拍板）

- D1 范围：跨节点可达性 + 部署策略改造（RollingUpdate）。
- D2 修复方案：**全部业务服务 nodeSelector 统一到 master**；worker 仅承载基础设施（elasticsearch/rabbitmq/minio，postgres/redis/etcd 已在 master；pod→pod overlay 跨节点通）。

## 待探索/风险（提案审核需核对）

- R1 资源：master 2C2G 承载全部业务服务（~12 个 pod）的资源可行性；多副本（RollingUpdate 需要 2 pod 并存）内存是否足够——零停机与 2C2G 的矛盾需在 design 给出取舍。
- R2 滚动更新落地：哪几个服务可双副本（如 gateway 入口零停机优先），其余 RollingUpdate 单副本（maxUnavailable 语义）的资源适配。
- R3 迁移后回归：下单链路（checkout→order→inventory→dtm）、商品链路（product→inventory）互调实证；压测可复跑。
- R4 基础设施跨节点调用：实证归属 postgres/redis/etcd 在 master，elasticsearch/rabbitmq/minio 在 worker；master 业务 pod 调 worker 基础设施（ES/rabbitmq/minio 的 ClusterIP）经 overlay 应可达，需实证。
- R5 只读约束：线上操作（迁移/rollout）是运维动作，非代码变更；验证不得污染业务数据。

## Checklist（反思审核基线）

- [ ] C1 全部业务服务（含 product/carts/checkout）统一 master 节点，无跨节点 hostNetwork 调用残留。
- [ ] C2 worker 上仅存基础设施，业务服务不再调度到 worker。
- [ ] C3 下单链路（checkout→order→inventory→dtm 回调）与商品链路（product→inventory）实证可达。
- [ ] C4 部署策略 RollingUpdate；网关（入口）零停机；其余服务明确 maxSurge/maxUnavailable 语义。
- [ ] C5 迁移与策略改动在线上实测（rollout 观察、压测回归），记录实证数字。
- [ ] C6 不修改业务代码与业务数据；运维改动可回滚。