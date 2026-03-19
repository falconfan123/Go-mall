# Jaeger 链路追踪集成

本项目使用 OpenTelemetry 和 Jaeger 实现分布式链路追踪。

## 架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Go-mall 链路追踪架构                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐   │
│   │  Order   │───▶│ Checkout │───▶│ Inventory│───▶│ Payment  │   │
│   │  Service │    │  Service │    │  Service │    │  Service │   │
│   └────┬─────┘    └────┬─────┘    └────┬─────┘    └────┬─────┘   │
│        │                │                │                │         │
│        └────────────────┴────────────────┴────────────────┘         │
│                                │                                     │
│                                ▼                                     │
│                    ┌───────────────────────┐                        │
│                    │  OpenTelemetry SDK    │                        │
│                    │  (每个服务内置)         │                        │
│                    └───────────┬───────────┘                        │
│                                │                                     │
│                                ▼                                     │
│   ┌────────────────────────────────────────────────────────────┐   │
│   │                    OTLP 协议 (gRPC/HTTP)                    │   │
│   └────────────────────────────────────────────────────────────┘   │
│                                │                                     │
│                                ▼                                     │
│   ┌─────────────┐    ┌─────────────────────────────────────────┐  │
│   │   Jaeger    │◀───│  OpenTelemetry Collector (可选)          │  │
│   │   Agent     │    │  (聚合和转发追踪数据)                    │  │
│   │ :14250      │    └─────────────────────────────────────────┘  │
│   └──────┬──────┘                                                 │
│          │                                                         │
│          ▼                                                         │
│   ┌─────────────────────────────────────────┐                      │
│   │            Jaeger UI                    │                      │
│   │         http://localhost:16686          │                      │
│   └─────────────────────────────────────────┘                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## 快速开始

### 1. 启动 Jaeger

项目已配置 Docker Compose，启动观测服务：

```bash
cd construct/observability
docker-compose up -d jaeger
```

### 2. 访问 Jaeger UI

打开浏览器访问: http://localhost:16686

### 3. 配置服务追踪

各服务的配置文件 (`etc/*.yaml`) 中已包含 Telemetry 配置：

```yaml
# 链路追踪
Telemetry:
  Name: order.rpc
  Endpoint: http://localhost:14268/api/traces
  Batcher: jaeger
  Sampler: 1.0
```

## 代码集成

### 1. 初始化追踪

在服务启动时初始化 Jaeger：

```go
import "github.com/falconfan123/Go-mall/common/utils/tracing"

// 在 main.go 或初始化函数中
func main() {
    // 初始化追踪
    shutdown, err := tracing.InitJaeger(&tracing.Config{
        ServiceName: "order.rpc",
        JaegerEndpoint: "localhost:14250",
        SampleRate: 1.0,
        Enabled: true,
    })
    if err != nil {
        log.Error(err)
    }
    if shutdown != nil {
        defer shutdown(context.Background())
    }

    // ... 其他服务启动代码
}
```

### 2. 在业务代码中使用

```go
import (
    "github.com/falconfan123/Go-mall/common/utils/tracing"
    "go.opentelemetry.io/otel/attribute"
)

// 在业务逻辑中
func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
    // 开始追踪
    ctx, span := tracing.StartSpan(l.ctx, "CreateOrder", "create_order",
        attribute.Int64("user_id", req.UserId),
    )
    defer span.End()

    // 业务逻辑
    // ...

    // 记录错误
    if err != nil {
        tracing.RecordError(span, err)
        return nil, err
    }

    return result, nil
}
```

### 3. gRPC 拦截器

项目已集成 go-zero 的 gRPC 链路追踪支持。

### 4. HTTP 中间件

对于 HTTP 服务，可以使用 otelhttp 中间件：

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

handler := otelhttp.NewHandler(httpHandler, "httpHandler")
```

## 配置说明

### Telemetry 配置

| 字段 | 说明 | 示例 |
|------|------|------|
| Name | 服务名称 | order.rpc |
| Endpoint | 收集器地址 | http://localhost:14268/api/traces |
| Batcher | 批处理类型 | jaeger / zipkin |
| Sampler | 采样率 (0-1) | 1.0 (100%) |

### 采样策略

- `1.0`: 全量采样
- `0.1`: 10% 采样
- `0.01`: 1% 采样

生产环境建议使用 `0.1` 以减少开销。

## 查看追踪

### 1. 访问 Jaeger UI

http://localhost:16686

### 2. 选择服务

在左侧下拉框中选择要查看的服务。

### 3. 搜索追踪

点击 "Find Traces" 按钮查看追踪结果。

### 4. 查看详情

点击任意追踪条目，查看详细的调用链和时间信息。

## 端口说明

| 端口 | 协议 | 说明 |
|------|------|------|
| 14250 | gRPC | Jaeger Agent (推荐) |
| 14268 | HTTP | Jaeger Collector |
| 16686 | HTTP | Jaeger UI |
| 9411 | HTTP | Zipkin 兼容接口 |

## 常见问题

### Q: 看不到追踪数据？

1. 确认 Jaeger 服务已启动: `docker ps | grep jaeger`
2. 检查服务配置中的 Endpoint 是否正确
3. 检查采样率设置

### Q: 如何关闭追踪？

将配置文件中的 `Enabled` 设为 `false`，或设置 `Sampler: 0.0`。

## 相关文档

- [Jaeger 官方文档](https://www.jaegertracing.io/)
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [go-zero 链路追踪](https://go-zero.dev/docs/feature/observability/tracing)
