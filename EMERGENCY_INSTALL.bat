@echo off
setlocal enabledelayedexpansion

echo.
echo ================================================
echo    DelGuard - 紧急安装方案
echo ================================================
echo.

echo 正在执行紧急安装...

:: 1. 确保文件存在
echo 检查文件...
if not exist "C:\Users\21601\bin\delguard.exe" (
    echo 复制主程序...
    copy /y "%~dp0delguard.exe" "C:\Users\21601\bin\delguard.exe" >nul
)

:: 2. 创建正确的批处理文件
echo 创建命令文件...
echo @"C:\Users\21601\bin\delguard.exe" --cp %%*> "C:\Users\21601\bin\cp.bat"
echo @"C:\Users\21601\bin\delguard.exe" %%*> "C:\Users\21601\bin\del.bat"
echo @"C:\Users\21601\bin\delguard.exe" %%*> "C:\Users\21601\bin\rm.bat"

:: 3. 获取当前PATH并添加bin目录
echo 配置环境变量...
set "CURRENT_PATH="
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v PATH 2^>nul') do set "CURRENT_PATH=%%b"

:: 检查是否已包含
set "NEW_PATH=C:\Users\21601\bin"
if not "%CURRENT_PATH%"=="" (
    echo %CURRENT_PATH% | findstr /i "C:\Users\21601\bin" >nul
    if errorlevel 1 (
        set "NEW_PATH=C:\Users\21601\bin;%CURRENT_PATH%"
    )
)

:: 4. 设置PATH
reg add "HKCU\Environment" /v PATH /t REG_EXPAND_SZ /d "%NEW_PATH%" /f >nul

echo.
echo ✅ 安装完成！
echo.
echo 立即使用方法：
echo 1. 关闭此窗口
echo 2. 按 Win+R 输入 cmd 回车
echo 3. 输入: cp --help
echo.
echo 如果仍有问题，重启电脑即可解决
echo.
pause