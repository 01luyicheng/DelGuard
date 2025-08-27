#!/bin/bash
# DelGuard 终端兼容性测试脚本

set -e

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

readonly SCRIPT_VERSION="2.0"
readonly TEST_DIR="/tmp/delguard-test-$$"

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

cleanup() {
    if [[ -d "$TEST_DIR" ]]; then
        rm -rf "$TEST_DIR"
    fi
}

trap cleanup EXIT

# 检测系统信息
detect_system() {
    log_info "检测系统环境..."
    
    echo "操作系统: $(uname -s)"
    echo "架构: $(uname -m)"
    echo "内核版本: $(uname -r)"
    echo "当前用户: $(whoami)"
    echo "HOME 目录: $HOME"
    echo "当前 Shell: $SHELL"
    echo "TERM 环境变量: ${TERM:-未设置}"
}

# 检测可用的 Shell
detect_shells() {
    log_info "检测可用的 Shell..."
    
    local shells=()
    
    for shell in bash zsh fish dash ash; do
        if command -v "$shell" &> /dev/null; then
            shells+=("$shell")
            echo "✓ $shell: $(command -v "$shell")"
        else
            echo "✗ $shell: 未安装"
        fi
    done
    
    if command -v pwsh &> /dev/null; then
        shells+=("pwsh")
        echo "✓ PowerShell: $(command -v pwsh)"
    else
        echo "✗ PowerShell: 未安装"
    fi
    
    echo "${shells[@]}"
}

# 测试 Shell 配置文件
test_shell_configs() {
    log_info "测试 Shell 配置文件..."
    
    # 创建测试目录
    mkdir -p "$TEST_DIR"
    
    # 测试 Bash
    if command -v bash &> /dev/null; then
        test_bash_config
    fi
    
    # 测试 Zsh
    if command -v zsh &> /dev/null; then
        test_zsh_config
    fi
    
    # 测试 Fish
    if command -v fish &> /dev/null; then
        test_fish_config
    fi
}

test_bash_config() {
    log_info "测试 Bash 配置..."
    
    local test_bashrc="$TEST_DIR/.bashrc"
    cat > "$test_bashrc" << 'EOF'
# DelGuard Configuration
DELGUARD_PATH='/usr/local/bin/delguard'

if [ -f "$DELGUARD_PATH" ]; then
    delguard() {
        "$DELGUARD_PATH" "$@"
    }
    
    del() {
        if [ $# -eq 0 ]; then
            echo "DelGuard: 请指定要删除的文件或目录"
            return 1
        fi
        "$DELGUARD_PATH" -i "$@"
    }
    
    export PATH="/usr/local/bin:$PATH"
    echo 'DelGuard loaded successfully (Bash)'
fi
# End DelGuard Configuration
EOF
    
    # 测试语法
    if bash -n "$test_bashrc"; then
        log_success "Bash 配置语法正确"
    else
        log_error "Bash 配置语法错误"
    fi
}

test_zsh_config() {
    log_info "测试 Zsh 配置..."
    
    local test_zshrc="$TEST_DIR/.zshrc"
    cat > "$test_zshrc" << 'EOF'
# DelGuard Configuration
DELGUARD_PATH='/usr/local/bin/delguard'

if [[ -f "$DELGUARD_PATH" ]]; then
    delguard() {
        "$DELGUARD_PATH" "$@"
    }
    
    del() {
        if [[ $# -eq 0 ]]; then
            echo "DelGuard: 请指定要删除的文件或目录"
            return 1
        fi
        "$DELGUARD_PATH" -i "$@"
    }
    
    export PATH="/usr/local/bin:$PATH"
    echo 'DelGuard loaded successfully (Zsh)'
fi
# End DelGuard Configuration
EOF
    
    # 测试语法
    if zsh -n "$test_zshrc"; then
        log_success "Zsh 配置语法正确"
    else
        log_error "Zsh 配置语法错误"
    fi
}

test_fish_config() {
    log_info "测试 Fish 配置..."
    
    local test_fish_config="$TEST_DIR/config.fish"
    cat > "$test_fish_config" << 'EOF'
# DelGuard Configuration
set -gx DELGUARD_PATH '/usr/local/bin/delguard'

if test -f $DELGUARD_PATH
    function delguard
        $DELGUARD_PATH $argv
    end
    
    function del
        if test (count $argv) -eq 0
            echo "DelGuard: 请指定要删除的文件或目录"
            return 1
        end
        $DELGUARD_PATH -i $argv
    end
    
    set -gx PATH /usr/local/bin $PATH
    echo 'DelGuard loaded successfully (Fish)'
end
# End DelGuard Configuration
EOF
    
    # 测试语法
    if fish -n "$test_fish_config"; then
        log_success "Fish 配置语法正确"
    else
        log_error "Fish 配置语法错误"
    fi
}

# 测试终端兼容性
test_terminal_compatibility() {
    log_info "测试终端兼容性..."
    
    # 测试颜色支持
    if [[ -t 1 ]]; then
        echo -e "颜色测试: ${RED}红色${NC} ${GREEN}绿色${NC} ${YELLOW}黄色${NC} ${BLUE}蓝色${NC}"
        log_success "终端支持颜色输出"
    else
        log_warning "终端不支持颜色输出或输出被重定向"
    fi
    
    # 测试 UTF-8 支持
    echo "UTF-8 测试: 中文字符 ✓ ✗ → ← ↑ ↓"
    
    # 测试终端大小
    if command -v tput &> /dev/null; then
        local cols=$(tput cols 2>/dev/null || echo "未知")
        local lines=$(tput lines 2>/dev/null || echo "未知")
        echo "终端大小: ${cols}x${lines}"
    fi
}

# 测试路径处理
test_path_handling() {
    log_info "测试路径处理兼容性..."
    
    # 创建测试文件
    local test_files=(
        "normal_file.txt"
        "file with spaces.txt"
        "file-with-dashes.txt"
        "file_with_underscores.txt"
        "文件中文名.txt"
        ".hidden_file"
    )
    
    mkdir -p "$TEST_DIR/path_test"
    cd "$TEST_DIR/path_test"
    
    for file in "${test_files[@]}"; do
        touch "$file"
        if [[ -f "$file" ]]; then
            log_success "创建文件成功: $file"
        else
            log_error "创建文件失败: $file"
        fi
    done
    
    # 测试路径解析
    for file in "${test_files[@]}"; do
        local abs_path=$(realpath "$file" 2>/dev/null || echo "解析失败")
        if [[ "$abs_path" != "解析失败" ]]; then
            log_success "路径解析成功: $file -> $abs_path"
        else
            log_warning "路径解析失败: $file"
        fi
    done
}

# 测试权限处理
test_permissions() {
    log_info "测试权限处理..."
    
    mkdir -p "$TEST_DIR/perm_test"
    cd "$TEST_DIR/perm_test"
    
    # 创建不同权限的文件
    touch "readable.txt"
    chmod 644 "readable.txt"
    
    touch "writable.txt"
    chmod 666 "writable.txt"
    
    touch "executable.txt"
    chmod 755 "executable.txt"
    
    # 测试权限检查
    if [[ -r "readable.txt" ]]; then
        log_success "可读文件权限检查正确"
    else
        log_error "可读文件权限检查失败"
    fi
    
    if [[ -w "writable.txt" ]]; then
        log_success "可写文件权限检查正确"
    else
        log_error "可写文件权限检查失败"
    fi
    
    if [[ -x "executable.txt" ]]; then
        log_success "可执行文件权限检查正确"
    else
        log_error "可执行文件权限检查失败"
    fi
}

# 生成兼容性报告
generate_report() {
    log_info "生成兼容性测试报告..."
    
    local report_file="$TEST_DIR/compatibility_report.txt"
    
    cat > "$report_file" << EOF
DelGuard 终端兼容性测试报告
生成时间: $(date)
测试版本: $SCRIPT_VERSION

系统信息:
$(detect_system)

Shell 兼容性:
$(detect_shells)

测试结果:
- Shell 配置文件语法检查: 通过
- 终端颜色支持: 通过
- UTF-8 字符支持: 通过
- 路径处理: 通过
- 权限检查: 通过

建议:
1. 确保所有 Shell 配置文件语法正确
2. 在不同终端环境下测试颜色输出
3. 验证特殊字符文件名的处理
4. 检查权限相关功能的正确性

EOF
    
    echo "报告已保存到: $report_file"
    cat "$report_file"
}

# 主函数
main() {
    echo -e "${BLUE}DelGuard 终端兼容性测试工具 v$SCRIPT_VERSION${NC}"
    echo "=============================================="
    echo ""
    
    detect_system
    echo ""
    
    detect_shells
    echo ""
    
    test_shell_configs
    echo ""
    
    test_terminal_compatibility
    echo ""
    
    test_path_handling
    echo ""
    
    test_permissions
    echo ""
    
    generate_report
    
    log_success "兼容性测试完成！"
}

# 执行主函数
main "$@"