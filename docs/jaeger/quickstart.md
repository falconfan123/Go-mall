# Jaeger 快速开始指南

本指南将帮助你在项目中快速集成 Jaeger 链路追踪。

## 前置条件

1. Docker 和 Docker Compose 已安装
2. 项目依赖已安装

## 步骤 1: 启动 Jaeger

```bash
cd construct/observability
docker-compose up -d jaeger
```

验证 Jaeger 启动成功：

```bash
docker ps | grep jaeger
```

## 步骤 2: 在服务中启用追踪

### 2.1 添加依赖

在服务的 `go.mod` 中添加：

```go
require (
    github.com/falconfan123/Go-mall/common latest
)
```

### 2.2 配置 Telemetry

编辑服务的配置文件 (`etc/*.yaml`)：

```yaml
# 链路追踪
Telemetry:
  Name: order.rpc
  Endpoint: http://localhost:14268/api/traces
  Batcher: jaeger
  Sampler: 1.0
```

### 2.3 初始化追踪

在服务的 `main.go` 或 `svc/servicecontext.go` 中初始化：

```go
package main

import (
    "context"
    "log"

    "github.com/falconfan123/Go-mall/common/utils/tracing"
)

func main() {
    // 初始化 Jaeger 追踪
    shutdown, err := tracing.InitJaeger(&tracing.Config{
        ServiceName: "order.rpc",
        JaegerEndpoint: "localhost:14250",
        SampleRate: 1.0,
        Enabled: true,
    })
    if err != nil {
        log.Printf("failed to init jaeger: %v", err)
    }
    if shutdown != nil {
        defer shutdown(context.Background())
    }

    // 启动服务
    // ...
}
```

## 步骤 3: 在业务代码中使用追踪

### 3.1 简单示例

```go
import (
    "github.com/falconfan123/Go-mall/common/utils/tracing"
    "go.opentelemetry.io/otel/attribute"
)

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
    ctx, span := tracing.StartSpan(l.ctx, "CreateOrder", "create_order",
        attribute.Int64("user_id", req.UserId),
        attribute.String("order_type", "normal"),
    )
    defer span.End()

    // 业务逻辑
    order, err := l.svcCtx.OrderModel.CreateOrder(ctx, req)
    if err != nil {
        tracing.RecordError(span, err)
        return nil, err
    }

    return order, nil
}
```

### 3.2 追踪数据库操作

```go
import (
    "github.com/falconfan123/Go-mall/common/utils/tracing"
    "go.opentelemetry.io/otel/attribute"
)

func (l *OrderRepo) CreateOrder(ctx context.Context, order *Order) error {
    ctx, span := tracing.StartSpan(ctx, "OrderRepo.CreateOrder", "db_write",
        attribute.String("table", "orders"),
    )
    defer span.End()

    _, err := l.db.Insert(order)
    if err != nil {
        tracing.RecordError(span, err)
    }
    return err
}
```

### 3.3 追踪 RPC 调用

```go
import (
    "github.com/falconfan123/Go-mall/common/utils/tracing"
    "go.opentelemetry.io/otel/attribute"
)

func (l *CheckoutLogic) CallPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    ctx, span := tracing.StartSpan(ctx, "Checkout.CallPayment", "rpc_call",
        attribute.String("rpc_service", "payment.rpc"),
    )
    defer span.End()

    resp, err := l.paymentClient.Pay(ctx, req)
    if err != nil {
        tracing.RecordError(span, err)
    }
    return resp, err
}
```

## 步骤 4: 查看追踪结果

1. 打开浏览器访问 http://localhost:16686
2. 在左侧服务下拉框中选择 `order.rpc`
3. 点击 "Find Traces" 按钮
4. 点击任意追踪查看详细信息

## 验证

启动一个服务并发起请求，然后刷新 Jaeger UI 查看追踪数据。

## 下一步

- 调整采样率以优化性能
- 添加更多自定义属性
- 集成 OpenTelemetry Collector
