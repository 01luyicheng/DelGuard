#!/bin/bash
set -e

echo "=== DelGuard 跨平台构建脚本 ==="
echo

echo "1. 清理旧文件..."
rm -f delguard
rm -rf build
mkdir -p build

echo
echo "2. 构建当前平台版本..."
go build -ldflags "-s -w" -o delguard .

echo
echo "3. 交叉编译所有平台..."

# Windows
echo "构建 Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o build/delguard-windows-amd64.exe .

# Linux
echo "构建 Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o build/delguard-linux-amd64 .

echo "构建 Linux arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o build/delguard-linux-arm64 .

# macOS
echo "构建 macOS amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o build/delguard-darwin-amd64 .

echo "构建 macOS arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o build/delguard-darwin-arm64 .

echo
echo "4. 构建完成！生成的文件："
ls -la build/delguard-*

echo
echo "=== 构建成功 ==="
echo "可以运行 ./delguard --install 安装别名"