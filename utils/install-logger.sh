#!/bin/bash
# DelGuard Installation Logger and Error Handler (Unix Shell)
# Version: 2.1.1 Enhanced
# Provides unified logging, error handling, and user interaction functions

# Global configuration
DELGUARD_LOG_LEVEL="${DELGUARD_LOG_LEVEL:-INFO}"
DELGUARD_LOG_FILE=""
DELGUARD_LANGUAGE="${DELGUARD_LANGUAGE:-zh-cn}"
DELGUARD_COLOR_OUTPUT="${DELGUARD_COLOR_OUTPUT:-true}"
DELGUARD_INTERACTIVE="${DELGUARD_INTERACTIVE:-true}"
DELGUARD_INSTALLATION_ID=""
DELGUARD_START_TIME=""

# Error codes
declare -A ERROR_CODES=(
    ["DEPS_MISSING"]=1001
    ["BUILD_FAILED"]=1002
    ["INSTALL_FAILED"]=1003
    ["PERMISSION_DENIED"]=1004
    ["PATH_NOT_FOUND"]=1005
    ["NETWORK_ERROR"]=1006
    ["FILE_ERROR"]=1007
    ["VERIFICATION_FAILED"]=1008
    ["CONFIG_ERROR"]=1009
    ["UNKNOWN_ERROR"]=9999
)

# Initialize installation session
init_install_session() {
    local log_path="$1"
    local language="${2:-zh-cn}"
    local interactive="${3:-true}"
    
    DELGUARD_INSTALLATION_ID="$(date +%s | sha256sum | cut -c1-8 2>/dev/null || echo "$(date +%s)")"
    DELGUARD_LANGUAGE="$language"
    DELGUARD_INTERACTIVE="$interactive"
    DELGUARD_START_TIME="$(date '+%Y-%m-%d %H:%M:%S')"
    
    if [[ -z "$log_path" ]]; then
        local timestamp="$(date +%Y%m%d-%H%M%S)"
        log_path="/tmp/delguard-install-$timestamp-$DELGUARD_INSTALLATION_ID.log"
    fi
    
    DELGUARD_LOG_FILE="$log_path"
    
    # Create log file and write header
    cat > "$DELGUARD_LOG_FILE" << EOF
DelGuard Installation Log
========================
Installation ID: $DELGUARD_INSTALLATION_ID
Start Time: $DELGUARD_START_TIME
Platform: $(uname -s) $(uname -m)
User: $(whoami)
Shell: $SHELL
Language: $DELGUARD_LANGUAGE
Interactive: $DELGUARD_INTERACTIVE

EOF
    
    log_info "Installation session initialized"
}

# Color output functions
get_color_code() {
    local color="$1"
    if [[ "$DELGUARD_COLOR_OUTPUT" != "true" ]] || ! command -v tput >/dev/null 2>&1 || [[ ! -t 1 ]]; then
        echo ""
        return
    fi
    
    case "$color" in
        red)     tput setaf 1 ;;
        green)   tput setaf 2 ;;
        yellow)  tput setaf 3 ;;
        blue)    tput setaf 4 ;;
        magenta) tput setaf 5 ;;
        cyan)    tput setaf 6 ;;
        white)   tput setaf 7 ;;
        bold)    tput bold ;;
        reset)   tput sgr0 ;;
        *)       echo "" ;;
    esac
}

# Logging functions
log_message() {
    local level="$1"
    local message="$2"
    local timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    local log_entry="$timestamp [$level] $message"
    
    # Write to log file
    if [[ -n "$DELGUARD_LOG_FILE" ]]; then
        echo "$log_entry" >> "$DELGUARD_LOG_FILE" 2>/dev/null || true
    fi
    
    # Output to console based on level
    case "$level" in
        ERROR)
            echo "$(get_color_code red)✗ $message$(get_color_code reset)" >&2
            ;;
        WARN)
            echo "$(get_color_code yellow)⚠ $message$(get_color_code reset)" >&2
            ;;
        SUCCESS)
            echo "$(get_color_code green)✓ $message$(get_color_code reset)"
            ;;
        INFO)
            echo "$(get_color_code cyan)ℹ $message$(get_color_code reset)"
            ;;
        DEBUG)
            if [[ "$DELGUARD_LOG_LEVEL" == "DEBUG" ]]; then
                echo "$(get_color_code blue)[DEBUG] $message$(get_color_code reset)"
            fi
            ;;
        HEADER)
            echo "$(get_color_code bold)$(get_color_code magenta)$message$(get_color_code reset)"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

log_error() { log_message "ERROR" "$1"; }
log_warn() { log_message "WARN" "$1"; }
log_success() { log_message "SUCCESS" "$1"; }
log_info() { log_message "INFO" "$1"; }
log_debug() { log_message "DEBUG" "$1"; }
log_header() { log_message "HEADER" "$1"; }

# Load messages from JSON config
load_messages() {
    local config_file="$1"
    if [[ ! -f "$config_file" ]]; then
        log_warn "Message config file not found: $config_file"
        return 1
    fi
    
    # Simple JSON parsing for messages (fallback if jq not available)
    if command -v jq >/dev/null 2>&1; then
        # Use jq for proper JSON parsing
        MESSAGES_JSON="$(cat "$config_file")"
    else
        # Fallback: basic parsing
        log_debug "jq not available, using basic JSON parsing"
    fi
}

# Get localized message
get_message() {
    local key="$1"
    shift
    local args=("$@")
    
    # Default messages (fallback)
    case "$key" in
        "welcome")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "欢迎使用 DelGuard 安全删除工具安装程序"
            else
                echo "Welcome to DelGuard Safe Delete Tool Installer"
            fi
            ;;
        "platform_detected")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "检测到平台: ${args[0]} (${args[1]})"
            else
                echo "Platform detected: ${args[0]} (${args[1]})"
            fi
            ;;
        "checking_deps")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "正在检查系统依赖..."
            else
                echo "Checking system dependencies..."
            fi
            ;;
        "deps_ok")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "所有依赖项检查通过"
            else
                echo "All dependencies check passed"
            fi
            ;;
        "deps_missing")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "缺少必需的依赖项: ${args[0]}"
            else
                echo "Missing required dependencies: ${args[0]}"
            fi
            ;;
        "building")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "正在构建 DelGuard 可执行文件..."
            else
                echo "Building DelGuard executable..."
            fi
            ;;
        "build_success")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "构建成功完成"
            else
                echo "Build completed successfully"
            fi
            ;;
        "build_failed")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "构建失败: ${args[0]}"
            else
                echo "Build failed: ${args[0]}"
            fi
            ;;
        "installing")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "正在安装到: ${args[0]}"
            else
                echo "Installing to: ${args[0]}"
            fi
            ;;
        "install_success")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "安装成功完成！"
            else
                echo "Installation completed successfully!"
            fi
            ;;
        "install_failed")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "安装失败: ${args[0]}"
            else
                echo "Installation failed: ${args[0]}"
            fi
            ;;
        "next_steps")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "下一步操作:"
            else
                echo "Next steps:"
            fi
            ;;
        "restart_shell")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "重启终端或运行: source ~/.bashrc"
            else
                echo "Restart terminal or run: source ~/.bashrc"
            fi
            ;;
        "test_command")
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                echo "测试命令: delguard --version"
            else
                echo "Test command: delguard --version"
            fi
            ;;
        *)
            echo "$key"
            ;;
    esac
}

# Error handling with recovery suggestions
handle_error() {
    local error_code="$1"
    local error_message="$2"
    local context="$3"
    
    log_error "Error $error_code: $error_message"
    
    if [[ -n "$context" ]]; then
        log_info "Context: $context"
    fi
    
    # Provide recovery suggestions
    case "$error_code" in
        "${ERROR_CODES[DEPS_MISSING]}")
            log_info "Recovery suggestions:"
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                log_info "• 使用包管理器安装缺失的依赖"
                log_info "• 检查网络连接"
                log_info "• 确保有足够的权限"
            else
                log_info "• Install missing dependencies using package manager"
                log_info "• Check network connection"
                log_info "• Ensure sufficient permissions"
            fi
            ;;
        "${ERROR_CODES[BUILD_FAILED]}")
            log_info "Recovery suggestions:"
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                log_info "• 检查 Go 语言环境是否正确安装"
                log_info "• 确保项目目录完整"
                log_info "• 尝试手动运行: go mod tidy && go build"
            else
                log_info "• Check if Go environment is properly installed"
                log_info "• Ensure project directory is complete"
                log_info "• Try manual build: go mod tidy && go build"
            fi
            ;;
        "${ERROR_CODES[PERMISSION_DENIED]}")
            log_info "Recovery suggestions:"
            if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                log_info "• 使用 sudo 运行脚本"
                log_info "• 选择用户目录进行安装"
                log_info "• 检查目标目录权限"
            else
                log_info "• Run script with sudo"
                log_info "• Choose user directory for installation"
                log_info "• Check target directory permissions"
            fi
            ;;
    esac
    
    # Log final error summary
    log_error "Installation failed with error code: $error_code"
    log_info "Full installation log saved to: $DELGUARD_LOG_FILE"
    
    return "$error_code"
}

# Progress indicator
show_progress() {
    local current="$1"
    local total="$2"
    local message="$3"
    
    if [[ "$DELGUARD_INTERACTIVE" == "true" ]]; then
        local percent=$((current * 100 / total))
        local filled=$((percent / 2))
        local empty=$((50 - filled))
        
        printf "\r$(get_color_code cyan)[${"#" * filled}${" " * empty}] %d%% %s$(get_color_code reset)" "$percent" "$message"
        
        if [[ "$current" -eq "$total" ]]; then
            echo ""
        fi
    else
        log_info "$message ($current/$total)"
    fi
}

# User confirmation
confirm_action() {
    local message="$1"
    local default="${2:-n}"
    
    if [[ "$DELGUARD_INTERACTIVE" != "true" ]]; then
        return 0
    fi
    
    local prompt
    if [[ "$default" == "y" ]]; then
        prompt="$message [Y/n]: "
    else
        prompt="$message [y/N]: "
    fi
    
    while true; do
        printf "$(get_color_code yellow)$prompt$(get_color_code reset)"
        read -r response
        
        case "$response" in
            [Yy]|[Yy][Ee][Ss])
                return 0
                ;;
            [Nn]|[Nn][Oo])
                return 1
                ;;
            "")
                if [[ "$default" == "y" ]]; then
                    return 0
                else
                    return 1
                fi
                ;;
            *)
                if [[ "$DELGUARD_LANGUAGE" == "zh-cn" ]]; then
                    echo "请输入 y 或 n"
                else
                    echo "Please enter y or n"
                fi
                ;;
        esac
    done
}

# Cleanup function
cleanup_install_session() {
    local exit_code="${1:-0}"
    local end_time="$(date '+%Y-%m-%d %H:%M:%S')"
    
    if [[ -n "$DELGUARD_LOG_FILE" ]]; then
        cat >> "$DELGUARD_LOG_FILE" << EOF

Installation Summary
===================
End Time: $end_time
Duration: $(($(date +%s) - $(date -d "$DELGUARD_START_TIME" +%s 2>/dev/null || echo 0))) seconds
Exit Code: $exit_code
Status: $(if [[ "$exit_code" -eq 0 ]]; then echo "SUCCESS"; else echo "FAILED"; fi)

EOF
    fi
    
    if [[ "$exit_code" -eq 0 ]]; then
        log_success "Installation completed successfully!"
    else
        log_error "Installation failed with exit code: $exit_code"
    fi
    
    log_info "Installation log saved to: $DELGUARD_LOG_FILE"
}

# Export functions for use in other scripts
export -f init_install_session
export -f log_error log_warn log_success log_info log_debug log_header
export -f get_message handle_error show_progress confirm_action
export -f cleanup_install_session