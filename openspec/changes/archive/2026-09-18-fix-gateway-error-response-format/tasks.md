> 变更B：网关错误响应结构化。白名单 = services/gateway/ + test/rpc/scenarios/gateway_smoke/（越界即停）。
> 口径：上游错误（>=400 且 body 非 JSON）→ 结构化 JSON；已 JSON 错误/2xx 透传。

## 1. 实现（D1-A / D2 / D3）

- [x] 1.1 `services/gateway/gateway.go` 新增 `responseRecorder`（缓冲 status + body，实现 `http.ResponseWriter`）；验证：`go build` 编译通过
- [x] 1.2 新增错误包装中间件（`gw.Use`）：status `>=400` 且 `!json.Valid(body)` → 输出 `{"status_code","status_msg","trace_id"}`（Content-Type JSON，状态码不变）；已 JSON 错误与 `<400` 透传；验证：代码 + 单测（三路径：纯文本错误包装 / JSON 错误透传 / 2xx 透传）
- [x] 1.3 `trace_id` 请求级生成（时间 + 随机，32 位十六进制）；`status_msg` 经 `json.Marshal` 序列化（禁拼接）+ 长度上限 1024；**修正既有 `writeAuthError` 用 `json.Marshal`/`json.NewEncoder` 生成**；验证：代码 + 单测（含错误文本带引号/换行的场景）

## 2. 验证（D4）

- [x] 2.1 curl 复现：`curl -sS ".../api/v1/activity/token?activity_id=<id>" -H "Short-Token: $ST" -H "Long-Token: $LT"` 触发上游错误 → 响应体为 JSON 且含 status_code/status_msg；另 curl 验证 401 认证失败响应为合法 JSON 且不被二次包装；验证：原始 curl 输出
- [x] 2.2 复测 `cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=5 -run TestGatewayHTTPHappyPath ./scenarios/gateway_smoke/...`（错误路径已 JSON 化，不再因非 JSON body 崩）；验证：输出
- [x] 2.2b **新增自动化错误路径测试**（`gateway_smoke` 内）：构造上游 RPC 失败场景，断言响应 `json.Valid` 且含 status_code/status_msg/trace_id（CI 回归拦截）；验证：新用例 + 输出
- [x] 2.3 Rule 1 门禁：`make build && make lint && make test-unit` 全绿；验证：命令输出
- [x] 2.4 对照 spec 各 Scenario 逐项核对；验证：验收记录

## 3. 交付

- [x] 3.1 红线：`git diff --name-only` 白名单 = services/gateway/ + test/rpc/scenarios/gateway_smoke/；验证：命令输出
- [x] 3.2 提 PR → Build + Quality 绿 → 合并；验证：PR 号 + merge commit hash