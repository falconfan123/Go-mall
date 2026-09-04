## Purpose

让用户在本地终端（Ghostty 右半 pane）一目了然掌握两台线上服务器（gomall-1/gomall-2）的机器状态与本地观测栈健康：线上机器指标经加密 SSH 隧道进入本地 Prometheus，终端仪表盘取代只能看本机的 btop。

> **与 observability 基线的关系（豁免声明）**：基线能力 `observability` 中「Prometheus 抓取目标与服务暴露端口一致」的"各服务"指 go-mall 服务（服务端口 +1000 规则）。本能力的远程机器 target（隧道口 9101/9102）不属于该范畴；且本能力明确定义"隧道断开时远程 target 合法 DOWN（降级语义）"，该行为优先于基线「全部 target 健康抓取」场景对任意时刻全 UP 的字面预期——隧道断开期间 9101/9102 本地无人监听属合法状态。

## ADDED Requirements

### Requirement: 线上机器指标经隧道进入本地 Prometheus

两台线上服务器 SHALL 在服务器本机回环地址（127.0.0.1:9100）运行 node_exporter 暴露机器指标，云防火墙 SHALL NOT 放行 9100 公网访问。本地 SHALL 维持到两台服务器的 SSH 本地转发隧道（127.0.0.1:9101→gomall-1:9100、127.0.0.1:9102→gomall-2:9100）。本地 Prometheus SHALL 抓取这两个 target 并打上 `env=production` 与 `instance=gomall-1/gomall-2` 标签，与既有 16 个 job 并存。

#### Scenario: 隧道建立后远程 target UP
- **WHEN** 本地 Prometheus 运行且两条隧道已建立
- **THEN** prometheus.yml 中 gomall-1/gomall-2 两个 job 的 target 状态为 UP，可查询到 `node_load1`、`node_memory_MemAvailable_bytes` 等机器指标

#### Scenario: 公网不可直达 node_exporter
- **WHEN** 从公网尝试访问任一台服务器的 9100 端口
- **THEN** 连接被拒绝或超时（云防火墙拦截），node_exporter 仅可经 SSH 隧道访问

#### Scenario: 新旧 job 并存
- **WHEN** 查看本地 Prometheus targets 全集
- **THEN** 既有全部本地 job（当前 16 个）全部保留且状态不变，新增的 2 个远程机器 target 并列存在

### Requirement: 隧道常驻与断线自愈

SSH 隧道 SHALL 由本地常驻进程管理机制守护：转发进程退出后 SHALL 在可判定时限内（≤10 秒）自动重启恢复转发；SSH 会话 SHALL 配置活性探测与转发失败快速退出机制，使静默断链可被检出并触发退出-重启自愈。隧道端点分配（9101=gomall-1、9102=gomall-2）SHALL 保持稳定，不与其他本地监听端口冲突。

#### Scenario: 隧道进程被杀后自动恢复
- **WHEN** 手动终止本地隧道转发进程
- **THEN** 守护机制在数秒内重启转发，Prometheus 的 gomall target 从 DOWN 恢复为 UP

#### Scenario: 静默断链可自愈
- **WHEN** SSH 连接被网络静默丢弃（对端无 RST）
- **THEN** 活性探测在配置的超时窗口内检出失活并退出转发进程，守护机制重启后隧道恢复

### Requirement: 终端仪表盘呈现两台服务器状态

终端仪表盘 SHALL 按服务器分块展示每台的 UP 状态、CPU 使用率、内存使用率、网络收发速率、磁盘使用率与系统负载，指标数据 SHALL 来自本地 Prometheus（:9090）的即时查询 API。仪表盘 SHALL 自含循环刷新逻辑（不依赖外部 watch 等系统命令），可在 Ghostty 分屏 pane 中直接运行。

#### Scenario: 展示两台机器实时状态
- **WHEN** 仪表盘运行且隧道与 Prometheus 正常
- **THEN** 输出中 gomall-1 与 gomall-2 各有一个状态块，含非空且随负载变化的 CPU/内存/网络速率/磁盘/负载数值

#### Scenario: 循环刷新
- **WHEN** 仪表盘持续运行
- **THEN** 展示内容按固定间隔自动刷新，无需人工重启进程

### Requirement: 采集中断的降级呈现

隧道断开、机器不可达或指标查询失败时，仪表盘 SHALL 按以下判据对受影响项分级显示：查询失败或无数据，且该数据源**从未成功返回过** → 显示 `DOWN`；此前曾成功返回但当前失败或停止更新 → 显示 `STALE` 并附最后成功时间。任一数据源降级时，仪表盘 SHALL NOT 报错退出、崩溃或用过期数据冒充实时数据。恢复后 SHALL 自动回到正常展示。

#### Scenario: 隧道断开时明确降级
- **WHEN** 任一隧道断开或对应机器不可达
- **THEN** 该机器块显示 DOWN（或 STALE + 最后成功时间），另一台机器块与本地观测栈行继续正常展示，仪表盘进程不退出

#### Scenario: STALE 判据强制触发
- **WHEN** 某数据源此前已成功展示过数值，随后其查询失败或指标停止更新
- **THEN** 该项显示 `STALE` 并附最后成功时间（而非 DOWN，更非过期数值冒充实时）

#### Scenario: 恢复后自动回到正常展示
- **WHEN** 隧道恢复、target 重新 UP
- **THEN** 该机器块在下一刷新周期恢复数值展示（DOWN/STALE 标记消失），无人工干预

### Requirement: 本地观测栈健康行

仪表盘 SHALL 附带本地观测栈状态行：Prometheus targets 的 UP/总数、Loki 近 5 分钟 error 级日志计数、Jaeger 可达性，分别来自 :9090、:3100、:16686 的 HTTP API，不引入任何额外采集 agent。任一组件不可达时对应项 SHALL 显示 DOWN 而不影响仪表盘其余部分。

#### Scenario: 全栈健康时展示汇总行
- **WHEN** 本地三栈与全部 target 正常
- **THEN** 状态行显示 targets UP/总数、error 计数（可为 0）与 jaeger UP

#### Scenario: 单组件故障不拖垮仪表盘
- **WHEN** 任一观测栈组件（如 Loki）不可达
- **THEN** 对应项显示 DOWN，其余行继续正常展示，仪表盘进程不退出

### Requirement: 仪表盘用法与排障文档

项目文档 SHALL 提供仪表盘使用说明，覆盖：启动方式（Ghostty 分屏场景）、前置依赖（SSH 别名/密钥、本地观测栈运行状态）、DOWN/STALE 标记含义、常见排障（隧道不恢复、target DOWN 的排查路径）。`docs/servers.md` SHALL 含指向该文档的链接。

#### Scenario: 按文档可独立启用
- **WHEN** 一名新用户按文档前置检查与启动步骤操作
- **THEN** 无需口头协助即可在本地终端运行仪表盘并看到两台服务器状态

#### Scenario: 降级标记有据可查
- **WHEN** 仪表盘出现 DOWN/STALE 标记
- **THEN** 文档中可查到该标记的含义与对应排查路径
