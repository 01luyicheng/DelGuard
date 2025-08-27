#!/bin/bash

# DelGuard 一键安装脚本
# 支持 Linux/macOS/Windows(WSL)

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
VERSION="latest"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="delguard"

print_banner() {
    echo -e "${BLUE}"
    echo "  ____       _       _     ____                      _     "
    echo " |  _ \  ___| |_ ___| |_  / ___|  ___  _   _ _ __ | |_ "
    echo " | | | |/ _ \ __/ _ \ __| \___ \ / _ \| | | | '_ \| __|"
    echo " | |_| |  __/ ||  __/ |_   ___) | (_) | |_| | | | | |_ "
    echo " |____/ \___|\__\___|\__| |____/ \___/ \__,_|_| |_|\__|"
    echo -e "${NC}"
    echo "  一键安装脚本 - 跨平台文件安全删除工具"
    echo ""
}

show_help() {
    echo "使用方法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -v, --version VERSION    指定版本 (默认: latest)"
    echo "  -d, --dir DIRECTORY      安装目录 (默认: /usr/local/bin)"
    echo "  -h, --help               显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                       # 安装最新版本"
    echo "  $0 -v v1.0.0            # 安装指定版本"
    echo "  $0 -d ~/.local/bin        # 安装到用户目录"
}

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

check_requirements() {
    log "检查系统要求..."
    
    # 检查 curl 或 wget
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        error "需要 curl 或 wget 来下载文件"
        exit 1
    fi
    
    # 检查系统架构
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        i686|i386) ARCH="386" ;;
        *) error "不支持的架构: $ARCH"; exit 1 ;;
    esac
    
    # 检查操作系统
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $OS in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        *) error "不支持的操作系统: $OS"; exit 1 ;;
    esac
    
    log "检测到: $OS-$ARCH"
}

download_binary() {
    local version=$1
    local os=$2
    local arch=$3
    local temp_dir=$(mktemp -d)
    
    local filename="delguard_${os}_${arch}"
    if [ "$os" = "windows" ]; then
        filename="${filename}.exe"
    fi
    
    local download_url="https://github.com/yourusername/DelGuard/releases/download/${version}/${filename}"
    
    log "下载 DelGuard ${version}..."
    
    if command -v curl &> /dev/null; then
        curl -L -o "${temp_dir}/delguard" "${download_url}"
    else
        wget -O "${temp_dir}/delguard" "${download_url}"
    fi
    
    if [ ! -f "${temp_dir}/delguard" ]; then
        error "下载失败"
        exit 1
    fi
    
    chmod +x "${temp_dir}/delguard"
    echo "$temp_dir"
}

install_binary() {
    local temp_dir=$1
    local install_dir=$2
    
    log "安装到 ${install_dir}..."
    
    # 创建安装目录
    if [ ! -d "$install_dir" ]; then
        mkdir -p "$install_dir"
    fi
    
    # 移动二进制文件
    mv "${temp_dir}/delguard" "${install_dir}/${BINARY_NAME}"
    
    # 清理临时目录
    rm -rf "$temp_dir"
    
    log "安装完成！"
}

setup_path() {
    local install_dir=$1
    
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        warn "建议将 ${install_dir} 添加到 PATH 环境变量"
        echo "临时解决方案:"
        echo "  export PATH=\"\$PATH:${install_dir}\""
        echo ""
        echo "永久解决方案 (添加到 ~/.bashrc 或 ~/.zshrc):"
        echo "  echo 'export PATH=\"\$PATH:${install_dir}\"' >> ~/.bashrc"
        echo "  source ~/.bashrc"
    fi
}

main() {
    print_banner
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 检查是否以root权限运行
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ "$EUID" -ne 0 ]; then
        warn "需要root权限安装到系统目录"
        echo "使用 sudo $0 重新运行，或指定用户目录:"
        echo "  $0 -d ~/.local/bin"
        exit 1
    fi
    
    check_requirements
    
    local temp_dir=$(download_binary "$VERSION" "$OS" "$ARCH")
    install_binary "$temp_dir" "$INSTALL_DIR"
    setup_path "$INSTALL_DIR"
    
    log "DelGuard 已成功安装！"
    echo ""
    echo "使用方法:"
    echo "  delguard --help     # 查看帮助"
    echo "  delguard file.txt   # 安全删除文件"
    echo "  delguard --restore  # 恢复文件"
}

main "$@"