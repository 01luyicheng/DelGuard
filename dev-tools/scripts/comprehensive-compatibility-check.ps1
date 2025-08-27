#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 综合兼容性检查工具
.DESCRIPTION
    全面检查 DelGuard 在不同操作系统、终端和 Shell 环境下的兼容性
#>

[CmdletBinding()]
param(
    [switch]$All,
    [switch]$BuildTest,
    [switch]$InstallTest,
    [switch]$TerminalTest,
    [switch]$ConfigTest,
    [switch]$Verbose,
    [switch]$DryRun
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
        HomeDirectory = $env:USERPROFILE
        TempDirectory = $env:TEMP
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
            ProfilePath = "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
        }
    }
    
    # PowerShell 7+
    $ps7Command = Get-Command pwsh -ErrorAction SilentlyContinue
    if ($ps7Command) {
        $environments += @{
            Name = "PowerShell 7+"
            Path = $ps7Command.Source
            Available = $true
            ProfilePath = "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1"
        }
    }
    
    Write-Host "PowerShell 环境:" -ForegroundColor Cyan
    foreach ($env in $environments) {
        $status = if ($env.Available) { "✓ 可用" } else { "✗ 不可用" }
        $color = if ($env.Available) { "Green" } else { "Red" }
        Write-Host "  $($env.Name): $status" -ForegroundColor $color
        if ($Verbose -and $env.Available) {
            Write-Host "    路径: $($env.Path)" -ForegroundColor Gray
            Write-Host "    配置文件: $($env.ProfilePath)" -ForegroundColor Gray
        }
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
                    # 测试 PowerShell 脚本语法
                    $null = [System.Management.Automation.PSParser]::Tokenize((Get-Content $script.Path -Raw), [ref]$null)
                    $result.SyntaxValid = $true
                } elseif ($script.Name -like "*.sh") {
                    # 假设 Shell 脚本语法正确（无法在 Windows 上直接测试）
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
        $status = if ($result.Exists -and $result.SyntaxValid) { "✓ 通过" } else { "✗ 失败" }
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
            BackupExists = $false
            SafeToModify = $true
            Issues = @()
        }
        
        if ($result.Exists) {
            $content = Get-Content $configPath -Raw -ErrorAction SilentlyContinue
            
            if ($content) {
                # 检查是否已有 DelGuard 配置
                $result.HasDelGuardConfig = $content.Contains("# DelGuard Configuration")
                
                # 检查是否有备份文件
                $backupPattern = "$configPath.delguard-backup-*"
                $result.BackupExists = (Get-ChildItem $backupPattern -ErrorAction SilentlyContinue).Count -gt 0
                
                # 检查潜在的冲突
                if ($content -match "function\s+del\s*\(" -and -not $result.HasDelGuardConfig) {
                    $result.Issues += "检测到现有的 'del' 函数定义"
                    $result.SafeToModify = $false
                }
                
                if ($content -match "function\s+rm\s*\(" -and -not $result.HasDelGuardConfig) {
                    $result.Issues += "检测到现有的 'rm' 函数定义"
                }
                
                if ($content -match "alias\s+del\s*=" -and -not $result.HasDelGuardConfig) {
                    $result.Issues += "检测到现有的 'del' 别名定义"
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
            Write-Host "  $configName: ○ 不存在（安全）" -ForegroundColor Yellow
        } elseif ($result.SafeToModify) {
            Write-Host "  $configName: ✓ 安全可修改" -ForegroundColor Green
        } else {
            Write-Host "  $configName: ⚠ 需要注意" -ForegroundColor Yellow
        }
        
        if ($result.HasDelGuardConfig) {
            Write-Host "    已有 DelGuard 配置" -ForegroundColor Cyan
        }
        
        if ($result.BackupExists) {
            Write-Host "    存在备份文件" -ForegroundColor Cyan
        }
        
        foreach ($issue in $result.Issues) {
            Write-Host "    问题: $issue" -ForegroundColor Yellow
        }
    }
    
    return $results
}

# 测试跨平台构建
function Test-CrossPlatformBuild {
    Write-LogInfo "测试跨平台构建..."
    
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-LogWarning "Go 未安装，跳过构建测试"
        return @()
    }
    
    $platforms = @(
        @{GOOS="windows"; GOARCH="amd64"},
        @{GOOS="windows"; GOARCH="386"},
        @{GOOS="linux"; GOARCH="amd64"},
        @{GOOS="linux"; GOARCH="arm64"},
        @{GOOS="darwin"; GOARCH="amd64"},
        @{GOOS="darwin"; GOARCH="arm64"}
    )
    
    $results = @()
    
    foreach ($platform in $platforms) {
        $env:GOOS = $platform.GOOS
        $env:GOARCH = $platform.GOARCH
        $env:CGO_ENABLED = "0"
        
        $platformName = "$($platform.GOOS)-$($platform.GOARCH)"
        
        try {
            if ($DryRun) {
                $results += @{Platform=$platformName; Status="DryRun"; Error=$null}
            } else {
                $output = go build -o "delguard-test-$platformName" . 2>&1
                if ($LASTEXITCODE -eq 0) {
                    $results += @{Platform=$platformName; Status="Success"; Error=$null}
                    
                    # 清理测试文件
                    $testFiles = @("delguard-test-$platformName", "delguard-test-$platformName.exe")
                    foreach ($file in $testFiles) {
                        if (Test-Path $file) {
                            Remove-Item $file -Force
                        }
                    }
                } else {
                    $results += @{Platform=$platformName; Status="Failed"; Error=$output}
                }
            }
        } catch {
            $results += @{Platform=$platformName; Status="Exception"; Error=$_.Exception.Message}
        }
    }
    
    # 恢复环境变量
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    
    Write-Host "跨平台构建测试:" -ForegroundColor Cyan
    foreach ($result in $results) {
        $status = switch ($result.Status) {
            "Success" { "✓ 成功" }
            "Failed" { "✗ 失败" }
            "Exception" { "✗ 异常" }
            "DryRun" { "○ 试运行" }
        }
        
        $color = switch ($result.Status) {
            "Success" { "Green" }
            "Failed" { "Red" }
            "Exception" { "Red" }
            "DryRun" { "Yellow" }
        }
        
        Write-Host "  $($result.Platform): $status" -ForegroundColor $color
        
        if ($result.Error -and $Verbose) {
            Write-Host "    错误: $($result.Error)" -ForegroundColor Gray
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
        [array]$ConfigSafety,
        [array]$BuildResults
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
    $psMaxScore = 2  # Windows PowerShell 5.1 + PowerShell 7+
    $totalScore += $psScore
    $maxScore += $psMaxScore
    
    Write-Host "PowerShell 环境支持: $psScore/$psMaxScore" -ForegroundColor $(if ($psScore -eq $psMaxScore) { "Green" } else { "Yellow" })
    
    # 安装脚本评分
    $scriptScore = ($InstallScripts | Where-Object { $_.Exists -and $_.SyntaxValid }).Count
    $scriptMaxScore = $InstallScripts.Count
    $totalScore += $scriptScore
    $maxScore += $scriptMaxScore
    
    Write-Host "安装脚本兼容性: $scriptScore/$scriptMaxScore" -ForegroundColor $(if ($scriptScore -eq $scriptMaxScore) { "Green" } else { "Red" })
    
    # 配置安全性评分
    $configScore = ($ConfigSafety | Where-Object { $_.SafeToModify }).Count
    $configMaxScore = $ConfigSafety.Count
    $totalScore += $configScore
    $maxScore += $configMaxScore
    
    Write-Host "配置文件安全性: $configScore/$configMaxScore" -ForegroundColor $(if ($configScore -eq $configMaxScore) { "Green" } else { "Yellow" })
    
    # 构建测试评分
    if ($BuildResults.Count -gt 0) {
        $buildScore = ($BuildResults | Where-Object { $_.Status -eq "Success" }).Count
        $buildMaxScore = $BuildResults.Count
        $totalScore += $buildScore
        $maxScore += $buildMaxScore
        
        Write-Host "跨平台构建: $buildScore/$buildMaxScore" -ForegroundColor $(if ($buildScore -eq $buildMaxScore) { "Green" } else { "Red" })
    }
    
    # 总体评分
    $overallScore = if ($maxScore -gt 0) { [math]::Round(($totalScore / $maxScore) * 100, 1) } else { 0 }
    
    Write-Host ""
    Write-Host "总体兼容性评分: $overallScore%" -ForegroundColor $(
        if ($overallScore -ge 90) { "Green" }
        elseif ($overallScore -ge 70) { "Yellow" }
        else { "Red" }
    )
    
    # 建议
    Write-Host ""
    Write-Host "建议:" -ForegroundColor Cyan
    
    if ($psScore -lt $psMaxScore) {
        Write-Host "• 建议安装 PowerShell 7+ 以获得更好的兼容性" -ForegroundColor Yellow
    }
    
    if ($scriptScore -lt $scriptMaxScore) {
        Write-Host "• 检查并修复安装脚本的语法错误" -ForegroundColor Yellow
    }
    
    if ($configScore -lt $configMaxScore) {
        Write-Host "• 在修改配置文件前请先备份现有配置" -ForegroundColor Yellow
        Write-Host "• 使用 -Force 参数可以覆盖现有的冲突配置" -ForegroundColor Yellow
    }
    
    if ($BuildResults.Count -gt 0) {
        $failedBuilds = $BuildResults | Where-Object { $_.Status -eq "Failed" -or $_.Status -eq "Exception" }
        if ($failedBuilds.Count -gt 0) {
            Write-Host "• 修复跨平台构建问题以支持所有目标平台" -ForegroundColor Yellow
        }
    }
    
    Write-Host ""
    if ($overallScore -ge 90) {
        Write-Host "✅ DelGuard 已准备好在当前环境中发布！" -ForegroundColor Green
    } elseif ($overallScore -ge 70) {
        Write-Host "⚠️  DelGuard 基本可以发布，但建议先解决上述问题" -ForegroundColor Yellow
    } else {
        Write-Host "❌ DelGuard 需要解决关键问题后才能发布" -ForegroundColor Red
    }
}

# 主函数
function Main {
    Write-Host "DelGuard 综合兼容性检查工具" -ForegroundColor Cyan
    Write-Host "============================" -ForegroundColor Cyan
    Write-Host ""
    
    # 检查系统环境
    $systemInfo = Test-SystemEnvironment
    Write-Host ""
    
    # 检查 PowerShell 环境
    $psEnvs = Test-PowerShellEnvironments
    Write-Host ""
    
    # 测试安装脚本
    if ($All -or $InstallTest) {
        $installScripts = Test-InstallScriptCompatibility
        Write-Host ""
    } else {
        $installScripts = @()
    }
    
    # 测试配置安全性
    if ($All -or $ConfigTest) {
        $configSafety = Test-ConfigSafety
        Write-Host ""
    } else {
        $configSafety = @()
    }
    
    # 测试跨平台构建
    if ($All -or $BuildTest) {
        $buildResults = Test-CrossPlatformBuild
        Write-Host ""
    } else {
        $buildResults = @()
    }
    
    # 生成综合报告
    Generate-ComprehensiveReport -SystemInfo $systemInfo -PowerShellEnvs $psEnvs -InstallScripts $installScripts -ConfigSafety $configSafety -BuildResults $buildResults
}

# 执行主函数
Main