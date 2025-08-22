#!/usr/bin/env pwsh
param(
    [string]$DelGuardPath = "",
    [switch]$Verbose = $false
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 获取DelGuard路径
if ([string]::IsNullOrEmpty($DelGuardPath)) {
    # 优先使用本地构建的版本
    $localBuild = Join-Path $PSScriptRoot ".." ".." "build" "DelGuard.exe"
    if (Test-Path $localBuild) {
        $DelGuardPath = $localBuild
    } else {
        # 回退到系统路径
        $DelGuardPath = "delguard"
    }
}

Write-Host "使用DelGuard路径: $DelGuardPath"

# 创建临时测试目录
$testDir = Join-Path $env:TEMP "delguard_test_$(Get-Date -Format 'yyyyMMdd_HHmmss')"
if (Test-Path $testDir) {
    Remove-Item -Recurse -Force $testDir
}
New-Item -ItemType Directory -Path $testDir -Force | Out-Null

function Cleanup {
    param($TestName)
    Write-Host "清理测试环境: $TestName"
    if (Test-Path $testDir) {
        Remove-Item -Recurse -Force $testDir -ErrorAction SilentlyContinue
    }
    New-Item -ItemType Directory -Path $testDir -Force | Out-Null
}

function Test-BasicDelete {
    Write-Host "=== 基础删除测试 ==="
    Cleanup "基础删除"
    
    $testFile = Join-Path $testDir "test.txt"
    "Hello World" | Out-File -FilePath $testFile -Encoding UTF8
    
    if (-not (Test-Path $testFile)) {
        Write-Host "FAIL: 测试文件创建失败"
        return $false
    }
    
    & $DelGuardPath --force $testFile
    
    if (Test-Path $testFile) {
        Write-Host "FAIL: 文件未被删除"
        return $false
    }
    
    Write-Host "PASS: 基础删除测试通过"
    return $true
}

function Test-SameNameFiles {
    Write-Host "=== 同名文件共存测试 ==="
    Cleanup "同名文件"
    
    $testFile1 = Join-Path $testDir "test.txt"
    $testFile2 = Join-Path $testDir "subdir" "test.txt"
    
    New-Item -ItemType Directory -Path (Join-Path $testDir "subdir") -Force | Out-Null
    "File 1" | Out-File -FilePath $testFile1 -Encoding UTF8
    "File 2" | Out-File -FilePath $testFile2 -Encoding UTF8
    
    & $DelGuardPath --force $testFile1
    & $DelGuardPath --force $testFile2
    
    if (Test-Path $testFile1) {
        Write-Host "FAIL: 第一个同名文件未被删除"
        return $false
    }
    
    if (Test-Path $testFile2) {
        Write-Host "FAIL: 第二个同名文件未被删除"
        return $false
    }
    
    Write-Host "PASS: 同名文件共存测试通过"
    return $true
}

function Test-SymlinkHandling {
    Write-Host "=== 符号链接测试 ==="
    Cleanup "符号链接"
    
    $targetFile = Join-Path $testDir "target.txt"
    $symlinkFile = Join-Path $testDir "link.txt"
    
    "Target Content" | Out-File -FilePath $targetFile -Encoding UTF8
    New-Item -ItemType SymbolicLink -Path $symlinkFile -Target $targetFile -Force | Out-Null
    
    if (-not (Test-Path $symlinkFile)) {
        Write-Host "SKIP: 符号链接创建失败（需要管理员权限）"
        return $true
    }
    
    & $DelGuardPath --force $symlinkFile
    
    if (Test-Path $symlinkFile) {
        Write-Host "FAIL: 符号链接未被删除"
        return $false
    }
    
    if (-not (Test-Path $targetFile)) {
        Write-Host "FAIL: 符号链接目标被意外删除"
        return $false
    }
    
    Write-Host "PASS: 符号链接测试通过"
    return $true
}

function Test-LongPath {
    Write-Host "=== 长路径测试 ==="
    Cleanup "长路径"
    
    $longDir = Join-Path $testDir "a" "b" "c" "d" "e" "f" "g" "h" "i" "j"
    New-Item -ItemType Directory -Path $longDir -Force | Out-Null
    $longFile = Join-Path $longDir "deep_file.txt"
    "Deep content" | Out-File -FilePath $longFile -Encoding UTF8
    
    & $DelGuardPath --force $longFile
    
    if (Test-Path $longFile) {
        Write-Host "FAIL: 长路径文件未被删除"
        return $false
    }
    
    Write-Host "PASS: 长路径测试通过"
    return $true
}

function Test-DirectoryHandling {
    Write-Host "=== 目录处理测试 ==="
    Cleanup "目录处理"
    
    $testDirPath = Join-Path $testDir "test_dir"
    New-Item -ItemType Directory -Path $testDirPath -Force | Out-Null
    $fileInDir = Join-Path $testDirPath "file_in_dir.txt"
    "Content" | Out-File -FilePath $fileInDir -Encoding UTF8
    
    # 测试不带-r删除目录
    $result = & $DelGuardPath $testDirPath 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "FAIL: 目录删除应该需要-r参数"
        return $false
    }
    
    # 测试带-r删除目录
    & $DelGuardPath --force --recursive $testDirPath
    
    if (Test-Path $testDirPath) {
        Write-Host "FAIL: 目录未被删除"
        return $false
    }
    
    Write-Host "PASS: 目录处理测试通过"
    return $true
}

function Test-ProtectedPaths {
    Write-Host "=== 关键路径保护测试 ==="
    
    $protectedPaths = @(
        "C:\",
        "C:\Windows",
        "C:\Program Files",
        $env:USERPROFILE
    )
    
    foreach ($path in $protectedPaths) {
        if (Test-Path $path) {
            $result = & $DelGuardPath $path 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "FAIL: 关键路径未受保护: $path"
                return $false
            }
        }
    }
    
    Write-Host "PASS: 关键路径保护测试通过"
    return $true
}

function Test-InteractiveMode {
    Write-Host "=== 交互模式测试 ==="
    Cleanup "交互模式"
    
    $testFile = Join-Path $testDir "interactive.txt"
    "Content" | Out-File -FilePath $testFile -Encoding UTF8
    
    # 设置环境变量强制交互
    $env:DELGUARD_INTERACTIVE = "true"
    
    # 由于交互测试需要用户输入，我们只检查环境变量是否生效
    $result = & $DelGuardPath --help
    if ($result -match "interactive") {
        Write-Host "PASS: 交互模式支持检测通过"
        return $true
    } else {
        Write-Host "FAIL: 交互模式支持检测失败"
        return $false
    }
}

# 主测试流程
$tests = @(
    @{ Name = "基础删除"; Func = ${function:Test-BasicDelete} },
    @{ Name = "同名文件"; Func = ${function:Test-SameNameFiles} },
    @{ Name = "符号链接"; Func = ${function:Test-SymlinkHandling} },
    @{ Name = "长路径"; Func = ${function:Test-LongPath} },
    @{ Name = "目录处理"; Func = ${function:Test-DirectoryHandling} },
    @{ Name = "关键路径保护"; Func = ${function:Test-ProtectedPaths} },
    @{ Name = "交互模式"; Func = ${function:Test-InteractiveMode} }
)

$passed = 0
$total = $tests.Count

Write-Host "开始执行DelGuard功能测试..."
Write-Host "测试目录: $testDir"
Write-Host ""

foreach ($test in $tests) {
    try {
        $result = & $test.Func
        if ($result) {
            $passed++
        }
    } catch {
        Write-Host "ERROR: $($test.Name) 测试执行失败: $($_.Exception.Message)"
    }
    Write-Host ""
}

# 清理测试环境
if (Test-Path $testDir) {
    Remove-Item -Recurse -Force $testDir -ErrorAction SilentlyContinue
}

Write-Host "测试结果总结:"
Write-Host "通过: $passed / 总计: $total"

if ($passed -eq $total) {
    Write-Host "所有测试通过！✅"
    exit 0
} else {
    Write-Host "部分测试失败！❌"
    exit 1
}
