# 脚本说明文档

本文档记录 `scripts/` 目录下所有脚本的用途。

## 目录

- [脚本列表](#脚本列表)
- [使用说明](#使用说明)

## 脚本列表

### 1. check.sh

**用途**: 本地代码质量检查脚本

**功能**:
- 代码格式检查 (gofmt)
- 静态分析 (go vet, staticcheck)
- 代码风格检查 (golint, revive)
- 编译检查
- 单元测试（可选）

**使用方法**:
```bash
# 运行检查（跳过测试）
./scripts/check.sh --skip-tests

# 运行完整检查（包含测试）
./scripts/check.sh

# 自动修复格式问题
./scripts/check.sh --auto-fix
```

**相关命令**:
```bash
make lint          # 通过 Makefile 运行检查
make lint-fast     # 快速格式检查
make fmt           # 自动格式化代码
```

---

### 2. build.sh

**用途**: 构建所有微服务

**功能**:
- 构建所有 API 服务 (apis/)
- 构建所有 RPC 服务 (services/)

**使用方法**:
```bash
./scripts/build.sh
```

**相关命令**:
```bash
make build         # 通过 Makefile 构建
```

---

### 3. start-unified.sh

**用途**: Go-Mall 统一启动脚本（增强版）

**功能**:
- 自动清理旧进程和占用端口
- 启动所有微服务（auths, audit, users, inventory, product, carts, coupons, order, checkout, payment）
- 启动所有 API 网关（user-api, product-api, carts-api, order-api, checkout-api, payment-api, coupon-api, flash-api）
- 启动主网关（gateway）
- 启动前端（frontend）

**服务端口映射**:
| 服务 | 端口 | 类型 |
|------|------|------|
| auths | 10000 | RPC |
| audit | 10008 | RPC |
| users | 10001 | RPC |
| inventory | 10011 | RPC |
| product | 10002 | RPC |
| carts | 10003 | RPC |
| coupons | 10009 | RPC |
| order | 10004 | RPC |
| checkout | 10005 | RPC |
| payment | 10006 | RPC |
| user-api | 8001 | API |
| product-api | 8002 | API |
| carts-api | 8003 | API |
| order-api | 8004 | API |
| checkout-api | 8005 | API |
| payment-api | 8006 | API |
| coupon-api | 8009 | API |
| flash-api | 8008 | API |
| gateway | 8888 | Gateway |
| frontend | 3000 | Frontend |

**使用方法**:
```bash
# 启动核心服务（默认）
./scripts/start-unified.sh

# 停止所有服务
./scripts/start-unified.sh stop
```

---

### 4. update_configs.sh

**用途**: 批量更新数据库配置文件（Shell 版本）

**功能**:
- 更新 MySQL 连接信息（从 jjzzchtt:jjzzchtt 改为 root:fht3825099）
- 移除 Redis 密码配置

**使用注意**:
- 该脚本会修改所有 services 和 apis 目录下的 yaml 配置文件
- 不会修改 .prod.yaml 生产环境配置
- 不会修改 manifests 目录下的配置

**使用方法**:
```bash
./scripts/update_configs.sh
```

---

### 5. update_configs.py

**用途**: 批量更新数据库配置文件（Python 版本）

**功能**:
- 与 update_configs.sh 相同，但使用 Python 实现
- 更新 MySQL 连接信息
- 移除 Redis 密码配置

**使用方法**:
```bash
python3 ./scripts/update_configs.py
```

---

## Makefile 快捷命令

项目根目录的 Makefile 提供了常用命令的快捷方式：

```bash
# 代码质量
make lint          # 运行本地 CI 检查（推荐提交前运行）
make lint-fast     # 快速格式检查
make fmt           # 自动格式化代码
make vet           # 运行 go vet
make staticcheck   # 运行 staticcheck

# 构建和测试
make build         # 构建所有服务
make test          # 运行测试
make tidy          # 整理依赖

# 安装工具
make install-tools # 安装所需工具

# CI/CD
make ci            # 模拟 CI 检查
```
