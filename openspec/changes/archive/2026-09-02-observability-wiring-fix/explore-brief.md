# Explore Brief — 观测性接线修复

> 探索日期：2026-08-30
> 触发问题：用户问"目前能够查看服务器的各项数据吗 tracing log metrics 三项有没有部署上去"
> 结论：三项都有配置文件，但实际接线程度差异巨大，线上基本没真正部署上去。

---

## 背景 — 现状调查结论（实证）

### 三项观测性栈的定义位置
- **Logs**：`infrastructure/docker-compose.yaml`（Loki 2.9.2 + Promtail 2.9.2 + Grafana 10.2.2）
- **Traces**：`construct/observability/docker-compose.yaml`（Jaeger all-in-one 1.61.0）+ `k8s/infrastructure/04-jaeger.yaml`（all-in-one 1.48，badger ephemeral）
- **Metrics**：`construct/observability/docker-compose.yaml`（bitnami/prometheus + grafana，匿名 Admin）+ `construct/observability/conf/prometheus.yml`（13 个 service job + gateway）

### 本地启动情况（实证：`scripts/start-unified.sh`）
- `start-unified.sh:12` 定义 `LOKI_STACK="${PROJECT_ROOT}/infrastructure/docker-compose.yaml"`
- `start-unified.sh:469` 执行 `docker compose -f "$LOKI_STACK" up -d loki promtail grafana`
- **start-unified.sh 全文 grep 不到 `observability`/`jaeger`/`prometheus`** → 只起日志栈，不起 traces/metrics 栈，需手动 `cd construct/observability && docker-compose up -d`

### 日志采集链路（本地唯一通的部分）
- 所有服务 `etc/*.yaml` 的 `Log.Mode: console`（21 处实证，含 dev/prod）
- `scripts/service-supervisor.sh:36` 用 `exec bash -lc "$COMMAND" ) >> "$LOG_FILE" 2>&1` 把 stdout 重定向到 `scripts/logs/<svc>.log`
- `infrastructure/promtail/promtail-config.yaml` 抓 `/var/log/go-mall/*.log`，docker volume 映射 `../scripts/logs:/var/log/go-mall:ro`
- **本地链路通**；但 k8s 容器化后 console→容器 stdout，无 Promtail DaemonSet

### D1 — Metrics 完全断链（最严重）
- `services/order/etc/order.yaml:73-77` 自定义 `PrometheusExt: Host 0.0.0.0, Port 9104, Path /metrics`
- `services/order/internal/config/config.go:22` 有 `PrometheusExt PrometheusExtConf`，`:30` 有 `PrometheusExtConf` 结构体定义
- **但 `services/order/order.go` main 完全没读 PrometheusExt、没启动 prometheus http server**
- 实证：`grep -rn "PrometheusExt" services/order/` 仅 config.go 一处，main.go 无消费
- 结果：配置写了 9104 端口但无人 listen → Prometheus 抓 `localhost:9104` = connection refused
- prod `order.prod.yaml:22-26` 改用 go-zero 原生 `DevServer: Port 11004`，但 `Mode: pro`（非 dev mode）下 DevServer 行为存疑，且端口与 prometheus.yml 的 9104 对不上

### D2 — Tracing 工具是死代码
- `common/utils/tracing/tracing.go` 定义 `InitJaeger`（:32）和 `InitWithOtelCollector`（:80）
- 实证：`grep -rn "InitJaeger|InitWithOtelCollector" services/ app/` **无任何输出** → 从未被任何 main.go 调用
- 且 `InitJaeger` 内部 `:43-45` 硬编码 `jaeger.WithAgentHost("localhost"), jaeger.WithAgentPort("14250")`，忽略传入的 `cfg.JaegerEndpoint`
- 服务实际靠 **go-zero 原生 `Telemetry` 配置**自动接线（这部分对）：
  - dev `order.yaml:12-16`：`Batcher: zipkin, Endpoint: http://localhost:14268/api/traces, Sampler: 1.0`
  - k8s `k8s/services/order.yaml:23-27`：`Endpoint: http://jaeger-collector.go-mall.svc.cluster.local:14268/api/traces`
  - prod `order.prod.yaml:29-33`：`Endpoint: ${TRACES_ENDPOINT}, Batcher: zipkin, Sampler: 0.1`

### D3 — 端口规划混乱，dev/prod/采集三方对不上
| 服务 | prometheus.yml target | dev PrometheusExt.Port | prod DevServer.Port |
|------|----------------------|----------------------|---------------------|
| order | 9104 | 9104（死代码）| 11004 |
| users | 11000 | ? | ? |
| auths | 9101 | ? | ? |
| product | 9102 | ? | ? |
| carts | 9103 | ? | ? |
| checkout | 9105 | ? | ? |
| payment | 9106 | ? | ? |
| inventory | 9107 | ? | ? |
| audit | 9108 | ? | ? |
| coupons | 9109 | ? | ? |
| system | 9110 | ? | ? |
| activity | 9111 | ? | ? |
| admin | 9112 | ? | ? |
| search | 9113 | ? | ? |
| gateway | 8888 | — | — |

- dev 用 9xxx 系列自定义 PrometheusExt（未接线），prod 用 11xxx 系列 go-zero DevServer
- prometheus.yml 统一抓 9xxx → **线上部署 prometheus 全抓空**

### D4 — 线上观测性栈零部署
- `scripts/start-unified.sh` 只起 LOKI_STACK，不起 construct/observability
- `k8s/infrastructure/` 有 `04-jaeger.yaml` ✅ 但 **无 prometheus、无 loki/promtail**
- `k8s/services/` 只有 `order.yaml` 一个，其他服务 k8s manifest 缺失；order configmap 配了 telemetry，但其他服务无 k8s manifest 也就无 telemetry 配置
- `docs/servers.md` 只列 gomall-1（140.143.169.125）/gomall-2（120.53.106.56）的 SSH，**完全没提观测性栈部署在哪台**
- `order.prod.yaml:31` 用 `${TRACES_ENDPOINT}`，但 scripts/k8s 里 **找不到该环境变量定义**

### D5 — 文档与实际路径不一致
- `docs/jaeger/README.md:53` 和 `docs/jaeger/quickstart.md:13` 写 `cd construct/observability && docker-compose up -d jaeger`
- `docs/logging-system.md:50` 写 `cd infrastructure && docker-compose up -d`
- 实际栈确实分两处，但文档未交叉说明本地一键起全套需两个 compose

---

## 目标 / 非目标

### 目标（用户视角收益）
- 用户能在 Grafana/Jaeger/Prometheus UI 上**真正看到**三项数据，而非配置空转
- 本地与线上（至少一处）观测性可用，可据此排查 payment-500 等线上问题
- 消除死代码与端口规划混乱，降低后续维护认知负担

### 非目标（明确排除）
- 不引入 OpenTelemetry Collector 中间层（当前直连 Jaeger + Prometheus 已够）
- 不重写日志架构（Loki 链路本地已通，仅补线上缺口）
- 不做告警规则/仪表板预设（先通采集，可视化后续再做）
- 不在本 change 内补齐所有服务 k8s manifest（k8s 全量部署是独立大项）

---

## 待决策点（探索中浮现，倾向待定）

- **D1 修法**：
  - 选项 A：删 `PrometheusExt` 自定义方案，统一用 go-zero 原生 `DevServer`，prometheus.yml targets 改 11xxx 系列
  - 选项 B：把 `PrometheusExt` 真正接线到各服务 main.go（启动 prometheus http server）
  - 倾向：A（go-zero 原生更省代码，DevServer 自带 /metrics + pprof），但需确认 prod `Mode: pro` 下 DevServer 生效条件
- **D2 处置**：
  - 选项 A：删 `common/utils/tracing` 死代码包
  - 选项 B：改造它真正接入服务（与 go-zero 原生 Telemetry 二选一，重复）
  - 倾向：A（删除，避免与 go-zero 原生机制重复且误导后人）
- **D4 范围**（本 change 边界）：
  - 选项 A：仅修本地 + dev 配置接线，线上部署延后单独 change
  - 选项 B：本 change 同时补 k8s prometheus/loki manifest
  - 倾向：A（避免 change 膨胀，线上 k8s 全量观测性独立立项更清晰），但需在 proposal 明确"线上暂不部署"为非目标
- **TRACES_ENDPOINT 等环境变量**：本 change 是否补 .env / 部署脚本定义？倾向：若选 D4-A 则不在本 change 处理线上 env

---

## Checklist（后续提案审核逐项核对基线）

- [ ] proposal 是否明确列出三项现状实证（栈位置、启动情况、接线断裂）
- [ ] proposal 是否将 D1-D5 五个断裂点全部纳入范围或显式排除
- [ ] design 是否给出 D1（Metrics 接线）的具体方案，且与 go-zero 原生机制一致
- [ ] design 是否给出 D2（tracing 死代码）的处置决定（删 or 改造）
- [ ] design 是否给出 D3 端口统一规划表（单一端口来源，prometheus.yml 与服务配置一致）
- [ ] specs 是否覆盖三项能力的可观测需求（而非仅栈部署）
- [ ] tasks 是否按"先通本地三项采集 → 再补线上"的顺序，且每步可独立验证
- [ ] 是否明确"线上 k8s 全量观测性"为本 change 非目标（避免范围蔓延）
- [ ] 是否处理 `common/utils/tracing` 死代码包（删除 or 改造）
- [ ] 是否处理文档路径不一致（docs/jaeger vs docs/logging-system vs 实际两处 compose）
