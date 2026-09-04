## Purpose

让 Go-mall 各服务真正可被观测：每个服务暴露一致的 metrics 端点、向 trace collector 导出链路、输出可被集中采集的结构化日志，且本地环境可一键拉起完整观测栈。

## Requirements

### Requirement: 服务暴露 Prometheus metrics 端点

每个 RPC 服务与 HTTP 网关 SHALL 在独立端口暴露 `/metrics` 端点供 Prometheus 抓取。指标端口 MUST 遵循 `指标端口 = 服务端口 + 1000` 规则（如服务监听 10004 则指标端口为 11004）。指标端点 SHALL 在服务启动后即可达，且 `/metrics` 返回 Prometheus 文本格式指标。

#### Scenario: 服务启动后 metrics 端点可达
- **WHEN** 一个服务以服务端口 10006 启动
- **THEN** 端口 11006 上的 `/metrics` 返回 200 与 Prometheus 格式指标体

#### Scenario: 指标端口遵循加 1000 规则
- **WHEN** 任意服务的服务端口为 100xx
- **THEN** 该服务的指标端口为 110xx，且与 Prometheus 抓取配置中的 target 端口一致

#### Scenario: prod 模式下 metrics 端点仍可达
- **WHEN** 服务以生产模式（非 dev 模式）启动且指标端点配置为启用
- **THEN** `/metrics` 端点依然被暴露并返回指标

### Requirement: Prometheus 抓取目标与服务暴露端口一致

`prometheus.yml` 的所有抓取 target 端口 MUST 与各服务实际暴露的指标端口逐一对应，不存在 target 指向无人监听端口的情况。

#### Scenario: 全部 target 健康抓取
- **WHEN** 所有服务已启动且 Prometheus 运行
- **THEN** Prometheus 的每个 job target 状态为 UP，无 connection refused

#### Scenario: 抓取地址对容器内 Prometheus 可达
- **WHEN** Prometheus 以 bridge 网络容器运行而服务进程跑在宿主机
- **THEN** 宿主机服务的 target 使用容器可达地址（如 `host.docker.internal`），抓取成功返回指标

### Requirement: 服务向 trace collector 导出链路

每个服务 SHALL 通过配置的 trace collector 端点导出分布式链路，配置项包含 collector endpoint、batcher 类型与采样率。一次跨服务调用 SHALL 产生可在 Jaeger UI 检索到的完整 trace。

#### Scenario: 跨服务调用产生可检索 trace
- **WHEN** 网关发起一次经 checkout→order→inventory 的调用
- **THEN** Jaeger UI 中可按入口服务检索到包含各下游 span 的完整 trace

#### Scenario: 采样率配置生效
- **WHEN** 服务 trace 配置的采样率小于 1.0（如 0.1）
- **THEN** 该服务导出的 trace 数量明显少于全量，且采样行为由配置项控制

### Requirement: 服务输出可被集中采集的结构化日志

每个服务 SHALL 输出 JSON 格式日志并携带 service 标签（与 `scripts/logs/<service>.log` 文件名对应），使 Promtail 能按 service 标签抓取并推送到 Loki。日志文件路径 MUST 与 Promtail 的 `__path__` 配置匹配。

#### Scenario: 日志按服务标签进入 Loki
- **WHEN** order 服务写一条 error 级日志
- **THEN** Grafana/Loki 中以 `{service="order"}` 可检索到该条日志

### Requirement: 本地一键拉起完整观测栈

本地统一启动脚本 SHALL 在单条命令下拉起日志（Loki+Promtail+Grafana）、链路（Jaeger）、指标（Prometheus）三栈。停止命令 SHALL 清理全部三栈。

#### Scenario: 启动后三栈 UI 均可达
- **WHEN** 执行本地统一启动命令
- **THEN** Grafana(:3001)、Jaeger UI(:16686)、Prometheus(:9090) 均可访问

#### Scenario: 停止后三栈全部退出
- **WHEN** 执行统一停止命令
- **THEN** 三栈相关容器均不再运行

### Requirement: 不存在与原生机制重复的观测性死代码

代码库 MUST NOT 包含从未被调用且与框架原生观测性机制重复的初始化代码或死配置。

#### Scenario: 删除未引用的 tracing 工具包
- **WHEN** 检查 `common/utils/tracing/` 包
- **THEN** 该包已移除，全仓库无对其导出函数的引用，服务 trace 接线仅依赖框架原生配置

#### Scenario: 删除未被消费的 PrometheusExt 死配置
- **WHEN** 检查 auths/users/product/order/checkout/payment/inventory/audit/coupons/system/activity/admin 共 12 个服务的 etc yaml
- **THEN** 无 `PrometheusExt` 配置块残留（其中除 users 外的 11 个 config.go 亦无 `PrometheusExtConf` 字段残留），其 metrics 端点由 `DevServer` 提供

#### Scenario: search 活配置例外保留
- **WHEN** 检查 search 服务
- **THEN** 其承载 `/metrics` 与 `/stats/rag/*` 业务 API 的 stats HTTP server 正常运行，指标端口遵循"服务端口 + 1000"规则（8081→9081）
