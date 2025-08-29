#!/bin/bash

# DelGuard 安装验证脚本
# 验证安装完整性和功能正确性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 全局变量
DELGUARD_BINARY=""
CONFIG_DIR=""
INSTALL_DIR=""
ERRORS=0
WARNINGS=0

# 打印函数
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
    ((WARNINGS++))
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
    ((ERRORS++))
}

print_header() {
    echo -e "${BLUE}$1${NC}"
    echo "$(printf '=%.0s' {1..50})"
}

# 检测系统环境
detect_environment() {
    print_header "🔍 检测系统环境"
    
    # 检测操作系统
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        CONFIG_DIR="$HOME/.config/delguard"
        INSTALL_DIR="/usr/local/bin"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
        CONFIG_DIR="$HOME/.config/delguard"
        INSTALL_DIR="/usr/local/bin"
    else
        print_error "不支持的操作系统: $OSTYPE"
        return 1
    fi
    
    print_success "操作系统: $OS"
    
    # 检测架构
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) print_warning "未知架构: $ARCH" ;;
    esac
    
    print_success "系统架构: $ARCH"
    
    # 查找DelGuard二进制文件
    if command -v delguard &> /dev/null; then
        DELGUARD_BINARY=$(command -v delguard)
        print_success "找到DelGuard: $DELGUARD_BINARY"
    else
        print_error "未找到DelGuard二进制文件"
        return 1
    fi
    
    echo
}

# 验证二进制文件
verify_binary() {
    print_header "🔧 验证二进制文件"
    
    # 检查文件存在性
    if [[ ! -f "$DELGUARD_BINARY" ]]; then
        print_error "二进制文件不存在: $DELGUARD_BINARY"
        return 1
    fi
    print_success "二进制文件存在"
    
    # 检查执行权限
    if [[ ! -x "$DELGUARD_BINARY" ]]; then
        print_error "二进制文件没有执行权限"
        return 1
    fi
    print_success "具有执行权限"
    
    # 检查文件大小
    local file_size=$(stat -c%s "$DELGUARD_BINARY" 2>/dev/null || stat -f%z "$DELGUARD_BINARY" 2>/dev/null)
    if [[ $file_size -lt 1000000 ]]; then  # 小于1MB可能有问题
        print_warning "二进制文件大小异常: ${file_size} bytes"
    else
        print_success "文件大小正常: ${file_size} bytes"
    fi
    
    # 检查文件类型
    local file_type=$(file "$DELGUARD_BINARY" 2>/dev/null || echo "unknown")
    if [[ "$file_type" == *"executable"* ]] || [[ "$file_type" == *"ELF"* ]]; then
        print_success "文件类型正确: executable"
    else
        print_warning "文件类型可能异常: $file_type"
    fi
    
    echo
}

# 验证基本功能
verify_basic_functionality() {
    print_header "⚡ 验证基本功能"
    
    # 测试版本命令
    print_info "测试 --version 命令..."
    if version_output=$("$DELGUARD_BINARY" --version 2>&1); then
        print_success "版本命令正常: $version_output"
    else
        print_error "版本命令失败: $version_output"
    fi
    
    # 测试帮助命令
    print_info "测试 --help 命令..."
    if help_output=$("$DELGUARD_BINARY" --help 2>&1); then
        if [[ "$help_output" == *"Usage"* ]] || [[ "$help_output" == *"Commands"* ]]; then
            print_success "帮助命令正常"
        else
            print_warning "帮助输出格式异常"
        fi
    else
        print_error "帮助命令失败"
    fi
    
    # 测试子命令存在性
    local commands=("delete" "restore" "list" "empty")
    for cmd in "${commands[@]}"; do
        if "$DELGUARD_BINARY" "$cmd" --help &>/dev/null; then
            print_success "子命令 '$cmd' 可用"
        else
            print_error "子命令 '$cmd' 不可用"
        fi
    done
    
    echo
}

# 验证配置系统
verify_configuration() {
    print_header "⚙️ 验证配置系统"
    
    # 检查配置目录
    if [[ -d "$CONFIG_DIR" ]]; then
        print_success "配置目录存在: $CONFIG_DIR"
        
        # 检查配置文件
        local config_file="$CONFIG_DIR/config.yaml"
        if [[ -f "$config_file" ]]; then
            print_success "配置文件存在"
            
            # 验证配置文件格式
            if command -v python3 &>/dev/null; then
                if python3 -c "import yaml; yaml.safe_load(open('$config_file'))" 2>/dev/null; then
                    print_success "配置文件格式正确"
                else
                    print_warning "配置文件格式可能有问题"
                fi
            elif command -v yq &>/dev/null; then
                if yq eval '.' "$config_file" &>/dev/null; then
                    print_success "配置文件格式正确"
                else
                    print_warning "配置文件格式可能有问题"
                fi
            else
                print_info "跳过配置文件格式验证（缺少yaml解析器）"
            fi
        else
            print_warning "配置文件不存在，将使用默认配置"
        fi
        
        # 检查目录权限
        if [[ -w "$CONFIG_DIR" ]]; then
            print_success "配置目录可写"
        else
            print_error "配置目录不可写"
        fi
    else
        print_warning "配置目录不存在: $CONFIG_DIR"
    fi
    
    echo
}

# 验证PATH环境变量
verify_path() {
    print_header "🛤️ 验证PATH环境变量"
    
    # 检查是否在PATH中
    if command -v delguard &>/dev/null; then
        print_success "DelGuard在PATH中"
        
        local which_delguard=$(command -v delguard)
        if [[ "$which_delguard" == "$DELGUARD_BINARY" ]]; then
            print_success "PATH中的DelGuard指向正确位置"
        else
            print_warning "PATH中的DelGuard指向: $which_delguard (期望: $DELGUARD_BINARY)"
        fi
    else
        print_error "DelGuard不在PATH中"
    fi
    
    # 检查安装目录是否在PATH中
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        print_success "安装目录在PATH中: $INSTALL_DIR"
    else
        print_warning "安装目录不在PATH中: $INSTALL_DIR"
    fi
    
    echo
}

# 验证Shell别名
verify_aliases() {
    print_header "🔗 验证Shell别名"
    
    local shell_configs=()
    local current_shell=$(basename "$SHELL")
    
    # 根据当前Shell添加配置文件
    case "$current_shell" in
        bash)
            shell_configs+=("$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile")
            ;;
        zsh)
            shell_configs+=("$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.profile")
            ;;
        *)
            shell_configs+=("$HOME/.profile")
            ;;
    esac
    
    local aliases_found=false
    for config_file in "${shell_configs[@]}"; do
        if [[ -f "$config_file" ]] && grep -q "DelGuard 别名配置" "$config_file" 2>/dev/null; then
            print_success "在 $config_file 中找到别名配置"
            aliases_found=true
        fi
    done
    
    if [[ "$aliases_found" == false ]]; then
        print_warning "未找到Shell别名配置"
    fi
    
    # 测试别名是否生效（在当前会话中）
    local aliases=("del" "trash" "restore" "empty-trash")
    for alias_name in "${aliases[@]}"; do
        if alias "$alias_name" &>/dev/null; then
            print_success "别名 '$alias_name' 已生效"
        else
            print_info "别名 '$alias_name' 未在当前会话中生效（需要重新加载Shell）"
        fi
    done
    
    echo
}

# 验证权限
verify_permissions() {
    print_header "🔐 验证权限"
    
    # 检查二进制文件权限
    local binary_perms=$(stat -c "%a" "$DELGUARD_BINARY" 2>/dev/null || stat -f "%Lp" "$DELGUARD_BINARY" 2>/dev/null)
    if [[ "$binary_perms" =~ ^[0-9]*[1357]$ ]] || [[ "$binary_perms" =~ ^[0-9]*[1357][0-9]*$ ]]; then
        print_success "二进制文件具有执行权限: $binary_perms"
    else
        print_error "二进制文件权限异常: $binary_perms"
    fi
    
    # 检查配置目录权限
    if [[ -d "$CONFIG_DIR" ]]; then
        if [[ -r "$CONFIG_DIR" && -w "$CONFIG_DIR" ]]; then
            print_success "配置目录权限正常"
        else
            print_error "配置目录权限不足"
        fi
    fi
    
    # 检查是否需要sudo权限
    if [[ "$INSTALL_DIR" == "/usr/local/bin" ]] || [[ "$INSTALL_DIR" == "/usr/bin" ]]; then
        if [[ $EUID -eq 0 ]]; then
            print_info "当前以root权限运行"
        else
            print_info "系统安装，可能需要sudo权限进行更新"
        fi
    fi
    
    echo
}

# 功能测试
test_functionality() {
    print_header "🧪 功能测试"
    
    # 创建测试目录
    local test_dir="/tmp/delguard-test-$$"
    mkdir -p "$test_dir"
    
    # 创建测试文件
    local test_file="$test_dir/test-file.txt"
    echo "DelGuard测试文件" > "$test_file"
    
    print_info "创建测试文件: $test_file"
    
    # 测试删除功能
    print_info "测试删除功能..."
    if "$DELGUARD_BINARY" delete "$test_file" --force 2>/dev/null; then
        if [[ ! -f "$test_file" ]]; then
            print_success "删除功能正常"
        else
            print_error "删除功能异常：文件仍然存在"
        fi
    else
        print_warning "删除功能测试失败（可能需要配置回收站）"
    fi
    
    # 测试列表功能
    print_info "测试列表功能..."
    if "$DELGUARD_BINARY" list &>/dev/null; then
        print_success "列表功能正常"
    else
        print_warning "列表功能测试失败"
    fi
    
    # 清理测试文件
    rm -rf "$test_dir"
    print_info "清理测试文件"
    
    echo
}

# 生成报告
generate_report() {
    print_header "📊 验证报告"
    
    echo "验证完成时间: $(date)"
    echo "DelGuard位置: $DELGUARD_BINARY"
    echo "配置目录: $CONFIG_DIR"
    echo "安装目录: $INSTALL_DIR"
    echo
    
    if [[ $ERRORS -eq 0 && $WARNINGS -eq 0 ]]; then
        print_success "🎉 所有验证通过！DelGuard安装完全正常。"
        return 0
    elif [[ $ERRORS -eq 0 ]]; then
        print_warning "⚠️ 验证完成，有 $WARNINGS 个警告。DelGuard基本功能正常。"
        return 0
    else
        print_error "❌ 验证失败，发现 $ERRORS 个错误和 $WARNINGS 个警告。"
        echo
        echo "建议操作："
        echo "1. 重新运行安装脚本"
        echo "2. 检查系统权限"
        echo "3. 查看安装日志"
        echo "4. 联系技术支持"
        return 1
    fi
}

# 主函数
main() {
    echo "🛡️  DelGuard 安装验证"
    echo "======================"
    echo
    
    # 执行所有验证步骤
    detect_environment || exit 1
    verify_binary
    verify_basic_functionality
    verify_configuration
    verify_path
    verify_aliases
    verify_permissions
    test_functionality
    
    # 生成报告
    generate_report
}

# 运行主函数
main "$@"