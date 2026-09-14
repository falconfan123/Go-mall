## Context

- 见 `proposal.md` — Why 与 `explore-brief.md`（含 A1–A6 口径）。根因实证：R1（Integration：2 核 runner 上 `go run` 起 15 服务，首个服务 system 90s 端口等待超时，run #34128731051）；R2a（Quality：`GOWORK=off` 独立编译 dal 模块缺 `replace common` → `undefined biz.ErrReturnAlreadyLocked`，run #34128731043）；R3（main 无 branch protection，PR #69 Integration 红仍合并）。
- 现有脚本：`scripts/ci-rpc-stack.sh`（`go run` + `wait_for_port 90s`）、`scripts/go-ci-vet.sh`（每模块 `GOWORK=off`）、`scripts/test-integration.sh`（起栈后跑 RPC 集成测试）。

## Goals / Non-Goals

- **Goals**：Integration 稳定转绿（A3 判定口径）；Quality 全绿；失败可诊断（A4）；门禁分阶段生效（A2）不削弱强度（C5）。
- **Non-Goals**：不删测试/skip/continue-on-error；不改业务代码；不新增/删除流水线；Mock Consistency 不纳入（A1 已绿）。

## Decisions

### D1 集成修复：B（预编译二进制 + 超时上调 + 依赖探针），A 作回退，C 不采用

- **选定 B**：`ci-rpc-stack.sh` 增加二进制模式——先 `go build` 全部服务到 `.artifacts/bin/`，`start_service` 用 `exec .artifacts/bin/<svc> -f <cfg>` 替代 `go run`；`wait_for_port` 超时 90→180s；依赖就绪等待（etcd/rabbitmq/minio 健康探针）加强。
  - 理由：`go run` 每次启动都触发编译（15 服务 × go.work 全量），2 核 runner 上首个服务必然超 90s（R1 实证）；预编译一次、启动用二进制，彻底移除启动期编译。
- **被否方案代价**：
  - A（纯超时上调）：最小改动，但 2 核编译耗时不可控，仅靠加超时仍可能偶发红，且 15 服务各自 `go run` 编译总时长可能撑爆 45min 流水线超时——作为 B 的**回退**（若 A6 验证表明全量 build 本身不可行）。
  - C（分层/只留关键链路）：改动面大（2-3 天）、改变集成覆盖语义，非本次目标；其"失败上传全量诊断"部分并入 D4 采纳。
- **A6 关键假设（design 第一验证项，量化阈值）**：B 成立前提是"2 核 runner 上 `go build ./...` 全量能在超时内完成并稳定启动"。**go/no-go 阈值**：冷编译（无缓存）≤ 8 分钟、缓存命中（actions/cache 恢复后）≤ 2 分钟，且合计不挤占 45min 流水线预算的 40% 以上；任一超限 → 降级 A（纯超时 + 缓存复用）或部分服务预编译。验证：CI 加一次 `go build ./...` 步骤实测耗时（A6），用 `gh run view --log` 取该步骤 duration。

### D2 模块策略：`dal/go.mod` 加 `replace Go-mall/common => ../../common`，否 GOWORK 全开

- **选定**：在 `dal/go.mod` 增加 `replace github.com/falconfan123/Go-mall/common => ../../common`，对齐 services 模块既有做法（`services/*/go.mod` 均已 replace common/dal）。
  - 理由：最小改动；`GOWORK=off` 独立编译下 dal 模块可解析本地 common（CI 与本地一致）；不触碰 build.yml 矩阵。
- **被否方案代价**：改 GOWORK 策略（workflow 层不设 off、依赖 go.work）——影响 build.yml 服务矩阵构建与 go-ci-vet.sh 的逐模块逻辑，改动面大、风险高。
- **A5 外溢影响（design 验证项，随 D2 一并检查）**：
  1. `build.yml` 服务矩阵（各 services 模块已有 replace，加 dal replace 后 `GOWORK=off` 独立编译应无回归——用 `go build ./...` + `go vet ./...` 在本地模拟 GOWORK=off 验证）；
  2. `go mod tidy`（go-ci-vet.sh 每模块 tidy）加 replace 后结果一致、不触发多余 go.mod/go.sum 变更；
  3. lint/staticcheck 在 GOWORK 开关下行为；
  4. dal 模块未来独立发布的语义变化（本地 replace 不随发布，需文档注明）；
  5. 本地 go.work 共存时 `go build ./...` 无回归。

### D3 门禁：分阶段 required（防自锁）

- **阶段 1**：仅 Build + Quality 设为 required status checks（Integration 仍非 required，红不阻塞合并）。
- **阶段 2**：Integration 满足 A3 判定（14 天内 20 次 run 绿率 ≥90% + 同 commit 3 次全绿）后，加入 required。
- 理由：Integration 现必红，若阶段 1 就 required 会立即阻塞所有 PR（A2）。

### D4 失败诊断产物

- Integration/Quality 失败时 upload artifact（A4 清单）：各服务 stderr/stdout 日志（`scripts/logs/*.log`）、`docker ps -a`、依赖容器健康状态、`ss -ltnp` 端口快照、etcd/rabbitmq/minio 健康探针输出；`actions/upload-artifact` 保留期设 7 天；PR 经 Actions run 的 Artifacts 页可访问。

### D5 Govulncheck 历史失败核对（随 D2）

- 实证：#33077827869（8-27）、#29748971423（7-20）Govulncheck 独立失败（非 Mock 非 R2a）。design/apply 阶段先核对失败内容：若为依赖漏洞 false positive → 记录并更新依赖（不删检查）；若为模块依赖缺失 → 随 D2 一并修复。**不通过删 govulncheck 或 continue-on-error 掩盖**。

## Risks / Trade-offs

- **[2 核 runner 全量 build 耗时不达标]** → A6 验证项先行；不达标则降级 A（纯超时+缓存复用）或部分预编译，成败标准不变。
- **[B 方案改动 ci-rpc-stack.sh 引入回归]** → 本地可复现（C6）：本地跑 `scripts/ci-rpc-stack.sh` 二进制模式应等同线上 start-unified 的起栈行为。
- **[D2 replace 影响 go.mod 一致性]** → A5 验证项（tidy/vet/build 无回归）。
- **[阶段 1 门禁可能把部分当前红的 Quality job 暴露为 required]** → 阶段 1 前先确保 Build+Quality 全绿（本 change 修复 R2a 后），再设 required。
- **[Govulncheck 独立问题未判明]** → D5 核对，不掩盖。
- **[D5 若需升 major 依赖版本 → 违反"不改业务代码"非目标]** → D5 核对若为真实漏洞且需 major 升级，本 change 声明该例外（仅依赖清单变更 + 必要的兼容修正）或拆分为独立 change；此取舍在设计/apply 阶段 D5 核对时明确记录，不默认吞入。

## Migration Plan

1. **A6 验证**：CI 加 `go build ./...` 步骤（临时）实测 2 核全量 build 耗时与稳定性 → 定 D1 是否可走 B。
2. **D2 修复**：`dal/go.mod` 加 replace common → 本地模拟 GOWORK=off 验证 Go Vet/Govulncheck/Unit Tests 转绿（A5 检查项全跑）。
3. **D1 实施**：`ci-rpc-stack.sh` 二进制模式 + 超时 180s + 依赖探针 → 本地二进制模式跑通（C6）。
4. **D4 诊断产物**：integration/quality.yml 失败时上传 A4 清单。
5. **门禁阶段 1**：Build + Quality 设 required（先确认二者已绿）。
6. **观察与阶段 2**：Integration 连续观察至 A3 判定通过 → 加入 required。
7. **验收**：C1–C8 checklist 逐项核对。

**回滚**：所有改动为 CI/脚本/依赖清单，可 git revert；branch protection 为仓库设置可移除。

## Open Questions

- Govulncheck 历史失败的确切内容（D5 核对后再定是否纳入 scope，不改变方案主线）。
---

## A6 验证实测（2026-09-14，增量追加，不改写已冻结结论）

- **CI 2 核 runner 全量 build 耗时**：**146s**（run #34830158766，draft PR #71，go-ci-build.sh --all 15 服务）。
- **本地 Mac 等价计时**：**93s**（go.work 各模块 `go build ./...`，0 失败）。
- **go/no-go 判定**：达标（146s < 冷编译阈值 480s、缓存阈值 120s 略超但冷编译场景；146s << 45min 预算 40% 的 1080s）→ **D1 走 B（预编译二进制起服务）**。
- **新增根因证据**：本次 run 集成失败点为 `audit` 服务连 ES(9200) 30 次重试失败 → `failed to listen: audit`（依赖就绪不足），区别于 9-07 的 `system` 90s 编译超时——R1 扩展为"编译慢 + 依赖就绪"两类，均被 B 方案的预编译 + 依赖探针加强覆盖。
