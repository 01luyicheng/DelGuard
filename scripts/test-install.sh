#!/bin/bash

# DelGuard 安装测试脚本
# 自动化测试安装过程的各个环节

# 导入错误处理库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -f "$SCRIPT_DIR/lib/error-handler.sh" ]]; then
    source "$SCRIPT_DIR/lib/error-handler.sh"
    init_error_handler
else
    # 基本错误处理
    set -e
    print_info() { echo "[INFO] $1"; }
    print_success() { echo "[SUCCESS] $1"; }
    print_error() { echo "[ERROR] $1"; }
    print_header() { echo "$1"; echo "$(printf '=%.0s' {1..50})"; }
fi

# 测试配置
TEST_DIR="/tmp/delguard-install-test"
INSTALL_SCRIPT="$SCRIPT_DIR/install.sh"
VERIFY_SCRIPT="$SCRIPT_DIR/verify-install.sh"
REPAIR_SCRIPT="$SCRIPT_DIR/repair-install.sh"

# 测试结果
TESTS_PASSED=0
TESTS_FAILED=0
TEST_RESULTS=()

# 测试函数
run_test() {
    local test_name="$1"
    local test_function="$2"
    
    print_header "🧪 测试: $test_name"
    
    if $test_function; then
        print_success "测试通过: $test_name"
        ((TESTS_PASSED++))
        TEST_RESULTS+=("✅ $test_name")
    else
        print_error "测试失败: $test_name"
        ((TESTS_FAILED++))
        TEST_RESULTS+=("❌ $test_name")
    fi
    
    echo
}

# 准备测试环境
setup_test_environment() {
    print_info "准备测试环境..."
    
    # 创建测试目录
    mkdir -p "$TEST_DIR"
    
    # 备份现有安装
    if command -v delguard &>/dev/null; then
        local existing_delguard=$(command -v delguard)
        print_info "发现现有DelGuard: $existing_delguard"
    fi
    
    return 0
}

# 清理测试环境
cleanup_test_environment() {
    print_info "清理测试环境..."
    
    # 删除测试目录
    if [[ -d "$TEST_DIR" ]]; then
        rm -rf "$TEST_DIR"
        print_success "已删除测试目录"
    fi
    
    return 0
}

# 测试1: 脚本语法检查
test_script_syntax() {
    print_info "检查安装脚本语法..."
    
    local scripts=("$INSTALL_SCRIPT" "$VERIFY_SCRIPT" "$REPAIR_SCRIPT")
    
    for script in "${scripts[@]}"; do
        if [[ -f "$script" ]]; then
            if bash -n "$script" 2>/dev/null; then
                print_success "语法检查通过: $(basename "$script")"
            else
                print_error "语法错误: $(basename "$script")"
                return 1
            fi
        else
            print_warning "脚本不存在: $(basename "$script")"
        fi
    done
    
    return 0
}

# 测试2: 依赖检查
test_dependencies() {
    print_info "检查系统依赖..."
    
    local required_commands=("curl" "tar" "chmod" "mkdir" "grep" "sed")
    local missing_deps=()
    
    for cmd in "${required_commands[@]}"; do
        if command -v "$cmd" &>/dev/null; then
            print_success "依赖可用: $cmd"
        else
            print_error "缺少依赖: $cmd"
            missing_deps+=("$cmd")
        fi
    done
    
    if [[ ${#missing_deps[@]} -eq 0 ]]; then
        return 0
    else
        print_error "缺少 ${#missing_deps[@]} 个依赖: ${missing_deps[*]}"
        return 1
    fi
}

# 测试3: 网络连接
test_network_connectivity() {
    print_info "测试网络连接..."
    
    if curl -s --connect-timeout 10 --head "https://api.github.com" >/dev/null 2>&1; then
        print_success "GitHub API连接正常"
        return 0
    else
        print_error "无法连接到GitHub API"
        return 1
    fi
}

# 测试4: 权限检查
test_permissions() {
    print_info "测试权限..."
    
    # 测试临时目录写入权限
    local temp_file="/tmp/delguard-permission-test"
    if echo "test" > "$temp_file" 2>/dev/null; then
        rm -f "$temp_file"
        print_success "临时目录可写"
    else
        print_error "临时目录不可写"
        return 1
    fi
    
    return 0
}

# 测试5: 错误处理库
test_error_handler() {
    print_info "测试错误处理库..."
    
    if [[ -f "$SCRIPT_DIR/lib/error-handler.sh" ]]; then
        if bash -n "$SCRIPT_DIR/lib/error-handler.sh" 2>/dev/null; then
            print_success "错误处理库语法正确"
            return 0
        else
            print_error "错误处理库语法错误"
            return 1
        fi
    else
        print_error "错误处理库不存在"
        return 1
    fi
}

# 生成测试报告
generate_test_report() {
    print_header "📊 测试报告"
    
    echo "测试完成时间: $(date)"
    echo "通过测试: $TESTS_PASSED"
    echo "失败测试: $TESTS_FAILED"
    echo "总计测试: $((TESTS_PASSED + TESTS_FAILED))"
    echo
    
    echo "详细结果:"
    for result in "${TEST_RESULTS[@]}"; do
        echo "  $result"
    done
    echo
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        print_success "🎉 所有测试通过！安装脚本准备就绪。"
        return 0
    else
        print_error "❌ 有 $TESTS_FAILED 个测试失败，请修复后重试。"
        return 1
    fi
}

# 主函数
main() {
    print_header "🛡️  DelGuard 安装测试套件"
    
    # 准备测试环境
    setup_test_environment
    
    # 运行测试
    run_test "脚本语法检查" test_script_syntax
    run_test "系统依赖检查" test_dependencies
    run_test "网络连接测试" test_network_connectivity
    run_test "权限检查" test_permissions
    run_test "错误处理库测试" test_error_handler
    
    # 清理测试环境
    cleanup_test_environment
    
    # 生成报告
    generate_test_report
}

# 运行主函数
main "$@"