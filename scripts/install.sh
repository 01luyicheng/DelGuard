#!/bin/bash
# DelGuard Linux/macOS 安装脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
INSTALL_DIR="/usr/local/bin"
BACKUP_DIR="/usr/local/share/delguard/backup"
DELGUARD_BIN="./delguard"

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}❌ 此脚本需要root权限运行${NC}"
        echo -e "${YELLOW}请使用: sudo $0${NC}"
        exit 1
    fi
}

# 检查操作系统
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        SHELL_RC="/etc/bash.bashrc"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        SHELL_RC="/etc/bashrc"
    else
        echo -e "${RED}❌ 不支持的操作系统: $OSTYPE${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ 检测到操作系统: $OS${NC}"
}

# 安装DelGuard
install_delguard() {
    echo -e "${GREEN}🚀 开始安装 DelGuard...${NC}"
    
    # 检查DelGuard可执行文件
    if [[ ! -f "$DELGUARD_BIN" ]]; then
        echo -e "${RED}❌ 找不到 delguard 可执行文件，请先编译项目${NC}"
        echo -e "${YELLOW}运行: go build -o delguard .${NC}"
        exit 1
    fi
    
    # 创建备份目录
    mkdir -p "$BACKUP_DIR"
    
    # 复制DelGuard到系统目录
    cp "$DELGUARD_BIN" "$INSTALL_DIR/delguard"
    chmod +x "$INSTALL_DIR/delguard"
    echo -e "${GREEN}✅ DelGuard 已安装到 $INSTALL_DIR${NC}"
    
    # 备份原始rm命令
    if [[ -f "$INSTALL_DIR/rm" ]] && [[ ! -f "$BACKUP_DIR/rm.original" ]]; then
        cp "$INSTALL_DIR/rm" "$BACKUP_DIR/rm.original"
        echo -e "${GREEN}✅ 已备份原始rm命令${NC}"
    fi
    
    # 创建rm命令替换脚本
    cat > "$INSTALL_DIR/rm" << 'EOF'
#!/bin/bash
# DelGuard 安全删除脚本
exec /usr/local/bin/delguard delete "$@"
EOF
    chmod +x "$INSTALL_DIR/rm"
    
    # 创建别名（作为备用方案）
    ALIAS_LINE="alias rm='/usr/local/bin/delguard delete'"
    
    # 添加到系统shell配置
    if [[ -f "$SHELL_RC" ]]; then
        if ! grep -q "delguard delete" "$SHELL_RC"; then
            echo "" >> "$SHELL_RC"
            echo "# DelGuard 安全删除别名" >> "$SHELL_RC"
            echo "$ALIAS_LINE" >> "$SHELL_RC"
            echo -e "${GREEN}✅ 已添加rm别名到 $SHELL_RC${NC}"
        fi
    fi
    
    # 添加到用户shell配置
    for user_home in /home/*; do
        if [[ -d "$user_home" ]]; then
            user_bashrc="$user_home/.bashrc"
            user_zshrc="$user_home/.zshrc"
            
            # 添加到.bashrc
            if [[ -f "$user_bashrc" ]]; then
                if ! grep -q "delguard delete" "$user_bashrc"; then
                    echo "" >> "$user_bashrc"
                    echo "# DelGuard 安全删除别名" >> "$user_bashrc"
                    echo "$ALIAS_LINE" >> "$user_bashrc"
                fi
            fi
            
            # 添加到.zshrc
            if [[ -f "$user_zshrc" ]]; then
                if ! grep -q "delguard delete" "$user_zshrc"; then
                    echo "" >> "$user_zshrc"
                    echo "# DelGuard 安全删除别名" >> "$user_zshrc"
                    echo "$ALIAS_LINE" >> "$user_zshrc"
                fi
            fi
        fi
    done
    
    # 创建卸载信息
    cat > "$BACKUP_DIR/install_info.json" << EOF
{
    "install_date": "$(date -Iseconds)",
    "version": "1.0.0",
    "os": "$OS",
    "install_dir": "$INSTALL_DIR",
    "backup_dir": "$BACKUP_DIR"
}
EOF
    
    echo -e "${GREEN}🎉 DelGuard 安装完成！${NC}"
    echo -e "${YELLOW}现在可以使用以下命令：${NC}"
    echo -e "${BLUE}  rm <文件>          - 安全删除文件到回收站${NC}"
    echo -e "${BLUE}  delguard list      - 查看回收站内容${NC}"
    echo -e "${BLUE}  delguard restore   - 恢复文件${NC}"
    echo -e "${BLUE}  delguard status    - 查看状态${NC}"
    echo ""
    echo -e "${YELLOW}⚠️  请重新登录或运行 'source ~/.bashrc' 以使别名生效${NC}"
}

# 卸载DelGuard
uninstall_delguard() {
    echo -e "${YELLOW}🗑️  开始卸载 DelGuard...${NC}"
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        echo -e "${RED}❌ DelGuard 未安装或安装信息丢失${NC}"
        exit 1
    fi
    
    # 恢复原始rm命令
    if [[ -f "$BACKUP_DIR/rm.original" ]]; then
        cp "$BACKUP_DIR/rm.original" "$INSTALL_DIR/rm"
        echo -e "${GREEN}✅ 已恢复原始rm命令${NC}"
    fi
    
    # 删除DelGuard
    rm -f "$INSTALL_DIR/delguard"
    
    # 从shell配置中移除别名
    if [[ -f "$SHELL_RC" ]]; then
        sed -i '/# DelGuard 安全删除别名/d' "$SHELL_RC"
        sed -i '/delguard delete/d' "$SHELL_RC"
    fi
    
    # 从用户shell配置中移除别名
    for user_home in /home/*; do
        if [[ -d "$user_home" ]]; then
            for rc_file in "$user_home/.bashrc" "$user_home/.zshrc"; do
                if [[ -f "$rc_file" ]]; then
                    sed -i '/# DelGuard 安全删除别名/d' "$rc_file"
                    sed -i '/delguard delete/d' "$rc_file"
                fi
            done
        fi
    done
    
    # 删除备份目录
    rm -rf "$BACKUP_DIR"
    rmdir "/usr/local/share/delguard" 2>/dev/null || true
    
    echo -e "${GREEN}✅ DelGuard 已成功卸载${NC}"
}

# 显示帮助
show_help() {
    echo "DelGuard 安装脚本"
    echo ""
    echo "用法:"
    echo "  sudo $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help      显示此帮助信息"
    echo "  -u, --uninstall 卸载DelGuard"
    echo ""
    echo "示例:"
    echo "  sudo $0                # 安装DelGuard"
    echo "  sudo $0 --uninstall    # 卸载DelGuard"
}

# 主逻辑
main() {
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -u|--uninstall)
            check_root
            detect_os
            uninstall_delguard
            ;;
        "")
            check_root
            detect_os
            install_delguard
            ;;
        *)
            echo -e "${RED}❌ 未知选项: $1${NC}"
            show_help
            exit 1
            ;;
    esac
}

main "$@"