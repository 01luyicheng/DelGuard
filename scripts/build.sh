#!/bin/bash

# DelGuard 构建脚本
# 支持跨平台构建和发布

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="delguard"
BUILD_DIR="build"
DIST_DIR="dist"

# 获取版本信息
get_version() {
    if git describe --tags --always --dirty >/dev/null 2>&1; then
        echo $(git describe --tags --always --dirty)
    else
        echo "v0.1.0"
    fi
}

# 获取Git提交信息
get_git_commit() {
    if git rev-parse --short HEAD >/dev/null 2>&1; then
        echo $(git rev-parse --short HEAD)
    else
        echo "unknown"
    fi
}

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

# 检查依赖
check_dependencies() {
    log_info "检查构建依赖..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    log_success "Go 版本: $(go version)"
}

# 清理构建目录
clean_build() {
    log_info "清理构建目录..."
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    go clean
    log_success "构建目录已清理"
}

# 安装依赖
install_deps() {
    log_info "安装Go依赖..."
    go mod download
    go mod tidy
    log_success "依赖安装完成"
}

# 运行测试
run_tests() {
    log_info "运行测试..."
    if go test -v ./...; then
        log_success "所有测试通过"
    else
        log_error "测试失败"
        exit 1
    fi
}

# 构建单个平台
build_platform() {
    local goos=$1
    local goarch=$2
    local version=$3
    local git_commit=$4
    local build_time=$5
    
    local output_name="${PROJECT_NAME}-${version}-${goos}-${goarch}"
    if [ "$goos" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    log_info "构建 ${goos}/${goarch}..."
    
    GOOS=$goos GOARCH=$goarch go build \
        -ldflags "-X main.Version=$version -X main.BuildTime=$build_time -X main.GitCommit=$git_commit -s -w" \
        -o "$DIST_DIR/$output_name" \
        ./cmd/$PROJECT_NAME
    
    if [ $? -eq 0 ]; then
        log_success "构建完成: $output_name"
    else
        log_error "构建失败: ${goos}/${goarch}"
        return 1
    fi
}

# 构建所有平台
build_all() {
    local version=$(get_version)
    local git_commit=$(get_git_commit)
    local build_time=$(date -u '+%Y-%m-%d_%H:%M:%S')
    
    log_info "开始构建所有平台..."
    log_info "版本: $version"
    log_info "Git提交: $git_commit"
    log_info "构建时间: $build_time"
    
    mkdir -p "$DIST_DIR"
    
    # 支持的平台列表
    platforms=(
        "windows/amd64"
        "windows/386"
        "linux/amd64"
        "linux/386"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
    )
    
    local success_count=0
    local total_count=${#platforms[@]}
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r goos goarch <<< "$platform"
        if build_platform "$goos" "$goarch" "$version" "$git_commit" "$build_time"; then
            ((success_count++))
        fi
    done
    
    log_success "构建完成: $success_count/$total_count 个平台"
}

# 打包发布文件
package_releases() {
    log_info "打包发布文件..."
    
    if [ ! -d "$DIST_DIR" ]; then
        log_error "构建目录不存在，请先运行构建"
        exit 1
    fi
    
    cd "$DIST_DIR"
    
    for file in ${PROJECT_NAME}-*; do
        if [ -f "$file" ]; then
            if [[ "$file" == *"windows"* ]]; then
                # Windows 平台使用 zip
                zip "${file}.zip" "$file" ../README.md ../LICENSE 2>/dev/null || true
                log_success "打包完成: ${file}.zip"
            else
                # Unix 平台使用 tar.gz
                tar -czf "${file}.tar.gz" "$file" ../README.md ../LICENSE 2>/dev/null || true
                log_success "打包完成: ${file}.tar.gz"
            fi
        fi
    done
    
    cd ..
}

# 生成校验和
generate_checksums() {
    log_info "生成校验和文件..."
    
    cd "$DIST_DIR"
    
    # 生成 SHA256 校验和
    if command -v sha256sum &> /dev/null; then
        sha256sum *.zip *.tar.gz 2>/dev/null > checksums.sha256 || true
    elif command -v shasum &> /dev/null; then
        shasum -a 256 *.zip *.tar.gz 2>/dev/null > checksums.sha256 || true
    fi
    
    if [ -f "checksums.sha256" ]; then
        log_success "校验和文件已生成: checksums.sha256"
    fi
    
    cd ..
}

# 显示构建信息
show_build_info() {
    log_info "构建信息:"
    echo "  项目名称: $PROJECT_NAME"
    echo "  版本: $(get_version)"
    echo "  Git提交: $(get_git_commit)"
    echo "  构建时间: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
    echo "  Go版本: $(go version)"
}

# 主函数
main() {
    case "${1:-all}" in
        "clean")
            clean_build
            ;;
        "deps")
            install_deps
            ;;
        "test")
            run_tests
            ;;
        "build")
            check_dependencies
            install_deps
            build_all
            ;;
        "package")
            package_releases
            generate_checksums
            ;;
        "all")
            check_dependencies
            clean_build
            install_deps
            run_tests
            build_all
            package_releases
            generate_checksums
            log_success "完整构建流程完成！"
            ;;
        "info")
            show_build_info
            ;;
        *)
            echo "用法: $0 {clean|deps|test|build|package|all|info}"
            echo ""
            echo "命令说明:"
            echo "  clean   - 清理构建目录"
            echo "  deps    - 安装依赖"
            echo "  test    - 运行测试"
            echo "  build   - 构建所有平台"
            echo "  package - 打包发布文件"
            echo "  all     - 完整构建流程"
            echo "  info    - 显示构建信息"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"