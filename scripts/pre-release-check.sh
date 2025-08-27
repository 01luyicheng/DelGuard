#!/bin/bash

# DelGuard 发布前检查脚本 - Unix版本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;37m'
NC='\033[0m' # No Color

# 参数解析
VERBOSE=false
SKIP_BUILD=false
SKIP_TESTS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose) VERBOSE=true; shift ;;
        --skip-build) SKIP_BUILD=true; shift ;;
        --skip-tests) SKIP_TESTS=true; shift ;;
        *) echo "未知参数: $1"; exit 1 ;;
    esac
done

echo -e "${CYAN}🚀 DelGuard 发布前检查${NC}"
echo -e "${CYAN}===================${NC}"

# 1. 检查Go环境
echo -e "\n${YELLOW}📦 检查Go环境...${NC}"
if command -v go &> /dev/null; then
    go_version=$(go version)
    echo -e "${GREEN}✅ Go环境正常: $go_version${NC}"
else
    echo -e "${RED}❌ Go环境未找到${NC}"
    exit 1
fi

# 2. 检查项目结构
echo -e "\n${YELLOW}📁 检查项目结构...${NC}"
required_files=(
    "go.mod"
    "main.go"
    "README.md"
    "LICENSE"
    "CHANGELOG.md"
    "install.sh"
    "install.ps1"
)

required_dirs=(
    "config"
    "config/languages"
    "docs"
    "scripts"
    "tests"
)

for file in "${required_files[@]}"; do
    if [[ -f "$file" ]]; then
        echo -e "${GREEN}✅ $file${NC}"
    else
        echo -e "${RED}❌ $file 缺失${NC}"
        exit 1
    fi
done

for dir in "${required_dirs[@]}"; do
    if [[ -d "$dir" ]]; then
        echo -e "${GREEN}✅ $dir/${NC}"
    else
        echo -e "${RED}❌ $dir/ 缺失${NC}"
        exit 1
    fi
done

# 3. 检查语言文件
echo -e "\n${YELLOW}🌍 检查语言文件...${NC}"
lang_count=$(find config/languages -name "*.json" | wc -l)
if [[ $lang_count -gt 0 ]]; then
    echo -e "${GREEN}✅ 找到 $lang_count 个语言文件${NC}"
    find config/languages -name "*.json" -exec basename {} \; | while read file; do
        echo -e "${GRAY}  - $file${NC}"
    done
else
    echo -e "${YELLOW}⚠️ 未找到语言文件${NC}"
fi

# 4. 构建测试
if [[ "$SKIP_BUILD" != "true" ]]; then
    echo -e "\n${YELLOW}🔨 构建测试...${NC}"
    if go build -o delguard; then
        echo -e "${GREEN}✅ 构建成功${NC}"
        
        # 测试基本功能
        if ./delguard --help &> /dev/null; then
            echo -e "${GREEN}✅ 帮助功能正常${NC}"
        else
            echo -e "${YELLOW}⚠️ 帮助功能异常${NC}"
        fi
        
        rm -f delguard
    else
        echo -e "${RED}❌ 构建失败${NC}"
        exit 1
    fi
fi

# 5. 运行测试
if [[ "$SKIP_TESTS" != "true" ]]; then
    echo -e "\n${YELLOW}🧪 运行测试...${NC}"
    if go test -v ./...; then
        echo -e "${GREEN}✅ 所有测试通过${NC}"
    else
        echo -e "${YELLOW}⚠️ 部分测试失败，请检查${NC}"
    fi
fi

# 6. 检查安装脚本
echo -e "\n${YELLOW}📥 检查安装脚本...${NC}"
for script in "install.sh" "install.ps1"; do
    if [[ -f "$script" ]]; then
        if grep -q "github.com/01luyicheng/DelGuard" "$script"; then
            echo -e "${GREEN}✅ $script GitHub URL 正确${NC}"
        else
            echo -e "${YELLOW}⚠️ $script GitHub URL 需要验证${NC}"
        fi
    fi
done

# 7. 检查版本信息
echo -e "\n${YELLOW}📋 检查版本信息...${NC}"
if [[ -f "CHANGELOG.md" ]]; then
    if grep -q "\[未发布\]" "CHANGELOG.md"; then
        echo -e "${YELLOW}⚠️ CHANGELOG.md 包含未发布版本，建议更新${NC}"
    else
        echo -e "${GREEN}✅ CHANGELOG.md 版本信息正常${NC}"
    fi
fi

# 8. 安全检查
echo -e "\n${YELLOW}🔒 安全检查...${NC}"
if [[ -f "final_security_check.go" ]]; then
    if go run final_security_check.go; then
        echo -e "${GREEN}✅ 安全检查完成${NC}"
    else
        echo -e "${YELLOW}⚠️ 安全检查脚本执行异常${NC}"
    fi
fi

echo -e "\n${CYAN}🎉 发布前检查完成！${NC}"
echo -e "${CYAN}===================${NC}"

echo -e "\n${NC}📝 下一步操作建议:${NC}"
echo -e "${GRAY}1. 创建 GitHub 仓库 (如果尚未创建)${NC}"
echo -e "${GRAY}2. 推送代码到 GitHub${NC}"
echo -e "${GRAY}3. 验证安装脚本可访问性${NC}"
echo -e "${GRAY}4. 创建 GitHub Release${NC}"
echo -e "${GRAY}5. 测试一键安装命令${NC}"