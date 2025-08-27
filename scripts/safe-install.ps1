#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 安全安装脚本 - Windows 版本 v2.0
.DESCRIPTION
    安全地安装 DelGuard，不破坏现有系统配置
.PARAMETER Force
    强制重新安装，覆盖现有配置
.PARAMETER SystemWide
    系统级安装（需要管理员权限）
.PARAMETER Uninstall
    卸载 DelGuard
.PARAMETER Status
    检查安装状态
.PARAMETER DryRun
    试运行模式，不实际修改文件
.PARAMETER NoBackup
    不备份现有配置文件
.PARAMETER Verbose
    详细输出
#>

[CmdletBinding()]
param(
    [switch]$Force,
    [switch]$SystemWide,
    [switch]$Uninstall,
    [switch]$Status,
    [switch]$DryRun,
    [switch]$NoBackup,
    [switch]$Verbose
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# 脚本版本和常量
$SCRIPT_VERSION = "2.0"
$APP_NAME = "DelGuard"
$EXECUTABLE_NAME = "delguard.exe"
$REPO_URL = "https://github.com/01luyicheng/DelGuard"
$RELEASE_API = "https://api.github.com/repos/01luyicheng/DelGuard/releases/latest"

# 检测 PowerShell 版本和平台
$IsPS7Plus = $PSVersionTable.PSVersion.Major -ge 7
$IsWindowsOS = if ($IsPS7Plus) { $IsWindows } else { $true }

if (-not $IsWindowsOS) {
    Write-Error "此脚本仅适用于 Windows 系统。请在 Linux/macOS 上使用 safe-install.sh"
}

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
function Write-LogInfo { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Cyan }
function Write-LogSuccess { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-LogWarning { param([string]$Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
function Write-LogError { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

# 检测 PowerShell 配置文件
function Get-PowerShellProfiles {
    $profiles = @()
    
    if ($IsPS7Plus) {
        $profiles += @{
            Name = "PowerShell 7+ (Current User)"
            Path = "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1"
            Priority = 1
        }
    }
    
    $profiles += @{
        Name = "Windows PowerShell 5.1 (Current User)"
        Path = "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
        Priority = 2
    }
    
    if ($SystemWide) {
        if ($IsPS7Plus) {
            $profiles += @{
                Name = "PowerShell 7+ (All Users)"
                Path = "$env:ProgramFiles\PowerShell\7\Microsoft.PowerShell_profile.ps1"
                Priority = 3
            }
        }
        
        $profiles += @{
            Name = "Windows PowerShell 5.1 (All Users)"
            Path = "$env:WINDIR\System32\WindowsPowerShell\v1.0\Microsoft.PowerShell_profile.ps1"
            Priority = 4
        }
    }
    
    return $profiles | Sort-Object Priority
}

# 安全备份配置文件
function Backup-ConfigFile {
    param([string]$ConfigPath)
    
    if ((Test-Path $ConfigPath) -and (-not $NoBackup)) {
        $backupPath = "$ConfigPath.delguard-backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
        Copy-Item $ConfigPath $backupPath -Force
        Write-LogInfo "已备份配置文件: $backupPath"
        return $backupPath
    }
    return $null
}

# 检查现有配置
function Test-ExistingConfig {
    param([string]$ConfigPath)
    
    if (Test-Path $ConfigPath) {
        $content = Get-Content $ConfigPath -Raw -ErrorAction SilentlyContinue
        return $content -and $content.Contains("# DelGuard Configuration")
    }
    return $false
}

# 生成 PowerShell 配置内容
function New-PowerShellConfig {
    param([string]$ExecutablePath)
    
    $installDir = Split-Path $ExecutablePath -Parent
    
    return @"
# DelGuard Configuration
# Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
# Version: DelGuard $SCRIPT_VERSION for PowerShell

`$delguardPath = '$ExecutablePath'

if (Test-Path `$delguardPath) {
    # 定义安全的函数别名
    function global:delguard {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        & `$delguardPath @Arguments
    }
    
    function global:del {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "DelGuard: 请指定要删除的文件或目录" -ForegroundColor Yellow
            return
        }
        & `$delguardPath -i @Arguments
    }
    
    # 只有在用户明确要求时才覆盖 rm 命令
    if (`$env:DELGUARD_OVERRIDE_RM -eq '1') {
        function global:rm {
            param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
            & `$delguardPath -i @Arguments
        }
    }
    
    # 安全的复制函数
    function global:delguard-cp {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        & `$delguardPath --cp @Arguments
    }
    
    # 添加到 PATH（如果尚未添加）
    if (`$env:PATH -notlike "*$installDir*") {
        `$env:PATH = "$installDir;" + `$env:PATH
    }
    
    # 显示加载消息（每个会话只显示一次）
    if (-not `$global:DelGuardLoaded) {
        Write-Host 'DelGuard loaded successfully (PowerShell)' -ForegroundColor Green
        Write-Host 'Commands: delguard, del, delguard-cp' -ForegroundColor Cyan
        Write-Host 'Set `$env:DELGUARD_OVERRIDE_RM="1" to override rm command' -ForegroundColor Gray
        Write-Host 'Use --help for detailed help' -ForegroundColor Gray
        `$global:DelGuardLoaded = `$true
    }
} else {
    Write-Warning "DelGuard executable not found: `$delguardPath"
}
# End DelGuard Configuration
"@
}

# 主安装函数
function Install-DelGuard {
    Write-LogInfo "开始安装 $APP_NAME v$SCRIPT_VERSION..."
    
    if ($DryRun) {
        Write-LogInfo "试运行模式 - 将执行以下操作:"
        Write-LogInfo "1. 下载最新版本的 DelGuard"
        Write-LogInfo "2. 安装到: $INSTALL_DIR"
        Write-LogInfo "3. 配置 PowerShell 别名"
        Write-LogInfo "4. 添加到系统 PATH"
        return
    }
    
    # 创建安装目录
    if (-not (Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        Write-LogSuccess "创建安装目录: $INSTALL_DIR"
    }
    
    # 配置 PowerShell 配置文件
    $profiles = Get-PowerShellProfiles
    $configAdded = $false
    
    foreach ($profile in $profiles) {
        if (Add-SafeConfig $profile.Path (New-PowerShellConfig $EXECUTABLE_PATH) $profile.Name) {
            $configAdded = $true
        }
    }
    
    if ($configAdded) {
        Write-LogSuccess "$APP_NAME 安装完成！"
        Write-LogInfo "请重新启动 PowerShell 或运行以下命令重新加载配置:"
        Write-LogInfo ". `$PROFILE"
    } else {
        Write-LogWarning "未能添加配置，请手动配置或使用 -Force 参数"
    }
}

# 安全地添加配置
function Add-SafeConfig {
    param(
        [string]$ConfigPath,
        [string]$ConfigContent,
        [string]$ProfileName
    )
    
    if (Test-ExistingConfig $ConfigPath) {
        if (-not $Force) {
            Write-LogWarning "DelGuard 配置已存在于: $ProfileName"
            Write-LogWarning "使用 -Force 参数覆盖现有配置"
            return $false
        }
        
        Backup-ConfigFile $ConfigPath
        Remove-ExistingConfig $ConfigPath
    }
    
    $configDir = Split-Path $ConfigPath -Parent
    if (-not (Test-Path $configDir)) {
        New-Item -ItemType Directory -Path $configDir -Force | Out-Null
        Write-LogSuccess "创建配置目录: $configDir"
    }
    
    if (-not (Test-Path $ConfigPath)) {
        Set-Content $ConfigPath "# PowerShell Profile`n" -Encoding UTF8
    }
    
    Add-Content $ConfigPath "`n$ConfigContent`n" -Encoding UTF8
    Write-LogSuccess "已更新 $ProfileName"
    return $true
}

# 移除现有配置
function Remove-ExistingConfig {
    param([string]$ConfigPath)
    
    if (Test-Path $ConfigPath) {
        $content = Get-Content $ConfigPath -Raw -ErrorAction SilentlyContinue
        if ($content) {
            $newContent = $content -replace '(?s)# DelGuard Configuration.*?# End DelGuard Configuration\r?\n?', ''
            $newContent = $newContent.Trim()
            
            if ($newContent) {
                Set-Content $ConfigPath $newContent -Encoding UTF8
            } else {
                Set-Content $ConfigPath "# PowerShell Profile" -Encoding UTF8
            }
            Write-LogInfo "已移除现有 DelGuard 配置"
        }
    }
}

# 卸载函数
function Uninstall-DelGuard {
    Write-LogInfo "开始卸载 $APP_NAME..."
    
    $profiles = Get-PowerShellProfiles
    foreach ($profile in $profiles) {
        if (Test-ExistingConfig $profile.Path) {
            Backup-ConfigFile $profile.Path
            Remove-ExistingConfig $profile.Path
            Write-LogSuccess "已从 $($profile.Name) 移除配置"
        }
    }
    
    if (Test-Path $INSTALL_DIR) {
        Remove-Item $INSTALL_DIR -Recurse -Force -ErrorAction SilentlyContinue
        Write-LogSuccess "已删除安装目录: $INSTALL_DIR"
    }
    
    Write-LogSuccess "$APP_NAME 卸载完成！"
}

# 检查安装状态
function Check-InstallStatus {
    Write-LogInfo "检查 $APP_NAME 安装状态..."
    
    $installed = Test-Path $EXECUTABLE_PATH
    Write-Host "可执行文件: " -NoNewline
    if ($installed) {
        Write-Host "已安装 ($EXECUTABLE_PATH)" -ForegroundColor Green
    } else {
        Write-Host "未安装" -ForegroundColor Red
    }
    
    $profiles = Get-PowerShellProfiles
    foreach ($profile in $profiles) {
        $configured = Test-ExistingConfig $profile.Path
        Write-Host "$($profile.Name): " -NoNewline
        if ($configured) {
            Write-Host "已配置" -ForegroundColor Green
        } else {
            Write-Host "未配置" -ForegroundColor Yellow
        }
    }
}

# 显示帮助
function Show-Help {
    Write-Host @"
DelGuard 安全安装脚本 v$SCRIPT_VERSION

用法: .\safe-install.ps1 [选项]

选项:
    -Force         强制重新安装，覆盖现有配置
    -SystemWide    系统级安装（需要管理员权限）
    -Uninstall     卸载 DelGuard
    -Status        检查安装状态
    -DryRun        试运行模式，不实际修改文件
    -NoBackup      不备份现有配置文件
    -Verbose       详细输出
    -Help          显示此帮助信息

环境变量:
    DELGUARD_OVERRIDE_RM=1    允许 DelGuard 覆盖系统 rm 命令

示例:
    .\safe-install.ps1                # 标准安装
    .\safe-install.ps1 -Force         # 强制重新安装
    .\safe-install.ps1 -SystemWide    # 系统级安装
    .\safe-install.ps1 -DryRun        # 预览安装过程
    .\safe-install.ps1 -Uninstall     # 卸载 DelGuard

"@ -ForegroundColor Cyan
}

# 主函数
function Main {
    Write-Host "DelGuard 安全安装程序 v$SCRIPT_VERSION" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host ""
    
    if ($Status) {
        Check-InstallStatus
    } elseif ($Uninstall) {
        Uninstall-DelGuard
    } else {
        Install-DelGuard
    }
}

# 错误处理
trap {
    Write-LogError "安装过程中发生错误: $_"
    exit 1
}

# 执行主函数
Main