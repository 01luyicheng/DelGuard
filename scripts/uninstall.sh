#!/bin/bash

# DelGuard 卸载脚本
# 支持 Linux 和 macOS 系统

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
BINARY_NAME="delguard"
INSTALL_DIRS=("/usr/local/bin" "$HOME/.local/bin" "/usr/bin")
CONFIG_DIR="$HOME/.config/delguard"

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
    echo -e "${RED}[ERROR]${NC} $1"
}

# 查找已安装的二进制文件
find_installed_binary() {
    for dir in "${INSTALL_DIRS[@]}"; do
        if [ -f "$dir/$BINARY_NAME" ]; then
            echo "$dir/$BINARY_NAME"
            return 0
        fi
    done
    return 1
}

# 移除二进制文件
remove_binary() {
    local binary_path="$1"
    local binary_dir=$(dirname "$binary_path")
    
    log_info "移除二进制文件: $binary_path"
    
    if [ -w "$binary_dir" ]; then
        rm -f "$binary_path"
    else
        sudo rm -f "$binary_path"
    fi
    
    # 移除备份文件
    if [ -f "${binary_path}.backup" ]; then
        if [ -w "$binary_dir" ]; then
            rm -f "${binary_path}.backup"
        else
            sudo rm -f "${binary_path}.backup"
        fi
    fi
    
    log_success "二进制文件已移除"
}

# 移除配置文件
remove_config() {
    if [ -d "$CONFIG_DIR" ]; then
        log_info "移除配置目录: $CONFIG_DIR"
        
        read -p "是否保留配置文件和日志? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$CONFIG_DIR"
            log_success "配置目录已移除"
        else
            log_info "配置目录已保留"
        fi
    else
        log_info "未找到配置目录"
    fi
}

# 移除 Shell 别名
remove_aliases() {
    log_info "移除 Shell 别名..."
    
    local shell_configs=(
        "$HOME/.bashrc"
        "$HOME/.bash_profile"
        "$HOME/.zshrc"
        "$HOME/.config/fish/config.fish"
    )
    
    for config_file in "${shell_configs[@]}"; do
        if [ -f "$config_file" ]; then
            # 检查是否包含 DelGuard 别名
            if grep -q "DelGuard aliases" "$config_file" 2>/dev/null; then
                log_info "从 $config_file 移除别名..."
                
                # 创建临时文件
                temp_file=$(mktemp)
                
                # 移除 DelGuard 相关的行
                awk '
                /# DelGuard aliases/ { skip = 1; next }
                /^alias del=/ && skip { next }
                /^alias rm-safe=/ && skip { next }
                /^alias trash=/ && skip { next }
                /^alias restore=/ && skip { next }
                /^alias empty-trash=/ && skip { next }
                /^$/ && skip { skip = 0; next }
                { if (!skip) print }
                ' "$config_file" > "$temp_file"
                
                # 替换原文件
                mv "$temp_file" "$config_file"
                log_success "已从 $config_file 移除别名"
            fi
        fi
    done
}

# 清理回收站
clean_trash() {
    log_info "检查回收站..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        read -p "是否清空回收站? [y/N]: " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            "$BINARY_NAME" empty --force 2>/dev/null || true
            log_success "回收站已清空"
        fi
    fi
}

# 主函数
main() {
    echo "🗑️  DelGuard 卸载脚本"
    echo "===================="
    
    # 查找已安装的二进制文件
    if binary_path=$(find_installed_binary); then
        log_info "找到已安装的 DelGuard: $binary_path"
        
        # 确认卸载
        echo ""
        read -p "确认卸载 DelGuard? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "卸载已取消"
            exit 0
        fi
        
        # 清理回收站
        clean_trash
        
        # 执行卸载步骤
        remove_binary "$binary_path"
        remove_aliases
        remove_config
        
        log_success "DelGuard 已完全卸载"
        log_info "感谢使用 DelGuard！"
        
    else
        log_warning "未找到已安装的 DelGuard"
        log_info "可能的安装位置:"
        for dir in "${INSTALL_DIRS[@]}"; do
            echo "  - $dir/$BINARY_NAME"
        done
    fi
}

main "$@"