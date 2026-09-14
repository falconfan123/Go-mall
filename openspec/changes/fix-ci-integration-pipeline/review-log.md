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
