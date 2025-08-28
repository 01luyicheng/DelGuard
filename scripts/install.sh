#!/bin/bash
# DelGuard 通用安装脚本
# 自动检测操作系统并调用相应的安装脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 显示帮助信息
show_help() {
    echo "DelGuard 通用安装脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --help                 显示帮助信息"
    echo "  --silent               静默安装"
    echo "  --install-path PATH    安装路径"
    echo "  --user-service         安装为用户服务"
    echo ""
    echo "支持的操作系统:"
    echo "  - Windows (PowerShell)"
    echo "  - Linux (各种发行版)"
    echo "  - macOS"
    echo ""
    echo "示例:"
    echo "  $0                     # 交互式安装"
    echo "  $0 --silent            # 静默安装"
    echo "  $0 --user-service      # 安装为用户服务"
}

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
        OS="windows"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    else
        log_error "不支持的操作系统: $OSTYPE"
        log_info "支持的操作系统: Windows, Linux, macOS"
        exit 1
    fi
}

# 检查安装脚本是否存在
check_install_scripts() {
    case $OS in
        windows)
            if [[ ! -f "$SCRIPT_DIR/install_windows.ps1" ]]; then
                log_error "Windows安装脚本不存在: $SCRIPT_DIR/install_windows.ps1"
                exit 1
            fi
            ;;
        linux|macos)
            if [[ ! -f "$SCRIPT_DIR/install_unix.sh" ]]; then
                log_error "Unix安装脚本不存在: $SCRIPT_DIR/install_unix.sh"
                exit 1
            fi
            ;;
    esac
}

# 主函数
main() {
    # 解析命令行参数
    local args=()
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help)
                show_help
                exit 0
                ;;
            *)
                args+=("$1")
                shift
                ;;
        esac
    done
    
    echo -e "${CYAN}"
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║                DelGuard 通用安装程序                         ║
║                     版本: 2.0.0                             ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    # 检测操作系统
    detect_os
    log_info "检测到操作系统: $OS"
    
    # 检查安装脚本
    check_install_scripts
    
    # 调用相应的安装脚本
    case $OS in
        windows)
            log_info "启动Windows安装脚本..."
            if command -v powershell >/dev/null 2>&1; then
                powershell -ExecutionPolicy Bypass -File "$SCRIPT_DIR/install_windows.ps1" "${args[@]}"
            elif command -v pwsh >/dev/null 2>&1; then
                pwsh -ExecutionPolicy Bypass -File "$SCRIPT_DIR/install_windows.ps1" "${args[@]}"
            else
                log_error "未找到PowerShell，无法执行Windows安装脚本"
                log_info "请安装PowerShell或直接运行: install_windows.ps1"
                exit 1
            fi
            ;;
        linux|macos)
            log_info "启动Unix安装脚本..."
            chmod +x "$SCRIPT_DIR/install_unix.sh"
            "$SCRIPT_DIR/install_unix.sh" "${args[@]}"
            ;;
    esac
    
    log_success "安装完成！"
}

# 执行主函数
main "$@"