@echo off
setlocal enabledelayedexpansion

echo.
echo ================================================
echo    DelGuard - 立即修复cp命令
echo ================================================
echo.

:: 创建正确的批处理文件
echo 正在创建正确的命令文件...

:: 修复cp.bat
(
echo @"C:\Users\21601\bin\delguard.exe" --cp %%*
) > "C:\Users\21601\bin\cp.bat"

:: 修复del.bat  
(
echo @"C:\Users\21601\bin\delguard.exe" %%*
) > "C:\Users\21601\bin\del.bat"

:: 修复rm.bat
(
echo @"C:\Users\21601\bin\delguard.exe" %%*
) > "C:\Users\21601\bin\rm.bat"

:: 获取当前PATH
for /f "usebackq tokens=2,*" %%A in (`reg query "HKCU\Environment" /v PATH`) do set "OLD_PATH=%%B"

:: 检查是否已包含
set "NEW_PATH=!OLD_PATH!"
echo !NEW_PATH! | findstr /i "C:\Users\21601\bin" >nul
if errorlevel 1 (
    set "NEW_PATH=C:\Users\21601\bin;!OLD_PATH!"
    echo 正在添加到用户PATH...
    reg add "HKCU\Environment" /v PATH /t REG_EXPAND_SZ /d "!NEW_PATH!" /f >nul
    echo ✅ PATH已更新
) else (
    echo ✅ PATH已包含
)

echo.
echo ================================================
echo    安装完成！
echo ================================================
echo.
echo 请执行以下操作：
echo 1. 关闭此窗口
echo 2. 重新打开新的命令提示符
echo 3. 输入：cp --help
echo.
echo 如果仍有问题，请重启电脑
echo.
pause