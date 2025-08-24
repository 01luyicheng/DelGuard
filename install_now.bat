@echo off
setlocal enabledelayedexpansion

echo.
echo ================================================
echo    DelGuard System Installation
echo    Use cp command anywhere
echo ================================================
echo.

:: Check admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Requesting admin rights...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

:: Set variables
set "INSTALL_DIR=%USERPROFILE%\bin"
set "EXE_NAME=delguard.exe"
set "CURRENT_DIR=%~dp0"

:: Create user bin directory
echo Creating user directory...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

:: Copy executable
echo Copying executable...
copy /y "%CURRENT_DIR%%EXE_NAME%" "%INSTALL_DIR%\%EXE_NAME%" >nul

:: Create batch commands
echo Creating commands...
echo @"%INSTALL_DIR%\%EXE_NAME%" --cp %%* > "%INSTALL_DIR%\cp.bat"
echo @"%INSTALL_DIR%\%EXE_NAME%" %%* > "%INSTALL_DIR%\del.bat"
echo @"%INSTALL_DIR%\%EXE_NAME%" %%* > "%INSTALL_DIR%\rm.bat"

:: Add to user PATH
echo Adding to user PATH...
setx PATH "%PATH%;%INSTALL_DIR%" >nul 2>&1

echo.
echo ================================================
echo    Installation Complete!
echo ================================================
echo.
echo You can now use these commands anywhere:
echo   cp source.txt dest.txt
echo   del filename
echo   rm filename
echo.
echo Please restart Command Prompt or PowerShell.
echo.
pause