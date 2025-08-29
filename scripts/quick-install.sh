#!/bin/bash

# DelGuard 一键安装脚本
# 自动检测系统并选择合适的安装方式

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检测系统类型
detect_system() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        SYSTEM="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        SYSTEM="macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        SYSTEM="windows"
    else
        print_error "不支持的系统类型: $OSTYPE"
        exit 1
    fi
}

# 下载并运行安装脚本
install_delguard() {
    local script_url="https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts"
    
    case $SYSTEM in
        "linux"|"macos")
            print_info "下载 Unix/Linux 安装脚本..."
            curl -fsSL "$script_url/install.sh" | bash
            ;;
        "windows")
            print_info "请在 PowerShell 中运行以下命令:"
            echo "iwr -useb $script_url/install.ps1 | iex"
            ;;
        *)
            print_error "未知系统类型"
            exit 1
            ;;
    esac
}

main() {
    echo "🛡️  DelGuard 一键安装"
    echo "===================="
    echo
    
    detect_system
    print_info "检测到系统: $SYSTEM"
    
    install_delguard
}

main "$@"