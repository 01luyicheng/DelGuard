#!/bin/bash
set -e

# DelGuard 发布脚本
# 创建 GitHub Release 并上传所有平台二进制文件

# 配置
REPO="yourusername/DelGuard"
BINARY_NAME="delguard"
VERSION=${1:-"v1.0.0"}
BUILD_DIR="build"
GH_CLI="gh"

echo "=== DelGuard 发布脚本 ==="
echo "版本: $VERSION"
echo

# 检查依赖
check_dependencies() {
    echo "检查依赖..."
    
    if ! command -v go &> /dev/null; then
        echo "错误: Go 未安装"
        exit 1
    fi
    
    if ! command -v $GH_CLI &> /dev/null; then
        echo "错误: GitHub CLI (gh) 未安装"
        echo "安装方法: https://cli.github.com/"
        exit 1
    fi
    
    echo "✓ 依赖检查通过"
}

# 清理旧文件
clean() {
    echo "清理旧文件..."
    rm -rf $BUILD_DIR
    mkdir -p $BUILD_DIR
    echo "✓ 清理完成"
}

# 构建所有平台
build_all() {
    echo "构建所有平台..."
    
    # 构建参数
    LDFLAGS="-s -w -X main.version=$VERSION"
    
    # Windows
    echo "  构建 Windows amd64..."
    GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/${BINARY_NAME}-windows-amd64.exe" .
    
    echo "  构建 Windows arm64..."
    GOOS=windows GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/${BINARY_NAME}-windows-arm64.exe" .
    
    # Linux
    echo "  构建 Linux amd64..."
    GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/${BINARY_NAME}-linux-amd64" .
    
    echo "  构建 Linux arm64..."
    GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/${BINARY_NAME}-linux-arm64" .
    
    echo "  构建 Linux arm..."
    GOOS=linux GOARCH=arm go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/${BINARY_NAME}-linux-arm" .
    
    # macOS
    echo "  构建 macOS amd64..."
    GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/${BINARY_NAME}-darwin-amd64" .
    
    echo "  构建 macOS arm64..."
    GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/${BINARY_NAME}-darwin-arm64" .
    
    echo "✓ 构建完成"
}

# 创建压缩包
create_archives() {
    echo "创建压缩包..."
    
    cd $BUILD_DIR
    
    for file in ${BINARY_NAME}-*; do
        if [[ -f "$file" ]]; then
            echo "  压缩 $file..."
            if [[ "$file" == *.exe ]]; then
                zip "${file%.exe}.zip" "$file"
            else
                tar -czf "${file}.tar.gz" "$file"
            fi
        fi
    done
    
    cd ..
    echo "✓ 压缩完成"
}

# 生成校验和
generate_checksums() {
    echo "生成校验和..."
    
    cd $BUILD_DIR
    
    # 生成 SHA256 校验和
    if command -v sha256sum &> /dev/null; then
        sha256sum ${BINARY_NAME}-* > checksums.txt
    elif command -v shasum &> /dev/null; then
        shasum -a 256 ${BINARY_NAME}-* > checksums.txt
    else
        echo "警告: 无法生成 SHA256 校验和"
    fi
    
    cd ..
    echo "✓ 校验和生成完成"
}

# 创建 GitHub Release
create_release() {
    echo "创建 GitHub Release..."
    
    # 检查是否已登录
    if ! $GH_CLI auth status &> /dev/null; then
        echo "请先登录 GitHub CLI:"
        echo "$GH_CLI auth login"
        exit 1
    fi
    
    # 创建 Release
    $GH_CLI release create "$VERSION" \
        --title "DelGuard $VERSION" \
        --notes-file "CHANGELOG.md" \
        --latest
    
    echo "✓ Release 创建完成"
}

# 上传文件
upload_files() {
    echo "上传文件到 Release..."
    
    cd $BUILD_DIR
    
    # 上传所有文件
    $GH_CLI release upload "$VERSION" \
        ${BINARY_NAME}-*.zip \
        ${BINARY_NAME}-*.tar.gz \
        checksums.txt \
        ../install.sh \
        ../install.ps1 \
        ../README.md \
        ../LICENSE
    
    cd ..
    echo "✓ 文件上传完成"
}

# 显示结果
show_results() {
    echo
    echo "=== 发布完成 ==="
    echo "版本: $VERSION"
    echo "GitHub Release: https://github.com/$REPO/releases/tag/$VERSION"
    echo
    echo "已发布的文件:"
    ls -la $BUILD_DIR/${BINARY_NAME}-*
    echo
    echo "安装命令:"
    echo "  Linux/macOS: curl -fsSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash"
    echo "  Windows: iwr -useb https://raw.githubusercontent.com/$REPO/main/install.ps1 | iex"
}

# 主流程
main() {
    check_dependencies
    clean
    build_all
    create_archives
    generate_checksums
    
    echo
    read -p "是否创建 GitHub Release? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        create_release
        upload_files
    else
        echo "跳过 GitHub Release 创建"
    fi
    
    show_results
}

# 执行主流程
main