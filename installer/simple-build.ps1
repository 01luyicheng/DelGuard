<#
.SYNOPSIS
    DelGuard简单构建脚本（不需要WiX）

.DESCRIPTION
    构建DelGuard可执行文件和基本安装包的PowerShell脚本
    此版本不需要安装WiX工具集
#>

# 设置变量
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$InstallerDir = $PSScriptRoot
$OutputDir = Join-Path $InstallerDir "output"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "DelGuard 简单构建脚本（不需要WiX）" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 创建输出目录
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

Write-Host "`n1. 构建主程序..." -ForegroundColor Green
Set-Location (Join-Path $ProjectRoot "src\DelGuard")

try {
    dotnet restore
    if ($LASTEXITCODE -ne 0) { throw "主程序还原失败" }
    
    dotnet build -c Release
    if ($LASTEXITCODE -ne 0) { throw "主程序构建失败" }
    
    dotnet publish -c Release -r win-x64 --self-contained false
    if ($LASTEXITCODE -ne 0) { throw "主程序发布失败" }
    
    Write-Host "主程序构建完成" -ForegroundColor Green
}
catch {
    Write-Error "错误：$($_.Exception.Message)"
    exit 1
}

# 复制可执行文件到输出目录
$publishPath = "bin\Release\net8.0\win-x64\publish"
$exeSource = Join-Path $PWD $publishPath\delguard.exe
$exeDest = Join-Path $OutputDir delguard.exe

if (Test-Path $exeSource) {
    Copy-Item $exeSource $exeDest -Force
    Write-Host "可执行文件已复制到: $exeDest" -ForegroundColor Green
    
    # 复制配置文件
    $configSource = Join-Path $ProjectRoot "config\windows.json"
    $configDest = Join-Path $OutputDir windows.json
    if (Test-Path $configSource) {
        Copy-Item $configSource $configDest -Force
        Write-Host "配置文件已复制到: $configDest" -ForegroundColor Green
    }
    
    # 复制README文件
    $readmeSource = Join-Path $ProjectRoot "README.md"
    $readmeDest = Join-Path $OutputDir "README.md"
    if (Test-Path $readmeSource) {
        Copy-Item $readmeSource $readmeDest -Force
        Write-Host "文档已复制到: $readmeDest" -ForegroundColor Green
    }
    
    # 复制许可证文件
    $licenseSource = Join-Path $ProjectRoot "LICENSE"
    $licenseDest = Join-Path $OutputDir "LICENSE"
    if (Test-Path $licenseSource) {
        Copy-Item $licenseSource $licenseDest -Force
        Write-Host "许可证已复制到: $licenseDest" -ForegroundColor Green
    }
    
    # 创建安装脚本
    $installScript = @"
@echo off
echo 正在安装 DelGuard 到系统目录...
set "INSTALL_DIR=%ProgramFiles%\DelGuard"
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
copy /Y delguard.exe "%INSTALL_DIR%\"
copy /Y windows.json "%INSTALL_DIR%\config\"
if not exist "%INSTALL_DIR%\config" mkdir "%INSTALL_DIR%\config"
copy /Y windows.json "%INSTALL_DIR%\config\"
echo 将安装目录添加到系统 PATH...
setx PATH "%PATH%;%INSTALL_DIR%" /M
echo.
echo 安装完成！请重新打开命令提示符以使用 delguard 命令。
echo.
pause
"@
    
    $installScriptPath = Join-Path $OutputDir "install.bat"
    $installScript | Out-File -FilePath $installScriptPath -Encoding ascii
    
    # 创建卸载脚本
    $uninstallScript = @"
@echo off
echo 正在从系统中卸载 DelGuard...
set "INSTALL_DIR=%ProgramFiles%\DelGuard"
if exist "%INSTALL_DIR%" rmdir /S /Q "%INSTALL_DIR%"
echo 从系统 PATH 中移除安装目录...
REM 这里需要手动从系统环境变量中移除路径
echo.
echo 卸载完成！请手动从系统环境变量 PATH 中移除 %INSTALL_DIR%
echo.
pause
"@
    
    $uninstallScriptPath = Join-Path $OutputDir "uninstall.bat"
    $uninstallScript | Out-File -FilePath $uninstallScriptPath -Encoding ascii
    
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "构建完成！" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "输出目录: $OutputDir" -ForegroundColor Cyan
    Write-Host "可执行文件: $exeDest" -ForegroundColor Cyan
    Write-Host "安装脚本: $installScriptPath" -ForegroundColor Cyan
    Write-Host "卸载脚本: $uninstallScriptPath" -ForegroundColor Cyan
    
    $exeInfo = Get-Item $exeDest
    Write-Host "文件大小: $($exeInfo.Length) 字节" -ForegroundColor Cyan
    
    Write-Host "`n使用说明:" -ForegroundColor Yellow
    Write-Host "1. 直接使用: 运行 $exeDest [选项] [文件路径]" -ForegroundColor White
    Write-Host "2. 安装使用: 以管理员身份运行 $installScriptPath" -ForegroundColor White
    Write-Host "3. 卸载程序: 以管理员身份运行 $uninstallScriptPath" -ForegroundColor White
    
    Write-Host "\n构建脚本执行完毕！" -ForegroundColor Green
} else {
    Write-Error "错误：可执行文件未找到: $exeSource"
    exit 1
}

# 返回项目根目录
Set-Location $ProjectRoot

exit 0