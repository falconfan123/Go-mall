# CI/CD 重构 PRD

## 1. 背景与目标

### 1.1 背景
- 项目目前没有任何 CI/CD 自动化流程
- 代码质量检查依赖人工 Review，效率低下
- 每次提交到 GitHub 后才能发现构建问题，浪费 CI 资源
- 缺乏统一的代码规范执行

### 1.2 目标
1. 构建本地 CI 检查脚本，在本地先筛一遍问题
2. 创建 GitHub Actions 工作流，实现自动化构建、测试、lint
3. 确保本地检查与 GitHub Actions 检查保持一致
4. 优先在本地发现问题，减少 CI 资源浪费

### 1.3 设计原则
- **本地优先**: 常见问题（格式、编译）在本地解决
- **快速反馈**: 本地脚本 1-2 分钟内完成
- **分层检查**: 本地只做快速检查，GitHub Actions 做完整检查
- **渐进增强**: 工具缺失不阻塞，只给出警告

---

## 2. 技术方案

### 2.1 工具选型

| 工具 | 用途 | 本地安装 |
|------|------|----------|
| gofmt | 代码格式化 | 内置 |
| go vet | 静态分析 | 内置 |
| staticcheck | 深度静态分析 | `go install honnef.co/go/tools/cmd/staticcheck@latest` |
| golint | 代码风格检查 | `go install golang.org/x/lint/golint@latest` |
| revive | Go 代码检查器 | `go install github.com/mgechev/revive@latest` |

### 2.2 本地检查流程

```
本地开发流程:
1. 编写代码
2. 运行 make lint (或 ./scripts/local-ci.sh)
3. 修复本地发现的问题
4. git commit & push
5. GitHub Actions 执行完整检查
```

### 2.3 检查分层

| 层级 | 检查内容 | 执行时间 | 阻塞级别 |
|------|----------|----------|----------|
| L1 (本地) | gofmt, go vet, staticcheck(-checks=basic) | <1min | 阻塞提交 |
| L2 (本地) | golint, revive | 1-2min | 警告 |
| L3 (CI) | 完整 staticcheck, 测试 | 5-10min | 阻塞合并 |
| L4 (CI) | 安全扫描, 依赖检查 | 2-5min | 阻塞合并 |

---

## 3. 实现细节

### 3.1 本地检查脚本

**文件**: `scripts/local-ci.sh`

**功能**:
- 自动检测所需工具是否存在
- 支持 `--skip-tests` 跳过测试
- 支持 `--auto-fix` 自动修复格式问题
- 彩色输出，易于阅读

**使用方法**:
```bash
# 完整检查 (包含测试)
./scripts/local-ci.sh

# 跳过测试
./scripts/local-ci.sh --skip-tests

# 自动修复格式问题
./scripts/local-ci.sh --auto-fix
```

### 3.2 Makefile

**文件**: `Makefile`

提供便捷的命令:
```bash
make help        # 查看所有命令
make lint        # 运行本地 CI 检查
make fmt         # 格式化代码
make vet         # 运行 go vet
make staticcheck # 运行 staticcheck
make test        # 运行测试
make build       # 构建所有服务
make install-tools # 安装所需工具
```

### 3.3 GitHub Actions

**文件**: `.github/workflows/ci.yml`

**工作流结构**:
1. **local-check**: 模拟本地快速检查
2. **lint**: 完整代码质量检查
3. **test**: 单元测试
4. **build**: 逐个服务构建验证
5. **security**: 安全扫描
6. **deps**: 依赖一致性检查

**触发条件**:
- push 到 main, develop, feature/** 分支
- pull request 到 main, develop

---

## 4. 卡点与问题记录

### 4.1 发现的问题

#### 问题 1: 代码格式不规范
**状态**: ✅ 已修复
**描述**: 项目存在大量未格式化的 Go 代码
**解决**: 运行 `make fmt` 自动修复

#### 问题 2: 遗留依赖问题
**状态**: ⚠️ 部分修复
**描述**:
- ~~`apis/flash_sale/internal/svc/servicecontext.go` 引用了不存在的包 `github.com/falconfan123/Go-mall/services/users/usersclient`~~ ✅ 已修复为 `users`
- `apis/flash_sale` 和 `apis/order` 服务引用了不存在的 `order.OrderService` 类型

**影响**:
- `go vet` 检查会有警告
- 编译检查会失败

**建议**: 后续清理废弃服务或修复引用

#### 问题 3: go-zero 生成代码的 JSON tag 兼容性
**状态**: ⚠️ 待处理
**描述**: go-zero 生成的代码中使用了 `json:"page,default=1"` 这种格式，但与新版 Go/静态检查工具不兼容

**影响**:
- staticcheck 会报错 `invalid appearance of unknown =1 tag option`
- 这些是生成代码，不应手动修改

**建议**: 后续升级 go-zero 版本或调整代码生成配置

#### 问题 4: golint 警告
**状态**: ⚠️ 待处理
**描述**: 大量导出类型/函数缺少注释，变量命名不符合 Go 规范 (如 `paymentId` 应为 `paymentID`)

**影响**: 不阻塞提交，但影响代码质量

**建议**: 后续逐步完善注释和命名规范

#### 问题 5: 缺少测试
**状态**: ⚠️ 待处理
**描述**: 项目几乎没有单元测试
**注意**: 单元测试应当依附于 Swagger 文档，不能随便乱测

**建议**: 后续根据 API 规范补充测试用例

### 4.2 本地工具安装问题

| 工具 | 安装命令 | 状态 |
|------|----------|------|
| staticcheck | `go install honnef.co/go/tools/cmd/staticcheck@latest` | ✅ 已安装（已更新至兼容 Go 1.25） |
| golint | `go install golang.org/x/lint/golint@latest` | ✅ 已安装 |
| revive | `go install github.com/mgechev/revive@latest` | ✅ 已安装 |
| gci | `go install github.com/daixiang0/gci@latest` | ✅ 已安装 |
| errcheck | 包路径变更，需使用新版 | ⚠️ 跳过 |

---

## 5. 使用指南

### 5.1 开发流程

```
第一次使用:
1. 安装工具: make install-tools
2. 运行检查: make lint
3. 修复问题后提交

日常使用:
1. 开发完成后
2. 运行 make lint (或 ./scripts/local-ci.sh --skip-tests)
3. 如有错误，修复后重新检查
4. git add . && git commit -m "feat: xxx"
5. git push
6. GitHub Actions 自动运行完整检查
```

### 5.2 CI 失败处理

| 检查阶段 | 失败原因 | 处理方式 |
|----------|----------|----------|
| 本地检查 | 格式/编译问题 | 本地修复后重新提交 |
| local-check | 同上 | 本地修复后 force push |
| lint | 代码质量问题 | 根据警告修复或添加 `//nolint` |
| test | 测试失败 | 修复测试或标记为已知问题 |
| build | 服务构建失败 | 检查服务依赖和代码 |
| security | 安全问题 | 立即修复 |
| deps | 依赖不一致 | 运行 `go mod tidy` 并提交 |

---

## 6. 后续工作

### 6.1 短期 (1-2周)
1. [x] 修复遗留的 `usersclient` 引用问题（已改为 `users` 包）
2. [ ] 修复 `order.OrderService` 类型未定义问题
3. [ ] 解决 go-zero 生成代码的 JSON tag 兼容性问题
4. [ ] 清理废弃的 `flash_sale` 和 `order` 服务
5. [ ] 补充单元测试覆盖率到 30%（根据 Swagger 文档）
6. [ ] 添加 pre-commit hook 自动运行 `make lint`

### 6.2 中期 (1个月)
1. [ ] 完善代码注释，满足 golint 要求
2. [ ] 统一变量命名规范
3. [ ] 添加集成测试
4. [ ] 配置 Dependabot 自动更新依赖

### 6.3 长期
1. [ ] 引入 SonarQube 进行更深入的代码分析
2. [ ] 添加性能测试
3. [ ] 配置自动发布
4. [ ] 添加蓝绿部署配置

---

## 7. 文件清单

| 文件路径 | 说明 |
|----------|------|
| `scripts/local-ci.sh` | 本地 CI 检查脚本 |
| `Makefile` | 开发命令集合 |
| `.github/workflows/ci.yml` | GitHub Actions 工作流 |
| `PRD_CICDreform.md` | 本文档 |

---

## 8. 附录

### 8.1 快速参考

```bash
# 安装工具
make install-tools

# 本地检查
make lint          # 完整检查
make lint-fast    # 快速检查 (仅格式)
make fmt          # 自动格式化

# 构建测试
make build        # 构建所有服务
make test         # 运行测试
make ci           # 模拟 CI 检查 (跳过测试)

# GitHub Actions 触发
git push origin main  # 自动触发 CI
```

### 8.2 CI 状态查看

- GitHub仓库 -> Actions -> CI 工作流
- 每次 push/PR 都会有详细报告

### 8.3 常见问题

**Q: 本地检查通过了，但 CI 失败了怎么办?**
A: 可能本地工具版本与 CI 不一致，确保本地工具为最新版本: `make install-tools`

**Q: golint 警告太多，不想修怎么办?**
A: 可以暂时使用 `//golint:ignore` 注释跳过，但建议逐步完善

**Q: 依赖检查失败怎么办?**
A: 运行 `go mod tidy` 整理依赖，如有冲突根据错误提示解决
