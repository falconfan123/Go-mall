> 验证命令均在 Go-mall 仓库执行；涉及 GitHub 用 `gh`。任务粒度 ≤2h。

## 1. A6 关键假设验证（design 第一验证项）

- [ ] 1.1（验证命令修正）在 CI 加一步 `go build ./...`（先于起栈），**开 draft PR 触发**（integration.yml 仅在 pull_request→main / push→main 触发，临时分支 push 不触发），`gh run view <id> --log` 取该 step 的 duration 字段（成功 run 用 --log 而非 --log-failed）；验证：记录 build 步骤耗时秒数，对照 design A6 阈值（冷编译≤8min/缓存≤2min）判定 B 方案可行性
- [ ] 1.2 本地验证等价性：`GOTOOLCHAIN=go$(cat .go-version|tr -d ' ') go build ./...` 在本地计时，与 CI 结果对比；验证：命令完成且给出本地耗时（作为 CI 参考基线）

## 2. D2 模块修复（Quality 转绿）

- [ ] 2.1 `dal/go.mod` 增加 `replace github.com/falconfan123/Go-mall/common => ../../common`；验证：`cd dal && GOWORK=off go build ./...` 通过、不再报 `undefined biz.ErrReturnAlreadyLocked`
- [ ] 2.2 模拟 CI 独立模块编译全绿：`bash scripts/go-ci-vet.sh`（GOWORK=off 逐模块 vet）；验证：命令 exit 0、无 `undefined` 类错误
- [ ] 2.3 A5 外溢检查：`make lint`（staticcheck 等）通过；`GOWORK=off go mod tidy` 不产生多余 go.mod/go.sum 变更（`git diff --stat` 仅预期文件）；`go build ./...`（go.work 下）无回归；验证：三条命令均通过、diff 符合预期
- [ ] 2.4（验证命令修正）开 draft PR 触发一次 Quality 流水线，观察 6 个 job（mock-consistency/unit-tests/coverage/go-vet/govulncheck/quality-summary）；验证：`gh run view <id> --json jobs` 的 6 个 job 全 success

## 3. D1 集成脚本二进制化

- [ ] 3.1 `ci-rpc-stack.sh` 增加二进制模式：前置 `go build` 全部服务到 `.artifacts/bin/`，`start_service` 用 `exec .artifacts/bin/<svc> -f <cfg>` 替代 `go run`；验证：脚本含二进制启动路径且 `go run` 不再用于业务服务
- [ ] 3.2 端口等待超时 90→180s + 依赖就绪探针加强（etcd/rabbitmq/minio 健康探测循环）；验证：`grep -n "wait_for_port\|wait_dependency\|180" scripts/ci-rpc-stack.sh` 显示新值
- [ ] 3.3 本地复现（C6）：`GO_MALL_TEST_LOCAL_ONLY=1 bash scripts/ci-rpc-stack.sh`（或按脚本本地模式）二进制启动全部服务至全部端口监听；验证：`scripts/logs/*.log` 各服务就绪、`lsof -iTCP -sTCP:LISTEN -P` 显示 10000-10012/8081/8888 等端口监听（macOS 无 ss，用 lsof，与脚本一致）
- [ ] 3.4 本地跑集成测试：`GO_MALL_TEST_LOCAL_ONLY=1 bash scripts/test-integration.sh`；验证：测试报告 0 失败；若本地集成测试当前即失败，记录为基线后**必须追查至 0 失败或与已有 main 基线一致**，不得把真实失败当基线放过

## 4. D4 失败诊断产物

- [ ] 4.1 integration.yml / quality.yml 失败路径（`if: failure()`）增加 upload-artifact 步骤，打包 A4 清单（各服务 `scripts/logs/*.log`、`docker ps -a`、依赖健康、`ss -ltnp`、etcd/rabbitmq/minio 探针输出）；验证：workflow 含 artifact 上传步骤、`retention-days: 7`
- [ ] 4.2 人为触发一次失败（临时中断一个依赖）验证产物生成；验证：`gh run view <id>` 的 Artifacts 可下载、含全部清单文件（验证后还原）

## 5. 门禁阶段 1（Build + Quality required）

- [ ] 5.1 确认 Build/Quality 当前全绿（本 change 修复后）：`gh run list --workflow build.yml --limit 3` + `--workflow quality.yml --limit 3` 全 success；验证：全绿记录
- [ ] 5.2（验证命令修正）设置 main branch protection（required status checks context 名 = **check-run 名**，即 `Build`（build-summary）与 `Quality`（quality-summary），非 workflow 名，矩阵 job 如 `Build order` 不设）；需 repo admin 权限；验证：`gh api repos/falconfan123/Go-mall/branches/main/protection --jq '.required_status_checks.contexts'` 含 Build/Quality 且不含 Integration
- [ ] 5.3 防自锁验证：开一个临时 PR（Integration 可能红），确认不被 Build/Quality required 阻塞；验证：`gh pr checks <num> --required` 只列出 Build/Quality，PR 可合并（临时 PR 关闭）

## 6. 观察 + 门禁阶段 2（Integration required）

- [ ] 6.1（验证命令修正）【合并后跟踪任务，不阻塞本 change 的 apply/verify】观察 Integration：修复合并后按 run 序累计（A3 口径：14 天内 20 次 run 绿率 ≥90%，同 commit ≥3 次全绿）；低 PR 流量下用 `gh run rerun <run-id>` 对同一 commit 重复触发攒满 20 次；验证：`gh run list --workflow integration.yml --limit 20` 统计绿率，记录达标日期
- [ ] 6.2（验证命令修正）【合并后跟踪任务，不阻塞 apply/verify】Integration 满足 A3 判定后加入 required（阶段 2，check-run 名 = `Integration`）；验证：`gh api repos/falconfan123/Go-mall/branches/main/protection --jq '.required_status_checks.contexts'` 含 Integration

## 7. 验收与收尾

- [ ] 7.1（验证命令修正）C5 无削弱门禁 + C7 无业务代码改动：`git diff` 无 `continue-on-error`/`t.Skip`；`git diff --diff-filter=D -- '*_test.go'` 无删除测试文件；`git diff --name-only` 白名单核对（仅 .github/workflows/ scripts/ dal/go.mod go.sum 类，无 services/*/internal 业务代码）；验证：三条命令输出符合预期
- [ ] 7.2 D5 Govulncheck 历史失败核对：#33077827869/#29748971423 的失败内容分类（依赖漏洞 false positive → 更新依赖记录；模块缺失 → 已随 D2 修复）；验证：给出分类结论与处置记录
- [ ] 7.3 对照 spec 各 Scenario 与 explore-brief C1–C8 逐项核对，产出验收记录（含 run id 与绿率数字）；验证：验收记录文件包含全部证据