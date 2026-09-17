## Context

见 `proposal.md`（Why）与 explore-brief（2C 取证）。核心事实：
- go-zero gateway `buildGrpcHandler`（`gateway/server.go:220-222`）对上游响应 `w.WriteHeader(status); io.Copy(w, body)`——gRPC 错误 body（`rpc error: ...`）原样透出，非 JSON。
- 测试 `gateway_smoke_test.go:438` `json.Unmarshal` 因 `invalid character 'r'` 崩溃（rogue redis 隔离后该路径仍可被任意上游 RPC 错误触发）。
- 既有 `writeAuthError`（`gateway.go:133`）已输出 JSON（401）。

## Goals / Non-Goals

- **Goals**：网关对上游错误统一输出结构化 JSON（status_code/status_msg/trace_id）；HTTP 状态码与错误语义保留；已 JSON 错误不二次包装。
- **Non-Goals**：不改上游 RPC 逻辑；不改非错误路径；不改门禁。

## Decisions

### D1 方案对比

| 方案 | 做法 | 评价 |
|---|---|---|
| **A 响应包装中间件（选定）** | `gw.Use` 包装 ResponseWriter：缓冲 body；状态 `>=400` 且 body 非合法 JSON → 替换为结构化 JSON；已 JSON 错误透传 | 不改上游、不改 go-zero 依赖；统一出口；对既有 JSON 错误（writeAuthError）零影响 |
| B 改 go-zero gateway 源码/fork | 改依赖库 | 侵入依赖、升级冲突；否决 |
| C 上游服务自行 JSON 化错误 | 改 activity 等所有 RPC 返回 | 分散、遗漏面大；仅能解决单点；否决 |

- **选定 A**。

### D2 包装规则

- 状态 `>=400` 且 body **非合法 JSON**（`json.Valid` 判 false）→ 输出 `{"status_code":<status>,"status_msg":"<sanitized>","trace_id":"<tid>"}`。
- 状态 `>=400` 且 body **已是合法 JSON**（既有 `writeAuthError`）→ 原样透出（不二次包装）。
- 状态 `<400` → 原样透出。
- **Header 处理**：包装写入前 MUST 删除原 `Content-Length`（由标准库按新 body 重新计算/chunked），MUST 强制 `Content-Type: application/json`（覆盖上游非 JSON Content-Type）。
- **sanitize**：`status_msg` 一律经 `json.Marshal` 序列化字符串字段，**禁止字符串拼接**（防 `"`/换行破坏 JSON）；trim + 去控制字符；长度上限 1024 字符（防超大错误文本）。
- **空 body**：状态 `>=400` 且 body 为空/仅空白 → 仍包装，`status_msg` 用 `http.StatusText(status)` 兜底。
- **WriteHeader 幂等**：`responseRecorder` 只记录首次状态（多次 `WriteHeader` 忽略后续）。
- **trace_id**：请求级自造 32 位十六进制（时间戳 + 随机），不依赖 tracing context（避免未初始化时为空）；与 Zipkin 衔接为后续项。

### D3 实现要点

- `responseRecorder`：实现 `http.ResponseWriter`，缓冲 status + body；`next(rec, r)` 后统一 flush（包装或透传）。
- **修正既有 `writeAuthError`（`gateway.go:133`）**：现用 `fmt.Fprintf` 字符串拼接（未转义 `statusMsg`），若错误文本含 `"`/`\`/换行将产出非法 JSON，进而被新中间件二次包装、破坏认证错误契约。改为 `json.Marshal`/`json.NewEncoder` 生成，保证输出始终合法 JSON（"已 JSON 错误不二次包装"才可靠）。
- 中间件注册：在既有 CORS/auth 中间件之后，作用于所有下游响应；实现后验证 `responseRecorder` 能拦截 `writeAuthError` 输出（避免 auth 中间件直接写底层 writer）。
- 接口兼容：检查 go-zero gateway handler 路径不依赖 `http.Flusher`/`http.Hijacker`（网关响应为 JSON 小 body，无流式）；若实现遇类型断言需求再补。
- 不实现 `http.Pusher`。

### D4 验证

- 复测 `TestGatewayHTTPHappyPath`：任意上游 RPC 错误不再产生纯文本 body；断言错误响应 `json.Valid`。
- curl 复现 `activity/token`（触发 RPC 错误）：响应体 MUST 为 JSON 且含 status_code/status_msg。
- `make build && make lint && make test-unit` 全绿。

## Risks / Trade-offs

- **[包装改变 4xx/5xx 形状]** → 已 JSON 错误透传（D2），2xx 不动；仅影响"纯文本错误 body"（正是缺陷本体）。
- **[缓冲引入内存开销]** → 网关响应为小 JSON，bounded；无流式依赖。
- **[trace_id 无链路上下文]** → 仅请求级唯一标识；完整 trace 与 observability 基线衔接（可后续扩展）。

## Migration Plan

1. `gateway.go` 新增 responseRecorder + 错误包装中间件。
2. `gw.Use` 注册。
3. curl 复现验证 + 复测 gateway_smoke。
4. 门禁 + 提 PR。

**回滚**：仅 `services/gateway/gateway.go`，git revert 即回。