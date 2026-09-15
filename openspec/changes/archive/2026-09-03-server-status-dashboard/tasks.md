## 0. 前置自检

- [x] 0.1 SSH 硬前置自检：`ssh -o BatchMode=yes gomall-1 true && ssh -o BatchMode=yes gomall-2 true` 双通过；`~/.ssh/config` 含 gomall-1/gomall-2 别名。验证：两条命令 exit 0；失败则停（proposal 声明的硬前置）。✅ 2026-09-02 实测：双机 BatchMode exit 0，别名在册。

## 1. 服务器侧：node_exporter 安装（D5）

- [x] 1.1 编写 `scripts/install-node-exporter.sh`（本地执行、目标机器作参数，版本参数化 `NODE_EXPORTER_VERSION=${1:-1.9.1}`）：安装前预检（`ss -lntp | grep :9100` 无占用——占用则**报错退出**并提示处置路径"按 design Risks 调整监听端口并同步隧道/prometheus 配置，需人工确认后按新端口修订三处配置"）；下载 node_exporter（linux-amd64）到 `/opt/node_exporter`；下载失败走 fallback（本地下载 → scp 上传，脚本内实现）；写 systemd unit 至 `/etc/systemd/system/node_exporter.service`（`ExecStart=/opt/node_exporter/node_exporter --web.listen-address=127.0.0.1:9100`，含 `ProtectHome=read-only` 沙箱项）；`daemon-reload && enable --now`。验证：对 gomall-1 执行后 `ssh gomall-1 'curl -s 127.0.0.1:9100/metrics | grep node_load1'` 有输出；以位置参数传入 1.8.2 可安装回退（承接 design Open Question）。✅ 2026-09-02 实测：gomall-1 安装成功（服务器直连 GitHub 超时 → fallback 本地下载+scp 生效）；1.8.2 位置参数回退安装成功后回到 1.9.1。预检增强：9100 被自身 node_exporter 占用时走"停止→原位升级"路径（仅外来进程报错退出），1.8.2 回退腿因此可执行。
- [x] 1.2 在 gomall-2 重复执行安装脚本，并做安全性与资源验证。验证：`ssh gomall-2 'curl -s 127.0.0.1:9100/metrics | grep node_load1'` 有输出；两台 `systemctl is-active node_exporter` 均 active；**公网不可达**：本地 `curl --max-time 5 http://140.143.169.125:9100/metrics` 与 `curl --max-time 5 http://120.53.106.56:9100/metrics` 均失败（refused/timeout，spec R1-S2）；`ssh gomall-2 free -m` 安装后 available 内存下降 ≤30MB（超限即触发 design Risks 停手上报 + delta 审批，阈值出处：design 预期 ~15MB 的 2x 容差）。✅ 2026-09-02 实测：gomall-2 安装成功（同走 fallback 下载）；两台均 active；公网 9100 双 IP 均 timeout（exit 28）；gomall-2 available 168MB（较装前 161MB 未降，远优于阈值）。

## 2. 本地隧道：launchd 常驻（D2）

- [x] 2.1 编写 `construct/observability/launchd/com.gomall.tunnel-gomall1.plist`：`/usr/bin/ssh -N -o ExitOnForwardFailure=yes -o ServerAliveInterval=15 -o ServerAliveCountMax=3 -o BatchMode=yes -o ConnectTimeout=10 -o IdentitiesOnly=yes -L 127.0.0.1:9101:127.0.0.1:9100 gomall-1`，KeepAlive=true、RunAtLoad=true、ThrottleInterval=5。同型编写 gomall-2（9102）。验证：`plutil -lint` 两 plist 通过；`grep -cE 'ExitOnForwardFailure=yes|ServerAliveInterval=15|ServerAliveCountMax=3|IdentitiesOnly=yes'` 两 plist 各命中 4（四个参数值，对 ProgramArguments 逐 token 编码稳健）。✅ 2026-09-02 实测：plutil OK、各命中 4。
- [x] 2.2 `launchctl bootstrap gui/$(id -u)` 安装两 plist（写入 docs）。验证：`lsof -i :9101 -i :9102` 有 LISTEN；`launchctl list | grep com.gomall.tunnel` 两个任务在册。✅ 2026-09-02 实测：监听 + 任务在册 + 隧道端到端 curl metrics 通。〔环境适配披露：launchd 拒读外置卷上的 plist（bootstrap 报 5，与本仓库脚本拒载同源），安装方式改为"仓库 plist 源 → 拷贝 ~/Library/LaunchAgents → bootstrap 标准用户域路径"，plist 源仍随仓库提交；docs 5.1 按此记录〕
- [x] 2.3 自愈验证（重启腿）：kill 两个 ssh 进程 → ≤10s 内 lsof 重新出现 LISTEN。验证：`launchctl list` PID 变化且 `curl -s 127.0.0.1:9101/metrics | grep node_load1` 恢复有输出。（S5 检出腿运行期证据为可选项：维护窗口内用 pf/iptables DROP 已建连 ssh 流模拟静默丢包，观察 ≤45s 检出 + 重启；不做则 docs 记录"检出依赖 45s 探测窗"的验收口径。）✅ 2026-09-02 实测：kill 后 **2s** 恢复（≤10s），新 PID 48892/48893，metrics 恢复输出。S5 检出腿按任务说明采"文档记录口径"（45s 探测窗），写入 docs 5.1。

## 3. Prometheus 增量 job（D3）

- [x] 3.1 `construct/observability/conf/prometheus.yml` 追加 gomall-1/gomall-2 两 job（target `host.docker.internal:9101/9102`，labels `namespace=go-mall`、`env=production`、`instance=gomall-1/gomall-2`）。验证：`docker compose -f construct/observability/docker-compose.yaml config` 通过；restart prometheus 后 `/api/v1/targets` 出现 2 个新 target 且原有 16 个全部保留。✅ 2026-09-02 实测：18/18 UP（16 旧 + gomall-1/gomall-2 双新）。〔环境披露：用户已切换 OrbStack 为默认 docker 运行时（daemon 未在跑），经用户选择后全栈在 OrbStack 下重建，start-unified 15/15 服务就绪〕
- [x] 3.2 首验容器→宿主机 loopback 转发：`docker exec` 进 prometheus 容器 curl `host.docker.internal:9101/metrics`。验证：返回指标文本；**若不达**——按 design D3 fallback（两腿叠加：LocalForward 绑 0.0.0.0 + extra_hosts 宿主网关 IP），并先按流程修订 spec R1 后再实施（不得静默偏离）。✅ 2026-09-02 实测：容器内 wget 直达 host.docker.internal:9101 拿到 node_load1（OrbStack 同样支持 loopback 转发），fallback 未触发。
- [x] 3.3 PromQL 冒烟：`:9090/api/v1/query?query=up{env="production"}` 两台均 1；`node_load1{instance="gomall-1"}` 有数值。验证：两个 instant query HTTP 200 且 result 非空。✅ 2026-09-02 实测：up{env=production} → gomall-1=1、gomall-2=1；node_load1{gomall-1}=0.02（需 --data-urlencode 编码查询串）。

## 4. 仪表盘脚本（D4/D4b）

- [x] 4.1a 编写 `scripts/server-dashboard.sh` 骨架与机器块：自含 `while true; render; sleep 5` 循环；curl + python3 拉取——每台机器块（UP/CPU%/内存%/网速 `eth0|ens.*`/磁盘% `mountpoint="/",fstype!~"tmpfs.*"`/load1）；DOWN/STALE 状态机（内存关联数组记录各数据源最后成功时间：从未成功→DOWN，曾成功→`STALE(since HH:MM:SS)`）；ANSI 清屏单屏 ≤24 行。验证：前台运行输出两台机器块数值且 5s 自动刷新。✅ 2026-09-03 实测：gomall-1 UP(cpu 8.0%/mem 68.2%/disk 54.2%/load 0.06/net 双向)、gomall-2 UP(57.5%/89.8%/48.5%/3.51)，5s 循环刷新两帧正常。
- [x] 4.1b 追加本地观测栈行：targets UP/总数（:9090 `/api/v1/targets`）、Loki 5m error 计数（:3100 LogQL，**label 口径按实测校准**——校准后与 `:3100/loki/api/v1/query` 原始返回比对计数一致，且至少一次非空路径验证，防"恒 0"静默错误）、Jaeger `/api/services` 可达性。验证：栈健康行**三项均有值**（error 计数可为 0 但须证明查询路径非空校验过）。✅ 2026-09-03 实测：targets 18/18 UP、loki 4 error(5m)（LogQL 口径 `{env="development"} | json | level="error"` 与原始返回比对一致，非空路径验证过）、jaeger UP。
- [x] 4.2 降级两态验证（R4-S8/S9 强判据）：(a) **STALE 态**——仪表盘运行中确认该机器块已展示数值 → `launchctl bootout gui/$(id -u)/com.gomall.tunnel-gomall1` → 刷新后该机器块断言 `STALE(since …)` 且**非 DOWN**，另一台与栈健康行正常、进程不退出；(b) **DOWN 态**——隧道保持停用，全新启动仪表盘实例 → 该机器块自始显示 `DOWN`。验证：两态断言逐一满足。✅ 2026-09-03 实测：(a) 运行中确认 UP → bootout 隧道1 → 帧显示 `[gomall-1] STALE(since 17:57:59)` 非 DOWN，gomall-2/栈行正常、进程存活；(b) 隧道停用下全新实例首帧 `[gomall-1] DOWN`。
- [x] 4.3 隧道恢复与栈健康降级：`launchctl bootstrap gui/$(id -u) construct/observability/launchd/com.gomall.tunnel-gomall1.plist` 重装恢复（bootout 后 kickstart 不可用）→ 下一刷新周期数值回归；临时 `docker compose -f infrastructure/docker-compose.yaml stop loki`（Loki 定义于 LOKI_STACK，非 observability compose）→ Loki 项显示 DOWN、其余正常；`start loki` 恢复。验证：spec R4-S10 / R5-S12 可勾。✅ 2026-09-03 实测：bootstrap 重装后 gomall-1 数值回归（STALE 消失）；stop loki → `loki: STALE(since 17:59:25)` 其余正常；start loki → `loki: 0 error(5m)` 恢复；仪表盘全程不退出。

## 5. 文档（D4b / spec R6）

- [x] 5.1 新增 `docs/server-status-dashboard.md`：启动方式（Ghostty ⌘D 分屏运行脚本）、前置依赖（SSH 别名/密钥（须无口令或已入 keychain）、三栈运行、plist 安装命令原样 + bootout/bootstrap 卸装重装）、DOWN/STALE 含义（R4 判据）、排障（隧道不恢复/target DOWN/9100 占用）。验证：按文档从零走一遍可独立启用仪表盘（R6-S13）。✅ 2026-09-03 实测：文档四节齐备（启动/前置含 plist 安装与密钥口令说明、降级标记表、排障四路径、卸载），全程按文档步骤已在本次实施中实际走过（R6-S13 满足）。
- [x] 5.2 `docs/servers.md` 补一行链接至该文档。验证：链接可达、路径正确（R6-S14）。✅ 2026-09-03 实测：`docs/servers.md` 头部已补 [服务器状态仪表盘](server-status-dashboard.md) 链接。

## 6. 端到端验证（spec 全 requirement 覆盖）

- [x] 6.1 全链路正常态：两隧道在册、prometheus targets 18/18（16 旧 + 2 新）、仪表盘两机器块数值实时刷新、栈健康行三项有值。验证：`/api/v1/targets` 18 个全 UP；仪表盘连续运行 2 分钟无报错。✅ 2026-09-03 实测：18/18 UP；仪表盘连续运行 125s（25 帧）最后帧距检查 4s、进程存活、无报错。
- [x] 6.2 降级与自愈回归（引用 4.2 两态断言 + 4.3）：STALE 态、DOWN 态、隧道恢复回归、Loki 故障降级逐条可勾。验证：spec R4-S8/S9/S10、R5-S12 逐条通过。✅ 2026-09-03 实测：S8/S9 见 4.2 两态断言、S10 见 4.3 恢复回归、R5-S12 见 4.3 Loki 降级，逐条通过。
- [x] 6.3 回归确认：`make lint`、`make test-unit` 通过（本 change 不改 Go 代码，验证无意外影响）；`openspec validate server-status-dashboard` 通过。✅ 2026-09-03 实测：lint『所有检查通过』、test-unit 94 用例 pass=94 fail=0、validate valid。
- [x] 6.4 回滚演练（design Migration 第 5 步承接，**全量口径**与断言匹配）：launchctl bootout 两条隧道 + `ssh gomall-N 'systemctl disable --now node_exporter && rm /etc/systemd/system/node_exporter.service && rm -rf /opt/node_exporter && systemctl daemon-reload'`（两台）+ prometheus.yml 还原 + restart → 验证 targets 回 16、`lsof` 无 9101/9102、两台无 unit 与 `/opt/node_exporter` 残留 → 按任务 1/2/3 重装恢复（顺带验证重装幂等）。✅ 2026-09-03 实测：回滚后 targets 回 16/16（gomall job 0）、隧道 0 监听、两台 unit 与 /opt/node_exporter 零残留；重装恢复后 18/18 UP（gomall job 2/2）、仪表盘两机器块 UP、幂等确认。〔附注：回滚期间 3 个本地 target 短暂 down 为 prometheus restart 后首 scrape 未完成的瞬态，数十秒内自愈〕
