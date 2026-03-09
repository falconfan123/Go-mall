# Go-Mall 项目启动成功！

## 已启动的服务

### 基础设施服务 (Docker)
- **Consul**: http://localhost:8500 - 服务发现与注册中心
- **RabbitMQ**: http://localhost:5672 (管理界面: http://localhost:15672)
- **MySQL**: localhost:3306 - 数据库
- **Redis**: localhost:6379 - 缓存

### 应用服务
- **auths.rpc**: localhost:10000 - 认证服务
- **users.rpc**: localhost:10001 - 用户服务
- **user-api**: http://localhost:8001 - 用户 API 网关

## 测试 API

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

响应示例：
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
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

## 如何停止服务

```bash
pkill -f 'go run'
```

## 查看日志

```bash
# 认证服务日志
tail -f /tmp/auths.log

# 用户服务日志
tail -f /tmp/users.log

# 用户API日志
tail -f /tmp/user-api.log
```

## 数据库信息

- 数据库: mall
- 用户: root / fht3825099
- 测试用户已创建: test@example.com

## Consul 服务列表

访问 http://localhost:8500 查看所有已注册的服务。
