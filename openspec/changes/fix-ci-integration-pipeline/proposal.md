## Why

CI 长期失败：近 200 次 run 成功率 51.5%，失败高度集中在 Integration（几乎必红，2 核 runner 上 `go run` 起 15 服务超时）与 Quality（Go Vet/Govulncheck/Unit Tests 因 CI 模块隔离编译缺 `replace common` 而红）；同时 main 分支无 branch protection、无 required status checks，**Integration 红照常合并（PR #69 实证）**，门禁形同虚设。需要修复 CI 使其稳定可靠，并让门禁真实生效（或明确仅报告），否则 CI 红线既不拦合并也不可诊断，工程质量无从保证。

## What Changes

- **修 Integration（R1）**：`scripts/ci-rpc-stack.sh` 改为**预编译二进制起服务**（先 `go build` 全部服务到 `.artifacts/bin`，启动用二进制而非 `go run`，避开逐服务重复编译）+ 端口等待超时上调（90→180s）+ 依赖就绪等待加强。**不删测试、不 skip、不加 continue-on-error**。
- **修 Quality（R2a）**：`dal/go.mod` 增加 `replace Go-mall/common => ../../common`（对齐 services 模块既有做法），使 `GOWORK=off` 独立编译下 Go Vet/Govulncheck/Unit Tests 转绿；或按 design 定夺 GOWORK 策略。
- **失败可诊断（A4）**：Integration 失败时上传全量诊断 artifact（各服务日志、`docker ps -a`、依赖健康、端口监听快照、etcd/rabbitmq/minio 探针），PR 可访问、保留期明确。
- **门禁分阶段 required（A2，防自锁）**：阶段 1 仅将 **Build + Quality** 设为 required status checks（Integration 仍非 required，不阻塞合并）；阶段 2 在 Integration 修绿并连续观察 N 次全绿后，再将其加入 required。**禁止在阶段 1 把 Integration 设为 required**（会立即阻塞所有 PR）。

## Capabilities

### New Capabilities
- `ci-pipeline-reliability`: CI 流水线可靠性契约——Integration 稳定转绿（连续观察绿率 ≥90%）、Quality 门禁全绿、失败时产出可诊断日志、required status checks 分阶段生效且不阻塞误伤。

### Modified Capabilities
<!-- 无：本 change 为 CI/工程流程改进，不改变任何业务运行时行为 spec。 -->

## Impact

- **CI 配置**：`.github/workflows/{integration,quality}.yml`（诊断产物、required 依赖、构建策略）。
- **脚本**：`scripts/ci-rpc-stack.sh`（预编译二进制模式）、可能 `scripts/go-ci-vet.sh`/`go-ci-build.sh`。
- **依赖清单**：`dal/go.mod`（加 replace）、可能触发 `go.sum` 调整。
- **门禁**：GitHub branch protection（main）required status checks 分阶段配置。
- **风险**：预编译全量 build 在 2 核 runner 的耗时未知（A6 关键假设，design 阶段先实测）；GOWORK 策略改动对 build.yml 矩阵的影响（A5，design 验证项）；Govulncheck 历史独立失败（#33077827869/#29748971423）需核对是否纳入。
- **明确排除**：Mock Consistency 不纳入（A1 已绿）；不改业务代码。**唯一例外（经用户拍板，2026-09-14）**：删除 `dal/model/inventory/inventorymodel.go` 中 1 行不可达死代码（`BatchReturnInventoryAtomWithSession` naked block 外冗余 `return nil`，语义等价）——它是 `go vet` unreachable code 报错源，删除是达成 Quality 全绿（C2）的必要条件；来源为 Initial commit 254330c（2026-03-10）存量缺陷，非本 change 引入。