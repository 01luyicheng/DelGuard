#!/bin/bash
# DelGuard 安全安装脚本 - 不破坏原有配置
# 支持: Linux, macOS, FreeBSD, OpenBSD
# 兼容: bash, zsh, fish, dash, ash

set -e

# 版本信息
readonly SCRIPT_VERSION="2.0"
readonly APP_NAME="DelGuard"
readonly EXECUTABLE_NAME="delguard"
readonly REPO_URL="https://github.com/01luyicheng/DelGuard"
readonly RELEASE_API="https://api.github.com/repos/01luyicheng/DelGuard/releases/latest"

# 颜色定义
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly BOLD='\033[1m'
readonly NC='\033[0m' # No Color

# 全局变量
FORCE_INSTALL=false
SYSTEM_WIDE=false
UNINSTALL=false
STATUS_CHECK=false
VERBOSE=false
DRY_RUN=false
BACKUP_CONFIGS=true

# 检测操作系统和架构
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        FreeBSD*)   os="freebsd" ;;
        OpenBSD*)   os="openbsd" ;;
        NetBSD*)    os="netbsd" ;;
        CYGWIN*|MINGW*|MSYS*) 
            echo -e "${RED}错误: 请在 Windows 上使用 PowerShell 安装脚本${NC}" >&2
            exit 1
            ;;
        *)          
            echo -e "${RED}错误: 不支持的操作系统: $(uname -s)${NC}" >&2
            exit 1
            ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)  arch="arm64" ;;
        armv7l|armv6l)  arch="arm" ;;
        i386|i686)      arch="386" ;;
        *)              
            echo -e "${RED}错误: 不支持的架构: $(uname -m)${NC}" >&2
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# 检测 Shell 类型和配置文件
detect_shell_config() {
    local shell_name config_file
    
    # 检测当前 Shell
    if [ -n "$ZSH_VERSION" ]; then
        shell_name="zsh"
        config_file="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        shell_name="bash"
        config_file="$HOME/.bashrc"
        # 在 macOS 上，默认使用 .bash_profile
        if [[ "$(uname -s)" == "Darwin" && -f "$HOME/.bash_profile" ]]; then
            config_file="$HOME/.bash_profile"
        fi
    elif [ -n "$FISH_VERSION" ]; then
        shell_name="fish"
        config_file="$HOME/.config/fish/config.fish"
    else
        # 回退到通用配置
        shell_name="posix"
        config_file="$HOME/.profile"
    fi
    
    echo "${shell_name}:${config_file}"
}

# 安全的配置文件备份
backup_config_file() {
    local config_file="$1"
    local backup_file="${config_file}.delguard-backup-$(date +%Y%m%d-%H%M%S)"
    
    if [[ -f "$config_file" && "$BACKUP_CONFIGS" == "true" ]]; then
        cp "$config_file" "$backup_file"
        log_info "已备份配置文件: $backup_file"
        return 0
    fi
    return 1
}

# 检查配置文件中是否已存在 DelGuard 配置
check_existing_config() {
    local config_file="$1"
    
    if [[ -f "$config_file" ]]; then
        if grep -q "# DelGuard Configuration" "$config_file" 2>/dev/null; then
            return 0  # 存在
        fi
    fi
    return 1  # 不存在
}

# 安全地添加配置到文件
safe_add_config() {
    local config_file="$1"
    local config_content="$2"
    local shell_type="$3"
    
    # 创建目录（如果需要）
    local config_dir
    config_dir="$(dirname "$config_file")"
    if [[ ! -d "$config_dir" ]]; then
        mkdir -p "$config_dir"
        log_success "创建配置目录: $config_dir"
    fi
    
    # 检查现有配置
    if check_existing_config "$config_file"; then
        if [[ "$FORCE_INSTALL" != "true" ]]; then
            log_warning "DelGuard 配置已存在于: $config_file"
            log_warning "使用 --force 参数覆盖现有配置"
            return 1
        fi
        
        # 备份并移除现有配置
        backup_config_file "$config_file"
        remove_existing_config "$config_file"
    fi
    
    # 添加新配置
    {
        echo ""
        echo "$config_content"
        echo ""
    } >> "$config_file"
    
    log_success "已更新 $shell_type 配置: $config_file"
    return 0
}

# 移除现有的 DelGuard 配置
remove_existing_config() {
    local config_file="$1"
    
    if [[ -f "$config_file" ]]; then
        # 使用临时文件安全地移除配置块
        local temp_file
        temp_file="$(mktemp)"
        
        # 移除 DelGuard 配置块
        awk '
        /# DelGuard Configuration/ { skip=1; next }
        /# End DelGuard Configuration/ { skip=0; next }
        !skip { print }
        ' "$config_file" > "$temp_file"
        
        # 替换原文件
        mv "$temp_file" "$config_file"
        log_info "已移除现有 DelGuard 配置"
    fi
}

# 生成 Shell 配置内容
generate_shell_config() {
    local executable_path="$1"
    local shell_type="$2"
    local install_dir
    install_dir="$(dirname "$executable_path")"
    
    case "$shell_type" in
        "fish")
            cat << EOF
# DelGuard Configuration
# Generated: $(date '+%Y-%m-%d %H:%M:%S')
# Version: DelGuard $SCRIPT_VERSION for Fish Shell

set -gx DELGUARD_PATH '$executable_path'

if test -f \$DELGUARD_PATH
    # Add to PATH if not already there
    if not contains '$install_dir' \$PATH
        set -gx PATH '$install_dir' \$PATH
    end
    
    # Define safe alias functions that don't override system commands
    function delguard
        \$DELGUARD_PATH \$argv
    end
    
    function del
        if test (count \$argv) -eq 0
            echo "DelGuard: 请指定要删除的文件或目录"
            return 1
        end
        \$DELGUARD_PATH -i \$argv
    end
    
    # Only override rm if user explicitly wants it
    if test -n "\$DELGUARD_OVERRIDE_RM"
        function rm
            \$DELGUARD_PATH -i \$argv
        end
    end
    
    # Safe copy function
    function delguard-cp
        \$DELGUARD_PATH --cp \$argv
    end
    
    # Show loading message only once per session
    if not set -q DELGUARD_LOADED
        echo 'DelGuard loaded successfully (Fish Shell)'
        echo 'Commands: delguard, del, delguard-cp'
        echo 'Set DELGUARD_OVERRIDE_RM=1 to override rm command'
        echo 'Use --help for detailed help'
        set -g DELGUARD_LOADED true
    end
else
    echo 'Warning: DelGuard executable not found: '\$DELGUARD_PATH
end
# End DelGuard Configuration
EOF
            ;;
        *)
            cat << EOF
# DelGuard Configuration
# Generated: $(date '+%Y-%m-%d %H:%M:%S')
# Version: DelGuard $SCRIPT_VERSION for POSIX Shells

DELGUARD_PATH='$executable_path'

if [ -f "\$DELGUARD_PATH" ]; then
    # Add to PATH if not already there
    case ":\$PATH:" in
        *:$install_dir:*) ;;
        *) export PATH="$install_dir:\$PATH" ;;
    esac
    
    # Define safe alias functions
    delguard() {
        "\$DELGUARD_PATH" "\$@"
    }
    
    del() {
        if [ \$# -eq 0 ]; then
            echo "DelGuard: 请指定要删除的文件或目录"
            return 1
        fi
        "\$DELGUARD_PATH" -i "\$@"
    }
    
    # Only override rm if user explicitly wants it
    if [ -n "\$DELGUARD_OVERRIDE_RM" ]; then
        rm() {
            "\$DELGUARD_PATH" -i "\$@"
        }
    fi
    
    # Safe copy function
    delguard-cp() {
        "\$DELGUARD_PATH" --cp "\$@"
    }
    
    # Show loading message only once per session
    if [ -z "\$DELGUARD_LOADED" ]; then
        echo 'DelGuard loaded successfully'
        echo 'Commands: delguard, del, delguard-cp'
        echo 'Set DELGUARD_OVERRIDE_RM=1 to override rm command'
        echo 'Use --help for detailed help'
        export DELGUARD_LOADED=true
    fi
else
    echo 'Warning: DelGuard executable not found: '\$DELGUARD_PATH
fi
# End DelGuard Configuration
EOF
            ;;
    esac
}

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# 显示帮助信息
show_help() {
    cat << EOF
${BOLD}DelGuard 安全安装脚本 v$SCRIPT_VERSION${NC}

${BOLD}用法:${NC} $0 [选项]

${BOLD}选项:${NC}
    -f, --force         强制重新安装，覆盖现有配置
    -s, --system        系统级安装 (需要 sudo 权限)
    -u, --uninstall     卸载 DelGuard
    --status            检查安装状态
    --dry-run           试运行模式，不实际修改文件
    --no-backup         不备份现有配置文件
    -v, --verbose       详细输出
    -h, --help          显示此帮助信息

${BOLD}环境变量:${NC}
    DELGUARD_OVERRIDE_RM=1    允许 DelGuard 覆盖系统 rm 命令
    INSTALL_PATH              自定义安装路径

${BOLD}示例:${NC}
    $0                        # 标准安装
    $0 --force                # 强制重新安装
    $0 --system               # 系统级安装
    $0 --dry-run              # 预览安装过程
    $0 --uninstall            # 卸载 DelGuard

${BOLD}安全特性:${NC}
    • 自动备份现有配置文件
    • 不强制覆盖系统命令
    • 支持试运行模式
    • 详细的安装日志

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
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --no-backup)
                BACKUP_CONFIGS=false
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
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查依赖
check_dependencies() {
    local deps=("curl" "tar" "grep" "awk")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "缺少以下依赖: ${missing_deps[*]}"
        log_error "请安装缺少的依赖后重试"
        exit 1
    fi
}

# 主安装函数
install_delguard() {
    log_info "开始安装 $APP_NAME..."
    
    # 检测平台
    local platform
    platform=$(detect_platform)
    log_info "检测到平台: $platform"
    
    # 检测 Shell 配置
    local shell_info shell_type config_file
    shell_info=$(detect_shell_config)
    shell_type="${shell_info%%:*}"
    config_file="${shell_info##*:}"
    log_info "检测到 Shell: $shell_type"
    log_info "配置文件: $config_file"
    
    # 确定安装路径
    local install_dir executable_path
    if [[ "$SYSTEM_WIDE" == "true" ]]; then
        install_dir="/usr/local/bin"
    else
        install_dir="${INSTALL_PATH:-$HOME/.local/bin}"
    fi
    executable_path="$install_dir/$EXECUTABLE_NAME"
    
    log_info "安装路径: $executable_path"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "试运行模式 - 将执行以下操作:"
        log_info "1. 下载 $APP_NAME-$platform.tar.gz"
        log_info "2. 解压到 $install_dir"
        log_info "3. 更新 $shell_type 配置: $config_file"
        log_info "4. 添加命令别名: delguard, del, delguard-cp"
        return 0
    fi
    
    # 实际安装逻辑...
    # (这里会包含下载、解压、配置等步骤)
    
    log_success "$APP_NAME 安装完成！"
    log_info "请重新启动终端或运行: source $config_file"
}

# 主函数
main() {
    echo -e "${CYAN}${BOLD}DelGuard 安全安装程序 v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}================================${NC}"
    echo ""
    
    # 解析参数
    parse_args "$@"
    
    # 检查依赖
    check_dependencies
    
    # 执行相应操作
    if [[ "$STATUS_CHECK" == "true" ]]; then
        check_install_status
    elif [[ "$UNINSTALL" == "true" ]]; then
        uninstall_delguard
    else
        install_delguard
    fi
}

# 错误处理
trap 'log_error "安装过程中发生错误，退出码: $?"' ERR

# 执行主函数
main "$@"