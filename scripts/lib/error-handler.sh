#!/bin/bash

# DelGuard 错误处理库
# 提供统一的错误处理、日志记录和恢复机制

# 全局变量
ERROR_LOG_FILE="${DELGUARD_LOG_DIR:-/tmp}/delguard-install.log"
ERROR_COUNT=0
WARNING_COUNT=0
DEBUG_MODE=${DELGUARD_DEBUG:-false}

# 颜色定义
declare -A COLORS=(
    [RED]='\033[0;31m'
    [GREEN]='\033[0;32m'
    [YELLOW]='\033[1;33m'
    [BLUE]='\033[0;34m'
    [PURPLE]='\033[0;35m'
    [CYAN]='\033[0;36m'
    [WHITE]='\033[1;37m'
    [NC]='\033[0m'
)

# 初始化错误处理系统
init_error_handler() {
    # 创建日志目录
    local log_dir=$(dirname "$ERROR_LOG_FILE")
    mkdir -p "$log_dir" 2>/dev/null || true
    
    # 初始化日志文件
    {
        echo "=== DelGuard 安装日志 ==="
        echo "时间: $(date)"
        echo "系统: $(uname -a)"
        echo "用户: $(whoami)"
        echo "工作目录: $(pwd)"
        echo "=========================="
        echo
    } > "$ERROR_LOG_FILE" 2>/dev/null || true
    
    # 设置错误处理陷阱
    set -E
    trap 'handle_error $? $LINENO $BASH_LINENO "$BASH_COMMAND" "${FUNCNAME[@]}"' ERR
    trap 'cleanup_on_exit' EXIT
    trap 'handle_interrupt' INT TERM
}

# 日志记录函数
log_message() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # 写入日志文件
    echo "[$timestamp] [$level] $message" >> "$ERROR_LOG_FILE" 2>/dev/null || true
    
    # 调试模式下输出到stderr
    if [[ "$DEBUG_MODE" == "true" ]]; then
        echo "DEBUG: [$level] $message" >&2
    fi
}

# 打印函数
print_message() {
    local level="$1"
    local message="$2"
    local color="${3:-WHITE}"
    
    echo -e "${COLORS[$color]}[$level]${COLORS[NC]} $message"
    log_message "$level" "$message"
}

print_info() {
    print_message "INFO" "$1" "BLUE"
}

print_success() {
    print_message "SUCCESS" "$1" "GREEN"
}

print_warning() {
    print_message "WARNING" "$1" "YELLOW"
    ((WARNING_COUNT++))
}

print_error() {
    print_message "ERROR" "$1" "RED"
    ((ERROR_COUNT++))
}

print_debug() {
    if [[ "$DEBUG_MODE" == "true" ]]; then
        print_message "DEBUG" "$1" "PURPLE"
    fi
    log_message "DEBUG" "$1"
}

print_header() {
    local title="$1"
    local color="${2:-CYAN}"
    
    echo
    echo -e "${COLORS[$color]}$title${COLORS[NC]}"
    echo -e "${COLORS[$color]}$(printf '=%.0s' $(seq 1 ${#title}))${COLORS[NC]}"
    log_message "HEADER" "$title"
}

# 错误处理函数
handle_error() {
    local exit_code=$1
    local line_number=$2
    local bash_lineno=$3
    local last_command=$4
    shift 4
    local function_stack=("$@")
    
    # 记录错误信息
    {
        echo "=== 错误详情 ==="
        echo "退出码: $exit_code"
        echo "行号: $line_number"
        echo "命令: $last_command"
        echo "函数栈: ${function_stack[*]}"
        echo "时间: $(date)"
        echo "==============="
    } >> "$ERROR_LOG_FILE" 2>/dev/null || true
    
    print_error "脚本执行失败 (退出码: $exit_code, 行号: $line_number)"
    print_error "失败命令: $last_command"
    
    # 提供恢复建议
    suggest_recovery "$exit_code" "$last_command"
    
    # 不立即退出，让调用者决定如何处理
    return $exit_code
}

# 中断处理
handle_interrupt() {
    print_warning "安装被用户中断"
    log_message "INTERRUPT" "用户中断安装过程"
    cleanup_on_exit
    exit 130
}

# 退出清理
cleanup_on_exit() {
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        print_success "安装完成，退出码: $exit_code"
    else
        print_error "安装失败，退出码: $exit_code"
    fi
    
    # 生成错误报告
    generate_error_report
    
    # 清理临时文件
    cleanup_temp_files
}

# 恢复建议
suggest_recovery() {
    local exit_code=$1
    local failed_command=$2
    
    print_header "🔧 恢复建议" "YELLOW"
    
    case $exit_code in
        1)
            echo "• 检查网络连接"
            echo "• 验证下载URL是否正确"
            echo "• 尝试使用代理或VPN"
            ;;
        2)
            echo "• 检查文件权限"
            echo "• 尝试使用sudo运行"
            echo "• 验证目标目录是否存在"
            ;;
        126)
            echo "• 检查文件执行权限"
            echo "• 运行: chmod +x <文件>"
            ;;
        127)
            echo "• 检查命令是否存在"
            echo "• 安装缺失的依赖"
            ;;
        *)
            echo "• 查看详细日志: $ERROR_LOG_FILE"
            echo "• 尝试重新运行安装脚本"
            echo "• 联系技术支持"
            ;;
    esac
    
    if [[ "$failed_command" == *"curl"* ]]; then
        echo "• 网络相关问题，检查防火墙设置"
        echo "• 尝试使用wget替代curl"
    elif [[ "$failed_command" == *"tar"* ]]; then
        echo "• 文件解压问题，检查下载文件完整性"
        echo "• 尝试重新下载文件"
    elif [[ "$failed_command" == *"cp"* ]] || [[ "$failed_command" == *"mv"* ]]; then
        echo "• 文件操作问题，检查磁盘空间和权限"
        echo "• 确保目标目录可写"
    fi
}

# 生成错误报告
generate_error_report() {
    local report_file="${ERROR_LOG_FILE%.log}.report"
    
    {
        echo "=== DelGuard 安装报告 ==="
        echo "生成时间: $(date)"
        echo "错误数量: $ERROR_COUNT"
        echo "警告数量: $WARNING_COUNT"
        echo
        
        if [[ $ERROR_COUNT -gt 0 ]]; then
            echo "=== 错误摘要 ==="
            grep "\[ERROR\]" "$ERROR_LOG_FILE" 2>/dev/null | tail -10 || echo "无法读取错误日志"
            echo
        fi
        
        if [[ $WARNING_COUNT -gt 0 ]]; then
            echo "=== 警告摘要 ==="
            grep "\[WARNING\]" "$ERROR_LOG_FILE" 2>/dev/null | tail -5 || echo "无法读取警告日志"
            echo
        fi
        
        echo "=== 系统信息 ==="
        echo "操作系统: $(uname -s)"
        echo "架构: $(uname -m)"
        echo "内核版本: $(uname -r)"
        echo "Shell: $SHELL"
        echo "PATH: $PATH"
        echo
        
        echo "=== 环境变量 ==="
        env | grep -E "(DELGUARD|HOME|USER|TMPDIR)" | sort
        echo
        
        echo "完整日志文件: $ERROR_LOG_FILE"
        echo "=========================="
    } > "$report_file" 2>/dev/null || true
    
    if [[ $ERROR_COUNT -gt 0 ]] || [[ $WARNING_COUNT -gt 0 ]]; then
        print_info "错误报告已生成: $report_file"
    fi
}

# 清理临时文件
cleanup_temp_files() {
    local temp_patterns=(
        "/tmp/delguard-*"
        "/tmp/install-*"
        "$HOME/.delguard-temp*"
    )
    
    for pattern in "${temp_patterns[@]}"; do
        # 使用find来安全删除匹配的文件和目录
        find /tmp -maxdepth 1 -name "$(basename "$pattern")" -type d -mtime +1 2>/dev/null | \
        while read -r dir; do
            rm -rf "$dir" 2>/dev/null || true
        done
    done
}

# 验证系统要求
verify_system_requirements() {
    print_header "🔍 验证系统要求"
    
    local requirements_met=true
    
    # 检查操作系统
    case "$(uname -s)" in
        Linux|Darwin)
            print_success "操作系统支持: $(uname -s)"
            ;;
        *)
            print_error "不支持的操作系统: $(uname -s)"
            requirements_met=false
            ;;
    esac
    
    # 检查架构
    case "$(uname -m)" in
        x86_64|aarch64|arm64|armv7l)
            print_success "系统架构支持: $(uname -m)"
            ;;
        *)
            print_warning "未测试的架构: $(uname -m)"
            ;;
    esac
    
    # 检查必要命令
    local required_commands=("curl" "tar" "chmod" "mkdir")
    for cmd in "${required_commands[@]}"; do
        if command -v "$cmd" &>/dev/null; then
            print_success "命令可用: $cmd"
        else
            print_error "缺少必要命令: $cmd"
            requirements_met=false
        fi
    done
    
    # 检查磁盘空间
    local available_space=$(df /tmp 2>/dev/null | awk 'NR==2 {print $4}' || echo "0")
    if [[ $available_space -gt 100000 ]]; then  # 100MB
        print_success "磁盘空间充足"
    else
        print_warning "磁盘空间可能不足"
    fi
    
    # 检查网络连接
    if curl -s --connect-timeout 5 https://api.github.com >/dev/null 2>&1; then
        print_success "网络连接正常"
    else
        print_error "无法连接到GitHub"
        requirements_met=false
    fi
    
    if [[ "$requirements_met" == "true" ]]; then
        print_success "系统要求验证通过"
        return 0
    else
        print_error "系统要求验证失败"
        return 1
    fi
}

# 安全执行命令
safe_execute() {
    local description="$1"
    shift
    local command=("$@")
    
    print_info "执行: $description"
    print_debug "命令: ${command[*]}"
    
    # 记录命令执行
    log_message "EXECUTE" "$description: ${command[*]}"
    
    # 执行命令并捕获输出
    local output
    local exit_code
    
    if output=$("${command[@]}" 2>&1); then
        exit_code=0
        print_success "$description 完成"
        if [[ -n "$output" ]] && [[ "$DEBUG_MODE" == "true" ]]; then
            print_debug "输出: $output"
        fi
    else
        exit_code=$?
        print_error "$description 失败 (退出码: $exit_code)"
        if [[ -n "$output" ]]; then
            print_error "错误输出: $output"
        fi
        log_message "ERROR" "$description 失败: $output"
    fi
    
    return $exit_code
}

# 重试机制
retry_command() {
    local max_attempts="$1"
    local delay="$2"
    local description="$3"
    shift 3
    local command=("$@")
    
    local attempt=1
    while [[ $attempt -le $max_attempts ]]; do
        print_info "尝试 $attempt/$max_attempts: $description"
        
        if safe_execute "$description" "${command[@]}"; then
            return 0
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            print_warning "等待 ${delay}s 后重试..."
            sleep "$delay"
        fi
        
        ((attempt++))
    done
    
    print_error "$description 在 $max_attempts 次尝试后仍然失败"
    return 1
}

# 检查并创建目录
ensure_directory() {
    local dir_path="$1"
    local description="${2:-目录}"
    
    if [[ -d "$dir_path" ]]; then
        print_success "$description 已存在: $dir_path"
        return 0
    fi
    
    if safe_execute "创建$description" mkdir -p "$dir_path"; then
        print_success "$description 创建成功: $dir_path"
        return 0
    else
        print_error "$description 创建失败: $dir_path"
        return 1
    fi
}

# 备份文件
backup_file() {
    local file_path="$1"
    local backup_suffix="${2:-.backup.$(date +%Y%m%d_%H%M%S)}"
    
    if [[ ! -f "$file_path" ]]; then
        print_debug "文件不存在，无需备份: $file_path"
        return 0
    fi
    
    local backup_path="${file_path}${backup_suffix}"
    
    if safe_execute "备份文件" cp "$file_path" "$backup_path"; then
        print_success "文件已备份: $backup_path"
        return 0
    else
        print_error "文件备份失败: $file_path"
        return 1
    fi
}

# 验证文件完整性
verify_file_integrity() {
    local file_path="$1"
    local expected_size="$2"
    local expected_checksum="$3"
    
    if [[ ! -f "$file_path" ]]; then
        print_error "文件不存在: $file_path"
        return 1
    fi
    
    # 检查文件大小
    if [[ -n "$expected_size" ]]; then
        local actual_size=$(stat -c%s "$file_path" 2>/dev/null || stat -f%z "$file_path" 2>/dev/null)
        if [[ "$actual_size" -eq "$expected_size" ]]; then
            print_success "文件大小验证通过: $actual_size bytes"
        else
            print_error "文件大小不匹配: 期望 $expected_size, 实际 $actual_size"
            return 1
        fi
    fi
    
    # 检查校验和
    if [[ -n "$expected_checksum" ]] && command -v sha256sum &>/dev/null; then
        local actual_checksum=$(sha256sum "$file_path" | cut -d' ' -f1)
        if [[ "$actual_checksum" == "$expected_checksum" ]]; then
            print_success "文件校验和验证通过"
        else
            print_error "文件校验和不匹配"
            print_error "期望: $expected_checksum"
            print_error "实际: $actual_checksum"
            return 1
        fi
    fi
    
    return 0
}

# 获取错误统计
get_error_stats() {
    echo "errors:$ERROR_COUNT,warnings:$WARNING_COUNT"
}

# 重置错误计数
reset_error_stats() {
    ERROR_COUNT=0
    WARNING_COUNT=0
}

# 导出函数供其他脚本使用
export -f init_error_handler log_message print_info print_success print_warning print_error print_debug print_header
export -f handle_error handle_interrupt cleanup_on_exit suggest_recovery generate_error_report cleanup_temp_files
export -f verify_system_requirements safe_execute retry_command ensure_directory backup_file verify_file_integrity
export -f get_error_stats reset_error_stats