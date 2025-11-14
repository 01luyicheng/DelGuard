/*
 * 文件用途：DelGuard安装程序构建脚本
 * 作者：TREA
 * 功能说明：自动化构建DelGuard MSI安装包的批处理脚本
 * 输入：项目源码和WiX工具集
 * 输出：DelGuard.msi安装包
 * 依赖：WiX Toolset v3, .NET 8.0 SDK
 * 交互方式：命令行执行，生成安装包文件
 */

@echo off
setlocal enabledelayedexpansion

echo ========================================
echo DelGuard 安装程序构建脚本
echo ========================================

:: 设置变量
set "PROJECT_ROOT=%~dp0.."
set "INSTALLER_DIR=%~dp0"
set "OUTPUT_DIR=%INSTALLER_DIR%output"
set "CONFIGURATION=Release"
set "WIX_PATH=C:\Program Files (x86)\WiX Toolset v3.11\bin"

:: 检查WiX工具集
if not exist "%WIX_PATH%\candle.exe" (
    echo 错误：未找到WiX工具集，请先安装WiX Toolset v3.11
    echo 下载地址：https://wixtoolset.org/releases/
    exit /b 1
)

:: 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

:: 设置WiX环境
set "PATH=%WIX_PATH%;%PATH%"

echo.
echo 1. 构建主程序...
cd /d "%PROJECT_ROOT%"

dotnet restore
dotnet build -c %CONFIGURATION%
dotnet publish -c %CONFIGURATION% -r win-x64 --self-contained false

if errorlevel 1 (
    echo 错误：主程序构建失败
    exit /b 1
)

echo.
echo 2. 构建自定义操作...
cd /d "%INSTALLER_DIR%"

dotnet build DelGuard.CustomActions.csproj -c %CONFIGURATION%

if errorlevel 1 (
    echo 错误：自定义操作构建失败
    exit /b 1
)

echo.
echo 3. 编译WiX源文件...
candle.exe -nologo -out "%OUTPUT_DIR%\" -arch x64 Product.wxs

if errorlevel 1 (
    echo 错误：WiX编译失败
    exit /b 1
)

echo.
echo 4. 链接生成MSI...
light.exe -nologo -out "%OUTPUT_DIR%\DelGuard.msi" "%OUTPUT_DIR%\Product.wixobj" -ext WixUIExtension -ext WixUtilExtension

if errorlevel 1 (
    echo 错误：WiX链接失败
    exit /b 1
)

echo.
echo 5. 验证安装包...
signtool.exe verify /pa "%OUTPUT_DIR%\DelGuard.msi" >nul 2>&1
if errorlevel 1 (
    echo 警告：安装包未签名（可选）
) else (
    echo 安装包签名验证通过
)

echo.
echo ========================================
echo 构建完成！
echo ========================================
echo 安装包位置：%OUTPUT_DIR%\DelGuard.msi
echo 文件大小：
for %%I in ("%OUTPUT_DIR%\DelGuard.msi") do echo   %%~zI 字节
echo.
echo 安装说明：
echo   1. 双击 DelGuard.msi 运行安装程序
echo   2. 按照安装向导完成安装
echo   3. 安装完成后可在命令行使用 delguard 命令
echo.
echo 卸载说明：
echo   - 控制面板 -^> 程序和功能 -^> 找到 DelGuard -^> 卸载
echo   - 或使用命令：msiexec /x {产品代码}
echo.
echo 构建脚本执行完毕！

endlocal
exit /b 0