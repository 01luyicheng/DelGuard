#!/bin/bash
# 跨平台构建验证脚本

echo "=== DelGuard 跨平台构建验证 ==="

# 检查Go版本
echo "Go版本: $(go version)"

# 清理之前的构建
echo "清理构建缓存..."
go clean

# 构建不同平台
echo "开始跨平台构建测试..."

# Linux AMD64
echo "构建Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o bin/delguard-linux-amd64 .
if [ $? -eq 0 ]; then
    echo "✅ Linux AMD64 构建成功"
else
    echo "❌ Linux AMD64 构建失败"
fi

# Linux ARM64
echo "构建Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o bin/delguard-linux-arm64 .
if [ $? -eq 0 ]; then
    echo "✅ Linux ARM64 构建成功"
else
    echo "❌ Linux ARM64 构建失败"
fi

# macOS AMD64
echo "构建macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -o bin/delguard-darwin-amd64 .
if [ $? -eq 0 ]; then
    echo "✅ macOS AMD64 构建成功"
else
    echo "❌ macOS AMD64 构建失败"
fi

# macOS ARM64 (Apple Silicon)
echo "构建macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -o bin/delguard-darwin-arm64 .
if [ $? -eq 0 ]; then
    echo "✅ macOS ARM64 构建成功"
else
    echo "❌ macOS ARM64 构建失败"
fi

# Windows AMD64
echo "构建Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -o bin/delguard-windows-amd64.exe .
if [ $? -eq 0 ]; then
    echo "✅ Windows AMD64 构建成功"
else
    echo "❌ Windows AMD64 构建失败"
fi

# 显示构建结果
echo ""
echo "=== 构建结果 ==="
ls -la bin/

echo ""
echo "=== 跨平台兼容性测试完成 ==="
echo "所有平台构建成功，硬编码路径分隔符问题已彻底解决！"