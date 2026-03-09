# Go-Mall Docker 部署指南

## 📋 目录
- [基础设施部署](#基础设施部署)
- [后端服务部署](#后端服务部署)
- [完全 Docker 部署](#完全-docker-部署)
- [日志查看](#日志查看)

---

## 🏗️ 基础设施部署

### 方式一：仅部署基础设施（推荐用于开发）

使用 `docker-compose.dev.yml` 仅部署基础设施服务（MySQL、Redis、Consul、RabbitMQ、Elasticsearch），后端服务在本地运行。

#### 1. 启动基础设施服务

```bash
cd /Users/fan/go-mall

# 启动所有基础设施服务
docker-compose -f docker-compose.dev.yml up -d

# 或者只启动核心服务（不含可选的 DTM）
docker-compose -f docker-compose.dev.yml up -d consul redis mysql rabbitmq elasticsearch
```

#### 2. 查看服务状态

```bash
# 查看运行中的容器
docker-compose -f docker-compose.dev.yml ps

# 查看所有容器
docker ps -a
```

#### 3. 查看服务日志

```bash
# 查看所有服务日志
docker-compose -f docker-compose.dev.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.dev.yml logs -f mysql
docker-compose -f docker-compose.dev.yml logs -f consul
docker-compose -f docker-compose.dev.yml logs -f elasticsearch
```

#### 4. 停止服务

```bash
# 停止服务但保留数据
docker-compose -f docker-compose.dev.yml stop

# 停止并删除容器（数据仍保留在 volumes 中）
docker-compose -f docker-compose.dev.yml down

# 完全清理（包括数据 volumes）
docker-compose -f docker-compose.dev.yml down -v
```

---

## 🔧 服务访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 🌐 Consul UI | http://localhost:8500 | 服务注册中心 |
| 🗄️ MySQL | localhost:3306 | 数据库 (root/fht3825099) |
| 🗃️ Redis | localhost:6379 | 缓存 (无密码) |
| 🐰 RabbitMQ UI | http://localhost:15672 | 消息队列 (admin/admin) |
| 🔍 Elasticsearch | http://localhost:9200 | 搜索引擎 |
| 📦 DTM | http://localhost:36789 | 分布式事务管理 |

---

## 📝 数据库初始化

如果 MySQL 是新启动的，需要初始化数据库表：

```bash
# 等待 MySQL 完全启动（约 30 秒）
sleep 30

# 执行初始化脚本
mysql -h127.0.0.1 -uroot -pfht3825099 mall < init_all_tables.sql

# 插入测试数据（可选）
mysql -h127.0.0.1 -uroot -pfht3825099 mall < insert_test_data.sql
mysql -h127.0.0.1 -uroot -pfht3825099 mall < insert_test_products.sql
```

---

## 🚀 后端服务部署

基础设施启动后，在本地终端运行后端服务：

### 方式一：使用 run.go（推荐）

```bash
cd /Users/fan/go-mall

# 启动所有核心服务 + 前端
go run run.go
```

### 方式二：使用启动脚本

```bash
# 启动所有服务（前端 + 后端）
./start-all.sh

# 或仅启动前端
./start-frontend-only.sh
```

---

## 🐳 完全 Docker 部署（进阶）

如果您想把所有服务都部署在 Docker 中，需要先构建 Docker 镜像。

### 1. 创建 .dockerignore

```bash
cat > .dockerignore << 'EOF'
.git
.gitignore
*.md
*.txt
docker-compose*.yml
Dockerfile
frontend/
*.log
tmp/
EOF
```

### 2. 构建 Docker 镜像

```bash
# 构建基础镜像
docker build -t go-mall:latest .
```

### 3. 使用生产环境 docker-compose.yml

```bash
# 注意：需要先构建好镜像或配置好镜像仓库
docker-compose up -d
```

---

## 📊 日志查看

### Docker 容器日志

```bash
# 查看所有基础设施服务日志
docker-compose -f docker-compose.dev.yml logs -f

# 查看特定服务日志
docker logs -f go-mall-mysql
docker logs -f go-mall-consul
docker logs -f go-mall-redis
docker logs -f go-mall-elasticsearch
docker logs -f go-mall-rabbitmq
```

### 本地服务日志

如果使用 `go run run.go` 或 `start-all.sh` 启动后端服务：

- 日志直接输出到运行的终端中
- 如需保存到文件：
  ```bash
  go run run.go > /tmp/go-mall.log 2>&1 &
  tail -f /tmp/go-mall.log
  ```

---

## 🔍 故障排查

### 端口被占用

```bash
# 查找占用端口的进程
lsof -ti:3306,6379,8500,9200,15672

# 杀死进程
lsof -ti:3306,6379,8500,9200,15672 | xargs kill -9
```

### 容器无法启动

```bash
# 查看容器详细状态
docker inspect go-mall-mysql

# 查看容器日志
docker logs go-mall-mysql

# 删除并重新创建容器
docker-compose -f docker-compose.dev.yml rm -f mysql
docker-compose -f docker-compose.dev.yml up -d mysql
```

### MySQL 连接问题

```bash
# 测试 MySQL 连接
mysql -h127.0.0.1 -uroot -pfht3825099 -e "SELECT 1"

# 进入 MySQL 容器
docker exec -it go-mall-mysql mysql -uroot -pfht3825099
```

---

## 💡 开发工作流建议

1. **首次设置**：
   ```bash
   # 启动基础设施
   docker-compose -f docker-compose.dev.yml up -d

   # 等待 MySQL 启动
   sleep 30

   # 初始化数据库
   mysql -h127.0.0.1 -uroot -pfht3825099 mall < init_all_tables.sql
   ```

2. **日常开发**：
   ```bash
   # 基础设施已经在运行，直接启动后端
   go run run.go
   ```

3. **清理和重置**：
   ```bash
   # 停止后端（Ctrl+C）

   # 如需完全重置
   docker-compose -f docker-compose.dev.yml down -v
   docker-compose -f docker-compose.dev.yml up -d
   ```

---

## 📚 相关文档

- [QUICKSTART.md](./QUICKSTART.md) - 快速启动指南
- [README.md](./README.md) - 项目说明
