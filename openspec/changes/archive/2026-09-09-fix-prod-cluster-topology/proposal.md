## Why

线上 k3s 集群存在跨节点 RPC 不可达：`product/carts/checkout` 服务在 worker 节点，调用迁移至 master 的 hostNetwork 服务（`inventory/order/coupons/payment/dtm`）时**跨节点超时**（实证：`product→inventory` 报 `context canceled while waiting for connections`，gateway 调 `/api/v1/products` 返回 504），导致商品/购物车/下单链路不可用、核心链路无法压测。该 hostNetwork 拓扑源自前序 settlement-fallback-layers 变更——为满足 dtm 回调的单 `BusiHost` 约束（dtm 回调走 master 节点 IP:10004/10007/10009）而引入。同时全部业务 deployment 为 `Recreate` + 单副本，更新存在停机窗口，无滚动更新零停机能力。

## What Changes

- **统一业务节点拓扑**：将所有业务服务（`product/carts/checkout` 及现有 master 上服务）通过 `nodeSelector` 统一调度到 master 节点（`vm-4-3-ubuntu`），worker 节点仅承载基础设施（postgres/redis/elasticsearch/rabbitmq/minio/etcd）。消除跨节点 `pod→节点IP:hostNetwork端口` 的不可达路径。**保留约束**：hostNetwork + 单 BusiHost 回调架构不变（仅从"跨节点"收敛为"同节点"），不得改为普通 ClusterIP 组网破坏 dtm 回调链路。
- **部署策略改造**：业务 deployment 由 `Recreate` 改为 `RollingUpdate`；**网关（gateway，入口）双副本 + RollingUpdate 实现零停机**；其余服务按 master 资源评估明确 `maxSurge`/`maxUnavailable` 语义（资源不允许双副本时，明确"单副本滚动 = 仍有短暂窗口"的取舍，即零停机仅承诺给网关）。
- **回归验证（含量化证据）**：迁移后实证商品链路（`product→inventory`）、下单链路（`checkout→order→inventory→dtm` 回调）可达并**记录可达性/延迟数字**；核心链路压测复跑并记录 QPS/错误率；滚动更新期间记录 gateway 连接中断计数（预期 0）。运维改动提供回滚方式。

## Capabilities

### New Capabilities
- `cluster-service-topology`: 线上集群服务部署拓扑与更新可靠性——业务服务节点归属（互调可达、无跨节点 hostNetwork 依赖）与更新契约（网关滚动更新零停机；其余服务 RollingUpdate 语义明确、可含受限资源下的短暂窗口）。

### Modified Capabilities
<!-- 无：本 change 为运维/部署拓扑行为，不改变既有业务运行时能力 spec（async-settlement-reliability 等的业务逻辑不变）。 -->

## Impact

- **线上系统**：k8s deployment 配置（`nodeSelector`/`strategy`/`replicas`）、线上网络路径、可用性/更新行为。涉及 ~12 个业务服务 deployment（含网关双副本）。
- **代码**：无业务代码改动（纯运维拓扑/配置变更；如需为滚动更新调整资源 requests/limits，属配置）。
- **压测/验证链路**：下单与商品链路恢复后可压测。
- **回滚**：所有改动为 k8s 配置，可通过恢复原 `nodeSelector`/`strategy`/`replicas` 回退；不影响业务数据。
- **风险**：master 节点 2C2G 资源上限（R1）；基础设施跨节点调用（overlay）需实证（R4）；滚动更新双副本受资源约束（R2）。