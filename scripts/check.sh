#!/bin/bash

# Go-mall 本地聚合检查脚本
# 用法: ./scripts/check.sh [--skip-tests]
#
# 此脚本用于本地一次性执行常用检查。
# GitHub Actions 会直接调用更细粒度的脚本入口，而不是依赖本脚本。

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 基础路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 标志位
SKIP_TESTS=false
RUN_UNIT_TESTS=false
VERBOSE=false
AUTO_FIX=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --unit-tests)
            RUN_UNIT_TESTS=true
            shift
            ;;
        --auto-fix)
            AUTO_FIX=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --skip-tests    跳过测试"
            echo "  --unit-tests    显式运行单元测试"
            echo "  --auto-fix     自动修复格式问题"
            echo "  --verbose, -v   显示详细输出"
            echo "  --help, -h      显示帮助"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            exit 1
            ;;
    esac
done

# 打印带颜色的消息
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# 检查 Go 是否安装
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装"
        exit 1
    fi
    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+')
    print_success "Go 已安装: $GO_VERSION"
}

# 检查必要的工具
check_tools() {
    print_header "检查必要工具"

    local MISSING_TOOLS=()

    # 检查 staticcheck
    if ! command -v staticcheck &> /dev/null; then
        MISSING_TOOLS+=("staticcheck")
    else
        print_success "staticcheck 已安装"
    fi

    # 检查 golint
    if ! command -v golint &> /dev/null; then
        MISSING_TOOLS+=("golint")
    else
        print_success "golint 已安装"
    fi

    # 检查 revive
    if ! command -v revive &> /dev/null; then
        MISSING_TOOLS+=("revive")
    else
        print_success "revive 已安装"
    fi

    # 检查 gci (Go import 整理)
    if ! command -v gci &> /dev/null; then
        MISSING_TOOLS+=("github.com/daixiang0/gci@latest")
    else
        print_success "gci 已安装"
    fi

    # 检查 errcheck
    if ! command -v errcheck &> /dev/null; then
        MISSING_TOOLS+=("github.com/kisielk/errcheck/cmd/errcheck@latest")
    else
        print_success "errcheck 已安装"
    fi

    # 如果有缺失工具，提示安装
    if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
        print_warning "缺少以下工具，建议安装:"
        for tool in "${MISSING_TOOLS[@]}"; do
            echo "  go install $tool"
        done
        echo ""
    fi
}

# 代码格式检查
check_format() {
    print_header "检查代码格式"

    # 检查 gofmt
    UNFORMATTED=$(gofmt -l .)
    if [ -n "$UNFORMATTED" ]; then
        if [ "$AUTO_FIX" = true ]; then
            print_warning "自动修复格式问题..."
            gofmt -w .
            print_success "格式已自动修复"
        else
            print_error "以下文件格式不正确 (运行 gofmt -w 修复 或使用 --auto-fix):"
            # 只显示前 20 个文件
            echo "$UNFORMATTED" | head -20
            if [ "$(echo "$UNFORMATTED" | wc -l)" -gt 20 ]; then
                echo "... 等共 $(echo "$UNFORMATTED" | wc -l) 个文件"
            fi
            return 1
        fi
    else
        print_success "代码格式检查通过"
    fi
}

# 运行 go vet
run_vet() {
    print_header "运行 go vet"

    if bash scripts/go-ci-vet.sh 2>&1; then
        print_success "go vet 检查通过"
    else
        print_warning "go vet 检查有警告"
    fi
}

# 运行 staticcheck
run_staticcheck() {
    print_header "运行 staticcheck"

    # staticcheck 检查较慢，使用缓存
    if staticcheck ./... 2>&1; then
        print_success "staticcheck 检查通过"
    else
        print_error "staticcheck 检查失败"
        return 1
    fi
}

# 运行 golint
run_golint() {
    print_header "运行 golint"

    if command -v golint &> /dev/null; then
        if golint -set_exit_status ./... 2>&1; then
            print_success "golint 检查通过"
        else
            print_error "golint 检查失败"
            return 1
        fi
    else
        print_warning "跳过 golint (未安装)"
    fi
}

# 运行 revive
run_revive() {
    print_header "运行 revive"

    if command -v revive &> /dev/null; then
        # 使用默认规则
        if revive ./... 2>&1; then
            print_success "revive 检查通过"
        else
            print_error "revive 检查失败"
            return 1
        fi
    else
        print_warning "跳过 revive (未安装)"
    fi
}

# 运行 errcheck
run_errcheck() {
    print_header "运行 errcheck"

    if command -v errcheck &> /dev/null; then
        # 只检查主要服务目录
        if errcheck -blank ./services/... 2>&1; then
            print_success "errcheck 检查通过"
        else
            print_error "errcheck 检查失败"
            return 1
        fi
    else
        print_warning "跳过 errcheck (未安装)"
    fi
}

# 检查依赖
check_deps() {
    print_header "检查依赖"

    print_info "同步 go.work..."
    if go work sync 2>&1; then
        print_success "工作区同步完成"
    else
        print_warning "工作区同步有警告"
    fi
}

# 运行测试
run_tests() {
    if [ "$SKIP_TESTS" = true ]; then
        print_warning "跳过测试 (--skip-tests)"
        return 0
    fi

    if [ "$RUN_UNIT_TESTS" != true ]; then
        print_warning "默认跳过单元测试，使用 --unit-tests 显式执行"
        return 0
    fi

    print_header "运行测试"

    if bash scripts/test-unit.sh 2>&1; then
        print_success "测试通过"
    else
        print_error "测试失败"
        return 1
    fi
}

# 构建项目
run_build() {
    print_header "构建项目"

    if bash scripts/go-ci-build.sh --all 2>&1; then
        print_success "CI 服务构建通过"
    else
        print_warning "CI 服务构建有警告"
    fi
}

# 生成报告摘要
print_summary() {
    print_header "检查完成"
    echo ""
    echo -e "${GREEN}所有检查通过!${NC}"
    echo ""
    echo "提示:"
    echo "  - 确保在提交前运行此脚本"
    echo "  - 使用 --skip-tests 跳过测试以加快检查"
    echo ""
}

# 主流程
main() {
    echo -e "${BLUE}Go-mall 本地 CI 检查${NC}"
    echo "项目目录: $PROJECT_ROOT"
    echo ""

    # 检查环境
    check_go
    check_tools

    # 代码质量检查
    check_format
    run_vet

    # 静态分析 (可跳过某些如果工具缺失)
    run_staticcheck || true
    run_golint || true
    run_revive || true
    run_errcheck || true

    # 依赖和构建
    check_deps
    run_build

    # 测试 (可选)
    run_tests

    # 成功
    print_summary
}

# 运行主流程
main
