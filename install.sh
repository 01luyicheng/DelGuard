#!/bin/bash

# DelGuard 安装脚本 - Unix版本 (Linux/macOS)
# 自动下载并安装 DelGuard 安全删除工具

set -e

# 常量定义
readonly REPO_URL="https://github.com/01luyicheng/DelGuard"
readonly RELEASE_API="https://api.github.com/repos/01luyicheng/DelGuard/releases/latest"
readonly APP_NAME="DelGuard"
readonly EXECUTABLE_NAME="delguard"

# 颜色定义
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# 全局变量
FORCE_INSTALL=false
SYSTEM_WIDE=false
UNINSTALL=false
STATUS_CHECK=false
VERBOSE=false

# 路径配置
if [[ "$SYSTEM_WIDE" == "true" ]]; then
    INSTALL_DIR="/usr/local/bin"
    CONFIG_DIR="/etc/delguard"
else
    INSTALL_DIR="$HOME/.local/bin"
    CONFIG_DIR="$HOME/.config/delguard"
fi

EXECUTABLE_PATH="$INSTALL_DIR/$EXECUTABLE_NAME"
LOG_FILE="$CONFIG_DIR/install.log"

# 日志函数
log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local log_message="[$timestamp] [$level] $message"
    
    case "$level" in
        "ERROR")
            echo -e "${RED}$log_message${NC}" >&2
            ;;
        "WARNING")
            echo -e "${YELLOW}$log_message${NC}"
            ;;
        "SUCCESS")
            echo -e "${GREEN}$log_message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}$log_message${NC}"
            ;;
        *)
            echo "$log_message"
            ;;
    esac
    
    # 写入日志文件
    mkdir -p "$(dirname "$LOG_FILE")"
    echo "$log_message" >> "$LOG_FILE"
}

# 显示帮助信息
show_help() {
    cat << EOF
DelGuard 安装脚本

用法: $0 [选项]

选项:
    -f, --force         强制重新安装
    -s, --system        系统级安装 (需要 sudo)
    -u, --uninstall     卸载 DelGuard
    --status            检查安装状态
    -v, --verbose       详细输出
    -h, --help          显示此帮助信息

示例:
    $0                  # 标准安装
    $0 --force          # 强制重新安装
    $0 --system         # 系统级安装
    $0 --uninstall      # 卸载

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--force)
                FORCE_INSTALL=true
                shift
                ;;
            -s|--system)
                SYSTEM_WIDE=true
                INSTALL_DIR="/usr/local/bin"
                CONFIG_DIR="/etc/delguard"
                EXECUTABLE_PATH="$INSTALL_DIR/$EXECUTABLE_NAME"
                LOG_FILE="$CONFIG_DIR/install.log"
                shift
                ;;
            -u|--uninstall)
                UNINSTALL=true
                shift
                ;;
            --status)
                STATUS_CHECK=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log "ERROR" "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检测操作系统
detect_os() {
    case "$(uname -s)" in
        Linux*)
            echo "linux"
            ;;
        Darwin*)
            echo "darwin"
            ;;
        *)
            log "ERROR" "不支持的操作系统: $(uname -s)"
            exit 1
            ;;
    esac
}

# 检测系统架构
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7l)
            echo "arm"
            ;;
        i386|i686)
            echo "386"
            ;;
        *)
            log "ERROR" "不支持的架构: $(uname -m)"
            exit 1
            ;;
    esac
}

# 检查依赖
check_dependencies() {
    local deps=("curl" "tar" "grep")
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log "ERROR" "缺少依赖: $dep"
            exit 1
        fi
    done
}

# 检查网络连接
check_network() {
    if ! curl -s --head "https://api.github.com" > /dev/null; then
        log "ERROR" "无法连接到 GitHub，请检查网络连接"
        exit 1
    fi
}

# 检查权限
check_permissions() {
    if [[ "$SYSTEM_WIDE" == "true" ]]; then
        if [[ $EUID -ne 0 ]]; then
            log "ERROR" "系统级安装需要 root 权限，请使用 sudo"
            exit 1
        fi
    fi
    
    # 检查安装目录权限
    if [[ ! -w "$(dirname "$INSTALL_DIR")" ]]; then
        log "ERROR" "没有写入权限: $(dirname "$INSTALL_DIR")"
        exit 1
    fi
}

# 获取最新版本信息
get_latest_release() {
    log "INFO" "获取最新版本信息..."
    
    local response
    response=$(curl -s "$RELEASE_API")
    
    if [[ -z "$response" ]]; then
        log "ERROR" "无法获取版本信息"
        exit 1
    fi
    
    echo "$response"
}

# 下载文件
download_file() {
    local url="$1"
    local output="$2"
    
    log "INFO" "下载文件: $url"
    
    if [[ "$VERBOSE" == "true" ]]; then
        curl -L -o "$output" "$url"
    else
        curl -s -L -o "$output" "$url"
    fi
    
    if [[ $? -ne 0 ]]; then
        log "ERROR" "下载失败"
        exit 1
    fi
    
    log "SUCCESS" "下载完成: $output"
}

# 安装 DelGuard
install_delguard() {
    log "INFO" "开始安装 $APP_NAME..."
    
    # 检查现有安装
    if [[ -f "$EXECUTABLE_PATH" ]] && [[ "$FORCE_INSTALL" != "true" ]]; then
        log "INFO" "$APP_NAME 已经安装在 $EXECUTABLE_PATH"
        log "INFO" "使用 --force 参数强制重新安装"
        return 0
    fi
    
    # 获取系统信息
    local os=$(detect_os)
    local arch=$(detect_arch)
    
    log "INFO" "检测到系统: $os-$arch"
    
    # 获取最新版本
    local release_info=$(get_latest_release)
    local version=$(echo "$release_info" | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)
    
    if [[ -z "$version" ]]; then
        log "ERROR" "无法解析版本信息"
        exit 1
    fi
    
    log "INFO" "最新版本: $version"
    
    # 构建下载URL
    local asset_name="${APP_NAME}-${os}-${arch}.tar.gz"
    local download_url=$(echo "$release_info" | grep -o "\"browser_download_url\": \"[^\"]*${asset_name}\"" | cut -d'"' -f4)
    
    if [[ -z "$download_url" ]]; then
        log "ERROR" "未找到适合的安装包: $asset_name"
        exit 1
    fi
    
    log "INFO" "下载URL: $download_url"
    
    # 创建临时目录
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # 下载文件
    local archive_path="$temp_dir/$asset_name"
    download_file "$download_url" "$archive_path"
    
    # 解压文件
    log "INFO" "解压安装包..."
    tar -xzf "$archive_path" -C "$temp_dir"
    
    # 查找可执行文件
    local executable_source=$(find "$temp_dir" -name "$EXECUTABLE_NAME" -type f | head -1)
    
    if [[ -z "$executable_source" ]]; then
        log "ERROR" "在安装包中未找到可执行文件"
        exit 1
    fi
    
    # 创建安装目录
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    
    # 复制可执行文件
    cp "$executable_source" "$EXECUTABLE_PATH"
    chmod +x "$EXECUTABLE_PATH"
    
    log "SUCCESS" "已安装到: $EXECUTABLE_PATH"
    
    # 添加到 PATH
    add_to_path
    
    # 安装 shell 别名
    install_shell_aliases
    
    log "SUCCESS" "$APP_NAME $version 安装成功！"
    log "INFO" "可执行文件位置: $EXECUTABLE_PATH"
    log "INFO" "配置目录: $CONFIG_DIR"
    log "INFO" ""
    log "INFO" "使用方法:"
    log "INFO" "  delguard file.txt          # 删除文件到回收站"
    log "INFO" "  delguard -p file.txt       # 永久删除文件"
    log "INFO" "  delguard --help            # 查看帮助"
    log "INFO" ""
    log "INFO" "请重新启动终端或运行 'source ~/.bashrc' 以使用 delguard 命令"
}

# 添加到 PATH
add_to_path() {
    local shell_rc=""
    
    # 检测 shell 类型
    if [[ -n "$BASH_VERSION" ]]; then
        shell_rc="$HOME/.bashrc"
    elif [[ -n "$ZSH_VERSION" ]]; then
        shell_rc="$HOME/.zshrc"
    else
        shell_rc="$HOME/.profile"
    fi
    
    # 检查是否已在 PATH 中
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        log "INFO" "PATH 中已存在: $INSTALL_DIR"
        return 0
    fi
    
    # 添加到 shell 配置文件
    if [[ -f "$shell_rc" ]]; then
        if ! grep -q "$INSTALL_DIR" "$shell_rc"; then
            echo "" >> "$shell_rc"
            echo "# DelGuard PATH" >> "$shell_rc"
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$shell_rc"
            log "SUCCESS" "已添加到 PATH: $shell_rc"
        else
            log "INFO" "PATH 配置已存在: $shell_rc"
        fi
    fi
    
    # 更新当前会话的 PATH
    export PATH="$PATH:$INSTALL_DIR"
}

# 安装 shell 别名
install_shell_aliases() {
    local shell_rc=""
    
    # 检测 shell 类型
    if [[ -n "$BASH_VERSION" ]]; then
        shell_rc="$HOME/.bashrc"
    elif [[ -n "$ZSH_VERSION" ]]; then
        shell_rc="$HOME/.zshrc"
    else
        shell_rc="$HOME/.profile"
    fi
    
    if [[ -f "$shell_rc" ]]; then
        if ! grep -q "DelGuard 别名配置" "$shell_rc"; then
            cat >> "$shell_rc" << EOF

# DelGuard 别名配置
if command -v delguard &> /dev/null; then
    alias dg='delguard'
    alias del='delguard'
    alias rm='delguard'
fi
EOF
            log "SUCCESS" "已添加 shell 别名配置"
        else
            log "INFO" "shell 别名已存在"
        fi
    fi
}

# 卸载 DelGuard
uninstall_delguard() {
    log "INFO" "开始卸载 $APP_NAME..."
    
    # 删除可执行文件
    if [[ -f "$EXECUTABLE_PATH" ]]; then
        rm -f "$EXECUTABLE_PATH"
        log "SUCCESS" "已删除: $EXECUTABLE_PATH"
    fi
    
    # 删除安装目录（如果为空）
    if [[ -d "$INSTALL_DIR" ]] && [[ -z "$(ls -A "$INSTALL_DIR")" ]]; then
        rmdir "$INSTALL_DIR"
        log "SUCCESS" "已删除安装目录: $INSTALL_DIR"
    fi
    
    # 从 PATH 中移除
    remove_from_path
    
    # 移除 shell 别名
    remove_shell_aliases
    
    log "SUCCESS" "$APP_NAME 卸载完成"
    log "INFO" "配置文件保留在: $CONFIG_DIR"
    log "INFO" "如需完全清理，请手动删除配置目录"
}

# 从 PATH 中移除
remove_from_path() {
    local shell_configs=("$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile")
    
    for shell_rc in "${shell_configs[@]}"; do
        if [[ -f "$shell_rc" ]]; then
            # 移除 DelGuard PATH 配置
            sed -i.bak '/# DelGuard PATH/,+1d' "$shell_rc" 2>/dev/null || true
            log "INFO" "已从 $shell_rc 中移除 PATH 配置"
        fi
    done
}

# 移除 shell 别名
remove_shell_aliases() {
    local shell_configs=("$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile")
    
    for shell_rc in "${shell_configs[@]}"; do
        if [[ -f "$shell_rc" ]]; then
            # 移除 DelGuard 别名配置
            sed -i.bak '/# DelGuard 别名配置/,/^fi$/d' "$shell_rc" 2>/dev/null || true
            log "INFO" "已从 $shell_rc 中移除别名配置"
        fi
    done
}

# 检查安装状态
check_install_status() {
    echo -e "${CYAN}=== DelGuard 安装状态 ===${NC}"
    
    if [[ -f "$EXECUTABLE_PATH" ]]; then
        echo -e "${GREEN}✓ 已安装${NC}"
        echo -e "  位置: ${EXECUTABLE_PATH}"
        
        if command -v delguard &> /dev/null; then
            local version=$(delguard --version 2>/dev/null || echo "无法获取")
            echo -e "  版本: ${version}"
        else
            echo -e "${YELLOW}  版本: 无法获取${NC}"
        fi
    else
        echo -e "${RED}✗ 未安装${NC}"
    fi
    
    # 检查 PATH
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        echo -e "${GREEN}✓ 已添加到 PATH${NC}"
    else
        echo -e "${YELLOW}✗ 未添加到 PATH${NC}"
    fi
    
    # 检查别名
    if command -v delguard &> /dev/null; then
        echo -e "${GREEN}✓ 命令可用${NC}"
    else
        echo -e "${YELLOW}✗ 命令不可用${NC}"
    fi
    
    # 检查配置目录
    if [[ -d "$CONFIG_DIR" ]]; then
        echo -e "${GREEN}✓ 配置目录存在: $CONFIG_DIR${NC}"
    else
        echo -e "${YELLOW}✗ 配置目录不存在${NC}"
    fi
}

# 主函数
main() {
    echo -e "${CYAN}DelGuard 安装程序${NC}"
    echo -e "${CYAN}=================${NC}"
    echo ""
    
    # 解析参数
    parse_args "$@"
    
    # 检查依赖
    check_dependencies
    
    # 执行相应操作
    if [[ "$STATUS_CHECK" == "true" ]]; then
        check_install_status
        exit 0
    fi
    
    if [[ "$UNINSTALL" == "true" ]]; then
        uninstall_delguard
        exit 0
    fi
    
    # 检查权限和网络
    check_permissions
    check_network
    
    # 安装
    install_delguard
}

# 错误处理
trap 'log "ERROR" "安装过程中发生错误，请查看日志: $LOG_FILE"' ERR

# 执行主函数
main "$@"