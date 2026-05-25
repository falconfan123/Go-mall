# Makefile for Go-mall project
# 提供本地开发常用的命令

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m

# 默认目标
.PHONY: help

help:
	@echo -e "$(BLUE)Go-mall 开发命令$(NC)"
	@echo ""
	@echo "代码质量:"
	@echo "  make lint          - 运行本地 CI 检查 (推荐提交前运行)"
	@echo "  make lint-fast    - 快速格式检查"
	@echo "  make fmt          - 自动格式化代码"
	@echo "  make vet          - 运行 go vet"
	@echo "  make staticcheck  - 运行 staticcheck"
	@echo ""
	@echo "构建和测试:"
	@echo "  make build        - 构建所有服务"
	@echo "  make test         - 运行单元测试"
	@echo "  make test-unit    - 运行单元测试"
	@echo "  make test-integration - 运行集成测试"
	@echo "  make integration-up - 启动 CI 本地集成依赖与 RPC 服务"
	@echo "  make integration-down - 停止 CI 本地集成依赖与 RPC 服务"
	@echo "  make coverage     - 生成覆盖率报告"
	@echo "  make coverage-ci  - 校验覆盖率门槛"
	@echo "  make mock         - 生成 mocks"
	@echo "  make ci-build     - 按 CI 白名单构建服务"
	@echo "  make ci-vet       - 按 workspace 模块运行 go vet"
	@echo "  make configure-branch-protection - 配置 main 分支 CI 门禁和 auto-merge"
	@echo "  make submit-ci MSG='...' - 自动提交、建 PR、等待 CI 并自动合并"
	@echo "  make tidy         - 整理依赖"
	@echo "  make rag ARGS='doctor' - 运行仓库内 RAG CLI"
	@echo ""
	@echo "安装工具:"
	@echo "  make install-tools - 安装所需工具"
	@echo "  make install-rag  - 安装 ~/bin/rag 包装脚本"
	@echo ""
	@echo "CI/CD:"
	@echo "  make ci           - 模拟 CI 检查"

# 安装所需工具
install-tools:
	@echo -e "$(BLUE)安装开发工具...$(NC)"
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/lint/golint@latest
	go install github.com/mgechev/revive@latest
	go install github.com/daixiang0/gci@latest
	go install github.com/kisielk/errcheck/cmd/errcheck@latest
	go install go.uber.org/mock/mockgen@v0.6.0
	@echo -e "$(GREEN)工具安装完成$(NC)"

# 格式化代码
fmt:
	@echo -e "$(BLUE)格式化代码...$(NC)"
	gofmt -w .
	@echo -e "$(GREEN)格式化完成$(NC)"

# 快速格式检查
lint-fast:
	@echo -e "$(BLUE)快速格式检查...$(NC)"
	@UNFORMATTED=$$(gofmt -l .) && \
	if [ -n "$$UNFORMATTED" ]; then \
		echo -e "$(RED)以下文件需要格式化:$$NC"; \
		echo "$$UNFORMATTED" | head -20; \
		exit 1; \
	else \
		echo -e "$(GREEN)格式检查通过$(NC)"; \
	fi

# 运行 go vet
vet:
	@echo -e "$(BLUE)运行 go vet...$(NC)"
	@bash scripts/go-ci-vet.sh

# 运行 staticcheck
staticcheck:
	@echo -e "$(BLUE)运行 staticcheck...$(NC)"
	staticcheck ./...

# 运行本地 CI 检查 (跳过测试，因为测试需要服务运行)
lint:
	@echo -e "$(BLUE)运行本地 CI 检查...$(NC)"
	@bash scripts/check.sh --skip-tests

# 整理依赖
tidy:
	@echo -e "$(BLUE)整理依赖...$(NC)"
	go mod tidy

# 运行测试
test: test-unit

test-unit:
	@echo -e "$(BLUE)运行单元测试...$(NC)"
	@bash scripts/test-unit.sh

test-integration:
	@echo -e "$(BLUE)运行集成测试...$(NC)"
	@bash scripts/test-integration.sh

coverage:
	@echo -e "$(BLUE)生成覆盖率报告...$(NC)"
	@bash scripts/coverage.sh

coverage-ci:
	@echo -e "$(BLUE)校验覆盖率门槛...$(NC)"
	@bash scripts/coverage.sh ci

mock:
	@echo -e "$(BLUE)生成 mocks...$(NC)"
	@bash scripts/mockgen.sh

rag:
	@bash scripts/rag $(ARGS)

install-rag:
	@mkdir -p "$$HOME/bin"
	@install -m 0755 scripts/rag "$$HOME/bin/rag"
	@echo -e "$(GREEN)已安装到 $$HOME/bin/rag$(NC)"

# 构建所有服务 (只构建核心服务，跳过有问题的废弃服务)
# 排除: flash_sale (引用不存在的 usersclient), order (引用不存在的 order.OrderService)
build:
	@echo -e "$(BLUE)构建所有服务...$(NC)"
	@bash scripts/go-ci-build.sh --all
	@echo -e "$(GREEN)构建完成$(NC)"

ci-build:
	@echo -e "$(BLUE)按 CI 白名单构建服务...$(NC)"
	@bash scripts/go-ci-build.sh --all

ci-vet:
	@echo -e "$(BLUE)按 CI 白名单运行 go vet...$(NC)"
	@bash scripts/go-ci-vet.sh

integration-up:
	@echo -e "$(BLUE)启动 CI 本地集成环境...$(NC)"
	@bash scripts/ci-rpc-stack.sh start

integration-down:
	@echo -e "$(BLUE)停止 CI 本地集成环境...$(NC)"
	@bash scripts/ci-rpc-stack.sh stop

configure-branch-protection:
	@echo -e "$(BLUE)配置 main 分支保护与 CI 门禁...$(NC)"
	@bash scripts/configure-branch-protection.sh

submit-ci:
	@echo -e "$(BLUE)自动提交并等待 CI 合并...$(NC)"
	@bash scripts/submit-with-ci.sh "$(MSG)"

# 模拟 CI 检查 (跳过测试)
ci:
	@echo -e "$(BLUE)模拟 CI 检查...$(NC)"
	@bash scripts/check.sh --skip-tests

# 清理缓存
clean:
	@echo -e "$(BLUE)清理缓存...$(NC)"
	go clean -cache
	@echo -e "$(GREEN)清理完成$(NC)"
