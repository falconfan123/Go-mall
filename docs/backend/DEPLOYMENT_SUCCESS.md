# Go-Mall 项目部署成功！

## 项目概述

Go-Mall 是一个基于 Go-Zero 框架的微服务电商系统，现已成功部署并运行！

## 已启动的服务

### 基础设施服务 (Docker)

| 服务 | 地址 | 状态 |
|------|------|------|
| Consul | http://localhost:8500 | ✅ 运行中 |
| Elasticsearch | http://localhost:9200 | ✅ 运行中 |
| RabbitMQ | http://localhost:15672 | ✅ 运行中 |
| MySQL | localhost:3306 | ✅ 运行中 |
| Redis | localhost:6379 | ✅ 运行中 |

### 后端 RPC 服务

| 服务 | 端口 | 状态 |
|------|------|------|
| auths.rpc | 10000 | ✅ 已注册到 Consul |
| users.rpc | 10001 | ✅ 已注册到 Consul |
| audit.rpc | 10009 | ✅ 已注册到 Consul |
| inventory.rpc | 10005 | ✅ 已注册到 Consul |
| products.rpc | 10002 | ✅ 已注册到 Consul |

### API 网关服务

| 服务 | 地址 | 状态 |
|------|------|------|
| User API | http://localhost:8001 | ✅ 运行中 |
| Product API | http://localhost:8002 | ✅ 运行中 |

## 数据库信息

- 数据库名: mall
- 用户名: root
- 密码: fht3825099

## API 测试

### 用户注册

```bash
curl -X POST http://localhost:8001/douyin/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "confirmPassword": "password123"
  }'
```

### 用户登录

```bash
curl -X POST http://localhost:8001/douyin/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

## 管理界面

- **Consul UI**: http://localhost:8500 - 查看所有注册的服务
- **Elasticsearch**: http://localhost:9200 - 搜索服务状态
- **RabbitMQ Management**: http://localhost:15672 (用户名: admin, 密码: admin)

## 日志查看

```bash
# 认证服务日志
tail -f /tmp/auths.log

# 用户服务日志
tail -f /tmp/users.log

# 审计服务日志
tail -f /tmp/audit.log

# 库存服务日志
tail -f /tmp/inventory.log

# 产品服务日志
tail -f /tmp/product.log

# 用户API日志
tail -f /tmp/user-api.log

# 产品API日志
tail -f /tmp/product-api.log
```

## 停止服务

```bash
# 停止所有 Go 服务
pkill -f 'go run'

# 停止 Docker 容器
docker stop go-mall-consul go-mall-rabbitmq go-mall-elasticsearch
```

## 启动服务

```bash
# 启动所有服务
cd /Users/fan/go-mall
./start-all.sh
```

## 项目文件说明

- `start-all.sh` - 完整启动脚本
- `start-minimal.sh` - 最小化启动脚本（仅用户服务）
- `STARTED.md` - 初始启动说明
- `.env` - 环境变量配置

## 测试数据

数据库中已包含测试数据：
- 测试用户: test@example.com
- 测试产品: iPhone 15, MacBook Pro, Nike Air Max 等
- 测试分类: 电子产品, 服装, 数码配件

## 注意事项

1. 所有服务都已注册到 Consul，可以通过 http://localhost:8500 查看
2. Elasticsearch 用于产品搜索功能
3. RabbitMQ 用于异步消息处理
4. Redis 用于缓存和会话管理
