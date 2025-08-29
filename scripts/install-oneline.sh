#!/bin/bash
# DelGuard 一行命令安装脚本 (Linux/macOS)
# 使用方法：复制粘贴以下命令到终端即可
# curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install-oneline.sh | sudo bash

# 检查root权限
if [[ $EUID -ne 0 ]]; then
    echo "❌ 此脚本需要root权限运行"
    echo "请使用: curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install-oneline.sh | sudo bash"
    exit 1
fi

# 设置参数
OWNER="01luyicheng"  # GitHub用户名
REPO="DelGuard"
VERSION="v1.4.1"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# 检测操作系统和架构
detect_platform() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    else
        echo -e "${RED}❌ 不支持的操作系统: $OSTYPE${NC}"
        exit 1
    fi
    
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) echo -e "${RED}❌ 不支持的架构: $ARCH${NC}"; exit 1 ;;
    esac
}

# 主安装函数
install_delguard() {
    echo -e "${GREEN}🚀 正在安装 DelGuard $VERSION...${NC}"
    
    detect_platform
    
    local download_url="https://github.com/$OWNER/$REPO/releases/download/$VERSION/delguard-${OS}-${ARCH}"
    local install_dir="/usr/local/bin"
    local temp_file=$(mktemp)
    
    echo -e "${CYAN}📥 正在下载...${NC}"
    
    # 下载文件
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_file" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_file" "$download_url"
    else
        echo -e "${RED}❌ 未找到 curl 或 wget${NC}"
        exit 1
    fi
    
    if [[ ! -f "$temp_file" ]]; then
        echo -e "${RED}❌ 下载失败${NC}"
        exit 1
    fi
    
    chmod +x "$temp_file"
    
    # 安装
    mv "$temp_file" "$install_dir/delguard"
    
    # 创建rm命令替换
    cat > "$install_dir/rm" << 'EOF'
#!/bin/bash
exec /usr/local/bin/delguard delete "$@"
EOF
    chmod +x "$install_dir/rm"
    
    # 创建卸载脚本
    cat > "$install_dir/delguard-uninstall" << 'EOF'
#!/bin/bash
if [[ $EUID -ne 0 ]]; then
    echo "请使用: sudo $0"
    exit 1
fi
echo "🗑️  正在卸载 DelGuard..."
rm -f /usr/local/bin/delguard
rm -f /usr/local/bin/rm
rm -f /usr/local/bin/delguard-uninstall
echo "✅ DelGuard 已成功卸载"
EOF
    chmod +x "$install_dir/delguard-uninstall"
    
    echo -e "${GREEN}✅ DelGuard 安装完成！${NC}"
    echo -e "${YELLOW}📖 使用说明:${NC}"
    echo -e "  delguard --help    - 查看帮助"
    echo -e "  delguard list      - 查看回收站"
    echo -e "  delguard restore   - 恢复文件"
    echo -e "  delguard-uninstall - 卸载程序"
    echo -e "${YELLOW}⚠️  请重新登录或运行 'source ~/.bashrc' 以使别名生效${NC}"
}

# 执行安装
install_delguard