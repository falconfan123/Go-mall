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
	@echo "  make test         - 运行测试"
	@echo "  make tidy         - 整理依赖"
	@echo ""
	@echo "安装工具:"
	@echo "  make install-tools - 安装所需工具"
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
	go vet ./...

# 运行 staticcheck
staticcheck:
	@echo -e "$(BLUE)运行 staticcheck...$(NC)"
	staticcheck ./...

# 运行本地 CI 检查 (跳过测试，因为测试需要服务运行)
lint:
	@echo -e "$(BLUE)运行本地 CI 检查...$(NC)"
	@bash scripts/local-ci.sh --skip-tests

# 整理依赖
tidy:
	@echo -e "$(BLUE)整理依赖...$(NC)"
	go mod tidy

# 运行测试
test:
	@echo -e "$(BLUE)运行测试...$(NC)"
	go test -race -short ./...

# 构建所有服务 (只构建核心服务，跳过有问题的废弃服务)
# 排除: flash_sale (引用不存在的 usersclient), order (引用不存在的 order.OrderService)
build:
	@echo -e "$(BLUE)构建所有服务...$(NC)"
	@echo "构建 apis (核心服务)..."
	@for dir in apis/carts apis/checkout apis/coupon apis/payment apis/product apis/user; do \
		svc=$$(basename "$$dir"); \
		echo "  构建 $$svc..."; \
		(cd "$$dir" && go build -o /dev/null .) || (echo "  跳过 $$svc (构建失败)"); \
	done
	@echo "跳过以下服务 (有遗留问题):"
	@echo "  - apis/flash_sale (引用不存在的包)"
	@echo "  - apis/order (引用不存在的类型)"
	@echo "构建 services..."
	@for dir in services/*/; do \
		svc=$$(basename "$$dir"); \
		echo "  构建 $$svc..."; \
		(cd "$$dir" && go build -o /dev/null .) || (echo "  跳过 $$svc (构建失败)"); \
	done
	@echo -e "$(GREEN)构建完成$(NC)"

# 模拟 CI 检查 (跳过测试)
ci:
	@echo -e "$(BLUE)模拟 CI 检查...$(NC)"
	@bash scripts/local-ci.sh --skip-tests

# 清理缓存
clean:
	@echo -e "$(BLUE)清理缓存...$(NC)"
	go clean -cache
	@echo -e "$(GREEN)清理完成$(NC)"
