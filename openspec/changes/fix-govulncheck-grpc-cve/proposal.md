## Why

**本变更目标 = 让 `scripts/go-ci-vulncheck.sh` 在全部模块上 exit 0（当前漏洞库快照 0 命中）**，从而解锁 Quality required。当前阻塞漏洞两个（同批 2026-09-15T18:39:25Z 披露）：

安全漏洞修复：`GO-2026-6348`（Heap Memory Exhaustion (OOM) via HTTP/2 DATA Frame Fragmentation），模块 `google.golang.org/grpc`，**Found in v1.82.1 / Fixed in v1.83.1**（CVE-2026-84304 / GHSA-vp52-pcj8-j9qc，published 2026-09-15）。CI Quality run `35079947313` 的 `Govulncheck` job 因该漏洞失败（其余 5 job 全绿），导致 Quality required 红、阻塞 PR #74（F1）合并。

## What Changes

- 将 18 个 Go 模块的 `google.golang.org/grpc` 由 **v1.82.1 升级到 ≥ v1.83.1**（目标 **v1.83.2**，回落 v1.83.1）：
  `common`、`dal`、`test/rpc`、`services/{activity,admin,audit,auths,carts,checkout,coupons,gateway,inventory,order,payment,product,search,system,users}`。
- 逐模块 `go get google.golang.org/grpc@<目标版本>` + `go mod tidy`（GOWORK=off），记录带出的 indirect 依赖变化（如 `golang.org/x/net` 等，逐个列出并说明必要性）。
- **同批第二个 CVE（CVE 必需、非顺手升级）**：`github.com/rabbitmq/amqp091-go` **v1.10.0 → v1.13.0**（仅 `services/users/go.mod`，indirect；漏洞 `GO-2026-6372`，调用栈经 `RabbitMQEventPublisher.publish` → `zeromicro/go-queue/rabbitmq` 间接到达 amqp091-go）。
- 验证：`bash scripts/go-ci-vulncheck.sh`（无 Vulnerability、exit 0）+ `make build && make lint && make test-unit` + F1 分支上 coupons 集成测试。
- **回滚方式**：`git revert` 本次 commit + 恢复各 go.mod/go.sum 至升级前版本 + 重新 `go mod tidy`。

## Capabilities

### New Capabilities
<!-- 无：依赖安全升级，无系统运行时行为变化（skip_specs: true）。 -->

### Modified Capabilities
<!-- 无：不改任何 spec 级行为；grpc 版本升级不改变对外契约。 -->

## Impact

- **依赖清单**：18 个 go.mod + 对应 go.sum（grpc 版本 + indirect 变化）；`services/users/go.mod`（amqp091-go v1.10.0→v1.13.0）。
- **漏洞矩阵**：`GO-2026-6348`（grpc，Found v1.82.1 / Fixed v1.83.1，18 模块）→ 升 v1.83.2；`GO-2026-6372`（amqp091-go，Found v1.10.0 / Fixed v1.13.0，services/users）→ 升 v1.13.0。
- **代码**：无业务代码改动（除非升级引入 API 不兼容，需评估——v1.82→v1.83 同 minor，预期兼容）。
- **门禁**：Quality 的 Govulncheck 转绿 → PR #74 解锁合并。
- **Non-goals**：**不升与本目标（`scripts/go-ci-vulncheck.sh` 在全部模块 exit 0，当前漏洞库快照）无关的依赖**；不改 CI 配置/ruleset；不改业务代码。
- **风险**：grpc minor 升级可能带出 indirect 依赖变化（逐个记录）；若 v1.83.2 不兼容则回落 v1.83.1（官方修复版）。