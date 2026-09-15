> 本 change 为线上运维拓扑/配置变更（k8s deployment），无业务代码改动；验证命令均在 gomall 集群执行（kubectl 已通，default context 指向腾讯云 k3s）。任务粒度 ≤2h。线上操作前先读 `docs/servers.md`。

## 1. 前置：策略改造 + 归属留证

- [x] 1.1 留证当前节点标签归属与业务 pod 分布：`kubectl get node --show-labels`（确认 `go-mall/node-role=node1` 落在 vm-4-3-ubuntu）+ `kubectl get pods -n go-mall -o wide` 输出存档；**同时核对已迁服务（payment/order/inventory/coupons/dtm）deployment 均已带 `nodeSelector: {go-mall/node-role: node1}`**（防未来误调度到 worker）；验证：命令输出留档于验证记录
- [x] 1.2 将 product/carts/checkout 三个 deployment 的 `strategy` 改为 `RollingUpdate{maxSurge:0, maxUnavailable:1}`（避免以 Recreate 迁移）；验证：`kubectl get deploy product-rpc carts-rpc checkout-rpc -n go-mall -o jsonpath='{.items[*].spec.strategy}'` 显示 RollingUpdate

## 2. 节点迁移（跨节点可达性修复）

- [x] 2.1 给 product/carts/checkout 加 `nodeSelector: {go-mall/node-role: node1}` 并 rollout 至 Running；验证：`kubectl get pods -n go-mall -o wide | grep -E 'product|carts|checkout'` 全部 NODE=vm-4-3-ubuntu，worker 上无业务 pod
- [x] 2.2 迁移后链路可达实证：`kubectl port-forward -n go-mall svc/gateway 18888:8888` 后 `curl -s -o /dev/null -w '%{http_code} %{time_total}s' 'http://localhost:18888/api/v1/products?page=1&page_size=5'` 返回非 504（product→inventory 库存读取成功）；记录响应码与耗时（gateway 地址解析方式固定为 `kubectl port-forward svc/gateway` 到本地端口）
- [x] 2.3 基础设施可达实证：从 master 业务 pod（`kubectl exec deploy/payment-rpc -n go-mall`）验证访问 worker 上基础设施 Service ClusterIP 连通——`elasticsearch`（10.43.41.167:9200）、`rabbitmq`（10.43.17.51:5672）、`minio`（10.43.121.205:9000），以及 master 上 `postgres`（10.43.132.5:5432）、`redis`（10.43.206.14:6379）；验证：TCP 连通成功记录（`/dev/tcp/<ip>/<port>` 或 curl -v）

## 3. 链路与结算回归

- [x] 3.1 下单链路实证：触发一条测试下单（走 checkout→order→inventory 全链路，可用现有 e2e 探针或网关下单接口），订单状态流转到待结算/已结算；验证：orders 表该单状态 + 各服务日志无跨节点超时（`context canceled while waiting for connections`）
- [x] 3.2 dtm 结算 saga 三分支回归：构造一条结算事件（复用 `services/payment/e2e_fallback_probe_test.go` 或等价注入），saga 经 dtm 回调 order/inventory/coupons；验证：`SELECT gid,status FROM trans_global` 该 gid 30s 内到达 succeed/failed 终态（无 submitted 挂起），订单/库存一致

## 4. 部署策略改造（滚动更新）

- [x] 4.1 确认/补齐 gateway deployment 的 readinessProbe（TCP 探测 8888），`maxUnavailable:0` 前提成立；验证：`kubectl get deploy gateway -n go-mall -o jsonpath='{.spec.template.spec.containers[0].readinessProbe}'` 存在
- [x] 4.2 gateway 改 `replicas:2` + `RollingUpdate{maxSurge:1, maxUnavailable:0}` 并 rollout；验证：`kubectl get deploy gateway -n go-mall -o jsonpath='{.spec.replicas}{.spec.strategy}'`，两副本 Ready；若 OOM/内存不足（`kubectl top node` 内存告警）→ 回退单副本 + maxUnavailable:1 并记录取舍（design D4 预案）
- [x] 4.3 其余全部业务 deployment（product/carts/checkout/order/inventory/coupons/payment/users/auths/system/**dtm**）统一 `RollingUpdate{maxSurge:0, maxUnavailable:1}`（含 dtm——避免 saga 编排器 Recreate 中断）；验证：各 deployment strategy 显示 RollingUpdate，共 11 个

## 5. 零停机验证 + 压测回归

- [x] 5.1 gateway 滚动更新零停机验证：后台探活 `while true; do curl -s -o /dev/null -w '%{http_code} %{time_total}\n' http://localhost:18888/api/v1/products?page=1\&page_size=5 >> /tmp/gw-probe.log; sleep 0.2; done &`，同时 `kubectl rollout restart deploy/gateway -n go-mall`，结束后 kill 探活并 `awk '$1!=200' /tmp/gw-probe.log | wc -l` 统计；探活覆盖整个 rollout 周期 + 10s 缓冲；验证：**非 200 计数 = 0**
- [x] 5.2 核心链路压测回归（ab）：`ab -n 1000 -c 50 -p /tmp/bench-login.json -T application/json http://localhost:18888/api/v1/users/login`，以及 `/api/v1/products` 读接口；验证：记录 QPS/错误率（注明工具/并发/机器；**本 change 验收目标为链路可达性，CPU 饱和/低吞吐不构成验收失败**，压测仅作链路恢复实证、非容量测试）

## 6. 实证记录与收尾

- [x] 6.1 汇总验证记录（节点归属留证、链路可达性/延迟、saga 终态、滚动中断计数、压测数字 + 可复现命令），追加到 **`docs/resume_metrics_report.md`**（新增"拓扑修复实证"节，唯一文档路径）；验证：文档含全部数字与命令
- [x] 6.2 对照 spec 各 Scenario 逐项核对（节点同驻/商品链路/下单链路/基础设施可达/saga 三分支/网关零停机/其余滚动语义/实证/回滚），确认无未验收项；验证：核对清单随验证记录输出（建议按 `R1-S1..S4` 标注对应任务）
- [x] 6.3 记录 master 节点资源基线：迁移前与 gateway 双副本后各跑一次 `kubectl top node`（内存/CPU），对照 design D4 的 1.5Gi/1.64Gi 估算给出实测；验证：两次 `kubectl top node` 输出存档
- [x] 6.4 确认全程未修改业务代码与业务数据（仅 k8s 配置与只读验证），回滚路径（恢复 nodeSelector/strategy/replicas）记录在案；验证：`git status` 无业务代码改动（仅 openspec/文档），deployment 原配置可复现

## 7. 门禁（Rule 1 适用性说明）

- [x] 7.1 本 change 无业务代码改动，Rule 1（编译/单测/lint）不适用；如需校验仓库完整性，运行 `make lint` 确认不因本 change 引入回归；验证：`make lint` 通过