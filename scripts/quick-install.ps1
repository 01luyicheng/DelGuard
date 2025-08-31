# DelGuard 一键安装脚本 (Windows PowerShell)
# 从GitHub下载最新版本并自动安装

param(
    [string]$Version = "v1.5.9",
    [string]$Repo = "DelGuard",
    [string]$Owner = "01luyicheng",  # GitHub用户名
    [switch]$Force
)

$ErrorActionPreference = "Stop"

# 颜色定义
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Cyan = "Cyan"
    White = "White"
}

function Write-ColorMessage {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Test-Admin {
    return ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
}

function Get-LatestVersion {
    try {
        $apiUrl = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
        $response = Invoke-RestMethod -Uri $apiUrl -Headers @{"User-Agent"="DelGuard-Installer"}
        return $response.tag_name
    } catch {
        Write-ColorMessage "⚠️ 无法获取最新版本，使用指定版本: $Version" Yellow
        return $Version
    }
}

function Download-DelGuard {
    param([string]$version)
    
    $arch = "amd64"
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        $arch = "arm64"
    }
    
    $filename = "delguard-windows-$arch.exe"
    $downloadUrl = "https://github.com/$Owner/$Repo/releases/download/$version/$filename"
    $tempDir = [System.IO.Path]::GetTempPath()
    $downloadPath = Join-Path $tempDir "delguard-$version.exe"
    
    Write-ColorMessage "📥 正在下载 DelGuard $version..." Cyan
    Write-ColorMessage "下载地址: $downloadUrl" White
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $downloadPath -UseBasicParsing
        Write-ColorMessage "✅ 下载完成" Green
        return $downloadPath
    } catch {
        Write-ColorMessage "❌ 下载失败: $($_.Exception.Message)" Red
        throw
    }
}

function Install-DelGuard {
    param([string]$binaryPath)
    
    $installDir = "$env:ProgramFiles\DelGuard"
    $backupDir = "$installDir\backup"
    
    # 检查是否已安装
    if (Test-Path $installDir) {
        if (-not $Force) {
            Write-ColorMessage "⚠️ DelGuard 已安装，使用 --Force 参数重新安装" Yellow
            return $false
        }
        Write-ColorMessage "🔄 检测到现有安装，正在重新安装..." Yellow
    }
    
    # 创建目录
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
    
    # 复制可执行文件
    Copy-Item $binaryPath "$installDir\delguard.exe" -Force
    Write-ColorMessage "✅ DelGuard 已安装到 $installDir" Green
    
    # 添加到系统PATH
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($currentPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "Machine")
        Write-ColorMessage "✅ 已添加到系统PATH" Green
    }
    
    # 创建del命令替换脚本
    $delScript = @"
@echo off
REM DelGuard 安全删除脚本
"$installDir\delguard.exe" delete %*
"@
    
    $delScriptPath = "$installDir\del.bat"
    $delScript | Out-File -FilePath $delScriptPath -Encoding ASCII
    
    # 创建卸载脚本
    $uninstallScript = @"
@echo off
echo 正在卸载 DelGuard...
set "installDir=$installDir"
set "pathToRemove=$installDir"

:: 从PATH中移除
for /f "usebackq tokens=2,*" %%A in (`reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path`) do (
    set "currentPath=%%B"
)
set "newPath=!currentPath:%installDir%;=!"
set "newPath=!newPath:;%installDir%=!"
set "newPath=!newPath:%installDir%=!"
reg add "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path /t REG_SZ /d "!newPath!" /f >nul

:: 删除安装目录
rmdir /s /q "$installDir"

echo DelGuard 已成功卸载
pause
"@
    
    $uninstallScript | Out-File -FilePath "$installDir\uninstall.bat" -Encoding ASCII
    
    return $true
}

function Show-Usage {
    Write-ColorMessage ""
    Write-ColorMessage "🎯 DelGuard 一键安装脚本" Green
    Write-ColorMessage ""
    Write-ColorMessage "用法:"
    Write-ColorMessage "  .\quick-install.ps1 [选项]"
    Write-ColorMessage ""
    Write-ColorMessage "选项:"
    Write-ColorMessage "  -Version <版本>    指定版本 (默认: v1.4.1)" Cyan
    Write-ColorMessage "  -Force             强制重新安装" Cyan
    Write-ColorMessage "  -Owner <用户名>    GitHub用户名 (默认: 01luyicheng)" Cyan
    Write-ColorMessage ""
    Write-ColorMessage "示例:"
    Write-ColorMessage "  .\quick-install.ps1" White
    Write-ColorMessage "  .\quick-install.ps1 -Version v1.4.1" White
    Write-ColorMessage "  .\quick-install.ps1 -Force" White
}

# 主逻辑
if ($args -contains "-h" -or $args -contains "--help") {
    Show-Usage
    exit 0
}

# 检查管理员权限
if (-not (Test-Admin)) {
    Write-ColorMessage "❌ 需要管理员权限运行" Red
    Write-ColorMessage "请右键点击PowerShell并选择'以管理员身份运行'" Yellow
    exit 1
}

Write-ColorMessage "🚀 DelGuard 一键安装程序" Green
Write-ColorMessage "从GitHub下载并安装最新版本" White
Write-ColorMessage ""

# 获取最新版本
if ($Version -eq "latest") {
    $Version = Get-LatestVersion
} elseif ($Version -notlike "v*") {
    $Version = "v$Version"
}

Write-ColorMessage "📦 版本: $Version" Cyan

# 下载并安装
try {
    $binaryPath = Download-DelGuard -version $Version
    
    if (Install-DelGuard -binaryPath $binaryPath) {
        Write-ColorMessage ""
        Write-ColorMessage "🎉 安装完成！" Green
        Write-ColorMessage ""
        Write-ColorMessage "📖 使用说明:" Yellow
        Write-ColorMessage "  delguard --help    - 查看帮助" White
        Write-ColorMessage "  delguard list      - 查看回收站" White
        Write-ColorMessage "  delguard restore   - 恢复文件" White
        Write-ColorMessage ""
        Write-ColorMessage "⚠️  请重新打开命令提示符或PowerShell" Yellow
        Write-ColorMessage "   或运行: refreshenv" Yellow
    }
} catch {
    Write-ColorMessage "❌ 安装失败: $($_.Exception.Message)" Red
    exit 1
} finally {
    # 清理临时文件
    if ($binaryPath -and (Test-Path $binaryPath)) {
        Remove-Item $binaryPath -Force -ErrorAction SilentlyContinue
    }
}