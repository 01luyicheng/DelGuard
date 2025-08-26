@echo off
setlocal enabledelayedexpansion

:: DelGuard Enhanced Windows Batch Installer (Fixed Version)
:: Supports: Windows 7+ with Command Prompt
:: Author: DelGuard Team
:: Version: 2.1.1 (Fixed)

:: Configuration
set "INSTALL_PATH=%USERPROFILE%\bin"
set "FORCE=false"
set "QUIET=false"
set "UNINSTALL=false"
set "SYSTEM_WIDE=false"
set "SKIP_ALIASES=false"
set "CHECK_ONLY=false"

:: Parse command line arguments
:parse_args
if "%~1"=="" goto :args_done
if /i "%~1"=="--install-path" (
    set "INSTALL_PATH=%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="--force" (
    set "FORCE=true"
    shift
    goto :parse_args
)
if /i "%~1"=="--quiet" (
    set "QUIET=true"
    shift
    goto :parse_args
)
if /i "%~1"=="--uninstall" (
    set "UNINSTALL=true"
    shift
    goto :parse_args
)
if /i "%~1"=="--system-wide" (
    set "SYSTEM_WIDE=true"
    shift
    goto :parse_args
)
if /i "%~1"=="--skip-aliases" (
    set "SKIP_ALIASES=true"
    shift
    goto :parse_args
)
if /i "%~1"=="--check-only" (
    set "CHECK_ONLY=true"
    shift
    goto :parse_args
)
if /i "%~1"=="--help" (
    goto :show_help
)
echo Unknown option: %~1
goto :show_help

:args_done

:: Color output functions (Windows 10+ only, fallback for older versions)
if "%QUIET%"=="true" goto :skip_colors
ver | find "10." >nul 2>&1
if %errorlevel%==0 (
    set "COLOR_RED=[31m"
    set "COLOR_GREEN=[32m"
    set "COLOR_YELLOW=[33m"
    set "COLOR_CYAN=[36m"
    set "COLOR_MAGENTA=[35m"
    set "COLOR_RESET=[0m"
) else (
    set "COLOR_RED="
    set "COLOR_GREEN="
    set "COLOR_YELLOW="
    set "COLOR_CYAN="
    set "COLOR_MAGENTA="
    set "COLOR_RESET="
)
:skip_colors

:: Main execution
if "%CHECK_ONLY%"=="true" (
    goto :check_installation
)

if "%UNINSTALL%"=="true" (
    goto :uninstall_delguard
) else (
    goto :install_delguard
)

:show_help
echo DelGuard Enhanced Windows Batch Installer (Fixed)
echo.
echo Usage: %~nx0 [OPTIONS]
echo.
echo Options:
echo     --install-path PATH    Install to specific path (default: %%USERPROFILE%%\bin)
echo     --force               Force overwrite existing installation
echo     --quiet               Suppress output messages
echo     --uninstall           Remove DelGuard installation
echo     --system-wide         Install system-wide (requires admin rights)
echo     --skip-aliases        Skip alias configuration
echo     --check-only          Only check installation status
echo     --help                Show this help message
echo.
echo Examples:
echo     %~nx0                              # Install to default location
echo     %~nx0 --install-path C:\tools      # Install to custom location
echo     %~nx0 --system-wide                # Install system-wide
echo     %~nx0 --uninstall                  # Remove DelGuard
echo     %~nx0 --check-only                 # Check installation status
goto :eof

:print_info
if "%QUIET%"=="false" echo %COLOR_CYAN%[INFO] %~1%COLOR_RESET%
goto :eof

:print_success
if "%QUIET%"=="false" echo %COLOR_GREEN%[SUCCESS] %~1%COLOR_RESET%
goto :eof

:print_warning
if "%QUIET%"=="false" echo %COLOR_YELLOW%[WARNING] %~1%COLOR_RESET%
goto :eof

:print_error
if "%QUIET%"=="false" echo %COLOR_RED%[ERROR] %~1%COLOR_RESET%
goto :eof

:print_header
if "%QUIET%"=="false" echo %COLOR_MAGENTA%%~1%COLOR_RESET%
goto :eof

:check_installation
call :print_header "=== DelGuard Installation Status Check ==="
call :print_info "Checking DelGuard installation..."

:: Check common installation paths
set "FOUND_INSTALLATIONS="
set "CHECK_PATHS=%USERPROFILE%\bin\delguard.exe;%ProgramFiles%\DelGuard\delguard.exe;%LOCALAPPDATA%\DelGuard\delguard.exe"

for %%P in (%CHECK_PATHS%) do (
    if exist "%%P" (
        set "FOUND_INSTALLATIONS=!FOUND_INSTALLATIONS! %%P"
        call :print_success "Found installation: %%P"
        "%%P" --version >nul 2>&1 && (
            call :print_info "Version check passed"
        ) || (
            call :print_warning "Version check failed"
        )
    )
)

if "%FOUND_INSTALLATIONS%"==" " (
    call :print_warning "DelGuard is not installed"
    exit /b 1
)

:: Check PATH
where delguard >nul 2>&1 && (
    call :print_success "DelGuard found in PATH"
) || (
    call :print_warning "DelGuard not found in PATH"
)

call :print_info "Installation status check complete"
goto :eof

:build_delguard
call :print_info "Building DelGuard executable..."

if not exist "go.mod" (
    call :print_error "go.mod not found. Please run this script from DelGuard project root."
    exit /b 1
)

:: Check if executable already exists
if exist "delguard.exe" (
    call :print_success "DelGuard executable already exists"
    goto :eof
)

:: Check if Go is installed
where go >nul 2>&1 || (
    call :print_error "Go is not installed. Please install Go first."
    call :print_info "Visit: https://golang.org/dl/"
    exit /b 1
)

:: Build executable
set "CGO_ENABLED=0"
set "GOOS=windows"
set "GOARCH=amd64"

call :print_info "Building DelGuard executable..."

:: Try to build with main.go
if exist "main.go" (
    call :print_info "Building with main.go..."
    go build -ldflags "-s -w" -o "delguard.exe" "main.go"
    set "BUILD_RESULT=!errorlevel!"
) else (
    call :print_info "Building entire project..."
    go build -ldflags "-s -w" -o "delguard.exe"
    set "BUILD_RESULT=!errorlevel!"
)

if !BUILD_RESULT! neq 0 (
    call :print_error "Build failed. Please check Go installation and project dependencies."
    exit /b 1
)

if not exist "delguard.exe" (
    call :print_error "Build completed but executable not found."
    exit /b 1
)

call :print_success "DelGuard executable built successfully"
goto :eof

:install_delguard
call :print_header "=== DelGuard Enhanced Windows Installer (Fixed) ==="
call :print_info "Starting DelGuard installation..."

:: Determine install path
if "%SYSTEM_WIDE%"=="true" (
    if "%INSTALL_PATH%"=="%USERPROFILE%\bin" (
        set "INSTALL_PATH=%ProgramFiles%\DelGuard"
    )
    
    :: Check admin rights for system-wide installation
    net session >nul 2>&1 || (
        call :print_error "System-wide installation requires administrator privileges."
        call :print_info "Please run Command Prompt as Administrator or use user installation."
        exit /b 1
    )
)

call :print_info "Install path: %INSTALL_PATH%"

:: Build DelGuard
call :build_delguard
if !errorlevel! neq 0 (
    call :print_error "Build failed. Installation aborted."
    exit /b 1
)

:: Create install directory
if not exist "%INSTALL_PATH%" (
    mkdir "%INSTALL_PATH%" 2>nul
    if !errorlevel!==0 (
        call :print_success "Created directory: %INSTALL_PATH%"
    ) else (
        call :print_error "Failed to create directory: %INSTALL_PATH%"
        exit /b 1
    )
)

set "TARGET_EXE=%INSTALL_PATH%\delguard.exe"

:: Check if already installed
if exist "%TARGET_EXE%" (
    if "%FORCE%"=="false" (
        call :print_warning "DelGuard is already installed at: %TARGET_EXE%"
        call :print_warning "Use --force to overwrite"
        exit /b 1
    )
)

:: Copy executable
copy /y "delguard.exe" "%TARGET_EXE%" >nul 2>&1
if !errorlevel!==0 (
    call :print_success "Copied executable to: %TARGET_EXE%"
) else (
    call :print_error "Failed to copy executable"
    exit /b 1
)

:: Add to PATH
echo %PATH% | find /i "%INSTALL_PATH%" >nul
if !errorlevel!==1 (
    if "%SYSTEM_WIDE%"=="true" (
        :: Add to system PATH (requires admin)
        for /f "tokens=2*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v PATH 2^>nul') do set "SYSTEM_PATH=%%B"
        if defined SYSTEM_PATH (
            reg add "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v PATH /t REG_EXPAND_SZ /d "!SYSTEM_PATH!;%INSTALL_PATH%" /f >nul 2>&1
            if !errorlevel!==0 (
                call :print_success "Added to system PATH: %INSTALL_PATH%"
            ) else (
                call :print_warning "Failed to add to system PATH"
            )
        )
    ) else (
        :: Add to user PATH
        for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v PATH 2^>nul') do set "USER_PATH=%%B"
        if defined USER_PATH (
            reg add "HKCU\Environment" /v PATH /t REG_EXPAND_SZ /d "!USER_PATH!;%INSTALL_PATH%" /f >nul 2>&1
        ) else (
            reg add "HKCU\Environment" /v PATH /t REG_EXPAND_SZ /d "%INSTALL_PATH%" /f >nul 2>&1
        )
        if !errorlevel!==0 (
            call :print_success "Added to user PATH: %INSTALL_PATH%"
        ) else (
            call :print_warning "Failed to add to PATH. Please add manually: %INSTALL_PATH%"
        )
    )
)

:: Configure aliases (basic batch file approach)
if "%SKIP_ALIASES%"=="false" (
    call :configure_aliases
)

call :print_success "=== Installation Complete ==="
call :print_info "DelGuard has been installed successfully!"
call :print_info ""
call :print_info "NEXT STEPS:"
call :print_info "1. Restart Command Prompt to use new PATH"
call :print_info "2. Test with: delguard --version"
call :print_info "3. Check status with: %~nx0 --check-only"
call :print_info "4. For PowerShell users, consider using install.ps1 for better alias support"
call :print_info ""
call :print_success "Happy safe deleting!"

:: Test installation
call :print_info "Testing installation..."
"%TARGET_EXE%" --version >nul 2>&1
if !errorlevel!==0 (
    call :print_success "DelGuard is working correctly"
) else (
    call :print_warning "DelGuard may not be working properly"
)

goto :eof

:configure_aliases
call :print_info "Configuring basic aliases..."
call :print_info "Note: For full alias support, use PowerShell with install.ps1"

:: Create simple batch files for aliases (basic approach)
set "ALIAS_DIR=%INSTALL_PATH%"

:: Create del.bat
echo @echo off > "%ALIAS_DIR%\del.bat"
echo "%INSTALL_PATH%\delguard.exe" %%* >> "%ALIAS_DIR%\del.bat"

:: Create rm.bat  
echo @echo off > "%ALIAS_DIR%\rm.bat"
echo "%INSTALL_PATH%\delguard.exe" %%* >> "%ALIAS_DIR%\rm.bat"

:: Create cp.bat
echo @echo off > "%ALIAS_DIR%\cp.bat"
echo "%INSTALL_PATH%\delguard.exe" --copy %%* >> "%ALIAS_DIR%\cp.bat"

call :print_success "Created basic alias batch files"
call :print_info "Commands available: del.bat, rm.bat, cp.bat, delguard.exe"

goto :eof

:uninstall_delguard
call :print_info "Uninstalling DelGuard..."

:: Remove executable and aliases
set "PATHS_TO_CHECK=%USERPROFILE%\bin;%ProgramFiles%\DelGuard;%LOCALAPPDATA%\DelGuard"
if not "%INSTALL_PATH%"=="" set "PATHS_TO_CHECK=%PATHS_TO_CHECK%;%INSTALL_PATH%"

for %%P in (%PATHS_TO_CHECK%) do (
    if exist "%%P\delguard.exe" (
        del /f /q "%%P\delguard.exe" 2>nul && call :print_success "Removed executable: %%P\delguard.exe"
        del /f /q "%%P\del.bat" 2>nul
        del /f /q "%%P\rm.bat" 2>nul  
        del /f /q "%%P\cp.bat" 2>nul
        
        :: Remove empty directory
        rmdir "%%P" 2>nul && call :print_success "Removed directory: %%P"
    )
)

call :print_success "DelGuard uninstalled successfully!"
call :print_info "Please restart Command Prompt to complete the removal."
goto :eof