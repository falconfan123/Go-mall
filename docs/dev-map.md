# Go-mall 开发导航图 (Dev-Map)

> 帮助 AI 快速理解项目结构、架构决策和关键路径。
> 在开始任何开发前阅读此文档。

## 项目概览

Go-mall 是一个基于 **go-zero** 框架的电商微服务系统，采用 **go.work** workspace 管理多模块。

### 目录结构

```
go-mall/
├── cmd/server/          # 应用入口（HTTP 服务启动）
├── common/              # 公共库
│   ├── config/          # 配置管理（yaml 加载）
│   ├── consts/          # 常量（错误码、业务常量）
│   ├── middleware/      # HTTP 中间件（CORS, 限流, 客户端 IP）
│   ├── response/        # 统一响应格式
│   ├── types/           # 公共类型定义
│   └── utils/           # 工具库（加密、JWT、幂等、追踪）
├── dal/
│   ├── model/           # 数据模型层（MySQL, gorm 风格）
│   └── es/              # Elasticsearch 数据访问
├── services/            # 所有微服务
│   ├── auths/           # 认证 — 登录/注册/Token
│   ├── users/           # 用户信息管理
│   ├── products/        # 商品管理
│   ├── carts/           # 购物车
│   ├── checkout/        # 结算（分布式事务入口）
│   ├── orders/          # 订单
│   ├── payment/         # 支付
│   ├── inventory/       # 库存
│   ├── coupons/         # 优惠券
│   ├── audit/           # 审计日志（DDD 架构）
│   ├── activity/        # 秒杀活动
│   ├── admin/           # 管理后台（DDD 架构）
│   ├── search/          # 搜索服务
│   ├── gateway/         # API 网关
│   ├── productcatalogservice/  # 商品目录服务
│   ├── system/          # 系统配置
│   └── etc/             # etcd 配置
├── frontend/            # React + Vite + TailwindCSS
├── k8s/                 # Kubernetes 部署（Helm charts）
├── manifests/           # Docker Compose 配置
├── infrastructure/      # 监控（Grafana, Loki, Promtail）
├── scripts/             # 工具脚本
├── configs/             # 环境配置
├── test/                # 集成测试
├── postman/             # Postman 集合
└── .github/workflows/   # CI/CD 流水线
```

## 微服务一览

| 服务 | 协议 | 数据库 | 说明 |
|------|------|--------|------|
| auths | gRPC | MySQL | 认证（JWT token） |
| users | gRPC | MySQL | 用户 CRUD |
| products | gRPC | MySQL + ES | 商品信息 |
| carts | gRPC | MySQL | 购物车 |
| checkout | gRPC | MySQL + DTM | 结算（分布式事务入口） |
| orders | gRPC | MySQL | 订单 |
| payment | gRPC | MySQL | 支付 |
| inventory | gRPC | MySQL | 库存 |
| coupons | gRPC | MySQL | 优惠券 |
| audit | gRPC | MySQL + ES | 审计（DDD） |
| activity | gRPC | MySQL | 秒杀活动 |
| admin | gRPC | MySQL | 管理后台（DDD） |
| search | HTTP/gRPC | ES | 搜索 |
| gateway | HTTP | - | API 网关 |

## 核心架构决策

1. **go-zero 框架** — handler 层通过模板生成，logic 层手写业务逻辑
2. **go.work workspace** — 所有子模块在同一个 workspace 中管理
3. **DTM Saga** — 跨服务分布式事务的唯一方案（checkout → order + payment + inventory）
4. **DDD 分层（admin, audit）** — pb 结构体禁止泄漏到 logic 层
5. **Consul** — 服务注册与发现
6. **Elasticsearch** — 全文搜索 + 审计日志存储
7. **RabbitMQ** — 异步消息（审计、活动通知）
8. **MinIO** — 对象存储（商品图片等）

## 关键业务路径

```
用户浏览:
  Browser → Gateway → products (gRPC) → MySQL/ES

用户下单（核心分布式事务路径）:
  Browser → Gateway → checkout (gRPC)
    → DTM Saga 协调 → orders + payment + inventory
    → 全部成功 → commit
    → 任一失败 → 全局补偿回滚

用户搜索:
  Browser → Gateway → search (gRPC) → ES

管理员操作:
  Admin UI → Gateway → admin (gRPC) → MySQL
```

## 基础设施（OrbStack，禁止修改）

| 服务 | 地址 | 端口 |
|------|------|------|
| MySQL | 127.0.0.1 | 3306 |
| Redis | 127.0.0.1 | 6379 |
| Consul | 127.0.0.1 | 8500 |
| ES | 127.0.0.1 | 9200 |
| RabbitMQ | 127.0.0.1 | 5672 |
| DTM | 127.0.0.1 | 36789 (HTTP) / 36790 (gRPC) |
| MinIO | 127.0.0.1 | 9000 |

## 常见陷阱

1. **go.work 中不能同时 `use` 和 `replace` 同一模块** — replace 放各子模块 go.mod
2. **proto 生成的 pb 必须在 `services/<name>/pb/`** — 不在根目录
3. **`make mock` 必须在修改 interface 后运行** — mock 不会自动更新
4. **audit 和 admin 是 DDD 架构** — 不要用 pb 结构体污染 logic 层
5. **覆盖率基线 63.6%** — 新代码不能导致覆盖率跌破此线
