#!/bin/bash

# DelGuard 智能安装脚本
# 支持 Linux 和 macOS 系统自动检测和安装

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 全局变量
DELGUARD_VERSION="latest"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/delguard"
GITHUB_REPO="01luyicheng/DelGuard"
TEMP_DIR="/tmp/delguard-install"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
    else
        print_error "不支持的操作系统: $OSTYPE"
        exit 1
    fi
    print_info "检测到操作系统: $OS"
}

# 检测系统架构
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            print_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    print_info "检测到系统架构: $ARCH"
}

# 检查依赖
check_dependencies() {
    print_info "检查系统依赖..."
    
    # 检查必要的命令
    local deps=("curl" "tar")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            print_error "缺少依赖: $dep"
            print_info "请先安装 $dep 后重试"
            exit 1
        fi
    done
    
    print_success "依赖检查通过"
}

# 检查权限
check_permissions() {
    print_info "检查安装权限..."
    
    if [[ ! -w "$INSTALL_DIR" ]]; then
        print_warning "需要管理员权限安装到 $INSTALL_DIR"
        if [[ $EUID -ne 0 ]]; then
            print_info "请使用 sudo 运行此脚本"
            exit 1
        fi
    fi
    
    print_success "权限检查通过"
}

# 获取最新版本
get_latest_version() {
    print_info "获取最新版本信息..."
    
    if [[ "$DELGUARD_VERSION" == "latest" ]]; then
        DELGUARD_VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [[ -z "$DELGUARD_VERSION" ]]; then
            print_error "无法获取最新版本信息"
            exit 1
        fi
    fi
    
    print_info "目标版本: $DELGUARD_VERSION"
}

# 下载二进制文件
download_binary() {
    print_info "下载 DelGuard 二进制文件..."
    
    # 构建下载URL - 直接使用裸二进制文件
    local binary_name="delguard-${OS}-${ARCH}"
    local download_url="https://github.com/$GITHUB_REPO/releases/download/$DELGUARD_VERSION/$binary_name"
    
    # 创建临时目录
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    print_info "下载地址: $download_url"
    
    # 下载文件
    if ! curl -L -o "delguard" "$download_url"; then
        print_error "下载失败"
        cleanup
        exit 1
    fi
    
    # 设置执行权限
    chmod +x "delguard"
    
    print_success "下载完成"
}

# 安装二进制文件
install_binary() {
    print_info "安装 DelGuard..."
    
    # 直接使用下载的二进制文件
    if [[ ! -f "delguard" ]]; then
        print_error "找不到二进制文件"
        cleanup
        exit 1
    fi
    
    # 复制到安装目录
    if ! cp "delguard" "$INSTALL_DIR/delguard"; then
        print_error "安装失败"
        cleanup
        exit 1
    fi
    
    # 设置执行权限
    chmod +x "$INSTALL_DIR/delguard"
    
    print_success "DelGuard 已安装到 $INSTALL_DIR/delguard"
}

# 创建配置目录
create_config() {
    print_info "创建配置目录..."
    
    mkdir -p "$CONFIG_DIR"
    
    # 创建默认配置文件
    cat > "$CONFIG_DIR/config.yaml" << EOF
# DelGuard 配置文件
verbose: false
force: false
quiet: false

# 回收站设置
trash:
  auto_clean: false
  max_days: 30
  max_size: "1GB"

# 日志设置
log:
  level: "info"
  file: "$CONFIG_DIR/delguard.log"
EOF
    
    print_success "配置目录已创建: $CONFIG_DIR"
}

# 配置Shell别名
setup_shell_aliases() {
    print_info "配置Shell别名..."
    
    local shell_configs=()
    
    # 检测用户使用的Shell
    if [[ -n "$BASH_VERSION" ]] || [[ "$SHELL" == *"bash"* ]]; then
        shell_configs+=("$HOME/.bashrc" "$HOME/.bash_profile")
    fi
    
    if [[ -n "$ZSH_VERSION" ]] || [[ "$SHELL" == *"zsh"* ]]; then
        shell_configs+=("$HOME/.zshrc" "$HOME/.zprofile")
    fi
    
    # 添加通用配置文件
    shell_configs+=("$HOME/.profile")
    
    local alias_content="
# DelGuard 别名配置
alias del='delguard delete'
alias rm='delguard delete'
alias trash='delguard delete'
alias restore='delguard restore'
alias empty-trash='delguard empty'
"
    
    for config_file in "${shell_configs[@]}"; do
        if [[ -f "$config_file" ]]; then
            # 检查是否已经配置过
            if ! grep -q "DelGuard 别名配置" "$config_file" 2>/dev/null; then
                echo "$alias_content" >> "$config_file"
                print_info "已添加别名到: $config_file"
            fi
        fi
    done
    
    print_success "Shell别名配置完成"
}

# 验证安装
verify_installation() {
    print_info "验证安装..."
    
    # 检查二进制文件
    if [[ ! -f "$INSTALL_DIR/delguard" ]]; then
        print_error "二进制文件不存在"
        return 1
    fi
    
    # 检查执行权限
    if [[ ! -x "$INSTALL_DIR/delguard" ]]; then
        print_error "二进制文件没有执行权限"
        return 1
    fi
    
    # 测试运行
    if ! "$INSTALL_DIR/delguard" --version &>/dev/null; then
        print_warning "无法运行 delguard --version，但文件已安装"
    fi
    
    print_success "安装验证通过"
}

# 清理临时文件
cleanup() {
    if [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
        print_info "已清理临时文件"
    fi
}

# 显示安装完成信息
show_completion_info() {
    print_success "🎉 DelGuard 安装完成！"
    echo
    echo "📍 安装位置: $INSTALL_DIR/delguard"
    echo "📁 配置目录: $CONFIG_DIR"
    echo
    echo "🚀 快速开始:"
    echo "  delguard --help          # 查看帮助"
    echo "  delguard delete <file>   # 删除文件到回收站"
    echo "  delguard list           # 查看回收站内容"
    echo "  delguard restore <file> # 恢复文件"
    echo "  delguard empty          # 清空回收站"
    echo
    echo "💡 别名已配置 (重新打开终端后生效):"
    echo "  del <file>     # 等同于 delguard delete"
    echo "  rm <file>      # 等同于 delguard delete (安全替代)"
    echo "  restore <file> # 等同于 delguard restore"
    echo "  empty-trash    # 等同于 delguard empty"
    echo
    echo "📖 更多信息: https://github.com/$GITHUB_REPO"
}

# 主函数
main() {
    echo "🛡️  DelGuard 智能安装脚本"
    echo "================================"
    echo
    
    # 检查系统环境
    detect_os
    detect_arch
    check_dependencies
    check_permissions
    
    # 下载和安装
    get_latest_version
    download_binary
    install_binary
    
    # 配置
    create_config
    setup_shell_aliases
    
    # 验证和清理
    verify_installation
    cleanup
    
    # 显示完成信息
    show_completion_info
}

# 错误处理
trap cleanup EXIT

# 运行主函数
main "$@"