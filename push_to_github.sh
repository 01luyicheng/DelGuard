#!/bin/bash
# DelGuard GitHub 发布脚本 (Linux/macOS)
# 用于将 v1.4.1 版本推送到GitHub

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

# 参数设置
OWNER="your-username"  # 替换为你的GitHub用户名
REPO="DelGuard"
VERSION="v1.4.1"

# 帮助信息
show_help() {
    echo -e "${GREEN}🚀 DelGuard GitHub 发布脚本${NC}"
    echo ""
    echo "用法:"
    echo "  ./push_to_github.sh [选项]"
    echo ""
    echo "选项:"
    echo "  -o, --owner OWNER    GitHub用户名 (默认: your-username)"
    echo "  -v, --version VER    版本号 (默认: v1.4.1)"
    echo "  -f, --force          强制推送"
    echo "  -h, --help           显示帮助"
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -o|--owner)
            OWNER="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}❌ 未知选项: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

echo -e "${GREEN}🚀 DelGuard GitHub 发布脚本${NC}"
echo -e "${CYAN}版本: $VERSION${NC}"
echo -e "${CYAN}仓库: $OWNER/$REPO${NC}"
echo ""

# 检查Git是否已初始化
if [[ ! -d ".git" ]]; then
    echo -e "${YELLOW}📁 初始化Git仓库...${NC}"
    git init
    git remote add origin "https://github.com/$OWNER/$Repo.git"
else
    echo -e "${GREEN}✅ Git仓库已存在${NC}"
fi

# 检查远程仓库
if ! git remote get-url origin >/dev/null 2>&1; then
    git remote add origin "https://github.com/$OWNER/$REPO.git"
fi

# 检查是否有未提交的更改
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${YELLOW}📋 检测到未提交的更改:${NC}"
    git status --porcelain
    
    if [[ "$FORCE" != "true" ]]; then
        read -p "是否继续提交更改？(y/N): " response
        if [[ "$response" != "y" && "$response" != "Y" ]]; then
            echo -e "${RED}❌ 操作已取消${NC}"
            exit 1
        fi
    fi
fi

# 添加所有文件
echo -e "${YELLOW}📥 添加文件...${NC}"
git add .

# 提交更改
echo -e "${YELLOW}📝 提交更改...${NC}"
git commit -m "release: 发布 DelGuard $VERSION

- ✨ 新增一键安装功能
- 🔧 支持Windows、Linux、macOS一行命令安装
- 📦 提供完整安装脚本和一行命令脚本
- 🛡️ 智能平台检测和权限验证
- 📖 更新安装文档和使用指南
- 🚀 版本号更新至 $VERSION"

# 创建标签
echo -e "${YELLOW}🏷️  创建标签...${NC}"
git tag -a "$VERSION" -m "DelGuard $VERSION - 一键安装功能发布"

# 推送到GitHub
echo -e "${YELLOW}📤 推送到GitHub...${NC}"
git push -u origin main
git push origin "$VERSION"

echo -e "${GREEN}✅ 推送成功！${NC}"
echo ""
echo -e "${CYAN}🔗 GitHub仓库: https://github.com/$OWNER/$REPO${NC}"
echo -e "${CYAN}🏷️  发布标签: https://github.com/$OWNER/$REPO/releases/tag/$VERSION${NC}"
echo ""
echo -e "${YELLOW}📖 下一步:${NC}"
echo "1. 访问GitHub仓库创建Release"
echo "2. 上传构建好的二进制文件"
echo "3. 发布新版本通知用户"