#!/bin/bash
# DelGuard 快速安装脚本 - 跨平台一键安装
# 支持 Linux/macOS/Windows

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 帮助信息
show_help() {
    cat << EOF
🛡️  DelGuard 快速安装脚本

使用方法: $0 [选项]

选项:
    -v, --version VERSION   指定要安装的版本 (默认: latest)
    -d, --dir DIR         指定安装目录
    -f, --force            强制重新安装
    -h, --help            显示此帮助信息
    --dry-run             模拟安装，不实际执行
    --uninstall           卸载 DelGuard

示例:
    $0                          # 安装最新版本
    $0 --version v1.6.3         # 安装指定版本
    $0 --dir /opt/delguard      # 安装到指定目录
    $0 --force                  # 强制重新安装

环境变量:
    DELGUARD_VERSION        设置默认版本
    DELGUARD_INSTALL_DIR    设置默认安装目录
    DELGUARD_CONFIG_DIR     设置默认配置目录

EOF
}

# 解析参数
parse_args() {
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
            -f|--force)
                FORCE="--force"
                shift
                ;;
            --dry-run)
                DRY_RUN="--dry-run"
                shift
                ;;
            --uninstall)
                UNINSTALL="--uninstall"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo -e "${RED}❌ 未知选项: $1${NC}"
                show_help
                exit 1
                ;;
        esac
    done
}

# 主程序
main() {
    echo -e "${CYAN}🛡️  DelGuard 快速安装脚本${NC}"
    echo -e "==============================${NC}"
    echo
    
    # 解析参数
    parse_args "$@"
    
    # 设置默认值
    VERSION=${VERSION:-${DELGUARD_VERSION:-latest}}
    INSTALL_DIR=${INSTALL_DIR:-${DELGUARD_INSTALL_DIR:-}}
    
    # 构建安装脚本参数
    INSTALL_ARGS=()
    [[ -n "$VERSION" ]] && INSTALL_ARGS+=("--version" "$VERSION")
    [[ -n "$INSTALL_DIR" ]] && INSTALL_ARGS+=("--dir" "$INSTALL_DIR")
    [[ -n "$FORCE" ]] && INSTALL_ARGS+=("$FORCE")
    [[ -n "$DRY_RUN" ]] && INSTALL_ARGS+=("$DRY_RUN")
    [[ -n "$UNINSTALL" ]] && INSTALL_ARGS+=("$UNINSTALL")
    
    # 检测操作系统
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    echo -e "${CYAN}系统信息:${NC}"
    echo -e "  操作系统: $OS"
    echo -e "  架构: $ARCH"
    echo -e "  版本: $VERSION"
    [[ -n "$INSTALL_DIR" ]] && echo -e "  安装目录: $INSTALL_DIR"
    echo
    
    case "$OS" in
        Linux*|Darwin*)
            echo -e "${GREEN}✅ 检测到 Unix-like 系统 ($OS)${NC}"
            echo -e "${CYAN}正在下载并运行安装脚本...${NC}"
            
            # 构建curl命令
            CURL_CMD="curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.sh"
            if [[ ${#INSTALL_ARGS[@]} -gt 0 ]]; then
                CURL_CMD="$CURL_CMD | bash -s -- ${INSTALL_ARGS[*]}"
            else
                CURL_CMD="$CURL_CMD | bash"
            fi
            
            echo -e "${YELLOW}执行命令: $CURL_CMD${NC}"
            echo
            
            if [[ -z "$DRY_RUN" ]]; then
                eval "$CURL_CMD"
            else
                echo -e "${YELLOW}干运行模式 - 仅显示将要执行的命令${NC}"
            fi
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)
            echo -e "${GREEN}✅ 检测到 Windows 系统 ($OS)${NC}"
            echo -e "${CYAN}请使用 PowerShell 运行以下命令:${NC}"
            echo
            
            # 构建PowerShell命令
            PS_ARGS=""
            [[ -n "$VERSION" ]] && PS_ARGS="-Version '$VERSION'"
            [[ -n "$INSTALL_DIR" ]] && PS_ARGS="$PS_ARGS -InstallDir '$INSTALL_DIR'"
            [[ -n "$FORCE" ]] && PS_ARGS="$PS_ARGS -Force"
            [[ -n "$DRY_RUN" ]] && PS_ARGS="$PS_ARGS -DryRun"
            [[ -n "$UNINSTALL" ]] && PS_ARGS="$PS_ARGS -Uninstall"
            
            PS_CMD="iwr -useb https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.ps1 | iex"
            if [[ -n "$PS_ARGS" ]]; then
                PS_CMD="$PS_CMD; Install-DelGuard $PS_ARGS"
            fi
            
            echo -e "${YELLOW}$PS_CMD${NC}"
            echo
            ;;
        *)
            echo -e "${RED}❌ 不支持的操作系统: $OS${NC}"
            exit 1
            ;;
    esac
    
    echo
    echo -e "${GREEN}✅ 操作完成！${NC}"
    
    if [[ -z "$DRY_RUN" ]]; then
        case "$OS" in
            Linux*|Darwin*)
                echo -e "${CYAN}运行 'delguard --help' 开始使用${NC}"
                ;;
            *)
                echo -e "${CYAN}运行 'delguard.exe --help' 开始使用${NC}"
                ;;
        esac
    fi
}

# 运行主程序
main "$@"