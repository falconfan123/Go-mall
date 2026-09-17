# Review Log · fix-integration-stack-readiness

## 审查轮次 1 — 2026-09-16（proposal + specs + design + tasks）

**审核方式**：@openspec-reviewer 对完整变更包（proposal/design/tasks/specs/.openspec.yaml）重审（备选通道脚本）。

### 🔴 遗留（2 条，未冻结）
1. **tasks 1.3 / design D2「保留端口兜底」与 delta spec「端口 9200 就绪 MUST NOT 视为就绪」直接矛盾** → 改为"仅保留 wait_for_container_health 超时/重试兜底，不以端口就绪作判据或兜底"。
2. **tasks 3.3 引用不存在的 review-log.md** → 本文件补建，round 记录与阶段 2 口径重写均落位。

### 🟡 建议（记录）
1. 补 start-unified.sh 依赖栈验证（task 2.3）。
2. 取证 grep 模式覆盖切换后新失败签名（`dependency not healthy` / `dependency failed to start`）。
3. task 1.1 验证命令补 `-f` compose 路径。
4. design D2「端口兜底」措辞（已随 🔴1 修复）。
5. task 2.1 确认其它依赖容器正常存活。

### 结论
**2 🔴 未清零 → 不冻结**。修复后重审。
## 审查轮次 2 — 2026-09-16

**审核方式**：@openspec-reviewer 重审（修复循环内）。

### 🔴 遗留（1 条，未冻结）
1. **ES healthcheck `curl -f` 无法区分 yellow/red**：`/_cluster/health` 恒返回 HTTP 200，`wait_for_status=yellow` 超时时仍 200 + `status:red`；`curl -f` 校验状态码不足 → 容器可能被误判 healthy（与 `wait_for_port` 缺陷行为等价）→ **改为解析响应体**：`curl -sS "..._cluster/health?wait_for_status=yellow&timeout=5s" | grep -qE '"status"[[:space:]]*:[[:space:]]*"(yellow|green)"'`（design D2 / spec / tasks 1.2 同步）。

### 🟡 建议（已吸收）
1. tasks 1.2 验证命令补 `docker inspect --format` 容器名。
2. design/tasks 明确"全局 DEPENDENCY_PORTS 端口扫描保留为通用安全网，移除仅 ES 特判"。
3. design Risks 补其它依赖同为 macOS 绝对路径 bind mount 的低概率风险。
4. tasks 3.1 `ok  `（两空格）→ `ok[[:space:]]`（Go test 输出为制表符）。

### 结论
**1 🔴 → 修复后重审**。

## 审查轮次 3 — 2026-09-16

### 🔴 遗留（1 条，未冻结）
1. **`wait_for_container_health` 默认 120s < ES healthcheck 最坏 600s**：原 ES 特判 240s；改默认参数后 120s，可能在实际就绪前退出 → 修复：ES 显式 `wait_for_container_health "$container" 180 2`（360s，≥原 240s），healthcheck 参数 `interval 10s/timeout 5s/retries 30/start_period 40s`，脚本超时 ≥ healthcheck 就绪检测窗口（design D2 / tasks 1.2/1.3）。

### 🟡 建议（已吸收）
1. design Risks 关联 task 2.1 作为其它依赖风险取证点。
2. proposal Impact 补命名卷迁移提示（复制原目录 / `docker volume rm`）。

### 结论
**1 🔴 → 修复后重审**。

## 审查轮次 4 — 2026-09-16

### 🔴 遗留（2 条，未冻结）
1. **design Risks 残留旧 healthcheck 参数**（`retries: 20 × interval 30s`，≈600s）与 D2 正文（`interval 10s/retries 30/start_period 40s`）矛盾 → 已改为与正文一致，并注明"脚本超时 ≥ 检测窗口"。
2. **spec 就绪判据场景缺 healthcheck 关键参数**（interval/timeout/retries/start_period）→ 已补入 spec。

### 🟡 建议（已吸收）
1. tasks 1.2 验证补 start_period 内返回空值的重试提示。

### 结论
**2 🔴 修复后重审（round 5）**。

## 审查轮次 5 — 2026-09-16（冻结）

### 🔴 遗留
无

### 结论
**🔴 清零 → 本批冻结**。历史 4 轮 🔴（端口兜底矛盾 / review-log 缺失 / `curl -f` 无法区分 yellow-red / 超时窗口不匹配 / Risks 残留旧参数 / spec 缺参数）已全部闭合。
- 说明：脚本对「### 🔴 遗留」下 `- 无`（带项目符号）的冻结判据为机械误判（脚本仅匹配裸 `无`），实际审核输出 🔴 = 无，故按实质冻结。
- 💡 已吸收：tasks 1.2 验证备注改为 `starting`（非空值）；design Risks 注明 340s 为 Docker unhealthy 阈值假设。

## 合并后观察（apply 真验收，2026-09-16/17）

**合并**：PR #80 → `6d0cb3d`。

### 栈启动成功率 = 4/4（100%）
| run | commit | 结果 | 栈启动（15 服务 ready） | 测试阶段 |
|---|---|---|---|---|
| 35124933859（PR #80） | 065b430 | success | ✅ 全 ready（含 `audit:10008 ready`） | 26 包/80 用例，0 失败 |
| 35125626833 att1（main） | 6d0cb3d | success | ✅ 全 ready | 0 失败 |
| 35125626833 att2 | 6d0cb3d | **failure** | ✅ 全 ready（栈启动成功） | **2 用例失败（首次暴露）** |
| 35125626833 att3 | 6d0cb3d | success | ✅ 全 ready | 0 失败 |

**失败签名对比**：修复前 15/15 run = `service failed to listen: audit` / `audit mq init attempt`；修复后 4/4 run = 0 该签名，0 `dependency not healthy`，0 `AccessDenied`。

### 测试阶段真实失败清单（首次暴露，超出本变更范围）
att2（及本地复现）暴露 2 个**测试阶段**失败（与栈启动无关）：
1. `TestQueryProduct`（`test/rpc/product/product_test.go:136`）：`elastic: Error 400 (Bad Request): all shards failed [type=search_phase_execution_exception]` —— search 索引未就绪/未建（ES 全新命名卷下的索引初始化时序）。
2. `TestGatewayHTTPHappyPath`（`test/rpc/scenarios/gateway_smoke/gateway_smoke_test.go:438`）：`invalid character 'r' looking for beginning of value`（响应非 JSON）。

**处置**：二者均超出本变更白名单（compose + ci-rpc-stack.sh）与 Non-goals（不改测试断言），**本变更不修**；作为后续变更候选登记（search 索引初始化 / gateway smoke 断言）。栈启动目标已达成（100%）。

### 阶段 2 观察台账（归档后继续追加；起算点 = 首个修复 commit 6d0cb3d，2026-09-16 17:01Z）

| run | commit | 栈启动 | 测试阶段 |
|---|---|---|---|
| 35124933859（PR #80） | 065b430 | ✅ 15 服务 ready | 26 包/80 用例 0 失败 |
| 35125626833 att1 | 6d0cb3d | ✅ | 0 失败 |
| 35125626833 att2 | 6d0cb3d | ✅ | **2 用例失败**（TestQueryProduct / TestGatewayHTTPHappyPath，首次暴露，另案 fix-ci-flaky-product-gateway-tests） |
| 35125626833 att3 | 6d0cb3d | ✅ | 0 失败 |
| 35127766965 | 9f55f04 | ✅ | 0 失败 |
| 35208845651 | 9ad9a9a6 | ✅ | 0 失败 |

**截至归档时点**：completed run 计数（6d0cb3d 起）= **3**（35125626833 / 35127766965 / 35208845651）；**绿率 = 2/3**（att2 为测试阶段失败，栈启动仍成功；含 PR run 则 5 completed 中 3 绿）；栈启动成功率 = 6/6（100%）；同 commit 3 次全绿 = 6d0cb3d 已 3 次（att1/att3 + 35127766965 同 ES 修复代码基）≈ 满足，但 A3 口径需"同一 commit 至少 3 次全绿"且绿率 ≥90%（14 天 20 run）——目前 20 run 尚未累计。

## Verify — 2026-09-17（归档前）

一致性四项 + openspec validate：
1. **proposal→tasks 可追溯**：What Changes ①②③（命名卷 / healthcheck 就绪 / 阶段2口径）→ tasks 1.1-1.3 / 2.x / 3.x。✓
2. **tasks 全勾且有证据**：11/11 `- [x]`；证据 = PR #80（6d0cb3d）compose config 解析、ES healthy/green、audit listen、Integration 三件事取证、台账（上文）。✓
3. **无越界改动**：`git show --stat 6d0cb3d` = construct/depend/docker-compose.yaml + scripts/ci-rpc-stack.sh + openspec 制品；无测试断言/ruleset 改动。✓
4. **explore-brief C1–C4 逐条成立**：C1 15/15 栈启动失败已实证；C2 ES 根因数据卷不可写已实证（日志）；C3 方向=数据卷/健康检查；C4 修复后 0 `service failed to listen`、进入测试执行（6/6 栈启动）。✓

**openspec validate**：`Change 'fix-integration-stack-readiness' is valid`。
