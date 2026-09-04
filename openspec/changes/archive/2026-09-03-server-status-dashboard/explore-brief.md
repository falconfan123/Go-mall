# Explore Brief — server-status-dashboard（两台服务器状态进本地 Ghostty 右半侧）

> 探索日期：2026-09-01～02
> 触发问题：用户问"能否将两台服务器状态同步到本地 Ghostty 右半侧，取代现有本地 CPU/网络等显示，并通过 Prometheus 及其他观测设备做可观测性"
> 结论：可行——本地 Prometheus 为单一数据源，远程机器指标经 SSH 隧道 + node_exporter 进 :9090，右半 pane 自写脚本消费 API。

---

## 背景 — 现状调查结论（实证）

### 本地侧（2026-09-01/02 实测）
- Ghostty 纯外观配置（catppuccin-mocha + JetBrainsMono），**无状态栏机制**；tmux 已装，`status-right` 仅时间日期（`~/.tmux.conf:11`）
- **btop 已装**（`/opt/homebrew/bin/btop`）——现有右半 pane 的本地 CPU/网络显示即此类本地 TUI；btop 不支持远程主机，"取代"必然换工具
- 本地 Prometheus :9090 已通（`observability-wiring-fix` 已完成并归档于 2026-09-02），16/16 targets 全为本地服务（prometheus 自抓 + `host.docker.internal` 110xx/gateway 9888/search 9081），**无任何远程 target**
- 本地观测栈 API 齐备且可编程消费：Prometheus :9090（指标即时查询）、Loki :3100（日志检索）、Jaeger :16686（链路/服务列表）——仪表盘数据源三件套

### 线上侧（2026-09-02 实测）
- **SSH 已打通**：Mac `~/.ssh/config` 别名 `gomall-1`(140.143.169.125) / `gomall-2`(120.53.106.56)，root 密钥 `~/.ssh/id_ed25519`；`docs/servers.md` 已同步
- 两台 Ubuntu **2C2G**（主机名 `VM-4-3-ubuntu` / `VM-0-5-ubuntu`）
- **两台均为同一 k3s 集群节点**（propose 阶段实测修正）：gomall-1=control-plane/master（VM-4-3-ubuntu），gomall-2=worker（VM-0-5-ubuntu，k3s-agent active）；k3s v1.32.1+k3s1，containerd 运行时（无独立 docker）；集群 26 pods，**go-mall 业务已在 k3s 内运行 12～42 天**（auths/carts/checkout/dtm/es/etcd 等）；唯一 DaemonSet 为 svclb-traefik（无任何监控组件）
- 资源：gomall-1 available 584MB；**gomall-2 available 仅 161MB（紧张）**
- 两台 **9100（node_exporter）均未开放**（云防火墙未放行）
- 本地 `~/.kube/config` 仅指向 127.0.0.1:6443（本地集群），**无 gomall k3s 凭证**；6443 公网可达但需凭证
- 待验证开放问题（残余）：node_exporter 下载源在服务器上的可达性（apply 时验证）
- 上一 change（observability-wiring-fix）明确将"线上服务器（gomall-1/2）观测性栈部署"列为非目标延后——本 change 即其自然延伸，但边界只取"机器层指标采集 + 本地终端消费"，不做线上完整观测栈

---

## 目标 / 非目标

### 目标（用户视角收益）
- Ghostty 右半 pane 显示两台线上服务器状态（CPU/内存/网络/磁盘/UP），**取代本地 btop**
- 数据统一走本地 Prometheus :9090（node_exporter 机器指标为主）
- 本地观测栈自身健康（targets UP、Loki error、Jaeger 速率）可作为仪表盘附加行
- 隧道/机器不可达时仪表盘**明确降级显示**（DOWN/STALE，不误报不花屏）

### 非目标（明确排除）
- 服务器上部署完整观测栈（线上 Prometheus/Jaeger/Loki 服务端）——线上只装轻量采集 agent
- 告警规则、Grafana 仪表板预设
- k8s 内业务服务指标采集（propose 实测修正：业务服务已在 k3s 内运行——边界决策不变，机器层之外的 k8s 服务指标采集是独立 change）
- Ghostty/tmux 深度定制美化

---

## 待决策点（倾向已定，propose 时定稿）

- **D1 采集路径**：走 **A（SSH 隧道 + node_exporter）**——服务器装 node_exporter（公网零暴露），本地 autossh 常驻 `127.0.0.1:9101→gomall-1:9100`、`127.0.0.1:9102→gomall-2:9100`，prometheus.yml 增两 job 抓隧道口。**propose 时与 A''（node_exporter 以 k8s DaemonSet + hostNetwork 部署进 gomall-1 的 k3s；gomall-2 纯 agent 另装）对比后定稿**——A'' 依赖"k3s 内现状"验证结果，且 gomall-2 非 k3s 节点需单独处理，复杂度不对称
- **D2 展示层**：右 pane 自写脚本（`scripts/server-dashboard.sh`，`watch -n5` + curl :9090/:3100/:16686 API），纯 HTTP API 消费、无新依赖；tmux `status-right` 一行小件为可选附带（倾向不做，聚焦右 pane）
- **D3 仪表盘内容分层**：机器层（node_exporter：CPU/mem/net/disk/load/UP）必做；服务健康（targets up）、日志层（Loki error rate）、链路层（Jaeger span 速率）为加分项，propose 定范围
- **D4 隧道常驻性**：autossh + launchd 开机自启 vs 挂 `start-unified.sh` 按需拉起——待定（影响断线自愈语义）
- **D5 SSH 连通性**：已解决（见背景），但 change 内需声明 `~/.ssh/config` 别名 + 密钥为前置依赖

---

## Checklist（后续提案审核逐项核对基线）

- [ ] proposal 是否明确端到端数据链路（node_exporter → SSH 隧道 → 本地 prometheus → 右 pane 脚本）与端口分配（9101/9102 隧道口，不与现有 110xx/9888/9081 冲突）
- [ ] proposal 是否将线上完整观测栈 / 告警 / k8s 业务指标采集列为非目标（防范围蔓延）
- [ ] design 是否给出 D1 终选（A vs A''）及理由，且包含"k3s 内现状"的验证结论
- [ ] design 是否覆盖隧道常驻方案（autossh/launchd 或 start-unified 集成）与断线降级行为（仪表盘呈现 DOWN/STALE）
- [ ] prometheus.yml 增量是否与现有 16 job 兼容（新 job 命名清晰、隧道端口不冲突）
- [ ] 服务器安装方式是否随 D1 结论明确（systemd 二进制 vs k8s DaemonSet），且 2C2G 资源占用可接受
- [ ] 是否声明 SSH 依赖（`~/.ssh/config` 别名 gomall-1/gomall-2 + id_ed25519）为前置条件
- [ ] tasks 是否包含端到端验证（两台 targets UP → 右 pane 显示 CPU/网速 → 断隧道降级显示 → 恢复自愈）
- [ ] 仪表盘"其他观测"行（如有）数据源是否全为 HTTP API（9090/3100/16686），无额外 agent
- [ ] 每个任务可独立验证且 ≤2h
