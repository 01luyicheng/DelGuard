@echo off
title DelGuard 快速安装 - 系统级cp命令
echo.
echo ================================================
echo    DelGuard 系统级安装工具
echo    在任何位置使用 cp 命令
echo ================================================
echo.

:: 检查管理员权限
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo 需要管理员权限...
    echo 正在以管理员身份重新启动...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

:: 设置变量
set "INSTALL_DIR=%USERPROFILE%\bin"
set "EXE_NAME=delguard.exe"
set "CURRENT_DIR=%~dp0"

:: 创建用户bin目录
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

:: 复制可执行文件
echo 正在复制文件...
copy /y "%CURRENT_DIR%%EXE_NAME%" "%INSTALL_DIR%\%EXE_NAME%" >nul

:: 创建批处理命令
echo 正在创建命令...
echo @"%INSTALL_DIR%\%EXE_NAME%" --cp %%* > "%INSTALL_DIR%\cp.bat"
echo @"%INSTALL_DIR%\%EXE_NAME%" %%* > "%INSTALL_DIR%\del.bat"
echo @"%INSTALL_DIR%\%EXE_NAME%" %%* > "%INSTALL_DIR%\rm.bat"

:: 添加到用户PATH（无需管理员权限）
echo 正在添加到用户PATH...
setx PATH "%PATH%;%INSTALL_DIR%" >nul

echo.
echo ================================================
echo    安装完成！
echo ================================================
echo.
echo 现在你可以在任何位置使用：
echo   cp source.txt dest.txt    :: 安全复制文件
echo   del filename              :: 安全删除文件
echo   rm filename               :: 安全删除文件
echo.
echo 请重新打开命令提示符或PowerShell窗口。
echo.
pause