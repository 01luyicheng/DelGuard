#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 综合兼容性检查工具
#>

[CmdletBinding()]
param(
    [switch]$All,
    [switch]$BuildTest,
    [switch]$InstallTest,
    [switch]$ConfigTest,
    [switch]$Verbose
)

$ErrorActionPreference = 'Stop'

function Write-LogInfo { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Cyan }
function Write-LogSuccess { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-LogWarning { param([string]$Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
function Write-LogError { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

# 检查系统环境
function Test-SystemEnvironment {
    Write-LogInfo "检查系统环境..."
    
    $results = @{
        OS = [System.Environment]::OSVersion.VersionString
        Architecture = [System.Environment]::Is64BitOperatingSystem
        PowerShellVersion = $PSVersionTable.PSVersion.ToString()
        PowerShellEdition = $PSVersionTable.PSEdition
        ExecutionPolicy = Get-ExecutionPolicy
        IsAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
        UserName = $env:USERNAME
        ComputerName = $env:COMPUTERNAME
    }
    
    Write-Host "系统信息:" -ForegroundColor Cyan
    $results.GetEnumerator() | ForEach-Object {
        Write-Host "  $($_.Key): $($_.Value)" -ForegroundColor White
    }
    
    return $results
}

# 检查 PowerShell 环境
function Test-PowerShellEnvironments {
    Write-LogInfo "检查 PowerShell 环境..."
    
    $environments = @()
    
    # Windows PowerShell 5.1
    $ps51Path = "$env:WINDIR\System32\WindowsPowerShell\v1.0\powershell.exe"
    if (Test-Path $ps51Path) {
        $environments += @{
            Name = "Windows PowerShell 5.1"
            Path = $ps51Path
            Available = $true
        }
    }
    
    # PowerShell 7+
    $ps7Command = Get-Command pwsh -ErrorAction SilentlyContinue
    if ($ps7Command) {
        $environments += @{
            Name = "PowerShell 7+"
            Path = $ps7Command.Source
            Available = $true
        }
    }
    
    Write-Host "PowerShell 环境:" -ForegroundColor Cyan
    foreach ($env in $environments) {
        $status = if ($env.Available) { "可用" } else { "不可用" }
        $color = if ($env.Available) { "Green" } else { "Red" }
        Write-Host "  $($env.Name): $status" -ForegroundColor $color
    }
    
    return $environments
}

# 测试安装脚本兼容性
function Test-InstallScriptCompatibility {
    Write-LogInfo "测试安装脚本兼容性..."
    
    $scripts = @(
        @{Name = "safe-install.ps1"; Path = "scripts/safe-install.ps1"},
        @{Name = "safe-install.sh"; Path = "scripts/safe-install.sh"}
    )
    
    $results = @()
    
    foreach ($script in $scripts) {
        $result = @{
            Name = $script.Name
            Path = $script.Path
            Exists = Test-Path $script.Path
            SyntaxValid = $false
            Error = $null
        }
        
        if ($result.Exists) {
            try {
                if ($script.Name -like "*.ps1") {
                    $null = [System.Management.Automation.PSParser]::Tokenize((Get-Content $script.Path -Raw), [ref]$null)
                    $result.SyntaxValid = $true
                } else {
                    $result.SyntaxValid = $true
                }
            } catch {
                $result.Error = $_.Exception.Message
            }
        } else {
            $result.Error = "文件不存在"
        }
        
        $results += $result
    }
    
    Write-Host "安装脚本检查:" -ForegroundColor Cyan
    foreach ($result in $results) {
        $status = if ($result.Exists -and $result.SyntaxValid) { "通过" } else { "失败" }
        $color = if ($result.Exists -and $result.SyntaxValid) { "Green" } else { "Red" }
        Write-Host "  $($result.Name): $status" -ForegroundColor $color
        
        if ($result.Error -and $Verbose) {
            Write-Host "    错误: $($result.Error)" -ForegroundColor Gray
        }
    }
    
    return $results
}

# 测试配置文件安全性
function Test-ConfigSafety {
    Write-LogInfo "测试配置文件安全性..."
    
    $configPaths = @(
        "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1",
        "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1"
    )
    
    $results = @()
    
    foreach ($configPath in $configPaths) {
        $result = @{
            Path = $configPath
            Exists = Test-Path $configPath
            HasDelGuardConfig = $false
            SafeToModify = $true
            Issues = @()
        }
        
        if ($result.Exists) {
            $content = Get-Content $configPath -Raw -ErrorAction SilentlyContinue
            
            if ($content) {
                $result.HasDelGuardConfig = $content.Contains("# DelGuard Configuration")
                
                if ($content -match "function\s+del\s*\(" -and -not $result.HasDelGuardConfig) {
                    $result.Issues += "检测到现有的 'del' 函数定义"
                    $result.SafeToModify = $false
                }
            }
        }
        
        $results += $result
    }
    
    Write-Host "配置文件安全性检查:" -ForegroundColor Cyan
    foreach ($result in $results) {
        $configName = if ($result.Path -like "*WindowsPowerShell*") { "PowerShell 5.1" } else { "PowerShell 7+" }
        
        if (-not $result.Exists) {
            Write-Host "  $configName : 不存在（安全）" -ForegroundColor Yellow
        } elseif ($result.SafeToModify) {
            Write-Host "  $configName : 安全可修改" -ForegroundColor Green
        } else {
            Write-Host "  $configName : 需要注意" -ForegroundColor Yellow
        }
        
        foreach ($issue in $result.Issues) {
            Write-Host "    问题: $issue" -ForegroundColor Yellow
        }
    }
    
    return $results
}

# 生成综合报告
function Generate-ComprehensiveReport {
    param(
        [hashtable]$SystemInfo,
        [array]$PowerShellEnvs,
        [array]$InstallScripts,
        [array]$ConfigSafety
    )
    
    Write-Host ""
    Write-Host "DelGuard 综合兼容性检查报告" -ForegroundColor Cyan
    Write-Host "============================" -ForegroundColor Cyan
    Write-Host "生成时间: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
    Write-Host ""
    
    # 系统兼容性评分
    $totalScore = 0
    $maxScore = 0
    
    # PowerShell 环境评分
    $psScore = ($PowerShellEnvs | Where-Object { $_.Available }).Count
    $psMaxScore = 2
    $totalScore += $psScore
    $maxScore += $psMaxScore
    
    Write-Host "PowerShell 环境支持: $psScore/$psMaxScore" -ForegroundColor $(if ($psScore -eq $psMaxScore) { "Green" } else { "Yellow" })
    
    # 安装脚本评分
    if ($InstallScripts.Count -gt 0) {
        $scriptScore = ($InstallScripts | Where-Object { $_.Exists -and $_.SyntaxValid }).Count
        $scriptMaxScore = $InstallScripts.Count
        $totalScore += $scriptScore
        $maxScore += $scriptMaxScore
        
        Write-Host "安装脚本兼容性: $scriptScore/$scriptMaxScore" -ForegroundColor $(if ($scriptScore -eq $scriptMaxScore) { "Green" } else { "Red" })
    }
    
    # 配置安全性评分
    if ($ConfigSafety.Count -gt 0) {
        $configScore = ($ConfigSafety | Where-Object { $_.SafeToModify }).Count
        $configMaxScore = $ConfigSafety.Count
        $totalScore += $configScore
        $maxScore += $configMaxScore
        
        Write-Host "配置文件安全性: $configScore/$configMaxScore" -ForegroundColor $(if ($configScore -eq $configMaxScore) { "Green" } else { "Yellow" })
    }
    
    # 总体评分
    $overallScore = if ($maxScore -gt 0) { [math]::Round(($totalScore / $maxScore) * 100, 1) } else { 0 }
    
    Write-Host ""
    Write-Host "总体兼容性评分: $overallScore%" -ForegroundColor $(
        if ($overallScore -ge 90) { "Green" }
        elseif ($overallScore -ge 70) { "Yellow" }
        else { "Red" }
    )
    
    Write-Host ""
    if ($overallScore -ge 90) {
        Write-Host "DelGuard 已准备好在当前环境中发布！" -ForegroundColor Green
    } elseif ($overallScore -ge 70) {
        Write-Host "DelGuard 基本可以发布，但建议先解决上述问题" -ForegroundColor Yellow
    } else {
        Write-Host "DelGuard 需要解决关键问题后才能发布" -ForegroundColor Red
    }
}

# 主函数
function Main {
    Write-Host "DelGuard 综合兼容性检查工具" -ForegroundColor Cyan
    Write-Host "============================" -ForegroundColor Cyan
    Write-Host ""
    
    $systemInfo = Test-SystemEnvironment
    Write-Host ""
    
    $psEnvs = Test-PowerShellEnvironments
    Write-Host ""
    
    $installScripts = @()
    if ($All -or $InstallTest) {
        $installScripts = Test-InstallScriptCompatibility
        Write-Host ""
    }
    
    $configSafety = @()
    if ($All -or $ConfigTest) {
        $configSafety = Test-ConfigSafety
        Write-Host ""
    }
    
    Generate-ComprehensiveReport -SystemInfo $systemInfo -PowerShellEnvs $psEnvs -InstallScripts $installScripts -ConfigSafety $configSafety
}

# 执行主函数
Main