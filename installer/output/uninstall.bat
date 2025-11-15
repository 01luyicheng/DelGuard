@echo off
echo 正在卸载 DelGuard...
set "INSTALL_DIR=%ProgramFiles%\DelGuard"
if exist "%INSTALL_DIR%" rmdir /S /Q "%INSTALL_DIR%"
echo 从 PATH 环境变量中移除...
REM 用户需要手动从系统环境变量中移除
echo.
echo 请手动从系统 PATH 环境变量中移除 %INSTALL_DIR%
echo.
pause