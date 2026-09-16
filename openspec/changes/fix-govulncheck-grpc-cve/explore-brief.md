# Explore Brief · fix-govulncheck-grpc-cve

> 探索结论落盘，作为提案审核需求基线。

## 背景（实证）

- **漏洞**：`GO-2026-6348`（Heap Memory Exhaustion (OOM) via HTTP/2 DATA Frame Fragmentation），模块 `google.golang.org/grpc`，**Found in v1.82.1 / Fixed in v1.83.1**（别名 CVE-2026-84304 / GHSA-vp52-pcj8-j9qc，published 2026-09-15T18:39Z）。
- **CI 证据**：Quality run `35079947313` 6 job 中 5 绿，仅 `Govulncheck` 红（该 job 输出 `Vulnerability #1: GO-2026-6348 ... Found in: google.golang.org/grpc@v1.82.1 / Fixed in: v1.83.1`）。此前 2026-09-14 Quality（run #34833333861）Govulncheck 尚绿 → 新披露漏洞，时间因素。
- **影响面（18 个 go.mod 含 grpc v1.82.1，逐个列出）**：
  `common/go.mod`、`dal/go.mod`、`test/rpc/go.mod`、`services/{activity,admin,audit,auths,carts,checkout,coupons,gateway,inventory,order,payment,product,search,system,users}/go.mod`
- 镜像可用（goproxy.cn 有 v1.83.1、v1.83.2）。

## 目标 / 非目标

- **目标**：把 grpc 升级到官方修复版及以上（目标 v1.83.2，回落 v1.83.1），使 `bash scripts/go-ci-vulncheck.sh` 无 Vulnerability 输出且 exit 0；全量验证（build/lint/test-unit）通过；解锁 PR #74（Quality required）。
- **非目标**：不顺手升级其它依赖（除 grpc 升级**必需**带出的 indirect 依赖，如 golang.org/x/net 等，需逐个列出并说明必要性）；不改业务代码；不改 CI 配置/ruleset。

## 待决策点

- **D1 目标版本**：先试 **v1.83.2**（同 minor 最新补丁）；若与代码不兼容或验证不过 → 回落 **v1.83.1**（官方明确修复版）。以"全量验证通过"为最终判据，选择理由写进 review-log。
- **D2 升级方式**：逐模块 `go get google.golang.org/grpc@<目标版本>` + `go mod tidy`（GOWORK=off；模块清单以 `scripts/go_ci_modules.txt` 为准，另含 test/rpc）。indirect 变化逐个记录。
- **D3 回滚**：`git revert` + go.mod/go.sum 复原 + 重新 `go mod tidy`。

## Checklist（反思审核基线）

- [ ] C1 `bash scripts/go-ci-vulncheck.sh` 在 18 个模块 exit 0 且无 `GO-2026-6348` **且无 `GO-2026-6372`**（附原始输出）；**不得**用 govulncheck 豁免/allowlist/`//nolint` 掩盖（必须真升级）。
- [ ] C2 18 个 go.mod 的 grpc 版本均 ≥ v1.83.1（逐个核对）+ `services/users` 的 amqp091-go ≥ v1.13.0；带出的 indirect 依赖变化逐个列出。
- [ ] C3 `make build` / `make lint` / `make test-unit` 全绿。
- [ ] C4 `cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 ./coupons/...` 通过（F1 分支上）。
- [ ] C5 PR 的 Build + Quality 全绿（Govulncheck 转绿）→ 合并 → PR #74 checks 重跑并合并。
- [ ] C6 不越界（白名单 = go.mod/go.sum + 本变更 openspec 制品）；Non-goal 不顺手升其它依赖。