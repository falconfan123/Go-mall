> 依赖安全升级（grpc v1.82.1 → ≥v1.83.1）。白名单 = go.mod/go.sum + 本变更 openspec 制品。

## 1. 升级与验证

- [x] 1.1 逐模块升级（GOWORK=off）：**仅对 18 个含 grpc 的模块**（common/dal/test/rpc + 15 services）执行 `go get google.golang.org/grpc@v1.83.2 && go mod tidy`；`tools/rag` 不含 grpc 依赖**不升级**（只跑验证）；**同批第二个 CVE**：`services/users` 升 `github.com/rabbitmq/amqp091-go` v1.10.0 → v1.13.0（GO-2026-6372，CVE 必需；同为 `// indirect`，验证 require 行保留 v1.13.0 且 tidy 不回落）；`dal`/`services/gateway` 的 grpc 为 `// indirect`（go get 可升级，验证 require 行版本 ≥ v1.83.1 即可）；验证：18 个模块 go.mod 的 grpc 版本 ≥ v1.83.1
- [x] 1.2 记录带出的 indirect 依赖变化（`git diff` go.mod/go.sum，逐个列出如 golang.org/x/net 等并说明必要性）；验证：diff 清单完整、无"顺手升级"的非必需依赖
- [x] 1.3 漏洞清除：`bash scripts/go-ci-vulncheck.sh` → **18 个模块（go_ci_modules.txt 全部条目）0 命中、exit 0**（两个 CVE 均清）（附原始输出；不得用豁免/allowlist/`//nolint` 掩盖）；验证：命令输出无 `GO-2026-6348` **且无** `GO-2026-6372`
- [x] 1.4 全量门禁：`make build && make lint && make test-unit` 全绿；验证：三条命令输出
- [x] 1.5 F1 分支集成测试：`cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 ./coupons/...` 通过；验证：ok 输出
- [x] 1.6 版本选择定论：若 v1.83.2 全量验证通过则采用；否则回落 v1.83.1 并记录理由（写入 review-log）——**回落后必须重跑 1.1–1.5 并附新输出**（禁止只改版本号不重验）；验证：给出最终版本与依据 + 重跑输出
- [ ] 1.7 提 PR → 等 Build + Quality 全绿（Govulncheck 必须转绿）→ 合并；验证：PR 号 + merge commit hash
- [ ] 1.8 解锁 PR #74（F1）：其 checks 重跑 → Build/Quality 绿 → 合并；验证：PR #74 合并 commit hash
- [ ] 1.9a 实施期间若漏洞库新增命中：按同口径（真升级、记录 indirect、更新漏洞矩阵）纳入本变更并更新漏洞矩阵；验证：矩阵与实际 go.mod 一致
- [ ] 1.9 Integration 备注：若本地/CI Integration 仍因 ES 栈启动失败（`service failed to listen: audit`）而红，标注"归因 = 集成栈就绪问题（另案 fix-integration-stack-readiness），与本变更无关"并附日志片段；验证：归因记录

## 2. 收尾

- [ ] 2.1 红线核查：`git diff --name-only` 白名单 = go.mod/go.sum + 本变更 openspec；无业务代码、无 CI 配置改动；验证：命令输出
- [ ] 2.2 变更归档（验证通过后）：`/opsx-archive fix-govulncheck-grpc-cve`（skip_specs → 主 spec 不变）；验证：归档路径 + commit