@echo off
setlocal enabledelayedexpansion

echo === DelGuard 构建脚本 ===
echo 正在构建 DelGuard 安全删除工具...

REM 设置版本信息
set VERSION=1.0.0
set BUILD_TIME=%date% %time%

REM 创建构建目录
if not exist "build" mkdir build
if not exist "dist" mkdir dist

REM 设置Go环境变量
set GO111MODULE=on
set CGO_ENABLED=0

REM 构建Windows版本
echo 正在构建 Windows 版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-X main.version=%VERSION% -X main.buildTime=%BUILD_TIME%" -o build\delguard-windows-amd64.exe .

set GOARCH=386
go build -ldflags="-X main.version=%VERSION% -X main.buildTime=%BUILD_TIME%" -o build\delguard-windows-386.exe .

REM 构建Linux版本
echo 正在构建 Linux 版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-X main.version=%VERSION% -X main.buildTime=%BUILD_TIME%" -o build\delguard-linux-amd64 .

set GOARCH=386
go build -ldflags="-X main.version=%VERSION% -X main.buildTime=%BUILD_TIME%" -o build\delguard-linux-386 .

set GOARCH=arm64
go build -ldflags="-X main.version=%VERSION% -X main.buildTime=%BUILD_TIME%" -o build\delguard-linux-arm64 .

REM 构建macOS版本
echo 正在构建 macOS 版本...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-X main.version=%VERSION% -X main.buildTime=%BUILD_TIME%" -o build\delguard-darwin-amd64 .

set GOARCH=arm64
go build -ldflags="-X main.version=%VERSION% -X main.buildTime=%BUILD_TIME%" -o build\delguard-darwin-arm64 .

REM 创建压缩包
echo 正在创建发布包...
cd build

for %%f in (delguard-*) do (
    if exist "%%f" (
        echo 压缩 %%f...
        tar -czf "..\dist\%%f.tar.gz" "%%f"
    )
)

cd ..

REM 生成校验文件
echo 正在生成校验文件...
cd dist
for %%f in (*.tar.gz) do (
    certutil -hashfile "%%f" SHA256 > "%%f.sha256"
)
cd ..

echo.
echo === 构建完成 ===
echo 版本: %VERSION%
echo 构建时间: %BUILD_TIME%
echo.
echo 构建文件列表:
dir build\delguard-*
echo.
echo 发布文件列表:
dir dist\*.tar.gz
echo.
echo 校验文件:
dir dist\*.sha256
echo.
echo 构建成功！请检查 dist 目录获取发布文件。
pause