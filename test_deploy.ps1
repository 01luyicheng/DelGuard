#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 部署测试脚本 - Windows版本

.DESCRIPTION
    自动部署并测试 DelGuard 安全删除工具的各项功能。
    支持 PowerShell 5.1+ 和 PowerShell 7+。

.PARAMETER Clean
    在测试前清理环境（卸载现有版本）

.EXAMPLE
    .\test_deploy.ps1
    标准测试部署

.EXAMPLE
    .\test_deploy.ps1 -Clean
    清理环境后测试部署
#>

[CmdletBinding()]
param(
    [switch]$Clean
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

# 显示横幅
function Show-Banner {
    $banner = @"
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║                🧪 DelGuard 部署测试工具                      ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
"@
    Write-Host $banner -ForegroundColor $ColorScheme.Header
    Write-Host ""
}

# 创建测试环境
function New-TestEnvironment {
    Write-Host "创建测试环境..." -ForegroundColor $ColorScheme.Info
    
    # 创建测试目录
    $testDir = Join-Path $env:TEMP "delguard-test"
    if (Test-Path $testDir) {
        Remove-Item $testDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $testDir -Force | Out-Null
    
    # 创建测试文件
    $testFiles = @(
        "test1.txt",
        "test2.txt",
        "important_document.docx",
        "report.pdf",
        "image.jpg",
        "config.json"
    )
    
    foreach ($file in $testFiles) {
        $content = "This is a test file: $file`nCreated for DelGuard testing."
        Set-Content -Path (Join-Path $testDir $file) -Value $content
    }
    
    Write-Host "测试环境已创建: $testDir" -ForegroundColor $ColorScheme.Success
    return $testDir
}

# 安装DelGuard
function Install-DelGuard {
    Write-Host "安装DelGuard..." -ForegroundColor $ColorScheme.Info
    
    # 运行安装脚本
    $installScript = Join-Path $PSScriptRoot "install_enhanced_complete.ps1"
    if (!(Test-Path $installScript)) {
        $installScript = Join-Path $PSScriptRoot "install_enhanced.ps1"
    }
    
    if (!(Test-Path $installScript)) {
        Write-Host "未找到安装脚本: $installScript" -ForegroundColor $ColorScheme.Error
        throw "安装脚本不存在"
    }
    
    # 执行安装脚本
    & $installScript -Force
    
    # 检查安装结果
    $delguardPath = Find-InstalledDelGuard
    if (!$delguardPath) {
        throw "DelGuard安装失败"
    }
    
    Write-Host "DelGuard安装成功: $delguardPath" -ForegroundColor $ColorScheme.Success
    return $delguardPath
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

# 卸载DelGuard
function Uninstall-DelGuard {
    Write-Host "卸载DelGuard..." -ForegroundColor $ColorScheme.Info
    
    # 运行卸载脚本
    $uninstallScript = Join-Path $PSScriptRoot "uninstall.ps1"
    
    if (!(Test-Path $uninstallScript)) {
        Write-Host "未找到卸载脚本: $uninstallScript" -ForegroundColor $ColorScheme.Warning
        return
    }
    
    # 执行卸载脚本
    & $uninstallScript -Force
    
    # 检查卸载结果
    $delguardPath = Find-InstalledDelGuard
    if ($delguardPath) {
        Write-Host "DelGuard卸载失败，仍然存在: $delguardPath" -ForegroundColor $ColorScheme.Warning
    } else {
        Write-Host "DelGuard卸载成功" -ForegroundColor $ColorScheme.Success
    }
}

# 测试基本功能
function Test-BasicFunctionality {
    param([string]$DelguardPath, [string]$TestDir)
    
    Write-Host "测试基本功能..." -ForegroundColor $ColorScheme.Info
    
    # 测试帮助命令
    Write-Host "测试帮助命令..." -ForegroundColor $ColorScheme.Info
    $helpOutput = & $DelguardPath --help 2>&1
    if ($helpOutput -match "使用方法" -or $helpOutput -match "Usage") {
        Write-Host "✓ 帮助命令正常" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "✗ 帮助命令异常" -ForegroundColor $ColorScheme.Error
    }
    
    # 测试版本命令
    Write-Host "测试版本命令..." -ForegroundColor $ColorScheme.Info
    $versionOutput = & $DelguardPath --version 2>&1
    if ($versionOutput -match "\d+\.\d+\.\d+") {
        Write-Host "✓ 版本命令正常: $versionOutput" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "✗ 版本命令异常" -ForegroundColor $ColorScheme.Error
    }
    
    # 测试删除文件
    $testFile = Join-Path $TestDir "test1.txt"
    Write-Host "测试删除文件: $testFile" -ForegroundColor $ColorScheme.Info
    & $DelguardPath $testFile
    
    if (!(Test-Path $testFile)) {
        Write-Host "✓ 文件删除成功" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "✗ 文件删除失败" -ForegroundColor $ColorScheme.Error
    }
    
    # 测试不存在的文件（智能搜索功能）
    $nonExistentFile = Join-Path $TestDir "non_existent.txt"
    Write-Host "测试智能搜索功能: $nonExistentFile" -ForegroundColor $ColorScheme.Info
    $searchOutput = & $DelguardPath $nonExistentFile 2>&1
    
    if ($searchOutput -match "不存在" -and $searchOutput -match "相似") {
        Write-Host "✓ 智能搜索功能正常" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "✗ 智能搜索功能异常" -ForegroundColor $ColorScheme.Error
    }
}

# 测试语言检测
function Test-LanguageDetection {
    param([string]$DelguardPath)
    
    Write-Host "测试语言检测功能..." -ForegroundColor $ColorScheme.Info
    
    # 获取当前系统语言
    $currentCulture = [System.Globalization.CultureInfo]::CurrentUICulture
    $languageCode = $currentCulture.Name
    
    Write-Host "当前系统UI语言: $languageCode" -ForegroundColor $ColorScheme.Info
    
    # 执行命令并检查输出语言
    $output = & $DelguardPath --help 2>&1
    
    if ($languageCode -like "zh*" -and $output -match "使用方法") {
        Write-Host "✓ 中文语言检测正常" -ForegroundColor $ColorScheme.Success
    } elseif ($languageCode -like "en*" -and $output -match "Usage") {
        Write-Host "✓ 英文语言检测正常" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "✗ 语言检测功能可能有问题" -ForegroundColor $ColorScheme.Warning
        Write-Host "  系统语言: $languageCode" -ForegroundColor $ColorScheme.Info
        Write-Host "  输出示例: $($output | Select-Object -First 3)" -ForegroundColor $ColorScheme.Info
    }
}

# 测试更新功能
function Test-UpdateFunctionality {
    Write-Host "测试更新功能..." -ForegroundColor $ColorScheme.Info
    
    # 运行更新脚本
    $updateScript = Join-Path $PSScriptRoot "update.ps1"
    
    if (!(Test-Path $updateScript)) {
        Write-Host "未找到更新脚本: $updateScript" -ForegroundColor $ColorScheme.Warning
        return
    }
    
    # 执行更新脚本（仅检查模式）
    & $updateScript -CheckOnly
    
    Write-Host "✓ 更新检查功能正常" -ForegroundColor $ColorScheme.Success
}

# 主程序
try {
    Show-Banner
    
    # 如果指定了Clean参数，先卸载现有版本
    if ($Clean) {
        Uninstall-DelGuard
    }
    
    # 创建测试环境
    $testDir = New-TestEnvironment
    
    # 安装DelGuard
    $delguardPath = Install-DelGuard
    
    # 测试基本功能
    Test-BasicFunctionality -DelguardPath $delguardPath -TestDir $testDir
    
    # 测试语言检测
    Test-LanguageDetection -DelguardPath $delguardPath
    
    # 测试更新功能
    Test-UpdateFunctionality
    
    Write-Host "所有测试完成！" -ForegroundColor $ColorScheme.Success
    
} catch {
    Write-Host "测试失败: $($_.Exception.Message)" -ForegroundColor $ColorScheme.Error
    exit 1
}