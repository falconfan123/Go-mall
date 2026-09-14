# Explore Brief · fix-ci-integration-pipeline

> 探索会话（2026-09-13，CI 长期失败实证调查）结论落盘，作为提案审核的需求基线 checklist。

## 背景（实证）

- **CI 整体**：近 200 次 run 成功率 51.5%（成功 103 / 失败 91 / 取消 6）；平均时长 515s、中位 204s。按流水线：Build 53 次基本全绿、Integration 53 次**几乎全红**、Quality 53 次**半红**；另 41 次为已删除历史 workflow「CI」的残留 run。
- **流水线结构**：`.github/workflows/{build,integration,quality}.yml` 3 条（无第四条）。
  - Integration：`rpc-integration`（起依赖 docker compose + 分阶段 `go run` 起 15 个业务服务 + `make test-integration`）+ `integration-summary`（聚合，红为连锁）。
  - Quality：`mock-consistency / unit-tests / coverage / go-vet / govulncheck / quality-summary`（coverage 门禁 `make coverage-ci`，阈值 30%，实测总覆盖 65.6%）。

### Integration 必红根因（R1：环境编排在 2 核 runner 上不可行）

- `scripts/ci-rpc-stack.sh` 用 `go run` 起服务，`wait_for_port <port> 127.0.0.1 90 1`（**90 秒**端口等待超时）。
- 实证日志（run #34128731051，2026-09-07）：依赖容器全 Healthy → `waiting for go-mall-etcd` → `starting system on 10010`（13:42:29）→ 90s 后 `service failed to listen: system`（13:44:01）→ `make test-integration` Error 1。
- 根因：GitHub `ubuntu-latest` 2 核 runner 上 `go run` 首次编译 go.work 全量（common/dal + 15 服务）超 90s，首个服务 system 必然监听超时。**系统性、必然失败**，非偶发 flaky。

### Quality 半红根因（R2：CI 模块隔离与本地 go.work 差异）

- 实证（run #34128731043，2026-09-07）：`Go Vet` 与 `Govulncheck` 报 `dal/model/inventory/inventorymodel.go:60: undefined: biz.ErrReturnAlreadyLocked`；`Unit Tests`（make test-unit）Error 1（同源：包列表含 dal 模块编译失败）。
- 根因：`quality.yml` 的 Go Vet/Govulncheck 设 `GOWORK: 'off'` 独立编译各模块；`dal/go.mod` **缺 `replace Go-mall/common => ../../common`**（仅远程伪版本 `v0.0.0-20260312...` 引用），CI 拉取失败 → undefined。本地 go.work 纳入 common 故本地绿。
- 更早（2026-08-27 及以前）：Quality 半红为 **Mock Consistency** 失败（mock 与代码不同步，另一时期问题，run #33078759527）。

### 门禁有效性（R3：形同虚设）

- **main 分支无 branch protection**（`gh api .../branches/main/protection` → 404 "Branch not protected"），无任何 required status checks。
- 实证：PR #69（2026-08-27 合并）的 checks **RPC Integration FAILURE + Integration FAILURE**，其余（Build/Unit/Go Vet/Govulncheck/Quality/Coverage Gate）全 SUCCESS——**Integration 红照常合并**。
- 结论：Integration 红线既不拦合并也不被 required 化；门禁实质只对 Build/Quality（绿时）有报告意义。**要么修好让它真能拦，要么明确它只做报告不拦**（决策点）。

## 目标 / 非目标

- **目标**：Integration 流水线稳定转绿（连续 N 次绿、近 20 次绿率 ≥90%）；Quality 全绿（修 R2a 模块问题；Mock Consistency 为历史问题不纳入）；失败时 CI 产出可诊断日志；明确门禁定位（拦合并 or 仅报告，二选一拍板）。
- **非目标**：不删测试、不 skip、不用 continue-on-error 掩盖；不改业务代码；不新增/删除流水线。**Mock Consistency 不纳入本 change**（A1 实证：最后一次失败为 2026-08-27 #33078759527，最近 2026-09-07 #34128731043 已绿，为历史已修复问题）。

## 待决策点

- **D1 fix 深度**（三选一）：
  - **A 最小**：调 `wait_for_port` 超时（90→180s）+ 依赖就绪探针加强 + go build 缓存复用。改动面小（ci-rpc-stack.sh），风险：若纯靠加超时仍不稳（2 核编译波动），治标。
  - **B 中等（推荐）**：ci-rpc-stack.sh 支持**预编译二进制**模式（先 `go build` 全部服务到 `.artifacts/bin`，启动用二进制而非 `go run`，避开逐服务编译）+ 超时上调。本地与 CI 一致可复现。改动面中（脚本逻辑 + 前置 build）。
  - **C 彻底**：分层（单测/契约/集成）+ 集成只留关键链路 + 失败自动上传全量诊断（各服务日志/端口/etcd 状态）。改动面大（2-3 天），可能改变覆盖语义。
- **D2 Quality 修复**：`dal/go.mod` 加 `replace Go-mall/common => ../../common`（对齐 services 模块）vs 改用 go.work workspace 编译（GOWORK 不 off）——需在 CI 与本地一致性上权衡。**外溢影响（A5，design 阶段验证项）**：
  1. `build.yml` 服务矩阵构建是否受影响（各 services 模块已有 `replace common/dal`，dal 加 replace 后 GOWORK=off 独立编译是否仍通过）；
  2. `go mod tidy` 行为（CI 的 go-ci-vet.sh 每模块 `GOWORK=off go mod tidy`，加 replace 后 tidy 结果是否一致、是否触发 go.mod/go.sum 变更）；
  3. lint/静态检查（Makefile lint、staticcheck）在 GOWORK 开关下的行为；
  4. 未来 dal 模块独立发布的影响（replace 是本地路径，发布语义变化）；
  5. go.work 与模块 replace 共存时本地 `go build ./...` 是否无回归。
- **D3 门禁定位**：补 branch protection + required checks（让 CI 真拦）vs 明确 Integration 仅报告（降级为可选 job）。**分阶段 required（A2，防自锁）**：
  - **阶段 1**：仅把 **Build + Quality** 设为 required status checks（Integration 仍非 required，不阻塞任何 PR）。
  - **阶段 2**：Integration 修绿并连续观察 N 次全绿后，再将其加入 required。
  - 若阶段 1 就把 Integration 设为 required，会立即阻塞所有 PR（现 Integration 必红）。
- **D4（已核实，降级）**：Mock Consistency 为历史问题，不纳入本 change（A1 实证：最后一次红 2026-08-27 #33078759527；2026-09-07 #34128731043 该 job 已绿）。另注意 Govulncheck 在 2026-08-27 #33077827869 / 2026-07-20 #29748971423 独立失败（非 Mock 非 R2a），design 阶段需单独核对是否为依赖漏洞 false positive 或模块问题，若仍红纳入 scope。

## Checklist（反思审核基线）

- [ ] C1 Integration 稳定转绿，**可判定口径（A3）**：从本 change 首个修复 commit 合并之后开始计数；按 run 序（非按天）；观察窗口 = 修复合并后 14 天内满 20 次 Integration run，绿率 ≥90%；且**同一 commit 至少重复触发 3 次全绿**（验证非 flaky）。
- [ ] C2 Quality 全绿（Go Vet/Govulncheck/Unit Tests/Coverage Gate 均绿；Mock Consistency 已在 A1 证实绿，不作为本 change 验收项）。
- [ ] C3 失败时 CI 产出可诊断日志（A4 清单）：各业务服务 stderr/stdout 日志（`scripts/logs/*.log` 或 `.artifacts/ci-rpc-stack/*.log`）、`docker ps -a` 全量、依赖容器健康状态、端口监听快照（`ss -ltnp`）、etcd/rabbitmq/minio 健康探针结果——以上打包为 artifact 上传，PR 中通过 Actions run 的 Artifacts 页面可访问，保留期遵循 repo 默认（建议显式设为 7 天）。
- [ ] C4 门禁定位明确（**分阶段 required，A2**）：阶段 1 仅 Build+Quality 设 required；阶段 2 在 Integration 连续观察 N 次全绿后加入 required。禁止在阶段 1 把 Integration 设为 required（防自锁）。
- [ ] C5 无删测试/无 skip/无 continue-on-error 掩盖（diff 中无 `continue-on-error`、`skip`、删除测试）。
- [ ] C6 本地可复现（`scripts/ci-rpc-stack.sh` 二进制模式本地跑通，与 CI 一致）。
- [ ] C7 不引入业务代码改动（仅 CI/脚本/依赖清单变更）。
- [ ] C8（B 方案关键假设，A6）："2 核 runner 上 `go build ./...` 全量能在超时内完成并稳定启动"为 B 方案成立前提。验证方式：先在 CI 加一次 `go build ./...` 步骤观察耗时（或 A+B 组合：预编译二进制 + 超时上调 90→180s），若全量 build 超时/不稳，B 方案需降级为 A（纯超时）或 C（分层）。该验证作为 design 阶段第一验证项。

## 待探索/风险（A6 关联）

- R-A6：GitHub `ubuntu-latest` 2 核 runner 上 go build 全量（common/dal+15 服务）耗时未知。若 >180s，预编译二进制方案需评估是否可行（缓存命中后可显著缩短）；作为 design 阶段验证项，先实测耗时再定 D1。

## 发现项（本 change 不修，另开变更）

- 暂无真实业务/测试代码 bug 证据（本地单测/覆盖率/vet 全绿；CI 失败均为环境编排与模块隔离）。若实施中在可诊断日志里发现真实代码缺陷，另行记录。