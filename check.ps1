# DelGuard 项目状态检查脚本

param(
    [switch]$Verbose = $false,
    [switch]$Fix = $false
)

$ErrorActionPreference = 'Continue'

Write-Host "DelGuard 项目状态检查" -ForegroundColor Cyan
Write-Host "=====================" -ForegroundColor Cyan
Write-Host ""

$issues = @()
$warnings = @()

# 检查 Go 环境
Write-Host "检查 Go 环境..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✓ Go 版本: $goVersion" -ForegroundColor Green
} catch {
    $issues += "Go 未安装或不在 PATH 中"
    Write-Host "✗ Go 未找到" -ForegroundColor Red
}

# 检查项目文件
Write-Host "`n检查项目文件..." -ForegroundColor Yellow
$requiredFiles = @(
    "main.go",
    "config.go", 
    "core_delete.go",
    "go.mod",
    "go.sum",
    "README.md",
    "LICENSE",
    "install.ps1",
    "install.sh",
    "build.ps1",
    "build.sh"
)

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "✓ $file" -ForegroundColor Green
    } else {
        $issues += "缺少文件: $file"
        Write-Host "✗ $file" -ForegroundColor Red
    }
}

# 检查目录结构
Write-Host "`n检查目录结构..." -ForegroundColor Yellow
$requiredDirs = @(
    ".github/workflows",
    "docs",
    "config/languages"
)

foreach ($dir in $requiredDirs) {
    if (Test-Path $dir -PathType Container) {
        Write-Host "✓ $dir/" -ForegroundColor Green
    } else {
        $warnings += "建议创建目录: $dir"
        Write-Host "⚠ $dir/" -ForegroundColor Yellow
    }
}

# 检查 Go 模块
Write-Host "`n检查 Go 模块..." -ForegroundColor Yellow
try {
    $modCheck = go mod verify
    Write-Host "✓ Go 模块验证通过" -ForegroundColor Green
} catch {
    $issues += "Go 模块验证失败"
    Write-Host "✗ Go 模块验证失败" -ForegroundColor Red
}

# 运行测试
Write-Host "`n运行测试..." -ForegroundColor Yellow
try {
    $testResult = go test -v ./... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 所有测试通过" -ForegroundColor Green
        if ($Verbose) {
            Write-Host $testResult -ForegroundColor Gray
        }
    } else {
        $issues += "测试失败"
        Write-Host "✗ 测试失败" -ForegroundColor Red
        Write-Host $testResult -ForegroundColor Red
    }
} catch {
    $issues += "无法运行测试"
    Write-Host "✗ 无法运行测试" -ForegroundColor Red
}

# 检查构建
Write-Host "`n检查构建..." -ForegroundColor Yellow
try {
    $buildResult = go build -o delguard-test.exe . 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 构建成功" -ForegroundColor Green
        Remove-Item delguard-test.exe -ErrorAction SilentlyContinue
    } else {
        $issues += "构建失败"
        Write-Host "✗ 构建失败" -ForegroundColor Red
        Write-Host $buildResult -ForegroundColor Red
    }
} catch {
    $issues += "无法构建项目"
    Write-Host "✗ 无法构建项目" -ForegroundColor Red
}

# 检查代码质量
Write-Host "`n检查代码质量..." -ForegroundColor Yellow
try {
    $vetResult = go vet ./... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ go vet 检查通过" -ForegroundColor Green
    } else {
        $warnings += "go vet 发现问题"
        Write-Host "⚠ go vet 发现问题" -ForegroundColor Yellow
        if ($Verbose) {
            Write-Host $vetResult -ForegroundColor Yellow
        }
    }
} catch {
    $warnings += "无法运行 go vet"
    Write-Host "⚠ 无法运行 go vet" -ForegroundColor Yellow
}

# 检查安装脚本
Write-Host "`n检查安装脚本..." -ForegroundColor Yellow
if (Test-Path "install.ps1") {
    try {
        $syntax = powershell -NoProfile -Command "& { . .\install.ps1 -WhatIf }" 2>&1
        Write-Host "✓ PowerShell 安装脚本语法正确" -ForegroundColor Green
    } catch {
        $warnings += "PowerShell 安装脚本可能有语法问题"
        Write-Host "⚠ PowerShell 安装脚本语法检查失败" -ForegroundColor Yellow
    }
}

# 生成报告
Write-Host "`n" + "="*50 -ForegroundColor Cyan
Write-Host "检查报告" -ForegroundColor Cyan
Write-Host "="*50 -ForegroundColor Cyan

if ($issues.Count -eq 0) {
    Write-Host "✓ 项目状态良好，没有发现严重问题！" -ForegroundColor Green
} else {
    Write-Host "✗ 发现 $($issues.Count) 个问题需要修复：" -ForegroundColor Red
    foreach ($issue in $issues) {
        Write-Host "  - $issue" -ForegroundColor Red
    }
}

if ($warnings.Count -gt 0) {
    Write-Host "`n⚠ 发现 $($warnings.Count) 个警告：" -ForegroundColor Yellow
    foreach ($warning in $warnings) {
        Write-Host "  - $warning" -ForegroundColor Yellow
    }
}

# 修复建议
if ($Fix -and ($issues.Count -gt 0 -or $warnings.Count -gt 0)) {
    Write-Host "`n修复建议：" -ForegroundColor Cyan
    
    if ($issues -contains "Go 未安装或不在 PATH 中") {
        Write-Host "1. 安装 Go: https://golang.org/dl/" -ForegroundColor Gray
    }
    
    if ($warnings -contains "建议创建目录: .github/workflows") {
        Write-Host "2. 创建 GitHub Actions 目录: mkdir -p .github/workflows" -ForegroundColor Gray
    }
    
    if ($warnings -contains "建议创建目录: docs") {
        Write-Host "3. 创建文档目录: mkdir docs" -ForegroundColor Gray
    }
    
    if ($warnings -contains "建议创建目录: config/languages") {
        Write-Host "4. 创建语言包目录: mkdir -p config/languages" -ForegroundColor Gray
    }
}

Write-Host "`n项目准备状态：" -ForegroundColor Cyan
if ($issues.Count -eq 0) {
    Write-Host "🚀 项目已准备好发布！" -ForegroundColor Green
} else {
    Write-Host "🔧 需要修复问题后才能发布" -ForegroundColor Red
}

# 返回适当的退出码
if ($issues.Count -gt 0) {
    exit 1
} else {
    exit 0
}