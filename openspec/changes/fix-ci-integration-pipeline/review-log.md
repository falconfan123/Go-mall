# Review Log · fix-ci-integration-pipeline

## 整体反思审核（proposal/specs/design/tasks 一次性生成后整体审核）

**审核方式**：@openspec-reviewer 对照 explore-brief（含 A1–A6）与 C1–C8 checklist 逐条核（证据真实性抽查通过：wait_for_port 90s / GOWORK=off / dal 无 replace / Quality 6 job 均属实）。

### 🔴 遗留
🔴 无（冻结）。

### 🟡 修复（本轮一次修完，8 条）
1. tasks 1.1 触发条件错误（push 不触发 integration.yml；--log-failed 无成功 step 耗时）→ 改 draft PR 触发 + `--log` 取 duration。
2. tasks 2.4 Quality job 数不符（实际 6 个）→ 改 6 job + draft PR 触发。
3. tasks 1.2 GOTOOLCHAIN 写法 → 统一 `GOTOOLCHAIN=go$(...)`（对齐脚本既有）。
4. tasks 6.1/6.2 跨 14 天违反 ≤2h 粒度 → 标注"合并后跟踪任务，不阻塞 apply/verify"+ 攒 run 机制（gh run rerun）。
5. design A6 go/no-go 无量化阈值 → 钉死：冷编译 ≤8min / 缓存命中 ≤2min / 不挤占 45min 预算 40%。
6. tasks 7.1 验收覆盖不全 → 补 `--diff-filter=D '*_test.go'` 删测试检测 + `git diff --name-only` 业务代码白名单（C7）。
7. spec Requirement 3 Quality 诊断物理不成立（runner 无 docker 栈）→ 限定 Quality 场景聚焦 vet/govulncheck 输出落盘 + runner 快照，空 artifact 显式记录不静默。
8. design D5 与"不改业务代码"潜在冲突 → Risks 补：Govulncheck 若需 major 升级则声明例外或拆分独立 change。

### 💡 修复（3 条）
1. tasks 5.2 required context 名 = check-run 名（`Build`/`Quality` 非 workflow 名）+ 注明需 repo admin 权限。
2. tasks 3.3 本地验证 `ss`→`lsof -iTCP -sTCP:LISTEN -P`（macOS 无 ss）。
3. tasks 3.4 集成测试基线措辞收紧（真实失败不得当基线放过）。

### 结论
整体审核 🔴 清零，proposal/specs/design/tasks 冻结。制品与 explore-brief A1–A6 / C1–C8 完整承接，"不削弱门禁"贯穿四层。

## apply 增量：A6 验证结论（2026-09-14）

- CI 全量 build **146s**（run #34830158766）、本地 **93s** → 达标，**D1 确认走 B**。
- 新根因：audit 连 ES 失败（30 次重试）→ failed to listen（依赖就绪），与 system 编译超时并列，B 方案覆盖两者。
- 附：draft PR #71 触发方式验证通过（integration.yml 仅 pull_request 触发，需 draft PR）。

## apply 增量：D6 例外（2026-09-14，用户拍板选 A）

- 用户拍板：本 change 允许唯一业务代码例外 = 删除 `dal/model/inventory/inventorymodel.go:102` unreachable code（naked block 外冗余 `return nil`，语义等价）。
- 来源核实：`git show 254330c`（Initial commit，2026-03-10）确认该结构存量存在，非 settlement 遗留。
- 成因：R2a 编译错误（undefined biz）使 vet 在编译阶段失败、分析未执行；修复 dal 编译后 vet 才暴露存量 unreachable——"编译错误掩盖 vet 分析"。
- 要求：删除后 go-ci-vet.sh 全模块无第二处 unreachable（有则停下报告）；make test-unit 无回归。

## apply 增量：第 2 组完成（2026-09-14）

- D2（dal replace common）+ D6（删 unreachable 1 行）+ unit-tests redis service 三处修复后，**Quality 6 job 全绿**（run #34833333861：Go Vet/Unit Tests/Govulncheck/Mock Consistency/Coverage Gate/Quality 全 success）。
- Unit Tests 原失败根因补充：CI 无 redis → idempotency 测试 PingCtx 内部 logx.Must panic（非编译/测试断言失败），加 redis:7 service 解决（CI 依赖提供，不削弱门禁）。
