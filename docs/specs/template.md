# [Feature Name] — 设计规范 (SPEC)

> 在写任何代码之前，先写这份 SPEC。不清晰的设计建在沙子上。

## 1. 目标

- 解决什么问题？
- 用户场景是什么？
- 预期效果是什么？

## 2. 范围

- **包含：** 明确列出所有功能点
- **不包含（明确排除）：** 列明本期不做什么
- **依赖：** 本功能依赖哪些已有服务/模块

## 3. 涉及的服务

| 服务 | 修改类型 | 说明 |
|------|---------|------|
| auths | 新增/修改/不变 | xxx |
| orders | 新增/修改/不变 | xxx |
| ... | ... | ... |

## 4. 接口定义

### REST API（Gateway）

```
POST /api/v1/xxx
GET  /api/v1/xxx/{id}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 订单 ID |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### gRPC 接口（服务间）

```protobuf
service XxxService {
  rpc CreateXxx(CreateXxxRequest) returns (CreateXxxResponse);
}
```

## 5. 数据模型

### 数据库表

```sql
CREATE TABLE xxx (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  ...
);
```

### ES 索引（如需要）

### MQ 消息结构（如需要）

## 6. 错误码

| 错误码 | 含义 | 处理方式 |
|--------|------|---------|
| 10001 | 参数错误 | 返回前端提示 |
| ... | ... | ... |

## 7. 边界条件

- **正常流程：** 最常用路径
- **异常流程：**
  - 网络超时 → 重试？
  - 数据不存在 → 返回 404？
  - 并发冲突 → 乐观锁？
- **极端情况：**
  - QPS 高峰 → 限流？
  - 数据量级 → 分页？
  - 幂等性 → 防重复提交？

## 8. "完成"的定义（AI 必须全部验证）

- [ ] 编译通过（`go build ./...`）
- [ ] 单元测试通过（`make test-unit`）
- [ ] Lint 通过（`make lint`）
- [ ] 覆盖率不低于 63.6%
- [ ] Mock 已更新（`make mock`）
- [ ] API 文档同步更新
- [ ] Gatekeeper 通过（`make gatekeeper`）
