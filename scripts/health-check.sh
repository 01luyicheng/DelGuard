#!/bin/bash

# DelGuard 健康检查脚本
# 定期检查DelGuard安装状态和运行健康度

# 导入错误处理库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -f "$SCRIPT_DIR/lib/error-handler.sh" ]]; then
    source "$SCRIPT_DIR/lib/error-handler.sh"
    init_error_handler
else
    # 基本错误处理
    print_info() { echo "[INFO] $1"; }
    print_success() { echo "[SUCCESS] $1"; }
    print_warning() { echo "[WARNING] $1"; }
    print_error() { echo "[ERROR] $1"; }
    print_header() { echo "$1"; echo "$(printf '=%.0s' {1..50})"; }
fi

# 配置
HEALTH_CHECK_LOG="$HOME/.delguard-health.log"
CONFIG_DIR="$HOME/.config/delguard"
INSTALL_DIR="/usr/local/bin"

# 健康检查项目
CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNING=0

# 记录健康检查结果
log_health_check() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $1" >> "$HEALTH_CHECK_LOG"
}

# 检查DelGuard二进制文件
check_binary() {
    print_header "🔧 检查DelGuard二进制文件"
    
    local delguard_path=$(command -v delguard 2>/dev/null)
    
    if [[ -z "$delguard_path" ]]; then
        print_error "DelGuard未找到或不在PATH中"
        log_health_check "ERROR: DelGuard binary not found"
        ((CHECKS_FAILED++))
        return 1
    fi
    
    print_success "DelGuard位置: $delguard_path"
    
    # 检查文件权限
    if [[ ! -x "$delguard_path" ]]; then
        print_error "DelGuard没有执行权限"
        log_health_check "ERROR: DelGuard not executable"
        ((CHECKS_FAILED++))
        return 1
    fi
    
    # 检查文件完整性
    local file_size=$(stat -c%s "$delguard_path" 2>/dev/null || stat -f%z "$delguard_path" 2>/dev/null)
    if [[ $file_size -lt 1000000 ]]; then  # 小于1MB可能有问题
        print_warning "DelGuard文件大小异常: ${file_size} bytes"
        log_health_check "WARNING: DelGuard file size unusual: $file_size bytes"
        ((CHECKS_WARNING++))
    else
        print_success "文件大小正常: ${file_size} bytes"
    fi
    
    ((CHECKS_PASSED++))
    log_health_check "SUCCESS: DelGuard binary check passed"
    return 0
}

# 检查基本功能
check_functionality() {
    print_header "⚡ 检查DelGuard功能"
    
    local delguard_path=$(command -v delguard)
    
    # 测试版本命令
    if version_output=$("$delguard_path" --version 2>&1); then
        print_success "版本命令正常: $version_output"
        log_health_check "SUCCESS: Version command works: $version_output"
    else
        print_error "版本命令失败"
        log_health_check "ERROR: Version command failed"
        ((CHECKS_FAILED++))
        return 1
    fi
    
    # 测试帮助命令
    if "$delguard_path" --help &>/dev/null; then
        print_success "帮助命令正常"
        log_health_check "SUCCESS: Help command works"
    else
        print_error "帮助命令失败"
        log_health_check "ERROR: Help command failed"
        ((CHECKS_FAILED++))
        return 1
    fi
    
    # 测试子命令
    local commands=("delete" "restore" "list" "empty")
    for cmd in "${commands[@]}"; do
        if "$delguard_path" "$cmd" --help &>/dev/null; then
            print_success "子命令 '$cmd' 可用"
        else
            print_warning "子命令 '$cmd' 可能有问题"
            log_health_check "WARNING: Subcommand '$cmd' may have issues"
            ((CHECKS_WARNING++))
        fi
    done
    
    ((CHECKS_PASSED++))
    log_health_check "SUCCESS: DelGuard functionality check passed"
    return 0
}

# 检查配置系统
check_configuration() {
    print_header "⚙️ 检查配置系统"
    
    # 检查配置目录
    if [[ ! -d "$CONFIG_DIR" ]]; then
        print_warning "配置目录不存在: $CONFIG_DIR"
        log_health_check "WARNING: Config directory missing: $CONFIG_DIR"
        ((CHECKS_WARNING++))
        
        # 尝试创建配置目录
        if mkdir -p "$CONFIG_DIR" 2>/dev/null; then
            print_success "已创建配置目录"
            log_health_check "SUCCESS: Created config directory"
        else
            print_error "无法创建配置目录"
            log_health_check "ERROR: Cannot create config directory"
            ((CHECKS_FAILED++))
            return 1
        fi
    else
        print_success "配置目录存在: $CONFIG_DIR"
    fi
    
    # 检查配置文件
    local config_file="$CONFIG_DIR/config.yaml"
    if [[ ! -f "$config_file" ]]; then
        print_warning "配置文件不存在，将使用默认配置"
        log_health_check "WARNING: Config file missing, using defaults"
        ((CHECKS_WARNING++))
    else
        print_success "配置文件存在"
        
        # 检查配置文件权限
        if [[ ! -r "$config_file" ]]; then
            print_error "配置文件不可读"
            log_health_check "ERROR: Config file not readable"
            ((CHECKS_FAILED++))
            return 1
        fi
    fi
    
    # 检查目录权限
    if [[ ! -w "$CONFIG_DIR" ]]; then
        print_error "配置目录不可写"
        log_health_check "ERROR: Config directory not writable"
        ((CHECKS_FAILED++))
        return 1
    fi
    
    ((CHECKS_PASSED++))
    log_health_check "SUCCESS: Configuration check passed"
    return 0
}

# 检查回收站功能
check_trash_functionality() {
    print_header "🗑️ 检查回收站功能"
    
    local delguard_path=$(command -v delguard)
    
    # 创建测试文件
    local test_dir="/tmp/delguard-health-test"
    local test_file="$test_dir/test-file.txt"
    
    mkdir -p "$test_dir"
    echo "DelGuard健康检查测试文件" > "$test_file"
    
    # 测试删除功能
    if "$delguard_path" delete "$test_file" --force &>/dev/null; then
        if [[ ! -f "$test_file" ]]; then
            print_success "删除功能正常"
            log_health_check "SUCCESS: Delete functionality works"
        else
            print_error "删除功能异常：文件仍然存在"
            log_health_check "ERROR: Delete function failed - file still exists"
            ((CHECKS_FAILED++))
        fi
    else
        print_warning "删除功能测试失败（可能需要配置回收站）"
        log_health_check "WARNING: Delete function test failed"
        ((CHECKS_WARNING++))
    fi
    
    # 测试列表功能
    if "$delguard_path" list &>/dev/null; then
        print_success "列表功能正常"
        log_health_check "SUCCESS: List functionality works"
    else
        print_warning "列表功能可能有问题"
        log_health_check "WARNING: List functionality may have issues"
        ((CHECKS_WARNING++))
    fi
    
    # 清理测试文件
    rm -rf "$test_dir" 2>/dev/null || true
    
    ((CHECKS_PASSED++))
    return 0
}

# 检查系统资源
check_system_resources() {
    print_header "💻 检查系统资源"
    
    # 检查磁盘空间
    local config_disk_usage=$(df "$CONFIG_DIR" 2>/dev/null | awk 'NR==2 {print $5}' | sed 's/%//')
    if [[ -n "$config_disk_usage" ]] && [[ $config_disk_usage -lt 90 ]]; then
        print_success "配置目录磁盘空间充足 (使用率: ${config_disk_usage}%)"
        log_health_check "SUCCESS: Disk space sufficient: ${config_disk_usage}%"
    else
        print_warning "配置目录磁盘空间可能不足 (使用率: ${config_disk_usage}%)"
        log_health_check "WARNING: Disk space may be insufficient: ${config_disk_usage}%"
        ((CHECKS_WARNING++))
    fi
    
    # 检查内存使用
    if command -v free &>/dev/null; then
        local mem_usage=$(free | awk 'NR==2{printf "%.1f", $3*100/$2}')
        print_info "系统内存使用率: ${mem_usage}%"
        log_health_check "INFO: Memory usage: ${mem_usage}%"
    fi
    
    # 检查日志文件大小
    local log_file="$CONFIG_DIR/delguard.log"
    if [[ -f "$log_file" ]]; then
        local log_size=$(stat -c%s "$log_file" 2>/dev/null || stat -f%z "$log_file" 2>/dev/null)
        local log_size_mb=$((log_size / 1024 / 1024))
        
        if [[ $log_size_mb -gt 100 ]]; then  # 大于100MB
            print_warning "日志文件较大: ${log_size_mb}MB，建议清理"
            log_health_check "WARNING: Log file large: ${log_size_mb}MB"
            ((CHECKS_WARNING++))
        else
            print_success "日志文件大小正常: ${log_size_mb}MB"
        fi
    fi
    
    ((CHECKS_PASSED++))
    return 0
}

# 检查更新
check_updates() {
    print_header "🔄 检查更新"
    
    local delguard_path=$(command -v delguard)
    local current_version=$("$delguard_path" --version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "unknown")
    
    print_info "当前版本: $current_version"
    
    # 检查最新版本
    if command -v curl &>/dev/null; then
        local latest_version=$(curl -s "https://api.github.com/repos/01luyicheng/DelGuard/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "unknown")
        
        if [[ "$latest_version" != "unknown" ]]; then
            print_info "最新版本: $latest_version"
            
            if [[ "$current_version" != "$latest_version" ]]; then
                print_warning "有新版本可用: $latest_version"
                log_health_check "WARNING: New version available: $latest_version (current: $current_version)"
                ((CHECKS_WARNING++))
            else
                print_success "版本是最新的"
                log_health_check "SUCCESS: Version is up to date"
            fi
        else
            print_warning "无法检查最新版本"
            log_health_check "WARNING: Cannot check latest version"
            ((CHECKS_WARNING++))
        fi
    else
        print_warning "curl不可用，跳过版本检查"
        ((CHECKS_WARNING++))
    fi
    
    ((CHECKS_PASSED++))
    return 0
}

# 生成健康报告
generate_health_report() {
    print_header "📊 健康检查报告"
    
    local total_checks=$((CHECKS_PASSED + CHECKS_FAILED + CHECKS_WARNING))
    
    echo "检查完成时间: $(date)"
    echo "总检查项目: $total_checks"
    echo "通过检查: $CHECKS_PASSED"
    echo "失败检查: $CHECKS_FAILED"
    echo "警告检查: $CHECKS_WARNING"
    echo
    
    # 计算健康分数
    local health_score=0
    if [[ $total_checks -gt 0 ]]; then
        health_score=$(( (CHECKS_PASSED * 100) / total_checks ))
    fi
    
    echo "健康分数: ${health_score}%"
    
    # 记录总体结果
    log_health_check "SUMMARY: Health score: ${health_score}% (Passed: $CHECKS_PASSED, Failed: $CHECKS_FAILED, Warnings: $CHECKS_WARNING)"
    
    if [[ $CHECKS_FAILED -eq 0 ]]; then
        if [[ $CHECKS_WARNING -eq 0 ]]; then
            print_success "🎉 DelGuard运行状态完全正常！"
            return 0
        else
            print_warning "⚠️ DelGuard基本正常，但有 $CHECKS_WARNING 个警告项目"
            return 0
        fi
    else
        print_error "❌ DelGuard有 $CHECKS_FAILED 个严重问题需要修复"
        echo
        echo "建议操作："
        echo "1. 运行修复脚本: $SCRIPT_DIR/repair-install.sh"
        echo "2. 重新安装DelGuard"
        echo "3. 查看详细日志: $HEALTH_CHECK_LOG"
        return 1
    fi
}

# 主函数
main() {
    print_header "🛡️  DelGuard 健康检查"
    
    # 初始化日志
    log_health_check "=== DelGuard Health Check Started ==="
    
    # 执行所有检查
    check_binary
    check_functionality
    check_configuration
    check_trash_functionality
    check_system_resources
    check_updates
    
    # 生成报告
    generate_health_report
    
    log_health_check "=== DelGuard Health Check Completed ==="
}

# 运行主函数
main "$@"