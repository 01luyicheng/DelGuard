#!/bin/bash
# DelGuard 一键安装脚本 (Linux/macOS)
# 从GitHub下载最新版本并自动安装

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 默认配置
VERSION="v1.4.1"
REPO="DelGuard"
OWNER="your-username"  # 需要替换为实际的GitHub用户名
FORCE=false

# 帮助信息
show_help() {
    echo -e "${GREEN}🎯 DelGuard 一键安装脚本${NC}"
    echo ""
    echo "用法:"
    echo "  sudo $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -v, --version VERSION    指定版本 (默认: v1.4.1)"
    echo "  -f, --force              强制重新安装"
    echo "  -o, --owner OWNER        GitHub用户名 (默认: your-username)"
    echo "  -h, --help               显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  sudo $0"
    echo "  sudo $0 -v v1.4.1"
    echo "  sudo $0 -f"
    echo "  sudo $0 --version latest"
}

# 输出带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 检查root权限
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_message $RED "❌ 此脚本需要root权限运行"
        print_message $YELLOW "请使用: sudo $0"
        exit 1
    fi
}

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        ARCH=$(uname -m)
        case $ARCH in
            x86_64) ARCH="amd64" ;;
            aarch64) ARCH="arm64" ;;
            armv7l) ARCH="arm" ;;
        esac
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        ARCH=$(uname -m)
        case $ARCH in
            x86_64) ARCH="amd64" ;;
            arm64) ARCH="arm64" ;;
        esac
    else
        print_message $RED "❌ 不支持的操作系统: $OSTYPE"
        exit 1
    fi
    
    print_message $GREEN "✅ 检测到操作系统: $OS $ARCH"
}

# 获取最新版本
get_latest_version() {
    local api_url="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
    
    if command -v curl >/dev/null 2>&1; then
        curl -s -H "Accept: application/vnd.github.v3+json" "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "$VERSION"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- --header="Accept: application/vnd.github.v3+json" "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "$VERSION"
    else
        print_message $YELLOW "⚠️ 无法获取最新版本，使用指定版本: $VERSION"
        echo "$VERSION"
    fi
}

# 下载DelGuard
download_delguard() {
    local version=$1
    local filename="delguard-${OS}-${ARCH}"
    local download_url="https://github.com/$OWNER/$REPO/releases/download/$version/$filename"
    local temp_dir=$(mktemp -d)
    local download_path="$temp_dir/delguard"
    
    print_message $CYAN "📥 正在下载 DelGuard $version..."
    print_message $WHITE "下载地址: $download_url"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$download_path" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$download_path" "$download_url"
    else
        print_message $RED "❌ 未找到 curl 或 wget"
        exit 1
    fi
    
    if [[ ! -f "$download_path" ]]; then
        print_message $RED "❌ 下载失败"
        exit 1
    fi
    
    chmod +x "$download_path"
    print_message $GREEN "✅ 下载完成"
    echo "$download_path"
}

# 安装DelGuard
install_delguard() {
    local binary_path=$1
    local install_dir="/usr/local/bin"
    local backup_dir="/usr/local/share/delguard/backup"
    
    # 检查是否已安装
    if [[ -f "$install_dir/delguard" ]]; then
        if [[ "$FORCE" != "true" ]]; then
            print_message $YELLOW "⚠️ DelGuard 已安装，使用 -f 参数重新安装"
            return 1
        fi
        print_message $YELLOW "🔄 检测到现有安装，正在重新安装..."
    fi
    
    # 创建备份目录
    mkdir -p "$backup_dir"
    
    # 安装DelGuard
    cp "$binary_path" "$install_dir/delguard"
    chmod +x "$install_dir/delguard"
    print_message $GREEN "✅ DelGuard 已安装到 $install_dir"
    
    # 备份原始rm命令
    if [[ -f "$install_dir/rm" ]] && [[ ! -f "$backup_dir/rm.original" ]]; then
        cp "$install_dir/rm" "$backup_dir/rm.original"
        print_message $GREEN "✅ 已备份原始rm命令"
    fi
    
    # 创建rm命令替换脚本
    cat > "$install_dir/rm" << 'EOF'
#!/bin/bash
# DelGuard 安全删除脚本
exec /usr/local/bin/delguard delete "$@"
EOF
    chmod +x "$install_dir/rm"
    
    # 创建卸载脚本
    cat > "$install_dir/delguard-uninstall" << EOF
#!/bin/bash
# DelGuard 卸载脚本
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

if [[ \$EUID -ne 0 ]]; then
    echo -e "\${RED}❌ 需要root权限运行\${NC}"
    echo -e "\${YELLOW}请使用: sudo \$0\${NC}"
    exit 1
fi

echo -e "\${YELLOW}🗑️  正在卸载 DelGuard...\${NC}"

# 恢复原始rm命令
if [[ -f "/usr/local/share/delguard/backup/rm.original" ]]; then
    cp "/usr/local/share/delguard/backup/rm.original" "/usr/local/bin/rm"
    echo -e "\${GREEN}✅ 已恢复原始rm命令\${NC}"
fi

# 删除DelGuard
rm -f "/usr/local/bin/delguard"
rm -f "/usr/local/bin/rm"
rm -f "/usr/local/bin/delguard-uninstall"

# 删除备份目录
rm -rf "/usr/local/share/delguard"

echo -e "\${GREEN}✅ DelGuard 已成功卸载\${NC}"
EOF
    chmod +x "$install_dir/delguard-uninstall"
    
    # 创建安装信息
    cat > "$backup_dir/install_info.json" << EOF
{
    "install_date": "$(date -Iseconds)",
    "version": "$VERSION",
    "os": "$OS",
    "arch": "$ARCH",
    "install_dir": "$install_dir",
    "backup_dir": "$backup_dir"
}
EOF
    
    return 0
}

# 清理函数
cleanup() {
    if [[ -n "$TEMP_BINARY" ]] && [[ -f "$TEMP_BINARY" ]]; then
        rm -f "$TEMP_BINARY"
        rmdir "$(dirname "$TEMP_BINARY")" 2>/dev/null || true
    fi
}

# 设置清理陷阱
trap cleanup EXIT

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -o|--owner)
            OWNER="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            print_message $RED "❌ 未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 主程序
print_message $GREEN "🚀 DelGuard 一键安装程序"
print_message $WHITE "从GitHub下载并安装最新版本"
echo ""

# 检查依赖
check_root
detect_os

# 处理版本号
if [[ "$VERSION" == "latest" ]]; then
    VERSION=$(get_latest_version)
elif [[ "$VERSION" != v* ]]; then
    VERSION="v$VERSION"
fi

print_message $CYAN "📦 版本: $VERSION"

# 下载并安装
if TEMP_BINARY=$(download_delguard "$VERSION"); then
    if install_delguard "$TEMP_BINARY"; then
        print_message $GREEN ""
        print_message $GREEN "🎉 安装完成！"
        print_message $GREEN ""
        print_message $YELLOW "📖 使用说明:"
        print_message $BLUE "  delguard --help    - 查看帮助"
        print_message $BLUE "  delguard list      - 查看回收站"
        print_message $BLUE "  delguard restore   - 恢复文件"
        print_message $BLUE "  delguard status    - 查看状态"
        print_message $BLUE "  delguard-uninstall - 卸载程序"
        print_message $YELLOW ""
        print_message $YELLOW "⚠️  请重新登录或运行 'source ~/.bashrc' 以使别名生效"
    fi
fi