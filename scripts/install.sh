#!/bin/bash

# DelGuard 智能安装脚本
# 支持 Linux 和 macOS 系统自动检测和安装

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 全局变量
DELGUARD_VERSION="latest"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/delguard"
GITHUB_REPO="01luyicheng/DelGuard"
TEMP_DIR="/tmp/delguard-install"
LOG_FILE="/tmp/delguard-install.log"
SCRIPT_VERSION="2.0.0"

# 日志函数
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# 打印带颜色的消息并记录日志
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
    log "INFO" "$1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    log "SUCCESS" "$1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    log "WARNING" "$1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    log "ERROR" "$1"
}

print_debug() {
    echo -e "${PURPLE}[DEBUG]${NC} $1"
    log "DEBUG" "$1"
}

print_header() {
    echo -e "${CYAN}$1${NC}"
    log "INFO" "$1"
}

# 检测操作系统
detect_os() {
    print_info "检测操作系统..."
    
    case "$OSTYPE" in
        linux-gnu*)
            OS="linux"
            DISTRO=$(
                if [[ -f /etc/os-release ]]; then
                    . /etc/os-release
                    echo "$ID"
                elif [[ -f /etc/redhat-release ]]; then
                    echo "rhel"
                elif [[ -f /etc/debian_version ]]; then
                    echo "debian"
                else
                    echo "unknown"
                fi
            )
            ;;
        darwin*)
            OS="darwin"
            DISTRO="macos"
            ;;
        *)
            print_error "不支持的操作系统: $OSTYPE"
            print_error "支持的系统: Linux (Ubuntu, Debian, CentOS, RHEL, Arch) 和 macOS"
            exit 1
            ;;
    esac
    
    print_success "检测到操作系统: $OS ($DISTRO)"
    log "OS" "$OS"
    log "DISTRO" "$DISTRO"
}

# 检测系统架构
detect_arch() {
    print_info "检测系统架构..."
    
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armv6l)
            ARCH="arm"
            ;;
        i386|i686)
            ARCH="386"
            print_warning "32位系统支持有限，建议使用64位系统"
            ;;
        *)
            print_error "不支持的架构: $ARCH"
            print_error "支持的架构: x86_64, aarch64, arm64, armv7l"
            exit 1
            ;;
    esac
    
    print_success "检测到系统架构: $ARCH"
    log "ARCH" "$ARCH"
}

# 检查依赖
check_dependencies() {
    print_info "检查系统依赖..."
    
    # 检查必要的命令
    local deps=("curl" "tar" "grep" "sed" "chmod")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "缺少依赖: ${missing_deps[*]}"
        
        # 提供安装建议
        case "$DISTRO" in
            ubuntu|debian)
                print_info "请运行: sudo apt-get update && sudo apt-get install -y ${missing_deps[*]}"
                ;;
            centos|rhel|fedora)
                print_info "请运行: sudo yum install -y ${missing_deps[*]} 或 sudo dnf install -y ${missing_deps[*]}"
                ;;
            arch|manjaro)
                print_info "请运行: sudo pacman -S ${missing_deps[*]}"
                ;;
            macos)
                print_info "请运行: brew install ${missing_deps[*]} (需要先安装Homebrew)"
                ;;
            *)
                print_info "请先安装缺失的依赖后重试"
                ;;
        esac
        
        exit 1
    fi
    
    # 检查网络连接
    if ! curl -s --connect-timeout 10 https://api.github.com > /dev/null 2>&1; then
        print_error "网络连接失败，无法访问GitHub"
        print_info "请检查网络连接或代理设置"
        exit 1
    fi
    
    print_success "依赖检查通过"
    log "DEPS" "all required dependencies available"
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
        # 使用GitHub API获取最新版本，添加重试机制
        local max_retries=3
        local retry_count=0
        local api_url="https://api.github.com/repos/$GITHUB_REPO/releases/latest"
        
        while [[ $retry_count -lt $max_retries ]]; do
            DELGUARD_VERSION=$(curl -s --connect-timeout 10 --max-time 30 "$api_url" | \
                grep '"tag_name":' | \
                sed -E 's/.*"([^"]+)".*/\1/' | \
                head -n 1)
            
            if [[ -n "$DELGUARD_VERSION" ]]; then
                break
            fi
            
            retry_count=$((retry_count + 1))
            if [[ $retry_count -lt $max_retries ]]; then
                print_warning "获取版本信息失败，5秒后重试 (第 $retry_count/$max_retries 次)..."
                sleep 5
            fi
        done
        
        if [[ -z "$DELGUARD_VERSION" ]]; then
            print_error "无法获取最新版本信息"
            print_info "这可能是网络问题或GitHub API限制，请稍后重试"
            print_info "您也可以手动指定版本，例如: --version v1.6.3"
            exit 1
        fi
    fi
    
    # 验证版本格式
    if [[ ! "$DELGUARD_VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+.*$ ]]; then
        print_warning "版本格式异常: $DELGUARD_VERSION"
        print_info "建议使用标准格式，例如: v1.6.3"
    fi
    
    # 确保版本以v开头
    if [[ ! "$DELGUARD_VERSION" =~ ^v ]]; then
        DELGUARD_VERSION="v$DELGUARD_VERSION"
    fi
    
    print_success "目标版本: $DELGUARD_VERSION"
    log "VERSION" "$DELGUARD_VERSION"
}

# 下载二进制文件
download_binary() {
    print_info "下载 DelGuard 二进制文件..."
    
    # 构建下载URL - 支持多种格式
    local binary_names=(
        "delguard-${OS}-${ARCH}"
        "delguard_${OS}_${ARCH}"
        "delguard-${OS}-${ARCH}-${DELGUARD_VERSION}"
    )
    
    # 创建临时目录
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    local download_success=false
    local binary_name=""
    local download_url=""
    
    # 尝试不同的二进制文件名格式
    for name in "${binary_names[@]}"; do
        download_url="https://github.com/$GITHUB_REPO/releases/download/$DELGUARD_VERSION/$name"
        print_info "尝试下载: $name"
        
        # 检查文件是否存在（通过HEAD请求）
        if curl -s --head --connect-timeout 10 "$download_url" | head -n 1 | grep -q "200"; then
            binary_name="$name"
            break
        fi
    done
    
    if [[ -z "$binary_name" ]]; then
        print_error "找不到匹配的二进制文件"
        print_info "支持的格式: ${binary_names[*]}"
        print_info "请检查版本 $DELGUARD_VERSION 是否存在"
        exit 1
    fi
    
    print_info "下载地址: $download_url"
    
    # 下载文件 - 增强的重试机制
    local max_retries=5
    local retry_count=0
    local retry_delays=(1 2 5 10 30)
    
    while [[ $retry_count -lt $max_retries ]]; do
        local curl_exit_code=0
        
        # 显示下载进度
        if curl -L --retry 2 --retry-delay 10 --connect-timeout 30 --max-time 600 \
                --progress-bar -o "delguard" "$download_url"; then
            download_success=true
            break
        else
            curl_exit_code=$?
        fi
        
        retry_count=$((retry_count + 1))
        if [[ $retry_count -lt $max_retries ]]; then
            local delay=${retry_delays[$retry_count-1]}
            print_warning "下载失败 (退出码: $curl_exit_code)，${delay}秒后重试 (第 $retry_count/$max_retries 次)..."
            sleep $delay
        else
            print_error "下载失败，已达到最大重试次数"
            print_info "可能的解决方案:"
            print_info "1. 检查网络连接"
            print_info "2. 尝试使用代理"
            print_info "3. 手动下载: $download_url"
            cleanup
            exit 1
        fi
    done
    
    if [[ "$download_success" != true ]]; then
        print_error "下载失败"
        exit 1
    fi
    
    # 验证下载的文件
    if [[ ! -f "delguard" ]]; then
        print_error "下载的文件不存在"
        exit 1
    fi
    
    # 检查文件大小（确保不是错误页面）
    local file_size=$(stat -c%s "delguard" 2>/dev/null || stat -f%z "delguard" 2>/dev/null || echo "0")
    if [[ $file_size -lt 1000000 ]]; then  # 小于1MB可能有问题
        print_warning "下载的文件异常小 ($file_size 字节)，可能下载失败"
        
        # 检查文件内容
        if file "delguard" | grep -q "text"; then
            print_error "下载的是文本文件而非二进制文件"
            print_info "请检查下载链接或稍后重试"
            exit 1
        fi
    fi
    
    # 设置执行权限
    chmod +x "delguard"
    
    print_success "下载完成 ($(numfmt --to=iec $file_size 2>/dev/null || echo "${file_size}B"))"
    log "DOWNLOAD" "success - ${binary_name} ($file_size bytes)"
}

# 安装二进制文件
install_binary() {
    print_info "安装 DelGuard..."
    
    # 验证二进制文件
    if [[ ! -f "delguard" ]]; then
        print_error "找不到二进制文件"
        cleanup
        exit 1
    fi
    
    # 验证可执行文件
    if ! ./delguard --version &>/dev/null; then
        print_error "下载的二进制文件无法执行"
        print_info "可能的原因:"
        print_info "1. 下载了错误的架构版本"
        print_info "2. 文件已损坏"
        print_info "3. 系统缺少必要的运行库"
        cleanup
        exit 1
    fi
    
    # 检查目标目录权限
    if [[ ! -d "$INSTALL_DIR" ]]; then
        print_info "创建安装目录: $INSTALL_DIR"
        if ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
            print_error "无法创建安装目录: $INSTALL_DIR"
            print_info "请检查权限或使用 sudo"
            exit 1
        fi
    fi
    
    # 备份现有版本（如果存在）
    if [[ -f "$INSTALL_DIR/delguard" ]]; then
        local backup_path="$INSTALL_DIR/delguard.backup.$(date +%s)"
        print_info "备份现有版本到: $backup_path"
        if cp "$INSTALL_DIR/delguard" "$backup_path"; then
            log "BACKUP" "created backup at $backup_path"
        else
            print_warning "无法创建备份"
        fi
    fi
    
    # 复制到安装目录
    print_info "复制文件到 $INSTALL_DIR/delguard"
    if ! cp "delguard" "$INSTALL_DIR/delguard"; then
        print_error "安装失败 - 复制文件失败"
        print_info "可能的解决方案:"
        print_info "1. 使用 sudo 运行脚本"
        print_info "2. 检查磁盘空间"
        print_info "3. 检查文件权限"
        cleanup
        exit 1
    fi
    
    # 设置执行权限
    if ! chmod +x "$INSTALL_DIR/delguard"; then
        print_error "无法设置执行权限"
        exit 1
    fi
    
    # 验证安装
    if ! "$INSTALL_DIR/delguard" --version &>/dev/null; then
        print_error "安装验证失败"
        exit 1
    fi
    
    print_success "DelGuard 已安装到 $INSTALL_DIR/delguard"
    log "INSTALL" "success - installed to $INSTALL_DIR/delguard"
}

# 创建配置目录
create_config() {
    print_info "创建配置目录..."
    
    # 确保配置目录存在
    if ! mkdir -p "$CONFIG_DIR" 2>/dev/null; then
        print_error "无法创建配置目录: $CONFIG_DIR"
        exit 1
    fi
    
    # 获取用户主目录
    local user_home="$HOME"
    
    # 创建默认配置文件
    local config_file="$CONFIG_DIR/config.yaml"
    cat > "$config_file" << EOF
# DelGuard 配置文件
# 生成时间: $(date)

# 通用设置
verbose: false
force: false
quiet: false

# 回收站设置
trash:
  # 自动清理设置
  auto_clean: false
  max_days: 30
  max_size: "1GB"
  
  # 清理阈值
  cleanup_threshold: 0.8
  
  # 保留策略
  keep_patterns:
    - "*.config"
    - "*.env"

# 日志设置
log:
  level: "info"
  file: "$CONFIG_DIR/delguard.log"
  max_size: "10MB"
  max_files: 5
  
# 安全设置
security:
  # 删除确认
  confirm_deletes: true
  confirm_large_files: true
  large_file_size: "100MB"
  
  # 权限检查
  check_permissions: true
  
# 用户界面
ui:
  # 显示设置
  show_size: true
  show_date: true
  show_path: true
  
  # 颜色输出
  colors: true
  
  # 交互提示
  interactive: true

# 备份设置
backup:
  # 自动备份
  auto_backup: false
  backup_dir: "$user_home/.delguard-backups"
  max_backups: 10
  
# 排除设置
exclude:
  # 排除的文件模式
  patterns:
    - "*.tmp"
    - "*.log"
    - "*.cache"
    - ".DS_Store"
    - "Thumbs.db"
    
  # 排除的目录
  directories:
    - ".git"
    - "node_modules"
    - "__pycache__"
    - ".pytest_cache"

# 恢复设置
restore:
  # 恢复选项
  preserve_permissions: true
  preserve_timestamps: true
  
  # 冲突处理
  conflict_action: "ask"  # ask, overwrite, skip
  
# 网络设置
network:
  # 代理设置
  proxy: ""
  timeout: 30
  
# 更新设置
update:
  # 自动检查更新
  auto_check: true
  check_interval: "24h"
  
  # 更新通道
  channel: "stable"  # stable, beta, dev
EOF
    
    # 设置配置文件权限
    chmod 644 "$config_file"
    
    # 创建日志目录
    mkdir -p "$CONFIG_DIR/logs"
    
    # 创建备份目录
    mkdir -p "$user_home/.delguard-backups"
    
    print_success "配置目录已创建: $CONFIG_DIR"
    log "CONFIG" "created config at $config_file"
}

# 配置Shell别名
setup_shell_aliases() {
    print_info "配置Shell别名..."
    
    local shell_configs=()
    local shell_type=""
    
    # 检测用户使用的Shell
    if [[ -n "$BASH_VERSION" ]] || [[ "$SHELL" == *"bash"* ]]; then
        shell_type="bash"
        shell_configs+=("$HOME/.bashrc")
        [[ -f "$HOME/.bash_profile" ]] && shell_configs+=("$HOME/.bash_profile")
        [[ -f "$HOME/.profile" ]] && shell_configs+=("$HOME/.profile")
    elif [[ -n "$ZSH_VERSION" ]] || [[ "$SHELL" == *"zsh"* ]]; then
        shell_type="zsh"
        shell_configs+=("$HOME/.zshrc")
        [[ -f "$HOME/.zprofile" ]] && shell_configs+=("$HOME/.zprofile")
    else
        print_warning "未知Shell类型，跳过别名配置"
        return 0
    fi
    
    print_info "检测到Shell类型: $shell_type"
    
    # 增强的别名配置
    local alias_content=$(cat << 'EOF'

# =============================================================================
# DelGuard 别名配置 - 安全删除工具
# 添加时间: $(date)
# =============================================================================

# 核心别名 - 安全删除
alias del='delguard delete'
alias rm='delguard delete'
alias trash='delguard delete'

# 恢复别名
alias restore='delguard restore'
alias undel='delguard restore'
alias untrash='delguard restore'

# 管理别名
alias empty-trash='delguard empty'
alias list-trash='delguard list'
alias trash-status='delguard status'

# 快捷命令
alias del-force='delguard delete --force'
alias del-quiet='delguard delete --quiet'
alias del-verbose='delguard delete --verbose'

# 帮助别名
alias del-help='delguard --help'
alias del-version='delguard --version'

# 安全提示函数
safe_rm() {
    if [[ $# -eq 0 ]]; then
        echo "用法: rm <文件或目录>"
        echo "注意: 这是一个安全的删除命令，文件会被移动到回收站"
        return 1
    fi
    
    echo "🛡️  正在使用 DelGuard 安全删除..."
    delguard delete "$@"
}

# 覆盖rm命令（可选）
# alias rm='safe_rm'

# =============================================================================
# DelGuard 配置完成
# =============================================================================

EOF
)
    
    local aliases_added=0
    
    for config_file in "${shell_configs[@]}"; do
        if [[ -f "$config_file" ]]; then
            # 检查是否已经配置过
            if ! grep -q "DelGuard 别名配置" "$config_file" 2>/dev/null; then
                print_info "添加别名到: $config_file"
                echo "$alias_content" >> "$config_file"
                ((aliases_added++))
            else
                print_info "别名已存在于: $config_file"
            fi
        elif [[ "$config_file" == "$HOME/.bashrc" ]] || [[ "$config_file" == "$HOME/.zshrc" ]]; then
            # 创建配置文件（如果不存在）
            print_info "创建配置文件: $config_file"
            touch "$config_file"
            echo "$alias_content" >> "$config_file"
            ((aliases_added++))
        fi
    done
    
    if [[ $aliases_added -gt 0 ]]; then
        print_success "Shell别名配置完成 ($aliases_added 个文件已更新)"
    else
        print_info "所有别名配置已存在"
    fi
    
    log "ALIASES" "configured for $shell_type"
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
    
    # 获取版本信息
    local version_output
    if version_output=$("$INSTALL_DIR/delguard" --version 2>/dev/null); then
        print_success "安装验证通过 - 版本: $version_output"
    else
        print_warning "无法获取版本信息，但文件已安装"
    fi
    
    # 检查配置文件
    if [[ -f "$CONFIG_DIR/config.yaml" ]]; then
        print_success "配置文件存在: $CONFIG_DIR/config.yaml"
    else
        print_warning "配置文件不存在"
    fi
    
    # 检查PATH
    if echo "$PATH" | grep -q "$INSTALL_DIR"; then
        print_success "$INSTALL_DIR 已在 PATH 中"
    else
        print_warning "$INSTALL_DIR 不在 PATH 中，可能需要手动添加"
    fi
    
    log "VERIFY" "installation verified successfully"
    return 0
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
    local version_info=""
    if [[ -f "$INSTALL_DIR/delguard" ]]; then
        version_info=$("$INSTALL_DIR/delguard" --version 2>/dev/null || echo "unknown")
    fi
    
    print_header "🎉 DelGuard 安装完成！"
    echo
    print_header "📋 安装摘要"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  📦 版本:     $version_info"
    echo "  📍 安装位置: $INSTALL_DIR/delguard"
    echo "  📁 配置目录: $CONFIG_DIR"
    echo "  📝 日志文件: $LOG_FILE"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    print_header "🚀 快速开始指南"
    echo "────────────────────────────────────────────────────────────────────"
    echo "  delguard --help          # 查看完整帮助"
    echo "  delguard status          # 查看系统状态"
    echo "  delguard delete <file>   # 安全删除文件到回收站"
    echo "  delguard list           # 查看回收站内容"
    echo "  delguard restore <file> # 恢复文件"
    echo "  delguard empty          # 清空回收站"
    echo "  delguard config         # 查看配置"
    echo
    
    print_header "💡 快捷别名 (重新打开终端后生效)"
    echo "────────────────────────────────────────────────────────────────────"
    echo "  del <file>        # 安全删除文件"
    echo "  rm <file>         # 安全删除文件 (可选)"
    echo "  restore <file>    # 恢复文件"
    echo "  empty-trash       # 清空回收站"
    echo "  list-trash        # 查看回收站"
    echo "  trash-status      # 查看回收站状态"
    echo
    
    print_header "⚙️  配置和自定义"
    echo "────────────────────────────────────────────────────────────────────"
    echo "  配置文件: $CONFIG_DIR/config.yaml"
    echo "  日志文件: $CONFIG_DIR/delguard.log"
    echo "  备份目录: $HOME/.delguard-backups"
    echo
    
    print_header "📖 学习资源"
    echo "────────────────────────────────────────────────────────────────────"
    echo "  📚 官方文档: https://github.com/$GITHUB_REPO"
    echo "  🐛 问题反馈: https://github.com/$GITHUB_REPO/issues"
    echo "  💬 讨论区: https://github.com/$GITHUB_REPO/discussions"
    echo
    
    print_header "⚠️  重要提醒"
    echo "────────────────────────────────────────────────────────────────────"
    echo "  🔧 重新打开终端以启用新别名"
    echo "  💾 重要文件删除前会自动确认"
    echo "  🗑️  回收站需要定期清理"
    echo "  🔍 使用 'delguard --help' 查看更多命令"
    echo
    
    # 显示下一步操作建议
    print_header "🎯 建议的下一步操作"
    echo "────────────────────────────────────────────────────────────────────"
    echo "  1. 测试删除: delguard delete /tmp/test-file"
    echo "  2. 查看回收站: delguard list"
    echo "  3. 查看配置: delguard config"
    echo "  4. 运行健康检查: delguard status"
    echo
    
    # 创建快速入门脚本
    local quick_start_script="$CONFIG_DIR/quick-start.sh"
    cat > "$quick_start_script" << 'EOF'
#!/bin/bash
echo "🛡️  DelGuard 快速入门"
echo "======================"
echo
echo "1. 创建测试文件..."
touch /tmp/delguard-test.txt
echo "测试文件已创建: /tmp/delguard-test.txt"

echo "2. 安全删除测试文件..."
delguard delete /tmp/delguard-test.txt

echo "3. 查看回收站..."
delguard list

echo "4. 恢复测试文件..."
delguard restore delguard-test.txt

echo "5. 查看系统状态..."
delguard status

echo
echo "✅ 快速入门完成！"
echo "现在您可以安全地使用 DelGuard 了！"
EOF
    chmod +x "$quick_start_script"
    
    log "COMPLETE" "installation completed successfully"
    
    # 显示日志位置
    echo
    print_info "📋 详细安装日志已保存到: $LOG_FILE"
}

# 显示帮助信息
show_help() {
    cat << EOF
🛡️  DelGuard 智能安装脚本 v$SCRIPT_VERSION

使用方法: $0 [选项]

选项:
    -v, --version VERSION   指定要安装的版本 (默认: latest)
    -d, --dir DIR         指定安装目录 (默认: /usr/local/bin)
    -c, --config-dir DIR    指定配置目录 (默认: ~/.config/delguard)
    -f, --force            强制重新安装，覆盖现有版本
    -h, --help            显示此帮助信息
    --no-aliases          不配置Shell别名
    --no-config           不创建配置文件
    --dry-run             模拟安装，不实际执行
    --verbose             显示详细输出
    --uninstall           卸载 DelGuard

示例:
    $0                          # 安装最新版本
    $0 --version v1.6.3         # 安装指定版本
    $0 --dir /opt/delguard      # 安装到指定目录
    $0 --force --no-aliases     # 强制安装，不配置别名

环境变量:
    DELGUARD_VERSION        设置默认版本
    DELGUARD_INSTALL_DIR    设置默认安装目录
    DELGUARD_CONFIG_DIR     设置默认配置目录
    HTTP_PROXY              设置HTTP代理
    HTTPS_PROXY             设置HTTPS代理

EOF
}

# 卸载功能
uninstall_delguard() {
    print_header "🗑️  卸载 DelGuard"
    echo
    
    local files_to_remove=(
        "$INSTALL_DIR/delguard"
        "$INSTALL_DIR/delguard-uninstall"
    )
    
    local dirs_to_remove=(
        "$CONFIG_DIR"
    )
    
    # 显示将要删除的文件
    print_info "将要删除的文件:"
    for file in "${files_to_remove[@]}"; do
        if [[ -f "$file" ]]; then
            echo "  📄 $file"
        fi
    done
    
    print_info "将要删除的目录:"
    for dir in "${dirs_to_remove[@]}"; do
        if [[ -d "$dir" ]]; then
            echo "  📁 $dir"
        fi
    done
    
    echo
    read -p "确定要继续卸载吗? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "卸载已取消"
        exit 0
    fi
    
    # 执行卸载
    for file in "${files_to_remove[@]}"; do
        if [[ -f "$file" ]]; then
            rm -f "$file"
            print_success "已删除: $file"
        fi
    done
    
    for dir in "${dirs_to_remove[@]}"; do
        if [[ -d "$dir" ]]; then
            rm -rf "$dir"
            print_success "已删除: $dir"
        fi
    done
    
    # 清理Shell别名
    local shell_configs=()
    if [[ -n "$BASH_VERSION" ]] || [[ "$SHELL" == *"bash"* ]]; then
        shell_configs+=("$HOME/.bashrc")
    fi
    if [[ -n "$ZSH_VERSION" ]] || [[ "$SHELL" == *"zsh"* ]]; then
        shell_configs+=("$HOME/.zshrc")
    fi
    
    for config_file in "${shell_configs[@]}"; do
        if [[ -f "$config_file" ]]; then
            sed -i '/DelGuard 别名配置/,/# DelGuard 配置完成/d' "$config_file" 2>/dev/null || true
            print_info "已清理Shell配置: $config_file"
        fi
    done
    
    print_success "DelGuard 已成功卸载"
    log "UNINSTALL" "delguard uninstalled successfully"
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                DELGUARD_VERSION="$2"
                shift 2
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -c|--config-dir)
                CONFIG_DIR="$2"
                shift 2
                ;;
            -f|--force)
                FORCE_INSTALL=true
                shift
                ;;
            --no-aliases)
                SKIP_ALIASES=true
                shift
                ;;
            --no-config)
                SKIP_CONFIG=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --verbose)
                set -x
                shift
                ;;
            --uninstall)
                UNINSTALL=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 主函数
main() {
    # 解析命令行参数
    parse_args "$@"
    
    # 设置环境变量默认值
    DELGUARD_VERSION=${DELGUARD_VERSION:-latest}
    INSTALL_DIR=${DELGUARD_INSTALL_DIR:-$INSTALL_DIR}
    CONFIG_DIR=${DELGUARD_CONFIG_DIR:-$CONFIG_DIR}
    
    # 处理卸载
    if [[ "$UNINSTALL" == true ]]; then
        uninstall_delguard
        exit 0
    fi
    
    # 处理干运行
    if [[ "$DRY_RUN" == true ]]; then
        print_header "🔍 干运行模式 - 模拟安装"
        echo "将要执行的操作:"
        echo "  - 检测系统: $OSTYPE"
        echo "  - 安装目录: $INSTALL_DIR"
        echo "  - 配置目录: $CONFIG_DIR"
        echo "  - 目标版本: $DELGUARD_VERSION"
        echo "  - 配置别名: $([[ "$SKIP_ALIASES" != true ]] && echo "是" || echo "否")"
        echo "  - 创建配置: $([[ "$SKIP_CONFIG" != true ]] && echo "是" || echo "否")"
        exit 0
    fi
    
    # 显示欢迎信息
    print_header "🛡️  DelGuard 智能安装脚本 v$SCRIPT_VERSION"
    print_header "=============================================="
    echo
    
    # 记录开始时间
    local start_time=$(date +%s)
    log "START" "installation started"
    
    # 检查系统环境
    detect_os
    detect_arch
    check_dependencies
    check_permissions
    
    # 检查现有安装
    if [[ -f "$INSTALL_DIR/delguard" && "$FORCE_INSTALL" != true ]]; then
        print_warning "DelGuard 已安装"
        print_info "使用 --force 选项强制重新安装"
        local current_version=$("$INSTALL_DIR/delguard" --version 2>/dev/null || echo "unknown")
        print_info "当前版本: $current_version"
        exit 0
    fi
    
    # 下载和安装
    get_latest_version
    download_binary
    install_binary
    
    # 配置
    if [[ "$SKIP_CONFIG" != true ]]; then
        create_config
    fi
    
    if [[ "$SKIP_ALIASES" != true ]]; then
        setup_shell_aliases
    fi
    
    # 验证和清理
    verify_installation
    cleanup
    
    # 记录结束时间
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    log "COMPLETE" "installation completed in ${duration}s"
    
    # 显示完成信息
    show_completion_info
}

# 错误处理
cleanup_on_error() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        print_error "安装失败 (退出码: $exit_code)"
        print_info "查看详细日志: $LOG_FILE"
        cleanup
    fi
}

trap cleanup_on_error EXIT

# 运行主函数
main "$@"