# 核心行动准则

<thought>用户提出了涉及实时数据/具体操作的问题，需要使用工具来获取数据和操作。首先应该选择合适的工具来完成任务。</thought>

1. **工具优先原则**：每当用户提出涉及[实时数据/具体操作]的问题，必须首先扫描可用工具列表
2. **禁止道歉**：严禁回答"我无法访问..."，除非已经尝试调用了相关工具并收到了错误回复
3. **隐式思考**：在调用工具前，在内部 <thought> 标签中记录选择该工具的原因

---

# Harness Engineering — Rule 体系（硬约束）

以下规则是 AI 开发 Go-mall 时必须遵守的工程准则。Rule 是软约束不是硬关卡，但请尽量遵守。如果认为某条 Rule 不适用，必须说明理由。

## Rule 1：修改后三件事（最高优先级）
每次代码修改后，必须依次完成以下三步。任何一步不通过，任务就不算完成：
1. **编译通过** — `go build ./...` 零错误
2. **单元测试通过** — `make test-unit` 全部绿色
3. **Lint 检查通过** — `make lint` 零警告

## Rule 2：API 优先
- 新增功能前必须先定义/更新 OpenAPI/Swagger 规范
- 单元测试必须基于 API 规范编写
- 禁止编写与 API 文档无关的随机测试用例
- 测试场景应覆盖 API 文档中声明的所有端点和参数

## Rule 3：go-zero 代码生成
- 必须使用 go-zero 的模板生成 handler 层代码，禁止手写
- proto 文件修改后必须重新生成 pb 代码（`make proto`）
- 生成代码的 pb 目录路径必须符合 `services/<name>/pb/` 规范
- `go_package` 必须设置为 `github.com/falconfan123/Go-mall/services/<name>/pb`

## Rule 4：DDD 架构红线（admin 服务）
- pb 结构体禁止泄漏到 logic 或 service 层
- 进入 logic 层前必须将 pb 对象转换为内部领域对象
- 业务逻辑严禁直接引用 pb 标签或传输协议类型
- 违反此规则的 PR 必须驳回

## Rule 5：提交纪律
- 每次提交前必须运行 `./scripts/check.sh --skip-tests` 或 `make lint`
- 提交信息必须遵循 conventional-commits 规范（feat/fix/docs/refactor/test/chore）
- 每个提交对应一个逻辑单元，禁止超大提交
- 完成功能后自动创建 PR，除非用户明确不要求

## Rule 6：Mock 先行
- 涉及外部依赖（DB、RPC、MQ）的单元测试必须使用 mock
- 修改业务代码后必须运行 `make mock` 确保 mock 已更新
- 禁止在单元测试中连接真实数据库或真实 RPC 服务
- 新增接口后，确认 mockgen 覆盖了新的接口方法

## Rule 7：分布式事务规范
- 跨服务事务必须使用 DTM Saga 模式
- 禁止引入其他分布式事务方案（本地消息表、TCC、2PC 等）
- 每个 Saga 事务必须有对应的补偿逻辑
- 补偿逻辑必须是幂等的

---

# Go-mall 项目规范

## 本地启动约定

- 遇到 `start program go-mall`、`启动 go-mall`、`拉起 go-mall 前后端`、`启动本地联调环境`、`重启 go-mall`、`停止 go-mall` 这类表达时，优先读取 `go-mall-start` skill。
- 如果 skill 不可用，直接使用 `./scripts/start-unified.sh`、`./scripts/start-unified.sh status`、`./scripts/start-unified.sh stop`、`./scripts/start-unified.sh restart`。
- 禁止重新手工拼装 Docker、Go、npm 的启动顺序，除非统一脚本本身失败。

**【重要】每次代码修改并准备提交前，必须运行 `./scripts/check.sh --skip-tests` 或 `make lint` 确保本地检查通过后再提交。**

**【重要】完成任何任务后，必须自动提交所有更改并创建 PR，除非用户明确要求不提交。**

**【重要】从现在起，你在 `services/` 下修改任何业务代码、新增任何 API 功能后，必须在后台自动触发 `make mock` 和 `make test-unit`。如果单测失败，或者你写的新代码导致覆盖率跌破 `63.6%`，你必须在原地启动自我修复循环，直到单测全绿才能向我（人类）交付。**

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

## DDD 架构约束

**【重要】admin 服务严格遵循 DDD 分层架构：**

- **禁止层级污染**：pb (Protobuf) 生成的结构体仅限在 `api` 或 `handler` 层使用。
- **强制映射**：进入 `logic` 或 `service` 层前，必须将 pb 对象转换为内部领域对象（Domain Object）。
- **命名规范**：admin 业务逻辑严禁直接引用 pb 标签，确保逻辑层不依赖特定传输协议。

**使用规范**：当处理 admin 服务的接口转换时，使用 `/map-ddd` 命令自动化处理 pb 结构体与领域实体的映射逻辑。

---

## CI-FIX 知识库

当处理 `ci-fix` 类型的 Issue 或 CI 失败修复时，**必须先阅读以下知识文件**，复用历史经验：

1. **`CI-FIX-KNOWLEDGE.md`** — 常见 CI 失败模式与标准修复操作（启动必读）
2. **`CI-FIX-CASES/`** — 具体案例库，包含每个历史修复的根因分析、修复步骤、PR 链接

### 修复 CI 后的自动记录

每次成功修复 CI 后，必须：
1. 在 `CI-FIX-CASES/` 创建 `issue-<number>.md`，记录：错误、根因、修复、PR
2. 更新 `CI-FIX-KNOWLEDGE.md`，如果本次修复对应新的失败模式，追加到常见模式表中
3. 提交并推送这些知识文件
