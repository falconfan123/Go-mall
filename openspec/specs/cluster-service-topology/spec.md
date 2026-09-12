# cluster-service-topology Specification

## Purpose

线上 k3s 集群的业务服务部署拓扑契约：业务服务同节点互调可达（无跨节点 hostNetwork 依赖）、dtm 回调链路保持、更新流程具备明确的停机语义（网关零停机）。

## Requirements

### Requirement: 业务服务节点同驻与互调可达

所有业务服务（product/carts/checkout/order/inventory/coupons/payment/gateway 等）MUST 部署在 master 节点；worker 节点 MUST 不承载任何业务服务负载，仅承载基础设施（实证：elasticsearch/rabbitmq/minio 在 worker；postgres/redis/etcd 已在 master）与 k3s 系统组件。业务服务之间的 RPC 调用 MUST 可达，不得因跨节点访问 hostNetwork 端口而失败。

#### Scenario: 跨节点调用收敛为同节点可达
- **WHEN** 迁移前 worker 节点上的业务服务（如 product）调用 master 节点 hostNetwork 服务（如 inventory）
- **THEN** 迁移后该调用 MUST 成功（不再出现 `context canceled while waiting for connections` / gateway 504）

#### Scenario: 商品链路可达
- **WHEN** gateway 请求 `/api/v1/products`（product→inventory 取库存）
- **THEN** 响应 MUST 成功（非 504），且返回的库存来自 inventory 服务

#### Scenario: 下单链路可达
- **WHEN** checkout 服务提交订单（checkout→order→inventory 全链路）
- **THEN** 各服务 RPC MUST 成功，订单进入结算（dtm saga 到达终态）

#### Scenario: 业务服务访问基础设施可达
- **WHEN** master 节点的业务 pod 访问 worker 节点上的基础设施（elasticsearch/rabbitmq/minio 的 Service ClusterIP）或 master 上的基础设施（postgres/redis）
- **THEN** 连接 MUST 成功（overlay/ClusterIP 路径不受 nodeSelector 收敛影响）

### Requirement: dtm 结算回调链路保持

结算 saga 的分支回调（order/inventory/coupons 的 action/compensate）MUST 在收敛后的拓扑下可达并可完成；hostNetwork + 单 BusiHost 回调架构 MUST 保留（不得改为普通 ClusterIP 组网破坏 dtm 回调）。

#### Scenario: 结算 saga 三分支回调完成
- **WHEN** payment 经 dtm 提交结算 saga（含 order/inventory/coupons 分支）
- **THEN** 各分支（含补偿路径）MUST 执行成功或按预期失败，trans_global 到达 succeed/failed 终态，无 submitted 挂起，且订单/库存状态一致

### Requirement: 更新停机语义明确

业务服务部署策略 MUST 为 RollingUpdate；gateway（入口）MUST 在滚动更新期间保持零连接中断；其余服务 MUST 声明并遵循明确的 `maxSurge`/`maxUnavailable` 语义（声明位置 = 各 deployment 的 `strategy.rollingUpdate` 字段本身；在 master 资源不足以支持双副本时，MUST 明确"单副本滚动存在短暂窗口"的取舍，验收按该声明判定）。

#### Scenario: 网关滚动更新零停机
- **WHEN** 对 gateway 执行滚动更新（kubectl rollout restart）
- **THEN** 更新期间连续探活 MUST 无连接中断（错误计数 0），新旧副本平滑切换

#### Scenario: 其余服务滚动更新语义
- **WHEN** 对非网关业务服务执行滚动更新
- **THEN** 更新行为 MUST 符合其 deployment 声明的 `strategy.rollingUpdate`（maxSurge/maxUnavailable）语义，且更新后服务恢复可用、无业务数据损坏

### Requirement: 迁移结果可实证

拓扑收敛与更新契约的达成 MUST 以线上可复现的实证为准（链路可达性、延迟、压测 QPS/错误率、滚动期间连接中断计数），而非仅配置声明。

#### Scenario: 实证记录
- **WHEN** 完成迁移与策略改造
- **THEN** 报告 MUST 记录各链路可达性/延迟数字、压测数字、滚动更新中断计数，并附可复现的命令/SQL

### Requirement: 可回滚与数据保护

拓扑收敛与策略改造 MUST 可回滚（恢复原 `nodeSelector`/`strategy`/`replicas` 后行为回到迁移前），且整个迁移与验证过程 MUST 不修改业务代码、不污染业务数据。

#### Scenario: 回滚与数据不变
- **WHEN** 执行回滚（恢复迁移前配置）或验证过程中访问业务数据
- **THEN** 拓扑/更新行为 MUST 恢复到迁移前状态，业务数据 MUST 无任何变更（探针/压测仅使用测试数据）