# Go-mall 项目规范

**【重要】每次代码修改并准备提交前，必须运行 `./scripts/check.sh --skip-tests` 或 `make lint` 确保本地检查通过后再提交。**

The role of this file is to describe common mistakes and confusion points that agents might encounter as they work in this project. If you ever encounter something in the project that surprises you, please alert the developer working with you and indicate that this is the case in the AgentMD file to help prevent future agents from having the same issue

## go-zero 开发规范
1. 必须使用 go-zero 的模板生成代码，禁止手写 handler 层代码
2. 编写 go-zero 相关代码时，必须调用 mcp-zero 工具进行代码生成和相关操作

## gRPC Protobuf 代码生成规范

**所有服务的 protobuf 生成的 gRPC 代码必须放在 `pb` 子目录**（工业界标准做法）：

```
services/
  order/
    pb/              # ← gRPC 生成的代码放这里
      order.pb.go
      order_grpc.pb.go
    internal/
    order.go
```

- **原因**：避免 Go 导入路径解析歧义（避免与模块名 `services/order` 冲突）
- **导入路径**：`import "github.com/falconfan123/Go-mall/services/order/pb"`
- **proto 配置**：确保 proto 文件的 `option go_package = "github.com/falconfan123/Go-mall/services/服务名/pb";`

## 基础服务地址（OrbStack 部署，禁止修改）
| 服务名称 | 地址 | 端口 | 说明 |
|---------|------|------|------|
| MySQL | 127.0.0.1 | 3306 | 关系型数据库 |
| Redis | 127.0.0.1 | 6379 | 缓存服务 |
| Consul | 127.0.0.1 | 8500 | 服务注册与发现 |
| Elasticsearch | 127.0.0.1 | 9200 | 搜索引擎 |
| RabbitMQ | 127.0.0.1 | 5672 | 消息队列（AMQP） |
| RabbitMQ管理界面 | 127.0.0.1 | 15672 | 消息队列管理后台 |
| DTM | 127.0.0.1 | 36789 | 分布式事务服务（HTTP） |
| DTM | 127.0.0.1 | 36790 | 分布式事务服务（gRPC） |
| MinIO | 127.0.0.1 | 9000 | 对象存储服务（API） |
| MinIO管理界面 | 127.0.0.1 | 9001 | 对象存储管理后台 |

## 分布式事务规范
项目采用 DTM Saga 模式实现分布式事务，禁止使用其他分布式事务方案

## 前端开发规范
修改前端服务代码后，必须调用 Chrome DevTools MCP 工具进行功能测试、兼容性测试和性能测试，确保修改不会破坏现有功能的正常运行

## 接口测试规范
当访问前后端交互边界时，必须使用 Apifox-mcp 进行接口测试（包括添加、删除接口等操作），禁止从前端直接越过网关（gateway）访问接口，也禁止从后端直接越过网关访问接口

### Apifox 配置信息
- API Key: `afxp_1bd4b5Je1OTIXl6b6AlwXgwQ7qxzzttqqTlk`
- 项目 ID: `7907732`

## CI/CD 检查规范

### 检查流程
项目有两套独立的检查机制，**各管各的**：

| 检查类型 | 运行位置 | 检查内容 |
|---------|---------|---------|
| **本地检查** | 本地电脑 (`./scripts/check.sh`) | staticcheck, golint, revive, gofmt 等 |
| **GitHub Actions CI** | GitHub 自动流程 | build, test, security, deps |

### 本地检查（每次 PR 前必运行）

**重要**：每次提交 PR 前，必须先运行本地检查，确保没有问题后再提交。

```bash
# 运行本地 CI 检查（跳过测试，加快检查速度）
./scripts/check.sh --skip-tests

# 或使用 Makefile
make lint
```

本地检查包含：
1. 代码格式 (`gofmt`)
2. 静态分析 (`go vet`, `staticcheck`)
3. 代码风格 (`golint`, `revive`)
4. 编译检查
5. 单元测试（可选）

### 快速格式化
如果格式检查失败，可以自动修复：
```bash
./scripts/check.sh --auto-fix
# 或
make fmt
```

### 工具安装
首次使用需要安装检查工具：
```bash
make install-tools
```

### GitHub Actions CI
GitHub Actions 会运行标准 CI 检查：
- **build** - 验证各服务编译
- **test** - 单元测试
- **security** - go vet + 敏感信息检查
- **deps** - 依赖一致性 + 漏洞扫描

注意：由于项目使用了 `replace` 本地模块替换，某些本地检查工具可能无法在 CI 环境中运行，因此本地检查只应在本地执行。

## Go Workspace (go.work) 规范

### 常见错误与避免方法

**错误1：在 go.work 中同时使用 use 和 replace**
- 问题：`go work` 中对同一模块不能同时使用 `use` 和 `replace`，会导致 "workspace module is replaced at all versions" 错误
- 解决：go.work 只使用 `use` 指令，replace 指令放在各服务的 go.mod 中

**错误2：子模块导入路径错误**
- 问题：proto 文件生成的 go_package 配置错误，导致导入路径多了一个 `/order` 后缀
- 解决：确保 proto 文件的 `option go_package = ".";` 配置正确，所有导入路径应为 `github.com/falconfan123/Go-mall/services/服务名`，而不是 `.../services/服务名/服务名`

**错误3：子模块遗留的 go.mod**
- 问题：旧代码可能残留子模块目录，包含独立的 go.mod
- 解决：删除这些遗留目录，确保 proto 生成的文件在正确位置

### 正确配置步骤

1. **清理阶段**：删除根目录的 go.mod（如果有）
2. **go.work 配置**：只使用 `use` 指令
   ```go
   go 1.25.0

   use (
       ./common
       ./dal
       ./services/checkout
       ...
   )
   ```
3. **各服务 go.mod**：添加需要的 replace 指令
   ```go
   module github.com/falconfan123/Go-mall/services/order

   go 1.25.0

   replace github.com/falconfan123/Go-mall/common => ../../common
   replace github.com/falconfan123/Go-mall/dal => ../../dal
   ...
   ```
4. **修复导入路径**：确保所有导入路径与 module name 匹配

## 单元测试规范

**重要原则**：单元测试应当依附于 Swagger 文档，不能随便乱测。

具体要求：
1. 测试用例必须基于 API 规范（Swagger/OpenAPI 文档）编写
2. 测试输入输出应与 API 定义保持一致
3. 测试场景应覆盖 API 文档中声明的所有端点和参数
4. 禁止编写与 API 文档无关的随机测试用例
5. 在编写测试前，应先查阅对应的 Swagger 文档或 API 定义
