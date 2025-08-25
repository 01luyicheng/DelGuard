@echo off
echo === DelGuard 跨平台构建脚本 ===
echo.

echo 1. 清理旧文件...
if exist delguard.exe del delguard.exe
if exist build rmdir /s /q build
mkdir build

echo.
echo 2. 构建 Windows 版本...
go build -ldflags "-s -w" -o build/delguard-windows-amd64.exe .
if %ERRORLEVEL% NEQ 0 (
    echo 构建失败！
    exit /b 1
)

echo.
echo 3. 交叉编译其他平台...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o build/delguard-linux-amd64 .

set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w" -o build/delguard-darwin-amd64 .

set GOOS=linux
set GOARCH=arm64
go build -ldflags "-s -w" -o build/delguard-linux-arm64 .

set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-s -w" -o build/delguard-darwin-arm64 .

echo.
echo 4. 复制当前平台可执行文件...
copy build\delguard-windows-amd64.exe delguard.exe

echo.
echo 5. 构建完成！生成的文件：
dir build\delguard-*

echo.
echo === 构建成功 ===
echo 可以运行 delguard.exe --install 安装别名