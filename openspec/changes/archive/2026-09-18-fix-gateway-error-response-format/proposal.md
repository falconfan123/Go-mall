## Why

API 网关（`services/gateway`，go-zero gateway）在**上游 gRPC 调用失败时向调用方透出非结构化纯文本响应体**，构成 API 契约缺陷（即使 CI 不再复现该测试，缺陷本身成立）：

- **现象（实测 curl 复现）**：
  ```
  $ curl -sS "http://localhost:8888/api/v1/activity/token?activity_id=140" \
      -H "Short-Token: $ST" -H "Long-Token: $LT" -H 'X-Real-IP: 127.0.0.1'
  rpc error: code = Unknown desc = activity not found or not started     ← HTTP 500，纯文本，非 JSON
  ```
  根因：go-zero gateway `server.go:220-222` 对上游响应直接 `w.WriteHeader(resp.StatusCode); io.Copy(w, resp.Body)`，把 gRPC 错误 body（`rpc error: ...`）原样透出；`Content-Type: application/json` 已设但 body 非 JSON。
- **影响**：① 调用方按 JSON 解析（`json.Unmarshal`）直接报 `invalid character 'r'`（测试 438 行崩溃源）；② 错误语义丢失（只有 gRPC 原始文本，无结构化 code/message）；③ 无法携带 trace_id 供排障。

## What Changes

- **① 网关统一错误包装**：`services/gateway/gateway.go` 增加响应包装中间件——当上游响应状态 `>= 400` 且 body **非合法 JSON** 时，包装为结构化 JSON：`{"status_code":<http 状态>,"status_msg":"<错误正文>","trace_id":"<请求级 trace>"}`；body 已是合法 JSON（如既有 `writeAuthError`）则原样透出（不二次包装）；`< 400` 原样透出。
- **② 保留 HTTP 语义**：状态码不变；`Content-Type: application/json`。
- **③ 不吞错误**：`status_msg` 保留原始错误信息（sanitize 换行/引号），不丢失排障信息。

## Capabilities

### New Capabilities
- `gateway-error-response`: 网关错误响应契约——上游 RPC 错误 MUST 以结构化 JSON 返回（status_code/status_msg/trace_id），MUST NOT 透出纯文本 `rpc error:...` 非 JSON body。

### Modified Capabilities
<!-- 无。 -->

## Impact

- **服务代码**：`services/gateway/gateway.go`（响应包装中间件）。
- **测试**：`test/rpc/scenarios/gateway_smoke/`（复测 + 断言响应体为 JSON）。
- **风险**：包装中间件改变所有 4xx/5xx 响应体形状——既有已返回 JSON 的错误（`writeAuthError`）不受影响（检测 JSON 透传）；正常 2xx 不受影响。

## Non-Goals

- 不改上游服务（activity 等）的 RPC 返回逻辑（错误本体不变，只统一网关出口格式）。
- 不改 ruleset / 门禁 / Integration required。
- 不修 gateway 的其它非错误路径。