# gateway-error-response Specification

## Purpose

API 网关（`services/gateway`）错误响应契约：上游 gRPC 错误 MUST 以结构化 JSON 返回（`status_code`/`status_msg`/`trace_id`），MUST NOT 透出纯文本 `rpc error:...` 非 JSON body；已 JSON 错误与正常响应不受影响。

## Requirements

### Requirement: 上游错误响应结构化

当上游 gRPC 调用失败（HTTP 状态 `>= 400`）时，网关响应体 MUST 为合法 JSON，包含 `status_code`（HTTP 状态码）、`status_msg`（原始错误语义，经 `json.Marshal` 序列化，长度 ≤1024 字符，不吞信息、不破坏 JSON）、`trace_id`（请求级唯一标识，32 位十六进制自造，不依赖 tracing context）；MUST 删除原 `Content-Length` 并由标准库重算、强制 `Content-Type: application/json`；MUST NOT 返回 `rpc error: code = ...` 之类的纯文本 body。既有 `writeAuthError` MUST 改为 `json.Marshal`/`json.NewEncoder` 生成（不得字符串拼接），确保其输出始终合法 JSON（不触发二次包装）。

#### Scenario: 上游 RPC 错误返回结构化 JSON
- **WHEN** 任一上游 RPC 返回错误（如 `activity/Token` 返回 `activity not found or not started`）
- **THEN** 网关响应 MUST 为合法 JSON（`json.Valid`），含 `status_code`/`status_msg`/`trace_id`，HTTP 状态码保留，`Content-Type: application/json`，`Content-Length` 与新 body 一致；调用方 `json.Unmarshal` MUST 不报 `invalid character` 类错误

#### Scenario: 空 body 错误仍有结构化响应
- **WHEN** 上游返回 `>=400` 且 body 为空或仅空白
- **THEN** 仍 MUST 输出合法 JSON，`status_msg` 用 `http.StatusText(status)` 兜底

### Requirement: 不二次包装既有 JSON 错误

网关已返回合法 JSON 的错误响应（如认证错误 `writeAuthError`）MUST 原样透出，MUST NOT 被再次包装为嵌套 JSON；正常 `2xx` 响应 MUST 不受影响。

#### Scenario: JSON 错误与 2xx 原样透出
- **WHEN** 网关处理 ① 已 JSON 的错误响应（401 认证失败）② 正常 `2xx` 响应
- **THEN** 两者 MUST 原样透出（body 不变、状态码不变），不出现二次包装