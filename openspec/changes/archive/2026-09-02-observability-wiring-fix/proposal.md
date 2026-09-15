## Why

项目已有 Logs（Loki）/ Traces（Jaeger）/ Metrics（Prometheus）三套观测性栈的配置文件，但实际接线严重断裂：Metrics 端点配置了端口却无人 listen（`PrometheusExt` 是死配置，main.go 从未消费）；`common/utils/tracing` 整个包从未被任何服务调用；prometheus.yml 的抓取目标与服务实际暴露端口三方对不上；本地一键启动脚本只起日志栈不起 traces/metrics 栈。结果是"有配置无数据"——本地三项里只有日志能真正看到，线上则全部没跑起来。需要现在修复，因为观测性空转让 payment-500 等线上问题无从排查，且死代码与端口混乱持续误导维护者。

## What Changes

- **统一 Metrics 端点机制**：删除自定义 `PrometheusExt` **死配置**——etc yaml 块 12 个服务（auths/users/product/order/checkout/payment/inventory/audit/coupons/system/activity/admin），其中 11 个另在 config.go 有 `PrometheusExtConf` 字段与结构体（users 仅 yaml 块）；carts 无任何 PrometheusExt，不在清理范围。改用 go-zero 原生 `DevServer`（已验证在 `Mode: pro` 下仅依赖 `Enabled: true`，暴露 `/metrics` + `/healthz` + 可选 pprof）。**search 例外**：其 `PrometheusExt` 是活配置——stats HTTP server（承载 `/metrics` + `/stats/rag/overview|trend|breakdown|events` 四个业务 API）运行时消费该配置，机制保留，仅端口对齐 +1000 规则。
- **统一端口规划**：所有服务（dev 与 prod）采用 `指标端口 = 服务端口 + 1000` 规则（已在 prod 配置中确立：order 10004→11004，carts.prod.yaml 注明此规则），为缺失的 dev 配置补 `DevServer`，为 gateway（HTTP 服务）补 metrics 端点（8888→9888）。system/activity/admin/search/gateway 无 prod 配置文件，其 prod 侧 DevServer 随线上部署独立 change 处理。
- **重写 prometheus.yml**：抓取目标对齐 `指标端口 = 服务端口 + 1000` 规则，修正当前 9xxx/错位 11000 的混乱；抓取 host 改为 Prometheus 容器可达地址（`host.docker.internal`，当前 localhost 在容器内指向容器自身，宿主机服务一个都抓不到）。
- **删除死代码**：移除 `common/utils/tracing/` 包（`InitJaeger`/`InitWithOtelCollector` 从未被调用，且 `InitJaeger` 硬编码 localhost:14250 忽略入参），服务继续靠 go-zero 原生 `Telemetry` 配置自动接线 trace exporter。
- **本地一键启动补全**：`scripts/start-unified.sh` 当前只起 `LOKI_STACK`，补起 `construct/observability` 的 Jaeger + Prometheus，使本地三项栈一键全启。
- **修正文档路径**：`docs/jaeger/` 与 `docs/logging-system.md` 标注两处 compose 的关系，说明 start-unified 现已统一拉起。
- **BREAKING**：`PrometheusExt` 配置字段从 12 个服务移除（etc yaml 块；其中 11 个连带 config.go 字段，users 仅 yaml 块），search 活配置除外；依赖被删字段的外部脚本需迁移到 `DevServer`。search 的字段名与机制不变。

## Capabilities

### New Capabilities
- `observability`: 三项可观测性（logs/traces/metrics）的本地采集与展示能力——服务必须暴露一致的 metrics 端点、对接 trace collector、输出可被 Promtail 抓取的日志，且本地环境一键拉起完整观测栈。

### Modified Capabilities
<!-- 无既有 spec，首次引入 observability 能力 -->

## Impact

- **配置文件**：12 个服务 dev yaml 删 `PrometheusExt` 块并补/核对 `DevServer`（search 仅调 stats 端口；carts 与 gateway 仅补 DevServer、无 PrometheusExt 可删）；prod 配置 10 个已存在文件不动（DevServer 已配齐）；`construct/observability/conf/prometheus.yml` 重写 targets 端口与 host。
- **代码**：11 个服务 `internal/config/config.go` 删 `PrometheusExtConf` 字段与结构体定义（12 个死配置服务中除 users 外，grep 实证）；`common/utils/tracing/` 整个包删除。search 与 carts 的代码不动。
- **脚本**：`scripts/start-unified.sh`（补起 observability 栈 + 指标端口预检）、`scripts/service-supervisor.sh` 无需改动（日志重定向链路已通）。
- **依赖**：无新增依赖，go-zero v1.10.1 原生 DevServer + prometheus client 已在 go.mod。
- **非目标**：线上服务器（gomall-1/2）观测性栈部署、k8s 全量服务 manifest、k8s Prometheus/Loki DaemonSet、告警规则与仪表板预设、OpenTelemetry Collector 中间层、search stats server 业务 API（`/stats/rag/*`）迁移与配置字段重命名——均延后独立 change。
- **说明**：文档路径修正（What Changes 第 6 条）为文档一致性修正，非运行时行为，不设独立 spec requirement，由任务 5.6 承载与验证。
