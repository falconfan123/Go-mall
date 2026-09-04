## Context

explore-brief（2026-09-02 实证 + propose 阶段修正）已确立关键事实：

- 本地观测栈就绪：Prometheus :9090（16 job 全本地）、Loki :3100、Jaeger :16686，API 可编程消费；Ghostty 右半 pane 现为 btop（仅本机）
- 两台均为同一 k3s 集群节点：gomall-1=control-plane（VM-4-3-ubuntu，available 内存 584MB）、gomall-2=worker（VM-0-5-ubuntu，**available 仅 161MB**）；k3s v1.32.1+k3s1，containerd 运行时；26 pods，go-mall 业务已运行；唯一 DaemonSet 为 svclb-traefik（svclb 会按 LB 服务占宿主机端口——安装 node_exporter 前需核查 9100 占用）
- 两台 9100 公网未开放（云防火墙拦截）；SSH 别名 gomall-1/gomall-2 + id_ed25519 已打通
- 本地 `~/.kube/config` 无 gomall 集群凭证（仅指向 127.0.0.1）；6443 公网可达但需凭证
- macOS 无原生 `watch` 命令；本地未装 autossh；Prometheus 容器抓取宿主机沿用 `host.docker.internal` 模式

## Goals / Non-Goals

**Goals:**
- 两台线上服务器的机器指标（CPU/mem/net/disk/load）进入本地 Prometheus
- Ghostty 右半 pane 仪表盘取代 btop：两台机器块 + 本地观测栈健康行
- 隧道常驻自愈；采集中断时仪表盘按 DOWN/STALE 判据明确降级

**Non-Goals:**
- 线上完整观测栈（线上 Prometheus/Jaeger/Loki 服务端）、告警规则
- k8s 内业务服务指标采集（业务 pods 的 /metrics 接入是独立 change）
- Grafana 仪表板预设、tmux 深度美化

## Decisions

### D1. 采集路径终选：A（systemd node_exporter 两台）——否决 A''（k8s DaemonSet+hostNetwork）

| 维度 | A：systemd 二进制 | A''：k8s DaemonSet + hostNetwork |
|---|---|---|
| 覆盖面 | 两台对称（各自独立装） | 单 manifest 覆盖两节点（同集群）✅ |
| 前置 | 仅 SSH（已就绪） | 需分发集群 admin kubeconfig 到本地（6443 公网 + 凭证，安全面扩大）；本地 kubeconfig 现无该集群 |
| 资源 | node_exporter ~15MB/台 | 同等常驻内存，但增加 k3s 调度/镜像拉取；**gomall-2 available 仅 161MB**，kubelet 重调度时更脆弱 |
| 故障域 | 独立于 k3s（k3s 挂了监控仍在）✅ **此维度 A 优** | 与 k3s 生命周期绑定（k3s 异常时观测同时失效）——**此维度 A'' 为劣势** |
| 运维 | 手工升级/回滚每台 | kubectl 统一升级 ✅ |

**结论：选 A**。决定性理由：①观测链路不应与被观测对象的运行时同生共死（k3s 故障时恰是仪表盘最有用的时候）；②避免向本地分发公网可达的集群 admin 凭证；③gomall-2 内存余量不宜再承担集群组件故障重调度。A'' 的"统一管理"优势在本场景（两台固定机器、一个静态 exporter）收益有限。将来若 k8s 化监控统一立项可迁移至 DaemonSet（hostNetwork + NodeExporter）。

**k3s 内现状验证结论**（propose 阶段实测）：两台同为 v1.32.1+k3s1 双节点集群、业务 26 pods 在跑、无监控组件、svclb-traefik 存在——A 的安装前置（9100 端口占用预检、systemd 可用）已确认可行（systemd is-system-running=running）。

### D2. 隧道：launchd KeepAlive + 原生 ssh（否决 autossh）

- plist 两份（gomall-1/gomall-2 各一）随仓库提交于 `construct/observability/launchd/`，安装方式 `launchctl bootstrap gui/$(id -u)` + 文档化
- Program：`/usr/bin/ssh -N -L 127.0.0.1:9101:127.0.0.1:9100 gomall-1`（gomall-2 用 9102），配置项：
  - `-o ExitOnForwardFailure=yes`（转发失败即退出，触发 KeepAlive 重启）
  - `-o ServerAliveInterval=15 -o ServerAliveCountMax=3`（45s 检出静默断链）
  - `-o BatchMode=yes -o ConnectTimeout=10 -o IdentitiesOnly=yes`（launchd 环境无 SSH_AUTH_SOCK，密钥须无口令或已入库 keychain）
- KeepAlive=true + RunAtLoad=true + ThrottleInterval=5：进程退出（含被检出断链）即重启，开机自启
- **否决 autossh**：多一个 brew 依赖，launchd KeepAlive 已提供等价自愈语义；autossh 的监控端口机制反而在严防环境下多开一个口
- 服务器端 node_exporter 绑定 `127.0.0.1:9100`：SSH 转发目标走服务器侧 loopback，公网零暴露（云防火墙保持不放行 9100）
- launchd 隧道独立于 `start-unified.sh` 运行（服务器监控不依赖本地商城栈启停）

### D3. Prometheus 增量 job

```yaml
- job_name: 'gomall-1'
  static_configs:
    - targets: ['host.docker.internal:9101']
      labels: { namespace: go-mall, env: production, instance: gomall-1 }
- job_name: 'gomall-2'
  static_configs:
    - targets: ['host.docker.internal:9102']
      labels: { namespace: go-mall, env: production, instance: gomall-2 }
```

- target host 沿用既有 `host.docker.internal` 模式（Prometheus 在 bridge 容器内，9101/9102 是宿主机 loopback 上的隧道口）；**apply 首验**：容器内 curl 可达性——Docker Desktop 的 host.docker.internal 转发至宿主机 loopback 为预期行为。**fallback（单方案，两腿叠加生效）**：ssh LocalForward 绑定地址由 127.0.0.1 改为 `0.0.0.0`（限受信 LAN，文档记录安全注意项）**且** prometheus 容器加 `extra_hosts` 将 host.docker.internal 指向宿主机网关 IP——单独任一腿均不成立（仅绑 loopback 时网关 IP 不可达；仅改 hosts 指向网关 IP 命不中 loopback 监听）。**注意**：该 fallback 与冻结 spec R1 的"127.0.0.1:9101/9102"端点字面冲突，启用前须先按 OpenSpec 流程修订 R1（delta 审批），不得静默偏离
- 变更生效：`docker compose -f construct/observability/docker-compose.yaml restart prometheus`
- 与既有 16 job 并存（9090/9888/9081/11000-11012 与 9101/9102 零冲突）

### D4. 仪表盘脚本 `scripts/server-dashboard.sh`

- 自含循环：`while true; do render; sleep 5; done`（macOS 无 watch；刷新间隔 5s，与 prometheus scrape_interval 15s 错拍）
- 单进程单文件，依赖仅 `curl` + `python3`（JSON 解析/数值计算，本机已有）+ ANSI 清屏
- 数据源与查询形状（PromQL / LogQL）：
  - UP：`up{env="production"}`（:9090 `/api/v1/query`）
  - CPU%：`100 * (1 - avg by (instance)(rate(node_cpu_seconds_total{mode="idle"}[2m])))`
  - 内存%：`100 * (1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)`
  - 网速：`sum by (instance)(rate(node_network_receive_bytes_total{device=~"eth0|ens.*"}[2m]))` 及 transmit 同型（收紧到物理网卡正则，避免 k3s 节点 cni0/flannel/veth 双计 pod 流量）
  - 磁盘%：`(1 - node_filesystem_avail_bytes{mountpoint="/",fstype!~"tmpfs.*"} / node_filesystem_size_bytes{mountpoint="/",fstype!~"tmpfs.*"}) * 100`
  - load：`node_load1`
  - targets UP 数：`:9090/api/v1/targets` 计数 `health=="up"`
  - Loki error(5m)：`:3100/loki/api/v1/query`，LogQL `sum(count_over_time({namespace="go-mall"} | json | level="error" [5m]))`（label 匹配以实际 Loki 标签为准，apply 时校准）
  - Jaeger：`:16686/api/services` HTTP 200 即 UP
- **DOWN/STALE 判据实现**（对应 spec R4）：每个数据源维护"最后成功时间"（内存关联数组，进程生命周期内）：查询失败/无数据时，无成功记录 → `DOWN`；有记录 → `STALE(since HH:MM:SS)`；任何记录只在真实成功响应后写入，杜绝过期数据冒充实时
- 输出布局（ANSI，两机器块 + 栈健康行），单屏 ≤24 行，适配半窗宽度

### D4b. 文档交付锚点（承接 spec R6）

`docs/server-status-dashboard.md` 为 R6 指定文档，内容清单（即 R6 两个 scenario 的验收范围）：
1. 启动方式：Ghostty ⌘D 分屏 → 右半运行 `./scripts/server-dashboard.sh`
2. 前置依赖：`~/.ssh/config` 别名 gomall-1/gomall-2 与密钥（**密钥须无口令或已入库 keychain**——launchd 无 SSH_AUTH_SOCK，plist 加 `IdentitiesOnly=yes`）；本地三栈运行状态；launchd plist 安装步骤（`launchctl bootstrap` 命令原样给出）
3. DOWN/STALE 标记含义（spec R4 判据）
4. 排障路径：隧道不恢复（launchctl kickstart/bootout 重装）、target DOWN（容器内 curl 首验命令）、9100 占用
`docs/servers.md` 末尾补一行指向该文档。

### D5. node_exporter 服务器侧安装

- 版本锁定 `1.9.1`（propose 时最新稳定；若下载不可达回退 1.8.2 或走 fallback 链）
- 安装方式：`scripts/install-node-exporter.sh`（本地执行，`ssh gomall-N` 内完成）：预检（`ss -lntp | grep :9100` 无占用——防 svclb 冲突）→ 下载（服务器直连 GitHub，失败 fallback：本地下载 scp 上传）→ 解压到 `/opt/node_exporter` → 写 systemd unit（`ExecStart=/opt/node_exporter/node_exporter --web.listen-address=127.0.0.1:9100`）→ `daemon-reload && enable --now` → `curl 127.0.0.1:9100/metrics` 自验
- 非 root 运行不强制（node_exporter 官方可 root 运行取全量指标；2C2G 简化处置，unit 内做基本加固 `ProtectHome=read-only` 等沙箱项酌情）

## Risks / Trade-offs

- **[host.docker.internal→loopback 假设]** → 容器抓宿主机 loopback 上的隧道口依赖 Docker Desktop 转发行为；apply 首验，不达则按 D3 fallback（绑定 0.0.0.0 或 extra_hosts 宿主网关 IP）
- **[node_exporter 下载可达性]** → 腾讯云北京 VM 直连 GitHub 可能不通；install 脚本内置 fallback（本地下载 scp 上传），任务含前置验证
- **[gomall-2 内存 161MB]** → node_exporter ~15MB 可承受但需盯；若 apply 时实测确认 OOM 风险，**不做单台交付**——上报用户并按 OpenSpec 流程修订 spec R1（delta 审批）后再定范围（R4 的运行期降级不构成交付期单台豁免）
- **[9100 端口占用（svclb）]** → 安装前 `ss -lntp` 预检，占用则调整 node_exporter 监听端口并同步隧道/prometheus 配置
- **[SSH 密钥依赖]** → `~/.ssh/config` 别名与 id_ed25519 为硬前置；文档记录；密钥变更需同步 launchd plist（plist 用别名不内嵌 IP，密钥由 ssh config 管理即可）
- **[隧道端口冲突（本地）]** → 9101/9102 与本地既有端口（9090/3000/3001/3100/8888/9081/110xx/16686…）零重叠，launchd 启动即验证

## Migration Plan

1. 服务器侧：install-node-exporter.sh 装两台 → 自验 `/metrics`
2. 本地隧道：plist 入库 + launchctl 安装 → `lsof -i :9101/:9102` 验证
3. prometheus.yml +2 job → restart prometheus → targets 18 个（16+2）
4. 仪表盘脚本 → 手动运行验证 → 文档（`docs/server-status-dashboard.md` 按 D4b 内容清单 + `docs/servers.md` 链接）
5. 回滚：launchctl bootout（移除 plist）+ `ssh gomall-N systemctl disable --now node_exporter && rm /etc/systemd/system/node_exporter.service && rm -rf /opt/node_exporter && systemctl daemon-reload` + prometheus.yml 还原 + 删除 docs（纯增量回滚，无数据迁移、无残留）

## Open Questions

- Loki error 计数的 label 匹配式（`namespace="go-mall"` vs `env="development"`）以 apply 时实测标签为准（任务含校准步骤）
- node_exporter 1.9.1 若与 Ubuntu 24.04 内核 6.8 有采集兼容问题（预期无），回退 1.8.2 LTS 分支
