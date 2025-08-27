#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 一键更新脚本 - Windows版本

.DESCRIPTION
    自动检查并更新 DelGuard 安全删除工具到最新版本。
    支持 PowerShell 5.1+ 和 PowerShell 7+。

.PARAMETER Force
    强制更新，即使已经是最新版本

.PARAMETER CheckOnly
    仅检查更新，不执行更新操作

.EXAMPLE
    .\update.ps1
    检查并更新DelGuard

.EXAMPLE
    .\update.ps1 -Force
    强制更新到最新版本

.EXAMPLE
    .\update.ps1 -CheckOnly
    仅检查是否有更新可用
#>

[CmdletBinding()]
param(
    [switch]$Force,
    [switch]$CheckOnly
)

# 设置错误处理
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# 常量定义
$REPO_URL = "https://github.com/01luyicheng/DelGuard"
$RELEASE_API = "https://api.github.com/repos/01luyicheng/DelGuard/releases/latest"
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