#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 增强安装脚本 - Windows版本

.DESCRIPTION
    自动下载并安装 DelGuard 安全删除工具到系统中。
    支持 PowerShell 5.1+ 和 PowerShell 7+。
    增强功能：自动设置UTF-8编码、智能语言检测、环境兼容性检查。

.PARAMETER Force
    强制重新安装，即使已经安装

.PARAMETER SystemWide
    系统级安装（需要管理员权限）

.PARAMETER Uninstall
    卸载 DelGuard

.PARAMETER Status
    检查安装状态

.PARAMETER SetUtf8
    设置PowerShell为UTF-8编码（默认启用）

.PARAMETER NoSetUtf8
    不设置PowerShell为UTF-8编码

.EXAMPLE
    .\install_enhanced.ps1
    标准安装

.EXAMPLE
    .\install_enhanced.ps1 -Force
    强制重新安装

.EXAMPLE
    .\install_enhanced.ps1 -SystemWide
    系统级安装

.EXAMPLE
    .\install_enhanced.ps1 -Uninstall
    卸载 DelGuard
#>

[CmdletBinding()]
param(
    [switch]$Force,
    [switch]$SystemWide,
    [switch]$Uninstall,
    [switch]$Status,
    [switch]$SetUtf8 = $true,
    [switch]$NoSetUtf8
)

# 如果指定了NoSetUtf8，则覆盖SetUtf8的默认值
if ($NoSetUtf8) {
    $SetUtf8 = $false
}

# 设置错误处理
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# 常量定义
$REPO_URL = "https://github.com/01luyicheng/DelGuard"
$RELEASE_API = "https://api.github.com/repos/01luyicheng/DelGuard/releases/latest"
$APP_NAME = "DelGuard"
$EXECUTABLE_NAME = "delguard.exe"
$VERSION = "2.1.0"

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