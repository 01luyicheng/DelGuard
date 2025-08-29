# DelGuard Windows 安装脚本
# 需要管理员权限运行

param(
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

# 检查管理员权限
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "❌ 此脚本需要管理员权限运行" -ForegroundColor Red
    Write-Host "请右键点击PowerShell并选择'以管理员身份运行'" -ForegroundColor Yellow
    exit 1
}

$DelGuardPath = Join-Path $PSScriptRoot "..\delguard.exe"
$InstallDir = "$env:ProgramFiles\DelGuard"
$BackupDir = "$InstallDir\backup"

function Install-DelGuard {
    Write-Host "🚀 开始安装 DelGuard..." -ForegroundColor Green
    
    # 创建安装目录
    if (!(Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    if (!(Test-Path $BackupDir)) {
        New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
    }
    
    # 复制DelGuard可执行文件
    if (Test-Path $DelGuardPath) {
        Copy-Item $DelGuardPath "$InstallDir\delguard.exe" -Force
        Write-Host "✅ DelGuard 已复制到 $InstallDir" -ForegroundColor Green
    } else {
        Write-Host "❌ 找不到 delguard.exe，请先编译项目" -ForegroundColor Red
        exit 1
    }
    
    # 创建del命令替换脚本
    $DelScript = @"
@echo off
REM DelGuard 安全删除脚本
"$InstallDir\delguard.exe" delete %*
"@
    
    $DelScriptPath = "$InstallDir\del.bat"
    $DelScript | Out-File -FilePath $DelScriptPath -Encoding ASCII
    
    # 添加到系统PATH
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($CurrentPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "Machine")
        Write-Host "✅ 已添加到系统PATH" -ForegroundColor Green
    }
    
    # 创建卸载信息
    $UninstallInfo = @{
        InstallDate = Get-Date
        Version = "1.0.0"
        InstallDir = $InstallDir
    }
    $UninstallInfo | ConvertTo-Json | Out-File "$InstallDir\uninstall.json"
    
    Write-Host "🎉 DelGuard 安装完成！" -ForegroundColor Green
    Write-Host "现在可以使用以下命令：" -ForegroundColor Yellow
    Write-Host "  del <文件>     - 安全删除文件到回收站" -ForegroundColor Cyan
    Write-Host "  delguard list  - 查看回收站内容" -ForegroundColor Cyan
    Write-Host "  delguard restore <文件> - 恢复文件" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "⚠️  请重新打开命令提示符以使PATH生效" -ForegroundColor Yellow
}

function Uninstall-DelGuard {
    Write-Host "🗑️  开始卸载 DelGuard..." -ForegroundColor Yellow
    
    if (!(Test-Path $InstallDir)) {
        Write-Host "❌ DelGuard 未安装" -ForegroundColor Red
        exit 1
    }
    
    # 从PATH中移除
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $NewPath = $CurrentPath -replace [regex]::Escape(";$InstallDir"), ""
    $NewPath = $NewPath -replace [regex]::Escape("$InstallDir;"), ""
    $NewPath = $NewPath -replace [regex]::Escape("$InstallDir"), ""
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "Machine")
    
    # 删除安装目录
    Remove-Item $InstallDir -Recurse -Force
    
    Write-Host "✅ DelGuard 已成功卸载" -ForegroundColor Green
}

# 主逻辑
if ($Uninstall) {
    Uninstall-DelGuard
} else {
    Install-DelGuard
}