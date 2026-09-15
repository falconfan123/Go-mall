# Go-Mall 日志系统架构

## 概述

Go-Mall 项目采用 Loki + Grafana 集中式日志架构，替代原来散落在各个服务中的日志文件。

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Services  │────▶│   Promtail  │────▶│    Loki     │
│  (JSON Log) │     │ (收集日志)   │     │ (存储聚合)   │
└─────────────┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │   Grafana   │
                                        │  (可视化查询) │
                                        └─────────────┘
```

## 组件说明

### 1. 各服务日志 (JSON 格式)
- 所有服务配置 `Encoding: json` 输出 JSON 格式日志
- 日志文件保存在 `scripts/logs/` 目录

### 2. Promtail (日志抓取)
- 配置文件: `infrastructure/promtail/promtail-config.yaml`
- 自动抓取所有服务的日志文件
- 添加 service 标签便于分类查询

### 3. Loki (日志存储与聚合)
- 配置文件: `infrastructure/loki/loki-config.yaml`
- 接收 Promtail 推送的日志
- 提供日志查询 API

### 4. Grafana (可视化)
- 访问地址: http://localhost:3001
- 默认账号: admin / admin123
- 自动配置 Loki 数据源

## 快速开始

> **栈的位置说明**：日志栈（Loki/Promtail/Grafana）位于 `infrastructure/docker-compose.yaml`；
> 链路/指标栈（Jaeger/Prometheus）位于 `construct/observability/docker-compose.yaml`。
> 两者是**两个独立 compose**，统一启动脚本 `./scripts/start-unified.sh` 现已将它们一并拉起/停止，
> 不存在指向两处合一的统一 compose 文件。

### 启动日志系统

```bash
# 方式1: 使用统一启动脚本 (推荐) —— 同时拉起日志栈与 Jaeger/Prometheus 观测栈
./scripts/start-unified.sh

# 方式2: 手动启动 (备选，仅日志栈；观测栈需另执行 cd construct/observability && docker-compose up -d jaeger prometheus)
cd infrastructure
docker-compose up -d
```

### 停止日志系统

```bash
# 方式1: 使用统一脚本 —— 同时清理日志栈与观测栈
./scripts/start-unified.sh stop

# 方式2: 手动停止
cd infrastructure
docker-compose down
```

## 日志查询示例

### 查询特定服务的日志
```
{service="payment"}
```

### 查询包含错误的日志
```
{service="order"} |= "error"
```

### 查询特定时间范围
```
{service="gateway"} | json | level="error" | time > now - 1h
```

### 按请求追踪 (trace)
```
{service="payment"} | json | trace="your-trace-id"
```

## 服务日志标签

| 服务 | 标签值 |
|------|--------|
| Gateway | gateway |
| Payment | payment |
| Order | order |
| Product | product |
| Inventory | inventory |
| Users | users |
| Carts | carts |
| Checkout | checkout |
| Auths | auths |
| System | system |
| Activity | activity |
| Audit | audit |
| Coupons | coupons |

## 端口说明

| 服务 | 端口 | 说明 |
|------|------|------|
| Loki | 3100 | 日志 API |
| Grafana | 3001 | 可视化界面 |
| Promtail | 9080 | 抓取配置 |

## 注意事项

1. **日志目录权限**: 确保 `scripts/logs` 目录存在且有写权限
2. **Docker 依赖**: 需要 Docker 和 docker-compose
3. **首次启动**: 首次启动需要拉取 Docker 镜像，可能较慢
4. **日志保留**: 默认保留 7 天，可在 loki-config.yaml 中调整
