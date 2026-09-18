## 审查轮次 1 — 2026-09-17（冻结）

**审核方式**：@openspec-reviewer 对完整变更包重审（备选通道脚本）。

### 修复循环（同一门禁内）
- 首轮：🔴 2（responseRecorder Header 处理缺失：Content-Length 需删除 + Content-Type 强制 JSON；trace_id 来源/格式未约束）→ design D2 补 Header 处理 / json.Marshal sanitize / 空 body / WriteHeader 幂等 / trace_id 32hex 自造；spec 补空 body 场景 + trace_id 约束。
- 末轮：**🔴 无 → 冻结**。

### 🟡 记录（本轮吸收/后续）
- 已吸收：status_msg 一律 json.Marshal（禁字符串拼接）；Content-Length 删除重算；空 body 兜底；WriteHeader 幂等；trace_id 32hex 自造。
- 后续：gateway_smoke 是否纳入 CI 门禁（Integration 已跑 test/rpc 全包，覆盖）；Flusher/Hijacker 接口兼容在实现时核验；Zipkin trace 衔接另案。

### 结论
**🔴 清零 → 冻结**，可 apply（单门禁）。

## Verify — 2026-09-18（归档前）

一致性四项 + openspec validate：
1. **proposal→tasks 可追溯**：What Changes ①（错误包装）/ ②（HTTP 语义）/ ③（不吞错误）→ tasks 1.x（responseRecorder/中间件/writeAuthError）/ 2.x（curl/复测/错误路径测试）；spec 两 Requirement 对应。✓
2. **tasks 全勾且有证据**：10/10 `- [x]`；证据 = PR #88（76062d3）+ 上轮 apply 输出（见下）。
3. **无越界改动**：`git show --stat 76062d3` = services/gateway/gateway.go + test/rpc/scenarios/gateway_smoke/gateway_smoke_test.go + openspec；白名单内。✓
4. **基线成立**：本变更无独立 explore-brief；基线取自 `fix-ci-flaky-product-gateway-tests` 2C（原始 body = rpc error 纯文本），实现与基线一致（结构化 JSON 包装）。

**验收证据（引用已有输出，不重跑）**：
- 上游 RPC 错误（act_start_limit 已删）：`{"status_code":500,"status_msg":"rpc error: code = Unknown desc = activity not found or not started","trace_id":"63ae103790d1cc2c36e333b578a9b716"}` HTTP 500（结构化 JSON）。
- 401 认证错误：`{"status_code":10000,"status_msg":"认证信息为空"}` HTTP 401（合法 JSON，不二次包装）。
- 新增 `TestGatewayErrorResponseStructured` PASS；`TestGatewayHTTPHappyPath` 5/5 PASS；build/lint/test-unit 绿。

**openspec validate**：`Change 'fix-gateway-error-response-format' is valid`。
