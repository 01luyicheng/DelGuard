#!/bin/bash

# DelGuard 跨平台构建修复脚本
set -e

echo "🔧 修复 DelGuard 跨平台构建问题..."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 清理构建缓存
echo "🧹 清理构建缓存..."
go clean -cache
go clean -modcache || true

# 下载依赖
echo "📦 下载依赖..."
go mod download
go mod tidy

# 设置构建环境
export CGO_ENABLED=0
export GO111MODULE=on

# 定义目标平台
platforms=(
    "windows/amd64"
    "windows/386"
    "linux/amd64"
    "linux/386"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

# 创建构建目录
mkdir -p build/cross-platform

echo "🏗️  开始跨平台构建..."

success_count=0
total_count=${#platforms[@]}

for platform in "${platforms[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    
    echo "构建 $os/$arch..."
    
    output_name="delguard"
    if [ "$os" = "windows" ]; then
        output_name="delguard.exe"
    fi
    
    output_path="build/cross-platform/${os}-${arch}/${output_name}"
    mkdir -p "build/cross-platform/${os}-${arch}"
    
    # 设置环境变量并构建
    if GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags="-s -w" -o "$output_path" .; then
        echo "✅ $os/$arch 构建成功"
        
        # 验证二进制文件
        if [ -f "$output_path" ]; then
            file_size=$(stat -f%z "$output_path" 2>/dev/null || stat -c%s "$output_path" 2>/dev/null || echo "unknown")
            echo "   文件大小: $file_size bytes"
            ((success_count++))
        else
            echo "❌ $os/$arch 构建文件不存在"
        fi
    else
        echo "❌ $os/$arch 构建失败"
    fi
    echo ""
done

echo "📊 构建结果: $success_count/$total_count 成功"

if [ $success_count -eq $total_count ]; then
    echo "🎉 所有平台构建成功！"
    
    # 创建发布包
    echo "📦 创建发布包..."
    cd build/cross-platform
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        
        if [ -d "${os}-${arch}" ]; then
            echo "打包 $os-$arch..."
            
            # 复制安装脚本
            if [ "$os" = "windows" ]; then
                cp "../../scripts/safe-install.ps1" "${os}-${arch}/"
                cp "../../scripts/install.ps1" "${os}-${arch}/" 2>/dev/null || true
            else
                cp "../../scripts/safe-install.sh" "${os}-${arch}/"
                cp "../../scripts/install.sh" "${os}-${arch}/" 2>/dev/null || true
                chmod +x "${os}-${arch}/install.sh" 2>/dev/null || true
                chmod +x "${os}-${arch}/safe-install.sh"
            fi
            
            # 复制文档
            cp "../../README.md" "${os}-${arch}/" 2>/dev/null || true
            cp "../../LICENSE" "${os}-${arch}/" 2>/dev/null || true
            
            # 创建压缩包
            if command -v tar &> /dev/null; then
                tar -czf "delguard-${os}-${arch}.tar.gz" "${os}-${arch}/"
                echo "✅ 创建了 delguard-${os}-${arch}.tar.gz"
            fi
        fi
    done
    
    cd ../..
    echo "🎉 跨平台构建和打包完成！"
    echo "📁 构建文件位于: build/cross-platform/"
    
else
    echo "⚠️  部分平台构建失败，请检查错误信息"
    exit 1
fi