#!/bin/bash

# DelGuard 跨平台构建脚本

echo "🔨 DelGuard 跨平台构建开始..."

# 设置版本信息
VERSION="1.0.0"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# 创建构建目录
mkdir -p build

echo "📦 构建 Windows 版本..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/delguard-windows-amd64.exe .
if [ $? -eq 0 ]; then
    echo "✅ Windows 版本构建成功"
else
    echo "❌ Windows 版本构建失败"
    exit 1
fi

echo "📦 构建 macOS 版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/delguard-darwin-amd64 .
if [ $? -eq 0 ]; then
    echo "✅ macOS Intel 版本构建成功"
else
    echo "❌ macOS Intel 版本构建失败"
    exit 1
fi

GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/delguard-darwin-arm64 .
if [ $? -eq 0 ]; then
    echo "✅ macOS Apple Silicon 版本构建成功"
else
    echo "❌ macOS Apple Silicon 版本构建失败"
    exit 1
fi

echo "📦 构建 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/delguard-linux-amd64 .
if [ $? -eq 0 ]; then
    echo "✅ Linux 版本构建成功"
else
    echo "❌ Linux 版本构建失败"
    exit 1
fi

GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o build/delguard-linux-arm64 .
if [ $? -eq 0 ]; then
    echo "✅ Linux ARM64 版本构建成功"
else
    echo "❌ Linux ARM64 版本构建失败"
    exit 1
fi

echo ""
echo "🎉 所有平台构建完成！"
echo ""
echo "构建文件列表:"
ls -la build/

echo ""
echo "📋 构建信息:"
echo "版本: ${VERSION}"
echo "构建时间: ${BUILD_TIME}"
echo "Git提交: ${GIT_COMMIT}"