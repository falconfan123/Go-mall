# Review Log — server-status-dashboard

> 基线：explore-brief.md（Checklist 10 项，2026-09-02 实证 + propose 阶段修正注记）
> 审核方式：@openspec-reviewer 只读子智能体，逐批审核

---

## 批次 1：proposal

### 审查轮次 1（2026-09-02）

**结论：PASS（0 🔴，2 🟡，4 💡）**

- Checklist 前两项 ✅：端到端链路与端口分配明确（9101/9102 与现有 9090/9888/9081/11000-11012 零冲突，grep 实证）；非目标五项与基线逐项对应
- 事实性抽查全过：docs/servers.md 别名/密钥/IP 与 proposal 前置声明逐字一致；scripts 命名无重名
- 🟡1 已修复：Modified Capabilities 注记扩展——specs 批次须声明远程 target 对 observability 基线"target 恒 UP"语义的豁免与优先级
- 🟡2 已修复：Impact 补"node_exporter 下载源可达性"开放问题 + fallback（本地下载 scp）
- 💡 带入后续批次：design 需要求 ssh `ServerAliveInterval/ExitOnForwardFailure`（KeepAlive 只管进程退出）；tasks 补 9100 端口占用预检；target host 沿用 `host.docker.internal`；plist 入库随仓库

**proposal.md 冻结。**

---

## 批次 2：specs（specs/server-status-dashboard/spec.md）

### 审查轮次 1（2026-09-02）

**结论：PASS（0 🔴，3 🟡，3 💡）**

- traceability 双向：proposal 7 条目全部落位；无越权发明（R2 活性探测有批次 1 💡1 背书）
- 豁免声明充分（对照 observability 主 spec 逐字核对）；格式纪律全过（11 scenario 4 井号、每 requirement ≥2 scenario）
- 🟡1 已修复：R4 定义 DOWN/STALE 判据（从未成功→DOWN；曾成功后失败/停更→STALE+最后成功时间）+ 新增强制触发 STALE 的 scenario（消除"或"字死条款）
- 🟡2 已修复：R2 具体配置项名（ServerAliveInterval 等）改为行为表述，参数移交 design
- 🟡3 已修复：新增「仪表盘用法与排障文档」requirement（2 scenario），承接 proposal 的文档条目
- 💡 全部采纳：job 计数去硬编码（"当前 16 个"）；自愈时限上界 ≤10s；Purpose 豁免注记补"隧道断开期间 9101/9102 本地无人监听属合法状态"

**specs 批次冻结（6 requirement / 14 scenario）。**

---

## 批次 3：design

### 审查轮次 1（2026-09-02）

**结论：PASS（0 🔴，5 🟡，4 💡）**

- D1 终选成立：对比表 5 维度 + k3s 现状验证结论在案，与 explore-brief 倾向一致；propose 修正事实（两台均 k3s 节点）被正确钢人化，未继承过时表述
- 冻结对齐：D2/D3/D5 与 proposal 7 条逐条落位；无越权发明（无 tmux 小件/无 k8s 业务指标）；Non-Goals 三方一致
- 技术抽查通过：launchd 检出 45s + 重启 ≤10s（ThrottleInterval=5）满足 R2；PromQL 形状全对（instance 标签覆盖语义正确）；DOWN/STALE 实现与 R4 强判据闭环（staleness 标记→空结果→STALE 路径）
- 批次 1/2 四项 💡 全部承接（plist 入库/9100 预检/host 前缀/ssh 参数组）
- 🟡1 已修复：D1 对比表"故障域"行 ✅ 错标纠正（A 优，A'' 劣势）
- 🟡2 已修复：fallback 腿一标注与冻结 spec R1 字面冲突，启用前须 delta 审批
- 🟡3 已修复：fallback 两腿合并为单方案（叠加生效），修正"网关 IP 命不中 loopback 监听"的网络死角
- 🟡4 已修复：新增 D4b 文档锚点（docs/server-status-dashboard.md 内容清单 = R6 验收范围；servers.md 补链接）
- 🟡5 已修复：gomall-2 内存回退改为"不做单台交付 + delta 审批修订 R1"（原引用 R4 不成立）
- 💡 全部采纳：IdentitiesOnly + 密钥口令/keychain 前置说明；网速收紧 eth0|ens.* 防双计；回滚补产物清理；拼写修正

**design 批次冻结。**

---

## 批次 4：tasks

### 审查轮次 1（2026-09-02）

**结论：FAIL（1 🔴，5 🟡，5 💡）**

- checklist 后 8 项：7 项 ✅；SSH 前置 tasks 层缺自检（💡2）；每任务 ≤2h 大体满足
- 场景覆盖矩阵：13/14——R1-S2 无验证（🟡2）、R2-S5 检出腿无验证（🟡1）、R4-S9 判据缺陷（🔴1）
- 🔴1：4.2/6.2 重新落入"DOWN 或 STALE"或字表述——前置"已展示数值"下按 R4 判据必须 STALE，"永远显示 DOWN"的缺陷实现也能过验收
- 🟡：2.1 无验证子句；公网不可达无验证；bootout 后用 kickstart 恢复不可行；Migration 回滚无任务承接；LogQL 校准无验收判据
- 💡：4.1 粒度、0.1 前置自检、版本参数化、30MB 阈值溯源、9100 占用处置语义对齐

### 审查轮次 2（2026-09-02）

**结论：PASS（0 🔴，3 🟡，4 💡）**——11 项修复 9✅/2⚠️ 无❌；场景覆盖矩阵 14/14 完整（R1-S2→1.2 双 curl、R2-S5→2.1 静态断言+2.3 可选运行期、R4-S9→4.2 两态断言强制 STALE）

- 🟡1 已修复：2.1 grep 断言改 `grep -cE` 四参数值（对 ProgramArguments 逐 token 编码稳健，防假阴性误导出 sh -c 反模式）
- 🟡2 已修复：4.3 Loki 故障注入指向正确 compose（`-f infrastructure/docker-compose.yaml`，原命令指向不存在的服务）
- 🟡3 已修复：6.4 回滚演练改全量口径（与"targets 回 16、无 9101/9102"断言匹配，消除单腿范围矛盾）
- 💡 全部采纳：1.1 版本参数措辞对齐位置参数；unit 路径写死 `/etc/systemd/system/node_exporter.service`；4.1a 补 fstype 过滤保真；review-log 批次 2 场景计数 13→14 更正

**tasks 批次冻结（6 组任务，4 批全部冻结）。**

---

## 最终结论

**✅ propose 反思全部通过（4 批次冻结：proposal 1 轮 / specs 1 轮 / design 1 轮 / tasks 2 轮）。** 可进入 /opsx-apply。
