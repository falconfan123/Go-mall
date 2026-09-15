## Context

当前三项观测性栈已定义但接线断裂（详见 proposal.md - Why 与 explore-brief.md 的 D1–D5 实证）。关键现状约束：

- go-zero v1.10.1 的 `DevServer`（`internal/devserver/server.go`）暴露 `/metrics`、`/healthz`、可选 pprof；`StartAgent` 仅检查 `Enabled`，**不依赖 Mode**——`Mode: pro` 下 `Enabled: true` 即生效（已读源码验证）。`EnableMetrics` 默认 `true`。
- prod 配置已确立"指标端口 = 服务端口 + 1000"规则（`carts.prod.yaml` 注释明确，order 10004→11004 等 10 个服务已配 `DevServer`）；prod yaml 仅 10 个服务存在，system/activity/admin/search/gateway **无 prod 配置文件**（其 prod 侧 DevServer 随线上部署独立 change）。dev 配置中仅 auths（11000）、audit（11008）已有 DevServer，其余 12 个 dev 配置缺失（含 carts，其无任何 PrometheusExt 可清理）；12 个服务的 dev yaml 含 `PrometheusExt` 死配置块（其中 11 个 config.go 另有 `PrometheusExtConf` 字段与结构体，users 仅 yaml 块），**search 为活代码**（stats HTTP server 运行时消费，且承载 4 个 `/stats/rag/*` 业务 API）。
- `prometheus.yml` 抓取目标与上述规则完全错位：order 抓 9104（实际 11004）、users 抓 11000（实际应是 11001，11000 属于 auths）；且全部 target 用 `localhost`，而 Prometheus 跑在 bridge 网络容器内（docker-compose 无 `network_mode: host`），容器内 localhost 指向容器自身——宿主机上的服务进程一个都抓不到，当前唯一能通的是 9090 自抓。
- 日志链路本地已通（supervisor 重定向 stdout→`scripts/logs/<svc>.log`→Promtail→Loki），无需改动。
- `common/utils/tracing/tracing.go` 的 `InitJaeger`（硬编码 localhost:14250，忽略入参）与 `InitWithOtelCollector` 从未被任何 main.go 调用；服务靠 go-zero 原生 `Telemetry` 配置自动接线 trace exporter，此部分已正确。

## Goals / Non-Goals

**Goals:**
- 让本地 `start-unified.sh` 一键拉起三栈，且三项数据在各自 UI 真正可见
- 统一 metrics 端口来源，消除 prometheus.yml 与服务配置的错位
- 移除与 go-zero 原生机制重复的死代码

**Non-Goals:**
- 线上服务器（gomall-1/2）部署观测性栈——延后独立 change
- k8s 全量服务 manifest 与 k8s Prometheus/Loki DaemonSet
- 告警规则、Grafana 仪表板预设
- OpenTelemetry Collector 中间层（当前直连 Jaeger + Prometheus 已够）

## Decisions

### D1. Metrics 端点：统一用 go-zero 原生 DevServer，删死配置 PrometheusExt（search 例外保留）

**选择**：删除 12 个服务的 `PrometheusExt` 死配置——etc yaml 块范围：auths/users/product/order/checkout/payment/inventory/audit/coupons/system/activity/admin（grep 实证配置存在但 main.go 从未消费）；其中 11 个（除 users，其 config.go 无字段）连带删除 config.go 的 `PrometheusExtConf` 字段与结构体定义。这些服务统一用 go-zero 原生 `DevServer`。carts 无任何 PrometheusExt，不在清理范围（仅需补 DevServer）。**search 例外**：其 `PrometheusExt` 是活配置——`search.go:34` 启动 `StatsHTTPServer`，`servicecontext.go:33` 用 `c.PrometheusExt` 构造，该 server 除 `/metrics` 外还承载 `/stats/rag/overview|trend|breakdown|events` 四个业务 API；机制与字段名保留不动，仅将其端口对齐 +1000 规则（8081→9081）并纳入 prometheus.yml 抓取。

**理由**：
- DevServer 自带 `/metrics`（promhttp）、`/healthz`、可选 pprof，零额外代码；12 个服务的 `PrometheusExt` 是死配置，删除即消除误导。
- 已验证 DevServer 在 `Mode: pro` 下生效（`StartAgent` 仅判 `Enabled`），满足 prod 需求。
- prod 配置已用 DevServer，dev 跟随统一，避免 dev/prod 双机制。
- search 的 stats server 是活的业务入口，拆除或迁移 `/stats/rag/*` 超出本 change 范围（业务功能回归风险）；保留机制仅对齐端口是零风险最小处置。字段重命名（如 `StatsServer`）留待后续独立 change。

**备选（否决）**：
- 把 12 个服务的 `PrometheusExt` 真正接线到各 main.go（启动 prometheus http server）。否决因重复造轮子——go-zero 原生已提供等价能力，自维护一套增加认知负担与出错面。
- 将 search `/stats/rag/*` 迁至独立 rest 入口后删除其 PrometheusExt。否决因涉及 search 业务路由重构，超出观测性接线修复边界。

### D2. 删除 common/utils/tracing 死代码包

**选择**：移除整个 `common/utils/tracing/` 包。

**理由**：
- `grep` 实证无任何服务调用其导出函数；`InitJaeger` 还硬编码 localhost:14250 忽略入参，即便被调也是错的。
- 服务 trace 接线靠 go-zero 原生 `Telemetry` 配置自动接线 trace exporter。**apply 阶段实测修正**：dev yaml 的 Telemetry 并非如提案时判断的"已正确"——① 10 个服务的 `Batcher: zipkin` 指向 `http://localhost:14268/api/traces`（Jaeger 原生 collector 端点，对 zipkin JSON 返回 400，spans 全部发送失败）；② product/system/activity/search/gateway 五个 dev yaml 完全没有 Telemetry 块。本 change 已修正：dev 全部 15 个服务统一 `Endpoint: http://localhost:9411/api/v2/spans`（Jaeger all-in-one 的 zipkin 兼容入口，compose 已映射 9411），缺失的 5 个补齐 Telemetry 块。k8s/prod yaml 同样存在 14268+zipkin 的错配，随线上部署独立 change 修正（本 change 不动 prod `${TRACES_ENDPOINT}` 与 k8s manifest）。

**备选（否决）**：改造它真正接入服务。否决因与 go-zero 原生 `Telemetry` 机制完全重复，二选一必删自维护那套。

### D3. 端口规划：指标端口 = 服务端口 + 1000（110xx 规则）

**选择**：全服务（dev/prod）统一 `指标端口 = 服务端口 + 1000`，gateway（HTTP 8888）单独分配。

**端口映射表**（服务端口 / 指标端口）：

| 服务 | 服务端口 | 指标端口 |
|------|---------|---------|
| auths | 10000 | 11000 |
| users | 10001 | 11001 |
| product | 10002 | 11002 |
| carts | 10003 | 11003 |
| order | 10004 | 11004 |
| checkout | 10005 | 11005 |
| payment | 10006 | 11006 |
| inventory | 10007 | 11007 |
| audit | 10008 | 11008 |
| coupons | 10009 | 11009 |
| system | 10010 | 11010 |
| activity | 10011 | 11011 |
| admin | 10012 | 11012 |
| gateway | 8888 | 9888 |
| search | 8081* | 9081（stats server 端口，现值 9113） |

\* search 服务端口已核实 `etc/search.yaml:2` 为 8081（不在 100xx 体系，+1000 规则同样适用：8081→9081）；其指标端口由活代码 stats HTTP server 提供（D1 例外），非 DevServer。gateway 无 prod 配置文件，9888 为 dev 侧规划值。system/activity/admin 的 11010/11011/11012 亦为 dev 侧规划值（无 prod 配置，prod 随线上 change 创建）。

**抓取 host 方案**：prometheus.yml 全部宿主机服务 target 统一用 `host.docker.internal`（Prometheus 跑在 bridge 网络容器内，`localhost` 指向容器自身抓不到宿主机服务；macOS Docker Desktop 原生支持该域名）；Prometheus 自抓 target（9090）保留容器内地址。k8s/线上环境的抓取地址方案随线上部署独立 change 处理。

**理由**：prod 已用此规则，dev 跟随统一；prometheus.yml 全量重写对齐此表（端口 + host 双修正，见上），单一数据源消除三方错位。

### D4. 本地启动：start-unified.sh 补起 observability 栈

**选择**：`start-unified.sh` 在起 LOKI_STACK 后，追加 `docker compose -f construct/observability/docker-compose.yaml up -d jaeger prometheus`；stop 对应清理。

**理由**：当前用户需手敲两处 compose 才能起全栈，与"一键启动"承诺不符。补全后本地三栈可一键拉起。

**备选（否决）**：合并 observability 与 infrastructure 为单个 compose。否决因两栈职责不同（依赖栈 vs 观测栈），且 docker network 已通过 `go-mall-network`/`go-mall` 隔离管理，合并反而模糊边界。

## Risks / Trade-offs

- **[DevServer 端口占用风险]** → 110xx 端口需确认未被占用；apply 时 start-unified 增加端口预检（`port_in_use` 已有现成函数）。
- **[gateway metrics 机制不同]** → gateway 是 go-zero `rest` HTTP 服务，其 DevServer 配置路径与 zrpc 一致但需确认 `rest` 同样调用 `devserver.StartAgent`；apply 时验证，若不同则按 `rest` 文档处理。
- **[system/activity/admin/search/gateway 无 prod 配置]** → 这五个服务无 prod.yaml，本 change 不创建（避免越界补全 DB/Redis/MQ 全套配置、与"线上部署延后"非目标冲突）；其 prod 侧观测配置随线上部署独立 change 处理。dev 侧已由任务 2.2/2.3 覆盖。
- **[host.docker.internal 兼容性]** → 该域名在 macOS Docker Desktop 原生可用；Linux 宿主机需 compose 追加 `extra_hosts: ["host.docker.internal:host-gateway"]`，apply 时若本地为 Linux 则同步补上（spec Req2 验收以本地 macOS 环境为准）。
- **[DevServer 暴露 pprof 的安全权衡]** → dev 与 prod 统一 `EnablePprof: false`（与既有 auths/audit dev 块及任务 2.2 模板一致），如需调试临时开启后恢复；apply 时校验所有含 DevServer 的配置 `EnablePprof: false`。
- **[BREAKING PrometheusExt 移除]** → 12 个服务为死配置（grep 实证：字段定义于 config.go、块写于 etc yaml，但 main.go 无任何消费），删除即删旧用 DevServer，影响面可控；search 为活代码例外保留（D1），无外部脚本依赖任何服务的 `PrometheusExt` 字段。

## Migration Plan

1. **代码清理**：删 `common/utils/tracing/`；删 `PrometheusExtConf` 及相关 config 字段；确认编译通过。
2. **配置统一**：12 个服务 dev `etc/*.yaml` 补 `DevServer`（auths/audit 已有 11000/11008，核对即可）；search dev 将 stats 端口 9113 调整为 9081；gateway dev 补 DevServer（9888）。prod 配置已存在的 10 个文件不动；system/activity/admin/search/gateway 无 prod 文件，不创建。
3. **prometheus.yml 重写**：按端口映射表重写全部 target。
4. **start-unified.sh 补全**：加 observability 栈 up/down 分支 + 健康检查。
5. **验证**：本地 `./scripts/start-unified.sh` 起全栈 → curl 各服务 `/metrics` → Prometheus targets 全 UP → Grafana/Jaeger/Prometheus 三 UI 可达。
6. **回滚**：git revert 即可（纯配置 + 脚本 + 删代码，无数据迁移）。

## Open Questions

- gateway 作为 go-zero `rest` 服务，其 `DevServer` 是否与 zrpc 走同一 `StartAgent` 路径？apply 时读 `services/gateway` 代码确认，不影响 spec 与任务分解。（反思阶段源码预验证：`gateway/server.go:48` `rest.MustNewServer` → `rest/server.go:49` `c.SetUp()` → `serviceconf.go:77` `devserver.StartAgent(sc.DevServer)`，与 zrpc 同路径，配置 DevServer 确定生效；apply 时仅需实机验证。）
- search 的指标端口最终落地值（9081）是否与 `internal/stats/http.go` 现有逻辑兼容（仅配置端口变化，无代码改动预期）？apply 时启动验证 `/metrics` 与 `/stats/rag/*` 均可达。（提示：`infrastructure/promtail/promtail-config.yaml` 的 `grpc_listen_port: 9081` 数值相同，但 promtail 容器未发布该端口到宿主机，本地无冲突；未来 search 容器化接入同一网络时需留意。）
