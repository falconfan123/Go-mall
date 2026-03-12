# Go-Mall 系统架构图

## 系统整体架构

```mermaid
graph TB
    User["用户
    👤"] --> Frontend["前端
    🎨 (Vue/React)
    :3000"]

    Frontend --> API_Gateway["API 网关层
    Nginx Ingress
    (K8s)"]

    subgraph "API 层 (HTTP)"
        User_API["用户 API
        📱 :8001"]
        Product_API["商品 API
        📦 :8002"]
        Carts_API["购物车 API
        🛒 :8003"]
        Order_API["订单 API
        📋 :8004"]
        Checkout_API["结算 API
        💳 :8005"]
        Payment_API["支付 API
        💰 :8006"]
        Coupon_API["优惠券 API
        🎟️ :8009"]
    end

    subgraph "微服务层 (gRPC)"
        Auths_Service["认证服务
        🔐 :10000"]
        Users_Service["用户服务
        👥 :10001"]
        Product_Service["商品服务
        🏪 :10002"]
        Carts_Service["购物车服务
        🛍️ :10003"]
        Order_Service["订单服务
        📝 :10004"]
        Checkout_Service["结算服务
        🧾 :10005"]
        Payment_Service["支付服务
        💵 :10006"]
        Inventory_Service["库存服务
        📊 :10011"]
        Coupons_Service["优惠券服务
        🎫 :10009"]
        Audit_Service["审计服务
        📈 :10008"]
    end

    subgraph "基础设施层"
        Consul["服务发现
        Consul
        :8500"]
        MySQL[(MySQL 8.0
        🗄️ :3306)]
        Redis[(Redis 6.0
        🚀 :6379)]
        RabbitMQ["RabbitMQ
        🐰 :5672/:15672"]
        Elasticsearch[(Elasticsearch 8.x
        🔍 :9200)]
        DTM["分布式事务
        DTM
        :36789/:36790"]
    end

    subgraph "监控运维层"
        Prometheus["Prometheus
        📉"]
        Grafana["Grafana
        📊"]
        EFK["EFK Stack
        📝 (Elasticsearch+Fluentd+Kibana)"]
    end

    API_Gateway --> User_API
    API_Gateway --> Product_API
    API_Gateway --> Carts_API
    API_Gateway --> Order_API
    API_Gateway --> Checkout_API
    API_Gateway --> Payment_API
    API_Gateway --> Coupon_API

    User_API --> Auths_Service
    User_API --> Users_Service
    Product_API --> Product_Service
    Carts_API --> Carts_Service
    Order_API --> Order_Service
    Checkout_API --> Checkout_Service
    Payment_API --> Payment_Service
    Coupon_API --> Coupons_Service

    Auths_Service --> Consul
    Users_Service --> Consul
    Product_Service --> Consul
    Carts_Service --> Consul
    Order_Service --> Consul
    Checkout_Service --> Consul
    Payment_Service --> Consul
    Inventory_Service --> Consul
    Coupons_Service --> Consul
    Audit_Service --> Consul

    Auths_Service --> MySQL
    Auths_Service --> Redis
    Users_Service --> MySQL
    Users_Service --> Redis
    Product_Service --> MySQL
    Product_Service --> Redis
    Product_Service --> Elasticsearch
    Carts_Service --> Redis
    Carts_Service --> MySQL
    Order_Service --> MySQL
    Order_Service --> Redis
    Checkout_Service --> DTM
    Checkout_Service --> RabbitMQ
    Payment_Service --> MySQL
    Payment_Service --> RabbitMQ
    Inventory_Service --> Redis
    Inventory_Service --> MySQL
    Coupons_Service --> MySQL
    Coupons_Service --> Redis
    Audit_Service --> MySQL

    Checkout_Service --> Order_Service
    Checkout_Service --> Inventory_Service
    Checkout_Service --> Carts_Service
    Checkout_Service --> Coupons_Service
    Order_Service --> Product_Service
    Order_Service --> Inventory_Service
    Payment_Service --> Order_Service

    Prometheus --> Consul
    Prometheus --> Grafana
    EFK --> Elasticsearch
```

## 分层架构详解

### 1. 接入层
- **前端应用**: 用户界面，运行在 :3000
- **API 网关**: Nginx Ingress (K8s 环境)

### 2. API 层 (HTTP)
| 服务 | 端口 | 功能 |
|------|------|------|
| user-api | 8001 | 用户认证、信息管理、地址管理 |
| product-api | 8002 | 商品查询、分类、搜索 |
| carts-api | 8003 | 购物车增删改查 |
| order-api | 8004 | 订单创建、查询、取消 |
| checkout-api | 8005 | 订单结算、价格计算 |
| payment-api | 8006 | 支付、退款 |
| coupon-api | 8009 | 优惠券领取、使用 |

### 3. 微服务层 (gRPC)
| 服务 | RPC 端口 | 监控端口 | 功能 |
|------|----------|----------|------|
| auths | 10000 | 11000 | JWT 认证、Token 管理 |
| users | 10001 | 11001 | 用户信息、收货地址 |
| product | 10002 | 11002 | 商品 SKU、分类、库存信息 |
| carts | 10003 | 11003 | 购物车数据 |
| order | 10004 | 11004 | 订单、订单项 |
| checkout | 10005 | 11005 | 结算、价格计算 |
| payment | 10006 | 11006 | 支付流水、回调处理 |
| inventory | 10011 | 11011 | 库存扣减、回滚 |
| coupons | 10009 | 11009 | 优惠券规则、用户券 |
| audit | 10008 | 11008 | 操作审计日志 |

### 4. 基础设施层
| 组件 | 端口 | 用途 |
|------|------|------|
| Consul | 8500 | 服务注册与发现 |
| MySQL | 3306 | 主数据库 |
| Redis | 6379 | 缓存、购物车、分布式锁 |
| RabbitMQ | 5672/15672 | 异步消息、订单事件 |
| Elasticsearch | 9200/9300 | 商品搜索 |
| DTM | 36789/36790 | 分布式事务协调 |

### 5. 监控运维层
- **Prometheus**: 指标采集
- **Grafana**: 可视化监控
- **EFK Stack**: 日志收集与分析

## 核心业务流程

### 下单流程
```mermaid
sequenceDiagram
    participant User as 用户
    participant FE as 前端
    participant CAPI as 购物车API
    participant CHAPI as 结算API
    participant OAPI as 订单API
    participant PAPI as 支付API
    participant CART as 购物车服务
    participant CHECKOUT as 结算服务
    participant ORDER as 订单服务
    participant INVENTORY as 库存服务
    participant COUPON as 优惠券服务
    participant PAYMENT as 支付服务
    participant DTM as DTM
    participant MySQL as MySQL
    participant Redis as Redis
    participant RMQ as RabbitMQ

    User->>FE: 浏览购物车
    FE->>CAPI: 获取购物车列表
    CAPI->>CART: CartItemList
    CART->>Redis: 读取购物车
    CART-->>CAPI: 返回购物车数据
    CAPI-->>FE: 返回

    User->>FE: 点击结算
    FE->>CHAPI: Checkout
    CHAPI->>CHECKOUT: Checkout
    CHECKOUT->>DTM: 开启SAGA事务
    CHECKOUT->>COUPON: 验证优惠券
    CHECKOUT->>INVENTORY: 预扣库存(Try)
    INVENTORY->>Redis: Lua脚本原子扣减
    CHECKOUT->>ORDER: 创建订单
    ORDER->>MySQL: 插入订单
    CHECKOUT->>CART: 清空购物车
    CHECKOUT-->>CHAPI: 返回订单ID
    CHAPI-->>FE: 返回订单ID

    User->>FE: 选择支付方式
    FE->>PAPI: CreatePayment
    PAPI->>PAYMENT: CreatePayment
    PAYMENT->>MySQL: 插入支付流水
    PAYMENT->>RMQ: 发送支付事件
    PAYMENT-->>PAPI: 返回支付参数
    PAPI-->>FE: 返回

    FE->>User: 唤起第三方支付
    User->>FE: 支付完成回调
    FE->>PAPI: 支付回调通知
    PAPI->>PAYMENT: 确认支付成功
    PAYMENT->>ORDER: 更新订单状态
    ORDER->>RMQ: 发送订单创建事件
    ORDER->>DTM: 确认(Confirm)
    ORDER-->>PAYMENT: 完成
```

## 技术栈

### 开发框架
- **Go 1.20+**
- **Go-Zero 1.7.6**: 微服务框架
- **gRPC**: 服务间通信
- **JWT**: 身份认证

### 数据存储
- **MySQL 8.0**: 关系型数据库
- **Redis 6.0**: 缓存、会话、购物车、分布式锁
- **Elasticsearch 8.x**: 商品搜索引擎

### 消息与事务
- **RabbitMQ**: 异步消息队列
- **DTM**: 分布式事务 (TCC/SAGA模式)

### 服务治理
- **Consul**: 服务注册与发现

### 容器化与部署
- **Docker**: 容器化
- **Kubernetes (K8s)**: 容器编排
- **ArgoCD**: GitOps 持续部署
- **GitHub Actions**: CI/CD

### 监控与可观测性
- **Prometheus**: 指标采集
- **Grafana**: 监控可视化
- **OpenTelemetry**: 链路追踪
- **EFK Stack**: 日志管理 (Elasticsearch + Fluentd + Kibana)

## 项目目录结构
```
go-mall/
├── apis/              # API 层 (HTTP)
│   ├── user/
│   ├── product/
│   ├── carts/
│   ├── order/
│   ├── checkout/
│   ├── payment/
│   └── coupon/
├── services/          # 微服务层 (gRPC)
│   ├── auths/
│   ├── users/
│   ├── product/
│   ├── carts/
│   ├── order/
│   ├── checkout/
│   ├── payment/
│   ├── inventory/
│   ├── coupons/
│   └── audit/
├── common/            # 公共模块
│   ├── config/
│   ├── consts/
│   ├── middleware/
│   ├── response/
│   └── utils/
├── dal/               # 数据访问层
├── cmd/               # 命令行工具
├── construct/         # 基础设施配置
├── manifests/         # K8s 部署清单
├── frontend/          # 前端代码
├── scripts/           # 脚本工具
├── test/              # 测试
├── run.go             # 本地服务启动器
└── docker-compose.yml # Docker 编排
```
