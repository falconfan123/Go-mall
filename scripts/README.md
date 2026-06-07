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

**用途**: Go-Mall 本地联调唯一启动入口

**功能**:
- 启动 `construct/depend/docker-compose.yaml` 中的依赖容器
- 校准 Postgres / RabbitMQ 本地账号与初始化数据
- 按固定顺序启动全部后端 RPC 与 gateway
- 启动 `stripe listen --forward-to http://127.0.0.1:11112/stripe/webhook`
- 在 `frontend/` 执行 `npm run dev -- --host 127.0.0.1 --port 3000`
- 将后端和前端 PID 统一写入 `/tmp/go-mall-pids.txt`
- 输出日志到 `scripts/logs/*.log`
- 支持 `status` / `stop` / `restart`

**统一命令接口**:
```bash
# 首次使用前先登录 Stripe CLI
stripe login

# 启动依赖 + 后端 + 前端
./scripts/start-unified.sh

# 查看当前状态
./scripts/start-unified.sh status

# 停止统一脚本拉起的进程与依赖
./scripts/start-unified.sh stop

# 完整重启
./scripts/start-unified.sh restart
```

**后端端口映射**:
| 服务 | 端口 | 类型 |
|------|------|------|
| auths | 10000 | RPC |
| users | 10001 | RPC |
| product | 10002 | RPC |
| carts | 10003 | RPC |
| order | 10004 | RPC |
| checkout | 10005 | RPC |
| payment | 10006 | RPC |
| inventory | 10007 | RPC |
| audit | 10008 | RPC |
| coupons | 10009 | RPC |
| system | 10010 | RPC |
| activity | 10011 | RPC |
| admin | 10012 | RPC |
| gateway | 8888 | Gateway |
| frontend | 3000 | Vite |

**固定访问地址**:
| 服务 | 地址 | 说明 |
|------|------|------|
| Frontend | http://127.0.0.1:3000 | 本地前端联调入口 |
| Gateway | http://127.0.0.1:8888 | 网关入口 |
| RabbitMQ 管理台 | http://127.0.0.1:15672 | admin / admin |
| Gorse | http://127.0.0.1:8088 | 推荐服务控制台 |
| Grafana | http://127.0.0.1:3001 | 日志可视化 |
| Loki | http://127.0.0.1:3100 | 日志 API |
| Stripe webhook | http://127.0.0.1:11112/stripe/webhook | 本地 webhook 接收端 |

**Stripe 本地支付链路说明**:
- 统一启动脚本会自动拉起 `stripe listen --forward-to http://127.0.0.1:11112/stripe/webhook`
- 如未登录 Stripe CLI，启动会直接失败并提示先执行 `stripe login`
- 查看转发日志：`tail -f scripts/logs/stripe-listen.log`

---

### 4. ci-rpc-stack.sh

**用途**: GitHub Actions / 本地 CI 使用的无交互集成环境编排脚本

**功能**:
- 使用 `construct/depend/docker-compose.yaml` 启动 Postgres、Redis、RabbitMQ、etcd、Elasticsearch、gorse
- 校准 Postgres 和 RabbitMQ 测试账号
- 直接以 `go run` 启动本地 RPC 服务与 gateway，不依赖 `air`
- 为 CI 本地模式渲染临时配置，将 `localhost` 依赖地址覆写为 `127.0.0.1`，避免 IPv6 `::1` 解析问题
- 校验依赖端口 `2379/5432/5672/6379/9200/8088`
- 校验核心端口 `10000-10009/8888`
- 输出服务日志到 `scripts/logs/*.log`
- 输出依赖容器日志到 `.artifacts/dependency-logs/*.log`
- 支持在同一个脚本生命周期内完成“启动 -> 测试 -> 日志采集 -> 清理”

**使用方法**:
```bash
./scripts/ci-rpc-stack.sh start
./scripts/ci-rpc-stack.sh status
./scripts/ci-rpc-stack.sh snapshot-dependency-logs
./scripts/ci-rpc-stack.sh stop
./scripts/ci-rpc-stack.sh run-local-suite env GO_MALL_TEST_LOCAL=1 bash scripts/test-integration.sh
```

**相关命令**:
```bash
make integration-up
make integration-down
```

---

### 5. test-integration.sh

**用途**: 执行 RPC 集成测试，并生成 HTML/JUnit/摘要报告

**功能**:
- 本地模式下在同一个脚本进程中完成依赖启动、RPC/gateway 拉起、端口校验、测试执行、日志采集与清理
- 生成 `.artifacts/rpc-integration-report/{index.html,junit.xml,summary.txt}`
- 如果在 `go test` 前的启动阶段失败，仍会生成最小失败报告，保证 CI artifact 完整
- 保留 Kubernetes Job 模式作为兼容入口，但 CI 默认不使用它

**使用方法**:
```bash
GO_MALL_TEST_LOCAL=1 ./scripts/test-integration.sh
```

**兼容旧的 Kubernetes Job 模式**:
```bash
./scripts/test-integration.sh
```

### 6. go-ci-build.sh

**用途**: 根据 `scripts/build_services.txt` 按白名单构建服务模块

**使用方法**:
```bash
./scripts/go-ci-build.sh --all
./scripts/go-ci-build.sh users payment
```

---

### 7. go-ci-vet.sh / go-ci-vulncheck.sh

**用途**: 根据 `scripts/go_ci_modules.txt` 对 workspace 模块逐个执行 `go vet` 或 `govulncheck`

**使用方法**:
```bash
./scripts/go-ci-vet.sh
./scripts/go-ci-vulncheck.sh
```

---

### 8. update_configs.sh

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

### 9. update_configs.py

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

### 10. rag

**用途**: 仓库内 RAG CLI，先检索本仓库，再调用模型回答或进入自主修复循环

**子命令**:
- `index`: 重建 `.artifacts/rag/rag.db` 中的本地索引
- `ask`: 检索仓库上下文并生成带文件/行号引用的回答
- `loop`: 在独立 git worktree 中执行“检索 -> 改代码 -> 跑白名单命令 -> 失败再修复”的闭环
- `resume`: 继续上一个未完成 session
- `doctor`: 检查 Anthropic 兼容鉴权、本地写权限、git worktree 和默认白名单命令

**使用方法**:
```bash
./scripts/rag doctor
./scripts/rag index
./scripts/rag ask "checkout 服务的状态回滚逻辑在哪"
./scripts/rag loop "修复 checkout 单测失败" --dry-run

# Minimax / Anthropic-compatible gateway:
ANTHROPIC_BASE_URL=https://your-gateway.example.com
ANTHROPIC_AUTH_TOKEN=...
./scripts/rag ask "inventory 预扣减逻辑在哪"
```

**环境变量**:
```bash
ANTHROPIC_AUTH_TOKEN=...   # 可选，Anthropic-compatible Bearer token，优先于 API key
ANTHROPIC_API_KEY=...      # 可选，Anthropic 原生 x-api-key
ANTHROPIC_MODEL=...        # 可选，默认 claude-sonnet-4-20250514
ANTHROPIC_BASE_URL=...     # 可选，兼容代理
ANTHROPIC_VERSION=...      # 可选，默认 2023-06-01
```

说明：这不是新增 `minimax` backend，而是对现有 `anthropic` backend 增加兼容鉴权扩展；接口仍然是 `POST /v1/messages`。

---

### 11. configure-branch-protection.sh

**用途**: 为 GitHub 仓库默认分支配置 CI 门禁与 auto-merge 能力

**功能**:
- 打开仓库 `allow_auto_merge`
- 为 `main` 创建或更新名为 `main-ci-gate` 的规则集
- 强制默认分支必须通过 PR 合并
- 要求固定检查 `Build` / `Quality` / `Integration` 通过后才能合并

**使用方法**:
```bash
./scripts/configure-branch-protection.sh
make configure-branch-protection
```

---

### 12. submit-with-ci.sh

**用途**: 一条命令完成提交、推送、创建 PR、等待 CI、自动合并

**功能**:
- 检查 `git` / `gh` / `go` / `make` 环境
- 运行本地前置校验：`go work sync`、`make test-unit`、`make ci-build`
- 自动创建时间戳分支、提交全部改动并 rebase 到 `origin/main`
- 自动创建 PR、启用 `squash` auto-merge
- 轮询 required checks，直到通过或失败
- 合并完成后同步本地 `main`

**使用方法**:
```bash
MSG="ci: enable automated PR merge flow" ./scripts/submit-with-ci.sh
make submit-ci MSG="ci: enable automated PR merge flow"
```

**注意事项**:
- 默认会拦截 basename 以 `.tmp` 开头的文件，避免调试残留被自动提交
- 依赖 GitHub Actions 检查名 `Build` / `Quality` / `Integration`

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
make integration-up
make integration-down
make ci-build
make ci-vet
make tidy          # 整理依赖

# 安装工具
make install-tools # 安装所需工具

# CI/CD
make ci            # 模拟 CI 检查
```
