@echo off
echo 正在安装DelGuard系统级cp命令...

:: 获取管理员权限（如果需要）
:: 检查是否以管理员身份运行
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo 需要管理员权限来安装系统级命令...
    echo 请以管理员身份运行此脚本
    pause
    exit /b 1
)

:: 设置安装路径
set "INSTALL_DIR=C:\Program Files\DelGuard"
set "EXE_NAME=delguard.exe"

:: 创建安装目录
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

:: 复制可执行文件
copy /y "%~dp0%EXE_NAME%" "%INSTALL_DIR%\%EXE_NAME%"

:: 添加到系统PATH
setx PATH "%PATH%;%INSTALL_DIR%" /M

:: 创建系统级批处理命令
echo @"%INSTALL_DIR%\%EXE_NAME%" --cp %%* > "%INSTALL_DIR%\cp.bat"
echo @"%INSTALL_DIR%\%EXE_NAME%" %%* > "%INSTALL_DIR%\del.bat"
echo @"%INSTALL_DIR%\%EXE_NAME%" %%* > "%INSTALL_DIR%\rm.bat"

echo.
echo 安装完成！
echo.
echo 现在你可以在任何位置使用以下命令：
echo   cp source.txt dest.txt    :: 安全复制文件
echo   del filename              :: 安全删除文件
echo   rm filename               :: 安全删除文件
echo.
echo 请重新打开命令提示符或PowerShell窗口以使用新命令。
pause