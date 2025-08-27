#!/bin/bash

# DelGuard 构建脚本 - Unix版本
# 构建所有支持平台的二进制文件

set -e

# 默认参数
VERSION="v1.0.0"
RELEASE=false
CLEAN=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -r|--release)
            RELEASE=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -h|--help)
            echo "用法: $0 [选项]"
            echo "选项:"
            echo "  -v, --version VERSION   设置版本号 (默认: v1.0.0)"
            echo "  -r, --release          创建发布包"
            echo "  -c, --clean            清理构建目录"
            echo "  -h, --help             显示帮助信息"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            exit 1
            ;;
    esac
done

# 常量定义
readonly APP_NAME="DelGuard"
readonly BUILD_DIR="build"
readonly DIST_DIR="dist"

# 颜色定义
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# 支持的平台
declare -a PLATFORMS=(
    "windows:amd64:.exe"
    "windows:arm64:.exe"
    "windows:386:.exe"
    "linux:amd64:"
    "linux:arm64:"
    "linux:arm:"
    "linux:386:"
    "darwin:amd64:"
    "darwin:arm64:"
)

echo -e "${CYAN}DelGuard 构建脚本${NC}"
echo -e "${CYAN}=================${NC}"
echo -e "${GREEN}版本: $VERSION${NC}"
echo ""

# 清理构建目录
if [[ "$CLEAN" == "true" ]]; then
    echo -e "${YELLOW}清理构建目录...${NC}"
    rm -rf "$BUILD_DIR" "$DIST_DIR"
fi

# 创建构建目录
mkdir -p "$BUILD_DIR" "$DIST_DIR"

# 检查依赖
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未找到 Go 编译器${NC}"
    exit 1
fi

# 运行测试
echo -e "${YELLOW}运行测试...${NC}"
if go test -v ./...; then
    echo -e "${GREEN}测试通过！${NC}"
    echo ""
else
    echo -e "${RED}测试失败，停止构建${NC}"
    exit 1
fi

# 构建所有平台
for platform in "${PLATFORMS[@]}"; do
    IFS=':' read -r os arch ext <<< "$platform"
    
    output_name="${APP_NAME}-${os}-${arch}${ext}"
    output_path="${BUILD_DIR}/${output_name}"
    
    echo -e "${BLUE}构建 ${os}/${arch}...${NC}"
    
    # 设置环境变量
    export GOOS="$os"
    export GOARCH="$arch"
    export CGO_ENABLED=0
    
    # 构建命令
    ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$(date '+%Y-%m-%d %H:%M:%S')"
    
    if go build -ldflags "$ldflags" -o "$output_path" .; then
        if [[ -f "$output_path" ]]; then
            size=$(stat -f%z "$output_path" 2>/dev/null || stat -c%s "$output_path" 2>/dev/null || echo "0")
            size_kb=$((size / 1024))
            echo -e "  ${GREEN}✓ $output_name (${size_kb} KB)${NC}"
            
            # 创建发布包
            if [[ "$RELEASE" == "true" ]]; then
                archive_name="${APP_NAME}-${os}-${arch}"
                archive_path="${DIST_DIR}/${archive_name}"
                
                if [[ "$os" == "windows" ]]; then
                    # Windows 使用 ZIP
                    zip_path="${archive_path}.zip"
                    if command -v zip &> /dev/null; then
                        zip -j "$zip_path" "$output_path" README.md LICENSE > /dev/null
                        echo -e "  ${CYAN}✓ 创建发布包: $zip_path${NC}"
                    else
                        echo -e "  ${YELLOW}⚠ 未找到 zip 命令，跳过 Windows 发布包${NC}"
                    fi
                else
                    # Unix 使用 tar.gz
                    tar_path="${archive_path}.tar.gz"
                    tar -czf "$tar_path" -C "$BUILD_DIR" "$output_name" -C .. README.md LICENSE
                    echo -e "  ${CYAN}✓ 创建发布包: $tar_path${NC}"
                fi
            fi
        else
            echo -e "  ${RED}✗ 构建失败${NC}"
        fi
    else
        echo -e "  ${RED}✗ 构建失败${NC}"
    fi
done

# 重置环境变量
unset GOOS GOARCH CGO_ENABLED

echo ""
echo -e "${GREEN}构建完成！${NC}"
echo -e "${NC}构建文件位于: $BUILD_DIR${NC}"

if [[ "$RELEASE" == "true" ]]; then
    echo -e "${NC}发布包位于: $DIST_DIR${NC}"
fi

# 显示构建统计
if [[ -d "$BUILD_DIR" ]]; then
    file_count=$(find "$BUILD_DIR" -type f | wc -l)
    total_size=$(find "$BUILD_DIR" -type f -exec stat -f%z {} + 2>/dev/null | awk '{sum+=$1} END {print sum}' || \
                 find "$BUILD_DIR" -type f -exec stat -c%s {} + 2>/dev/null | awk '{sum+=$1} END {print sum}' || echo "0")
    total_size_mb=$((total_size / 1024 / 1024))
    
    echo ""
    echo -e "${CYAN}构建统计:${NC}"
    echo -e "  文件数量: $file_count"
    echo -e "  总大小: ${total_size_mb} MB"
    
    if [[ "$RELEASE" == "true" && -d "$DIST_DIR" ]]; then
        dist_count=$(find "$DIST_DIR" -type f | wc -l)
        echo -e "  发布包数量: $dist_count"
    fi
fi