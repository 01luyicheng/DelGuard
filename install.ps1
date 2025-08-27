#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 安装脚本 - Windows版本

.DESCRIPTION
    自动下载并安装 DelGuard 安全删除工具到系统中。
    支持 PowerShell 5.1+ 和 PowerShell 7+。

.PARAMETER Force
    强制重新安装，即使已经安装

.PARAMETER SystemWide
    系统级安装（需要管理员权限）

.PARAMETER Uninstall
    卸载 DelGuard

.PARAMETER Status
    检查安装状态

.EXAMPLE
    .\install.ps1
    标准安装

.EXAMPLE
    .\install.ps1 -Force
    强制重新安装

.EXAMPLE
    .\install.ps1 -SystemWide
    系统级安装

.EXAMPLE
    .\install.ps1 -Uninstall
    卸载 DelGuard
#>

[CmdletBinding()]
param(
    [switch]$Force,
    [switch]$SystemWide,
    [switch]$Uninstall,
    [switch]$Status
)

# 设置错误处理
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# 常量定义
$REPO_URL = "https://github.com/01luyicheng/DelGuard"
$RELEASE_API = "https://api.github.com/repos/01luyicheng/DelGuard/releases/latest"
$APP_NAME = "DelGuard"
$EXECUTABLE_NAME = "delguard.exe"

# 路径配置
if ($SystemWide) {
    $INSTALL_DIR = "$env:ProgramFiles\$APP_NAME"
    $CONFIG_DIR = "$env:ProgramData\$APP_NAME"
} else {
    $INSTALL_DIR = "$env:LOCALAPPDATA\$APP_NAME"
    $CONFIG_DIR = "$env:APPDATA\$APP_NAME"
}

$EXECUTABLE_PATH = Join-Path $INSTALL_DIR $EXECUTABLE_NAME
$LOG_FILE = Join-Path $CONFIG_DIR "install.log"

# 日志函数
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] [$Level] $Message"
    Write-Host $logMessage
    
    if (!(Test-Path (Split-Path $LOG_FILE))) {
        New-Item -ItemType Directory -Path (Split-Path $LOG_FILE) -Force | Out-Null
    }
    Add-Content -Path $LOG_FILE -Value $logMessage -Encoding UTF8
}

# 检查管理员权限
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 获取系统架构
function Get-SystemArchitecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        "x86" { return "386" }
        default { return "amd64" }
    }
}

# 检查网络连接
function Test-NetworkConnection {
    try {
        $response = Invoke-WebRequest -Uri "https://api.github.com" -Method Head -TimeoutSec 10
        return $response.StatusCode -eq 200
    } catch {
        return $false
    }
}

# 获取最新版本信息
function Get-LatestRelease {
    try {
        Write-Log "获取最新版本信息..."
        $response = Invoke-RestMethod -Uri $RELEASE_API -TimeoutSec 30
        return $response
    } catch {
        Write-Log "获取版本信息失败: $($_.Exception.Message)" "ERROR"
        throw "无法获取最新版本信息，请检查网络连接"
    }
}

# 下载文件
function Download-File {
    param([string]$Url, [string]$OutputPath)
    
    try {
        Write-Log "下载文件: $Url"
        $webClient = New-Object System.Net.WebClient
        $webClient.DownloadFile($Url, $OutputPath)
        Write-Log "下载完成: $OutputPath"
    } catch {
        Write-Log "下载失败: $($_.Exception.Message)" "ERROR"
        throw "下载失败: $($_.Exception.Message)"
    }
}

# 安装 DelGuard
function Install-DelGuard {
    Write-Log "开始安装 $APP_NAME..."
    
    # 检查管理员权限（系统级安装时）
    if ($SystemWide -and !(Test-Administrator)) {
        Write-Log "系统级安装需要管理员权限" "ERROR"
        throw "请以管理员身份运行 PowerShell"
    }
    
    # 检查网络连接
    if (!(Test-NetworkConnection)) {
        Write-Log "网络连接检查失败" "ERROR"
        throw "无法连接到 GitHub，请检查网络连接"
    }
    
    # 检查现有安装
    if ((Test-Path $EXECUTABLE_PATH) -and !$Force) {
        Write-Log "$APP_NAME 已经安装在 $EXECUTABLE_PATH"
        Write-Log "使用 -Force 参数强制重新安装"
        return
    }
    
    try {
        # 获取最新版本
        $release = Get-LatestRelease
        $version = $release.tag_name
        Write-Log "最新版本: $version"
        
        # 确定下载URL
        $arch = Get-SystemArchitecture
        $assetName = "$APP_NAME-windows-$arch.zip"
        $asset = $release.assets | Where-Object { $_.name -eq $assetName }
        
        if (!$asset) {
            Write-Log "未找到适合的安装包: $assetName" "ERROR"
            throw "未找到适合当前系统的安装包"
        }
        
        $downloadUrl = $asset.browser_download_url
        Write-Log "下载URL: $downloadUrl"
        
        # 创建临时目录
        $tempDir = Join-Path $env:TEMP "delguard-install"
        if (Test-Path $tempDir) {
            Remove-Item $tempDir -Recurse -Force
        }
        New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
        
        # 下载文件
        $zipPath = Join-Path $tempDir "$assetName"
        Download-File -Url $downloadUrl -OutputPath $zipPath
        
        # 解压文件
        Write-Log "解压安装包..."
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
        
        # 创建安装目录
        if (!(Test-Path $INSTALL_DIR)) {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        }
        
        # 复制文件
        $extractedExe = Get-ChildItem -Path $tempDir -Name $EXECUTABLE_NAME -Recurse | Select-Object -First 1
        if ($extractedExe) {
            $sourcePath = Join-Path $tempDir $extractedExe.FullName
            Copy-Item -Path $sourcePath -Destination $EXECUTABLE_PATH -Force
            Write-Log "已安装到: $EXECUTABLE_PATH"
        } else {
            throw "在安装包中未找到可执行文件"
        }
        
        # 添加到 PATH
        Add-ToPath -Path $INSTALL_DIR
        
        # 安装 PowerShell 别名
        Install-PowerShellAliases
        
        # 创建配置目录
        if (!(Test-Path $CONFIG_DIR)) {
            New-Item -ItemType Directory -Path $CONFIG_DIR -Force | Out-Null
        }
        
        # 清理临时文件
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        
        Write-Log "$APP_NAME $version 安装成功！"
        Write-Log "可执行文件位置: $EXECUTABLE_PATH"
        Write-Log "配置目录: $CONFIG_DIR"
        Write-Log ""
        Write-Log "使用方法:"
        Write-Log "  delguard file.txt          # 删除文件到回收站"
        Write-Log "  delguard -p file.txt       # 永久删除文件"
        Write-Log "  delguard --help            # 查看帮助"
        Write-Log ""
        Write-Log "请重新启动 PowerShell 以使用 delguard 命令"
        
    } catch {
        Write-Log "安装失败: $($_.Exception.Message)" "ERROR"
        throw
    }
}

# 添加到 PATH
function Add-ToPath {
    param([string]$Path)
    
    try {
        if ($SystemWide) {
            $envPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
            $target = "Machine"
        } else {
            $envPath = [Environment]::GetEnvironmentVariable("PATH", "User")
            $target = "User"
        }
        
        if ($envPath -notlike "*$Path*") {
            $newPath = "$envPath;$Path"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, $target)
            Write-Log "已添加到 PATH: $Path"
            
            # 更新当前会话的 PATH
            $env:PATH = "$env:PATH;$Path"
        } else {
            Write-Log "PATH 中已存在: $Path"
        }
    } catch {
        Write-Log "添加到 PATH 失败: $($_.Exception.Message)" "WARNING"
    }
}

# 安装 PowerShell 别名
function Install-PowerShellAliases {
    try {
        $profilePath = $PROFILE.CurrentUserAllHosts
        
        if (!(Test-Path $profilePath)) {
            New-Item -ItemType File -Path $profilePath -Force | Out-Null
        }
        
        $aliasContent = @"

# DelGuard 别名配置
if (Test-Path "$EXECUTABLE_PATH") {
    Set-Alias -Name delguard -Value "$EXECUTABLE_PATH" -Scope Global
    Set-Alias -Name dg -Value "$EXECUTABLE_PATH" -Scope Global
}
"@
        
        $currentContent = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
        if ($currentContent -notlike "*DelGuard 别名配置*") {
            Add-Content -Path $profilePath -Value $aliasContent -Encoding UTF8
            Write-Log "已添加 PowerShell 别名配置"
        } else {
            Write-Log "PowerShell 别名已存在"
        }
    } catch {
        Write-Log "配置 PowerShell 别名失败: $($_.Exception.Message)" "WARNING"
    }
}

# 卸载 DelGuard
function Uninstall-DelGuard {
    Write-Log "开始卸载 $APP_NAME..."
    
    try {
        # 删除可执行文件
        if (Test-Path $EXECUTABLE_PATH) {
            Remove-Item $EXECUTABLE_PATH -Force
            Write-Log "已删除: $EXECUTABLE_PATH"
        }
        
        # 删除安装目录（如果为空）
        if ((Test-Path $INSTALL_DIR) -and ((Get-ChildItem $INSTALL_DIR).Count -eq 0)) {
            Remove-Item $INSTALL_DIR -Force
            Write-Log "已删除安装目录: $INSTALL_DIR"
        }
        
        # 从 PATH 中移除
        Remove-FromPath -Path $INSTALL_DIR
        
        # 移除 PowerShell 别名
        Remove-PowerShellAliases
        
        Write-Log "$APP_NAME 卸载完成"
        Write-Log "配置文件保留在: $CONFIG_DIR"
        Write-Log "如需完全清理，请手动删除配置目录"
        
    } catch {
        Write-Log "卸载失败: $($_.Exception.Message)" "ERROR"
        throw
    }
}

# 从 PATH 中移除
function Remove-FromPath {
    param([string]$Path)
    
    try {
        if ($SystemWide) {
            $envPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
            $target = "Machine"
        } else {
            $envPath = [Environment]::GetEnvironmentVariable("PATH", "User")
            $target = "User"
        }
        
        if ($envPath -like "*$Path*") {
            $newPath = $envPath -replace [regex]::Escape(";$Path"), ""
            $newPath = $newPath -replace [regex]::Escape("$Path;"), ""
            $newPath = $newPath -replace [regex]::Escape($Path), ""
            [Environment]::SetEnvironmentVariable("PATH", $newPath, $target)
            Write-Log "已从 PATH 中移除: $Path"
        }
    } catch {
        Write-Log "从 PATH 中移除失败: $($_.Exception.Message)" "WARNING"
    }
}

# 移除 PowerShell 别名
function Remove-PowerShellAliases {
    try {
        $profilePath = $PROFILE.CurrentUserAllHosts
        
        if (Test-Path $profilePath) {
            $content = Get-Content $profilePath -Raw
            $newContent = $content -replace "(?s)# DelGuard 别名配置.*?(?=\r?\n\r?\n|\r?\n$|$)", ""
            $newContent = $newContent.Trim()
            
            if ($newContent) {
                Set-Content -Path $profilePath -Value $newContent -Encoding UTF8
            } else {
                Remove-Item $profilePath -Force
            }
            Write-Log "已移除 PowerShell 别名配置"
        }
    } catch {
        Write-Log "移除 PowerShell 别名失败: $($_.Exception.Message)" "WARNING"
    }
}

# 检查安装状态
function Get-InstallStatus {
    Write-Host "=== DelGuard 安装状态 ===" -ForegroundColor Cyan
    
    if (Test-Path $EXECUTABLE_PATH) {
        Write-Host "✓ 已安装" -ForegroundColor Green
        Write-Host "  位置: $EXECUTABLE_PATH" -ForegroundColor Gray
        
        try {
            $version = & $EXECUTABLE_PATH --version 2>$null
            Write-Host "  版本: $version" -ForegroundColor Gray
        } catch {
            Write-Host "  版本: 无法获取" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✗ 未安装" -ForegroundColor Red
    }
    
    # 检查 PATH
    $pathCheck = $env:PATH -split ';' | Where-Object { $_ -eq $INSTALL_DIR }
    if ($pathCheck) {
        Write-Host "✓ 已添加到 PATH" -ForegroundColor Green
    } else {
        Write-Host "✗ 未添加到 PATH" -ForegroundColor Yellow
    }
    
    # 检查别名
    if (Get-Alias delguard -ErrorAction SilentlyContinue) {
        Write-Host "✓ PowerShell 别名已配置" -ForegroundColor Green
    } else {
        Write-Host "✗ PowerShell 别名未配置" -ForegroundColor Yellow
    }
    
    # 检查配置目录
    if (Test-Path $CONFIG_DIR) {
        Write-Host "✓ 配置目录存在: $CONFIG_DIR" -ForegroundColor Green
    } else {
        Write-Host "✗ 配置目录不存在" -ForegroundColor Yellow
    }
}

# 主函数
function Main {
    try {
        Write-Host "DelGuard 安装程序" -ForegroundColor Cyan
        Write-Host "=================" -ForegroundColor Cyan
        Write-Host ""
        
        if ($Status) {
            Get-InstallStatus
            return
        }
        
        if ($Uninstall) {
            Uninstall-DelGuard
            return
        }
        
        Install-DelGuard
        
    } catch {
        Write-Log "操作失败: $($_.Exception.Message)" "ERROR"
        Write-Host ""
        Write-Host "安装失败！" -ForegroundColor Red
        Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "日志文件: $LOG_FILE" -ForegroundColor Gray
        exit 1
    }
}

# 执行主函数
Main