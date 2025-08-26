#!/usr/bin/env pwsh
# DelGuard 功能测试脚本

Write-Host "=== DelGuard 功能测试 ===" -ForegroundColor Cyan

# 测试1: 检查可执行文件
Write-Host "`n1. 检查 DelGuard 可执行文件..." -ForegroundColor Yellow
if (Test-Path "delguard.exe") {
    Write-Host "✓ delguard.exe 存在" -ForegroundColor Green
    
    # 测试版本信息
    try {
        $version = & .\delguard.exe --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ 版本信息: $version" -ForegroundColor Green
        } else {
            Write-Host "✗ 无法获取版本信息" -ForegroundColor Red
        }
    } catch {
        Write-Host "✗ 版本检查失败: $($_.Exception.Message)" -ForegroundColor Red
    }
} else {
    Write-Host "✗ delguard.exe 不存在" -ForegroundColor Red
}

# 测试2: 检查PowerShell配置文件
Write-Host "`n2. 检查 PowerShell 配置文件..." -ForegroundColor Yellow
$profilePath = "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1"
if (Test-Path $profilePath) {
    Write-Host "✓ PowerShell 配置文件存在: $profilePath" -ForegroundColor Green
    
    # 检查语法
    try {
        powershell -NoProfile -File $profilePath -ErrorAction Stop
        Write-Host "✓ PowerShell 配置文件语法正确" -ForegroundColor Green
    } catch {
        Write-Host "✗ PowerShell 配置文件语法错误: $($_.Exception.Message)" -ForegroundColor Red
    }
} else {
    Write-Host "✗ PowerShell 配置文件不存在" -ForegroundColor Red
}

# 测试3: 检查别名功能
Write-Host "`n3. 检查别名功能..." -ForegroundColor Yellow

$aliases = @("del", "rm", "cp", "delguard")
foreach ($alias in $aliases) {
    try {
        # 在新的PowerShell会话中测试别名
        $result = powershell -Command "& { . '$profilePath'; Get-Command $alias -ErrorAction Stop; Write-Output 'OK' }" 2>$null
        if ($result -contains "OK") {
            Write-Host "✓ 别名 '$alias' 可用" -ForegroundColor Green
        } else {
            Write-Host "✗ 别名 '$alias' 不可用" -ForegroundColor Red
        }
    } catch {
        Write-Host "✗ 别名 '$alias' 测试失败" -ForegroundColor Red
    }
}

# 测试4: 检查安装脚本
Write-Host "`n4. 检查安装脚本..." -ForegroundColor Yellow
$installScripts = @(
    "scripts\install.ps1",
    "scripts\install.sh", 
    "scripts\universal_install.ps1",
    "install.bat"
)

foreach ($script in $installScripts) {
    if (Test-Path $script) {
        Write-Host "✓ 安装脚本存在: $script" -ForegroundColor Green
    } else {
        Write-Host "✗ 安装脚本缺失: $script" -ForegroundColor Red
    }
}

# 测试5: 创建测试文件并测试删除功能
Write-Host "`n5. 测试删除功能..." -ForegroundColor Yellow
$testFile = "test_delguard_temp.txt"
try {
    # 创建测试文件
    "This is a test file for DelGuard" | Out-File -FilePath $testFile -Encoding UTF8
    if (Test-Path $testFile) {
        Write-Host "✓ 测试文件创建成功: $testFile" -ForegroundColor Green
        
        # 测试删除功能（dry-run模式）
        $result = & .\delguard.exe --dry-run $testFile 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ 删除功能测试通过（dry-run模式）" -ForegroundColor Green
        } else {
            Write-Host "✗ 删除功能测试失败" -ForegroundColor Red
        }
        
        # 清理测试文件
        Remove-Item $testFile -Force -ErrorAction SilentlyContinue
    }
} catch {
    Write-Host "✗ 删除功能测试失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Cyan
Write-Host "如果所有测试都通过，DelGuard 已正确安装并可以使用。" -ForegroundColor Green
Write-Host "重启 PowerShell 后可以使用 del, rm, cp, delguard 命令。" -ForegroundColor Yellow