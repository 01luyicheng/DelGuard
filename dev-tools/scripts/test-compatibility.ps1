# DelGuard 兼容性测试脚本
# 测试发布脚本和安装脚本在不同环境下的兼容性

param(
    [switch]$TestInstall = $false,
    [switch]$TestBuild = $false,
    [switch]$TestRelease = $false,
    [switch]$All = $false
)

$ErrorActionPreference = 'Continue'

Write-Host "DelGuard 兼容性测试" -ForegroundColor Cyan
Write-Host "===================" -ForegroundColor Cyan
Write-Host ""

# 测试系统环境
function Test-SystemEnvironment {
    Write-Host "系统环境检查:" -ForegroundColor Yellow
    
    # 操作系统信息
    Write-Host "  操作系统: $($PSVersionTable.OS)" -ForegroundColor Gray
    Write-Host "  PowerShell版本: $($PSVersionTable.PSVersion)" -ForegroundColor Gray
    Write-Host "  架构: $env:PROCESSOR_ARCHITECTURE" -ForegroundColor Gray
    
    # Go 环境
    try {
        $goVersion = go version
        Write-Host "  ✓ Go: $goVersion" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ Go 未安装或不可用" -ForegroundColor Red
        return $false
    }
    
    # Git 环境
    try {
        $gitVersion = git --version
        Write-Host "  ✓ Git: $gitVersion" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ Git 未安装或不可用" -ForegroundColor Red
        return $false
    }
    
    return $true
}

# 测试 CGO 和竞态检测支持
function Test-RaceDetection {
    Write-Host "`n竞态检测支持测试:" -ForegroundColor Yellow
    
    # 测试 CGO_ENABLED=0 (默认构建模式)
    $env:CGO_ENABLED = "0"
    try {
        go test -race -run=NonExistentTest ./... 2>$null
        Write-Host "  ✗ CGO_ENABLED=0 时不支持竞态检测 (预期行为)" -ForegroundColor Yellow
    } catch {
        Write-Host "  ✓ CGO_ENABLED=0 时正确拒绝竞态检测" -ForegroundColor Green
    }
    
    # 测试 CGO_ENABLED=1
    $env:CGO_ENABLED = "1"
    try {
        $output = go test -race -run=NonExistentTest ./... 2>&1
        if ($LASTEXITCODE -eq 0 -or $output -notlike "*build constraints exclude all Go files*") {
            Write-Host "  ✓ CGO_ENABLED=1 时支持竞态检测" -ForegroundColor Green
            $raceSupported = $true
        } else {
            Write-Host "  ✗ CGO_ENABLED=1 时仍不支持竞态检测" -ForegroundColor Red
            $raceSupported = $false
        }
    } catch {
        Write-Host "  ✗ 竞态检测测试失败" -ForegroundColor Red
        $raceSupported = $false
    }
    
    # 重置环境变量
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    
    return $raceSupported
}

# 测试构建脚本
function Test-BuildScript {
    Write-Host "`n构建脚本测试:" -ForegroundColor Yellow
    
    try {
        # 测试构建脚本语法
        $null = powershell -Command "& { . .\build.ps1; exit 0 }" -ErrorAction Stop
        Write-Host "  ✓ build.ps1 语法正确" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ build.ps1 语法错误: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    try {
        # 测试基本构建 (不实际构建，只检查参数)
        .\build.ps1 -Version "test" -WhatIf 2>$null
        Write-Host "  ✓ build.ps1 参数处理正常" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ build.ps1 参数处理失败" -ForegroundColor Red
        return $false
    }
    
    return $true
}

# 测试发布脚本
function Test-ReleaseScript {
    Write-Host "`n发布脚本测试:" -ForegroundColor Yellow
    
    try {
        # 测试发布脚本语法
        $null = powershell -Command "& { . .\release.ps1; exit 0 }" -ErrorAction Stop
        Write-Host "  ✓ release.ps1 语法正确" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ release.ps1 语法错误: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    try {
        # 测试试运行模式
        .\release.ps1 -Version "v0.0.0-test" -DryRun -Force
        Write-Host "  ✓ release.ps1 试运行模式正常" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ release.ps1 试运行失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    return $true
}

# 测试安装脚本
function Test-InstallScript {
    Write-Host "`n安装脚本测试:" -ForegroundColor Yellow
    
    try {
        # 测试 PowerShell 安装脚本语法
        $null = powershell -Command "& { . .\install.ps1; exit 0 }" -ErrorAction Stop
        Write-Host "  ✓ install.ps1 语法正确" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ install.ps1 语法错误: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    try {
        # 测试状态检查功能
        .\install.ps1 -Status
        Write-Host "  ✓ install.ps1 状态检查正常" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ install.ps1 状态检查失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    # 检查 Unix 安装脚本 (如果在 WSL 或有 bash)
    if (Get-Command bash -ErrorAction SilentlyContinue) {
        try {
            bash -n install.sh
            Write-Host "  ✓ install.sh 语法正确" -ForegroundColor Green
        } catch {
            Write-Host "  ✗ install.sh 语法错误" -ForegroundColor Red
            return $false
        }
    } else {
        Write-Host "  ⚠ 无法测试 install.sh (bash 不可用)" -ForegroundColor Yellow
    }
    
    return $true
}

# 测试网络连接
function Test-NetworkConnectivity {
    Write-Host "`n网络连接测试:" -ForegroundColor Yellow
    
    try {
        $response = Invoke-WebRequest -Uri "https://api.github.com" -Method Head -TimeoutSec 10
        if ($response.StatusCode -eq 200) {
            Write-Host "  ✓ GitHub API 连接正常" -ForegroundColor Green
        } else {
            Write-Host "  ✗ GitHub API 连接异常: $($response.StatusCode)" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "  ✗ GitHub API 连接失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/01luyicheng/DelGuard/releases/latest" -TimeoutSec 10
        Write-Host "  ✓ 版本信息获取正常" -ForegroundColor Green
    } catch {
        if ($_.Exception.Message -like "*404*") {
            Write-Host "  ⚠ 仓库还没有发布版本 (正常情况)" -ForegroundColor Yellow
        } else {
            Write-Host "  ✗ 版本信息获取失败: $($_.Exception.Message)" -ForegroundColor Red
            return $false
        }
    }
    
    return $true
}

# 测试执行策略
function Test-ExecutionPolicy {
    Write-Host "`n执行策略检查:" -ForegroundColor Yellow
    
    $policy = Get-ExecutionPolicy
    Write-Host "  当前执行策略: $policy" -ForegroundColor Gray
    
    switch ($policy) {
        "Restricted" {
            Write-Host "  ✗ 执行策略过于严格，可能无法运行脚本" -ForegroundColor Red
            Write-Host "    建议运行: Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser" -ForegroundColor Yellow
            return $false
        }
        "AllSigned" {
            Write-Host "  ⚠ 需要签名脚本，可能影响安装" -ForegroundColor Yellow
            return $true
        }
        "RemoteSigned" {
            Write-Host "  ✓ 执行策略适合" -ForegroundColor Green
            return $true
        }
        "Unrestricted" {
            Write-Host "  ✓ 执行策略允许所有脚本" -ForegroundColor Green
            return $true
        }
        "Bypass" {
            Write-Host "  ✓ 执行策略已绕过" -ForegroundColor Green
            return $true
        }
        default {
            Write-Host "  ⚠ 未知执行策略: $policy" -ForegroundColor Yellow
            return $true
        }
    }
}

# 主测试函数
function Run-CompatibilityTests {
    $results = @{
        SystemEnvironment = $false
        RaceDetection = $false
        ExecutionPolicy = $false
        NetworkConnectivity = $false
        BuildScript = $false
        ReleaseScript = $false
        InstallScript = $false
    }
    
    # 基础环境测试
    $results.SystemEnvironment = Test-SystemEnvironment
    $results.RaceDetection = Test-RaceDetection
    $results.ExecutionPolicy = Test-ExecutionPolicy
    $results.NetworkConnectivity = Test-NetworkConnectivity
    
    # 脚本测试
    if ($TestBuild -or $All) {
        $results.BuildScript = Test-BuildScript
    }
    
    if ($TestRelease -or $All) {
        $results.ReleaseScript = Test-ReleaseScript
    }
    
    if ($TestInstall -or $All) {
        $results.InstallScript = Test-InstallScript
    }
    
    # 显示测试结果
    Write-Host "`n" + "="*50 -ForegroundColor Cyan
    Write-Host "兼容性测试结果" -ForegroundColor Cyan
    Write-Host "="*50 -ForegroundColor Cyan
    
    $passCount = 0
    $totalCount = 0
    
    foreach ($test in $results.GetEnumerator()) {
        if ($test.Key -eq "BuildScript" -and !($TestBuild -or $All)) { continue }
        if ($test.Key -eq "ReleaseScript" -and !($TestRelease -or $All)) { continue }
        if ($test.Key -eq "InstallScript" -and !($TestInstall -or $All)) { continue }
        
        $totalCount++
        if ($test.Value) {
            Write-Host "✓ $($test.Key)" -ForegroundColor Green
            $passCount++
        } else {
            Write-Host "✗ $($test.Key)" -ForegroundColor Red
        }
    }
    
    Write-Host "`n测试通过: $passCount/$totalCount" -ForegroundColor $(if ($passCount -eq $totalCount) { "Green" } else { "Yellow" })
    
    if ($passCount -eq $totalCount) {
        Write-Host "🎉 所有测试通过！系统兼容性良好。" -ForegroundColor Green
    } else {
        Write-Host "⚠ 部分测试失败，请检查上述问题。" -ForegroundColor Yellow
    }
    
    return $passCount -eq $totalCount
}

# 执行测试
if ($All) {
    $TestBuild = $true
    $TestRelease = $true
    $TestInstall = $true
}

$success = Run-CompatibilityTests

if (!$success) {
    exit 1
}