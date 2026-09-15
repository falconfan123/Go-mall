## Why

本地三项观测栈已一键可用（observability-wiring-fix 已归档），但两台线上服务器 gomall-1/gomall-2（2C2G，k3s 双节点、跑着 go-mall 业务 pods）零机器指标、9100 未开放，线上资源与连通状态无从一眼掌握；本地 Ghostty 右半 pane 的 btop 只能看本机。现在补齐第一层线上观测：机器指标进本地 Prometheus，终端仪表盘取代 btop。

## What Changes

- **服务器侧**：两台安装 node_exporter（systemd 二进制方式，**绑定 127.0.0.1:9100**，云防火墙不放行，公网零暴露）
- **本地隧道**：launchd 常驻 `ssh -N -L`（KeepAlive 自动重连，零新依赖、不用 autossh）：`127.0.0.1:9101→gomall-1:9100`、`127.0.0.1:9102→gomall-2:9100`
- **Prometheus**：`construct/observability/conf/prometheus.yml` 增 2 个 job（target 9101/9102，labels `env=production`、`instance=gomall-1/gomall-2`），与现有 16 job 并存
- **仪表盘**：新增 `scripts/server-dashboard.sh`——Ghostty 右半 pane 运行，自含循环刷新（macOS 无原生 watch），输出两台机器块（UP/CPU/内存/网速/磁盘/load）+ 本地观测栈行（prometheus targets UP 数、Loki 5m error 数、Jaeger 存活），数据源全部为 HTTP API（:9090/:3100/:16686），无额外 agent
- **降级行为**：隧道断/机器宕时仪表盘显示 `DOWN/STALE`，不报错退出、不显示过期数据冒充实时
- **文档**：新增 `docs/server-status-dashboard.md`（用法/前置/排障），`docs/servers.md` 补链接
- BREAKING：无（prometheus.yml 为增量，服务器仅新增进程）

## Capabilities

### New Capabilities
- `server-status-dashboard`: 两台线上服务器的机器层指标采集（node_exporter）、SSH 隧道接入本地 Prometheus、终端仪表盘呈现与健康降级

### Modified Capabilities
<!-- 无既有 requirement 的行为变更；但需在 specs 批次显式处理与 observability 基线的字面关系：
     基线 Requirement「Prometheus 抓取目标与服务暴露端口一致」约束所有 target 不指向无人监听端口，
     而本能力定义"隧道断开时远程 target 合法 DOWN（降级语义）"——specs 批次须在新能力 spec 中声明
     该豁免与优先级（远程机器 target 不适用"+1000 规则"与"target 恒 UP"语义），避免与主 spec 字面冲突。 -->

## Impact

- **服务器（gomall-1/2）**：各新增 `/opt/node_exporter` 二进制 + systemd unit（常驻 ~15MB；gomall-2 可用内存仅 161MB，为已知紧张项，node_exporter 可承受）
- **本地**：prometheus.yml +2 job；新增 launchd plist（隧道常驻）；新增 `scripts/server-dashboard.sh`；`brew install autossh` 不再需要（launchd KeepAlive + 原生 ssh）
- **前置依赖**：`~/.ssh/config` 别名 gomall-1/gomall-2 + 密钥 id_ed25519（2026-09-02 已就绪）
- **已知开放问题**：node_exporter 在腾讯云北京区 VM 上的下载源可达性存疑（GitHub Releases 直连大概率不通）——apply 前置验证，fallback：本地下载后 scp 上传
- **非目标**：线上完整观测栈（线上 Prometheus/Jaeger/Loki 服务端）、告警规则、k8s 内业务服务指标采集、Grafana 仪表板预设、tmux 深度美化——均延后独立 change
