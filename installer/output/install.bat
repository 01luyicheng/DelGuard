@echo off
echo ???? DelGuard ?????...
set "INSTALL_DIR=%ProgramFiles%\DelGuard"
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
copy /Y delguard.exe "%INSTALL_DIR%\"
copy /Y windows.json "%INSTALL_DIR%\config\"
if not exist "%INSTALL_DIR%\config" mkdir "%INSTALL_DIR%\config"
copy /Y windows.json "%INSTALL_DIR%\config\"
echo ?????????? PATH...
setx PATH "%PATH%;%INSTALL_DIR%" /M
echo.
echo ?????????????????? delguard ???
echo.
pause
