# 事故复盘：2026-06 CI 修复链膨胀

## 概要

2026-06-14 至 2026-06-16，仓库进入了一个 CI 修复无限循环：PR → CI 失败 → Issue → 修复 PR → CI 又失败 → 又 Issue... 72小时内产生了 6 个 Issue、4 个修复 PR，但 main 分支没有任何合并。

## 时间线

### 第一阶段：大 PR 失控

**2026-06-14 — PR #57 feat(rag): add DeepSeek AI client support**

- **5,365 个文件变更**，+942,833 / -11,490
- 混入了多个无关提交：前端搜索/RAG 页面、工程规范文档、CI 模块修复、proto/package 修正、GOWORK/GOTOOLCHAIN 调整
- 进 PR 的文件包括 `frontend/node_modules/`、`frontend/dist/`、`.codegraph/`、生成产物等噪声

**根因：分支管理失焦。** 一个长期分支把多批无关改动一起带上来了，没有做范围控制。

### 第二阶段：CI 全面爆炸

PR #57 的 24 个 CI check 集体失败，原因分为四类：

| 类别 | 失败项 | 根因 |
|------|--------|------|
| **工具链** | Build (17个服务), Go Vet | GOWORK 不是绝对路径（Go 1.25+ 要求） |
| **生成同步** | Mock Consistency | `make mock` 生成代码与已提交不一致 |
| **质量门槛** | Coverage Gate, Go Vet | 新代码覆盖率不足、vet 报错 |
| **环境依赖** | Integration, RPC Integration | 外部服务（search 等）未就绪 |
| **安全基线** | Govulncheck | 依赖需要升级 |

### 第三阶段：自动修复链循环

**2026-06-15 → 2026-06-16**

```
PR #57 CI失败
  → Monitor 创建 Issue #58
  → SymphonyMac 尝试修 #58 → 失败 (max_turns=1, 无PTY)
  
PR #59 (知识库) CI失败  
  → Monitor 创建 Issue #60
  → SymphonyMac 修 #60 → PR #62 → CI又失败（同一根因：GOWORK）
  → Monitor 创建 Issue #63
  → SymphonyMac 修 #63 → PR #64 → CI又失败
  → Monitor 创建 Issue #65
  → SymphonyMac 修 #65 → PR #66 → CI又失败
```

**循环原因：**
1. **ci-fix-monitor 去重锁在 workflow run ID，不是 PR 号** — CI 重跑后 run ID 变，脚本认为有新失败
2. **修复 PR 自身的 CI 也失败** — 根因（GOWORK）没修，每个修复 PR 都在同一块石头上绊倒
3. **SymphonyMac 没有认知阻断** — 它不知道自己在修 fix-of-fix，只看到有新 Issue 就接

## 根因总结

### 1. PR 范围失控
5,365 文件、+94万行的 PR 不可能被有效 review。后期问题排查时也不知道失败来自哪部分改动。

**预防：**
- PR 单一职责，单 PR 文件变更控制在 ~50 个以内
- `.gitignore` 覆盖 `node_modules/`、`dist/`、build 产物、IDE 配置等
- 长期分支定期 rebase/reconcile main，避免一次性合并海量 commit

### 2. 工具链不稳定（GOWORK）
Go 1.25+ 要求 `GOWORK` 为绝对路径，但所有 CI 脚本 `cd` 到子目录后都没有重新设置。这是所有 Build 失败和 Go Vet 失败的根本原因。

**预防：**
- 所有 CI 脚本中 `cd` 到子目录前，`export GOWORK="$ROOT_DIR/go.work"`
- CI 脚本定期在 CI 环境中实际运行，而不是只在本地测试

### 3. 自动修复缺少人类兜底
ci-fix-monitor + SymphonyMac 的自动化链条缺少"认知阻断"机制。连续产生 fix-of-fix 时没有人停下来问"我们修对了方向吗"。

**预防：**
- 同一个 PR 自动创建 ci-fix Issue 超过 1 次后，暂停自动修复，转人工
- ci-fix-monitor 脚本增加循环检测：如果最近 N 个 Issue 都是对同一个 PR 的修复 PR，报警而不是继续创建
- 修复 PR 的 CI 首次失败时，直接标为"需要人工介入"，不继续自动修复

### 4. 新 Issue 不查 PR 已有 Issue
ci-fix-monitor 的 dedup 靠 workflow run ID，不是 PR 号。CI 重跑后 run ID 变，就重复创建 Issue。

**预防：**
- 创建新 Issue 前，查 GitHub 是否有 open ci-fix Issue 引用此 PR
- 有则更新已有 Issue，不新建

## 修复清单

| 修复 | 状态 | 位置 |
|------|------|------|
| `export GOWORK` 绝对路径 | 已修 PR #59 | `scripts/go-ci-build.sh`, `go-ci-vet.sh`, `go-ci-vulncheck.sh`, `test-unit.sh`, `coverage.sh` |
| ci-fix-monitor dedup 改 PR 级 | 已修 | `~/.hermes/scripts/ci-fix-monitor.sh` |
| SymphonyMac max_turns: 1→50 | 已修 | `orchestrator_state.json` |
| CI-FIX 知识库 | 已建 PR #59 | `CI-FIX-KNOWLEDGE.md`, `CI-FIX-CASES/` |
| 循环产物清理 | 已完成 | Issue #60 #61 #63 #65 关闭, PR #62 #64 #66 关闭 |

## 经验教训（给未来 agent）

1. **CI 失败先看根因分类** — 是工具链问题、生成同步问题、还是代码逻辑问题？GOWORK 类的问题修一个就可以解决十几个失败的 check。
2. **不要修 fix-of-fix** — 如果看到 Issue 引用的 PR 本身就是一个修复 PR，先查原始 PR 的 CI 状态。可能原始问题的修法就能解决一切。
3. **看 PR 范围** — 5,000+ 文件的 PR 本身就说明分支管理有问题。不要在这个 PR 上逐个修 CI，应该先拆分 PR。
4. **自动循环要手动熔断** — 3 个以上的连续 ci-fix Issue 说明自动化跑偏了，停下来人工判断。
