#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 一键卸载脚本 - Windows版本

.DESCRIPTION
    自动卸载 DelGuard 安全删除工具，并清理相关配置。
    支持 PowerShell 5.1+ 和 PowerShell 7+。

.PARAMETER KeepConfig
    保留配置文件，不完全清理

.PARAMETER Force
    强制卸载，不提示确认

.EXAMPLE
    .\uninstall.ps1
    标准卸载，会提示确认

.EXAMPLE
    .\uninstall.ps1 -Force
    强制卸载，不提示确认

.EXAMPLE
    .\uninstall.ps1 -KeepConfig
    卸载但保留配置文件
#>

[CmdletBinding()]
param(
    [switch]$KeepConfig,
    [switch]$Force
)

# 设置错误处理
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# 常量定义
$APP_NAME = "DelGuard"
$EXECUTABLE_NAME = "delguard.exe"

# 颜色定义
$ColorScheme = @{
    Success = 'Green'
    Error = 'Red'
    Warning = 'Yellow'
    Info = 'Cyan'
    Header = 'Magenta'
    Normal = 'White'
}

# 查找已安装的DelGuard
function Find-InstalledDelGuard {
    # 检查常见安装位置
    $possibleLocations = @(
        "$env:LOCALAPPDATA\$APP_NAME\$EXECUTABLE_NAME",
        "$env:ProgramFiles\$APP_NAME\$EXECUTABLE_NAME",
        "$env:USERPROFILE\bin\$EXECUTABLE_NAME",
        "$env:USERPROFILE\.local\bin\$EXECUTABLE_NAME"
    )
    
    foreach ($location in $possibleLocations) {
        if (Test-Path $location) {
            return $location
        }
    }
    
    # 尝试从PATH中查找
    $fromPath = Get-Command $EXECUTABLE_NAME -ErrorAction SilentlyContinue
    if ($fromPath) {
        return $fromPath.Source
    }
    
    return $null
}

# 查找配置目录
function Find-ConfigDir {
    # 检查常见配置位置
    $possibleLocations = @(
        "$env:APPDATA\$APP_NAME",
        "$env:ProgramData\$APP_NAME",
        "$env:LOCALAPPDATA\$APP_NAME\config"
    )
    
    foreach ($location in $possibleLocations) {
        if (Test-Path $location) {
            return $location
        }
    }
    
    return $null
}

# 从PATH中移除
function Remove-FromPath {
    param([string]$Path)
    
    try {
        # 检查用户PATH
        $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($userPath -like "*$Path*") {
            $newPath = $userPath -replace [regex]::Escape(";$Path"), ""
            $newPath = $newPath -replace [regex]::Escape("$Path;"), ""
            $newPath = $newPath -replace [regex]::Escape($Path), ""
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-Host "已从用户PATH中移除: $Path" -ForegroundColor $ColorScheme.Success
        }
        
        # 检查系统PATH
        $systemPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
        if ($systemPath -like "*$Path*") {
            $newPath = $systemPath -replace [regex]::Escape(";$Path"), ""
            $newPath = $newPath -replace [regex]::Escape("$Path;"), ""
            $newPath = $newPath -replace [regex]::Escape($Path), ""
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
            Write-Host "已从系统PATH中移除: $Path" -ForegroundColor $ColorScheme.Success
        }
    } catch {
        Write-Host "从PATH中移除失败: $($_.Exception.Message)" -ForegroundColor $ColorScheme.Warning
    }
}

# 移除PowerShell别名
function Remove-PowerShellAliases {
    $profilePaths = @(
        $PROFILE.CurrentUserAllHosts,
        $PROFILE.CurrentUserCurrentHost,
        $PROFILE.AllUsersAllHosts,
        $PROFILE.AllUsersCurrentHost
    )
    
    foreach ($profilePath in $profilePaths) {
        if (Test-Path $profilePath) {
            $content = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
            
            # 检查是否包含DelGuard配置
            if ($content -match "# DelGuard") {
                # 移除DelGuard相关配置
                $newContent = $content -replace "(?s)# DelGuard.*?(?=\r?\n\r?\n|\r?\n$|$)", ""
                $newContent = $newContent.Trim()
                
                if ($newContent) {
                    Set-Content -Path $profilePath -Value $newContent -Encoding UTF8
                } else {
                    Remove-Item $profilePath -Force
                }
                Write-Host "已从PowerShell配置文件移除DelGuard别名: $profilePath" -ForegroundColor $ColorScheme.Success
            }
        }
    }
}

# 显示横幅
function Show-Banner {
    $banner = @"
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║                🗑️ DelGuard 一键卸载工具                      ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
"@
    Write-Host $banner -ForegroundColor $ColorScheme.Header
    Write-Host ""
}

# 主程序
try {
    Show-Banner
    
    # 查找已安装的DelGuard
    $installedPath = Find-InstalledDelGuard
    if (-not $installedPath) {
        Write-Host "未找到已安装的DelGuard。" -ForegroundColor $ColorScheme.Warning
        exit 0
    }
    
    $installDir = Split-Path $installedPath -Parent
    Write-Host "已找到DelGuard: $installedPath" -ForegroundColor $ColorScheme.Info
    
    # 查找配置目录
    $configDir = Find-ConfigDir
    if ($configDir) {
        Write-Host "已找到配置目录: $configDir" -ForegroundColor $ColorScheme.Info
    }
    
    # 确认卸载
    if (-not $Force) {
        $confirmation = Read-Host "确认卸载DelGuard？(Y/N)"
        if ($confirmation -ne "Y" -and $confirmation -ne "y") {
            Write-Host "卸载已取消。" -ForegroundColor $ColorScheme.Warning
            exit 0
        }
    }
    
    # 停止可能正在运行的DelGuard进程
    $processes = Get-Process | Where-Object { $_.Path -eq $installedPath }
    if ($processes) {
        Write-Host "正在停止DelGuard进程..." -ForegroundColor $ColorScheme.Warning
        $processes | Stop-Process -Force
        Start-Sleep -Seconds 1
    }
    
    # 删除可执行文件
    Remove-Item $installedPath -Force
    Write-Host "已删除可执行文件: $installedPath" -ForegroundColor $ColorScheme.Success
    
    # 删除安装目录（如果为空）
    if ((Test-Path $installDir) -and ((Get-ChildItem $installDir).Count -eq 0)) {
        Remove-Item $installDir -Force
        Write-Host "已删除空安装目录: $installDir" -ForegroundColor $ColorScheme.Success
    }
    
    # 从PATH中移除
    Remove-FromPath -Path $installDir
    
    # 移除PowerShell别名
    Remove-PowerShellAliases
    
    # 处理配置目录
    if ($configDir -and -not $KeepConfig) {
        Remove-Item $configDir -Recurse -Force
        Write-Host "已删除配置目录: $configDir" -ForegroundColor $ColorScheme.Success
    } elseif ($configDir) {
        Write-Host "已保留配置目录: $configDir" -ForegroundColor $ColorScheme.Info
    }
    
    Write-Host "DelGuard卸载完成！" -ForegroundColor $ColorScheme.Success
    
} catch {
    Write-Host "卸载失败: $($_.Exception.Message)" -ForegroundColor $ColorScheme.Error
    exit 1
}