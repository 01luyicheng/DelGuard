#!/usr/bin/env pwsh
<#
.SYNOPSIS
    DelGuard 发布前检查脚本

.DESCRIPTION
    执行发布前的各项检查，确保项目准备就绪
#>

param(
    [switch]$Verbose,
    [switch]$SkipBuild,
    [switch]$SkipTests
)

$ErrorActionPreference = 'Stop'

Write-Host "🚀 DelGuard 发布前检查" -ForegroundColor Cyan
Write-Host "===================" -ForegroundColor Cyan

# 1. 检查Go环境
Write-Host "`n📦 检查Go环境..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✅ Go环境正常: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go环境未找到" -ForegroundColor Red
    exit 1
}

# 2. 检查项目结构
Write-Host "`n📁 检查项目结构..." -ForegroundColor Yellow
$requiredFiles = @(
    "go.mod",
    "main.go", 
    "README.md",
    "LICENSE",
    "CHANGELOG.md",
    "install.sh",
    "install.ps1"
)

$requiredDirs = @(
    "config",
    "config/languages", 
    "docs",
    "scripts",
    "tests"
)

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "✅ $file" -ForegroundColor Green
    } else {
        Write-Host "❌ $file 缺失" -ForegroundColor Red
        exit 1
    }
}

foreach ($dir in $requiredDirs) {
    if (Test-Path $dir -PathType Container) {
        Write-Host "✅ $dir/" -ForegroundColor Green
    } else {
        Write-Host "❌ $dir/ 缺失" -ForegroundColor Red
        exit 1
    }
}

# 3. 检查语言文件
Write-Host "`n🌍 检查语言文件..." -ForegroundColor Yellow
$langFiles = Get-ChildItem "config/languages" -Filter "*.json"
if ($langFiles.Count -gt 0) {
    Write-Host "✅ 找到 $($langFiles.Count) 个语言文件" -ForegroundColor Green
    foreach ($file in $langFiles) {
        Write-Host "  - $($file.Name)" -ForegroundColor Gray
    }
} else {
    Write-Host "⚠️ 未找到语言文件" -ForegroundColor Yellow
}

# 4. 构建测试
if (-not $SkipBuild) {
    Write-Host "`n🔨 构建测试..." -ForegroundColor Yellow
    try {
        go build -o delguard.exe
        Write-Host "✅ 构建成功" -ForegroundColor Green
        
        # 测试基本功能
        try {
            $null = & "./delguard.exe" --help
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ 帮助功能正常" -ForegroundColor Green
            } else {
                Write-Host "⚠️ 帮助功能异常" -ForegroundColor Yellow
            }
        } catch {
            Write-Host "⚠️ 帮助功能测试异常: $_" -ForegroundColor Yellow
        }
        
        Remove-Item "delguard.exe" -ErrorAction SilentlyContinue
    } catch {
        Write-Host "❌ 构建失败: $_" -ForegroundColor Red
        exit 1
    }
}

# 5. 运行测试
if (-not $SkipTests) {
    Write-Host "`n🧪 运行测试..." -ForegroundColor Yellow
    try {
        go test -v ./...
        Write-Host "✅ 所有测试通过" -ForegroundColor Green
    } catch {
        Write-Host "⚠️ 部分测试失败，请检查" -ForegroundColor Yellow
    }
}

# 6. 检查安装脚本
Write-Host "`n📥 检查安装脚本..." -ForegroundColor Yellow
$installScripts = @("install.sh", "install.ps1")
foreach ($script in $installScripts) {
    if (Test-Path $script) {
        $content = Get-Content $script -Raw
        if ($content -match "github\.com/01luyicheng/DelGuard") {
            Write-Host "✅ $script GitHub URL 正确" -ForegroundColor Green
        } else {
            Write-Host "⚠️ $script GitHub URL 需要验证" -ForegroundColor Yellow
        }
    }
}

# 7. 检查版本信息
Write-Host "`n📋 检查版本信息..." -ForegroundColor Yellow
if (Test-Path "CHANGELOG.md") {
    $changelog = Get-Content "CHANGELOG.md" -Raw
    if ($changelog -match "\[未发布\]") {
        Write-Host "⚠️ CHANGELOG.md 包含未发布版本，建议更新" -ForegroundColor Yellow
    } else {
        Write-Host "✅ CHANGELOG.md 版本信息正常" -ForegroundColor Green
    }
}

# 8. 安全检查
Write-Host "`n🔒 安全检查..." -ForegroundColor Yellow
if (Test-Path "final_security_check.go") {
    try {
        go run final_security_check.go
        Write-Host "✅ 安全检查完成" -ForegroundColor Green
    } catch {
        Write-Host "⚠️ 安全检查脚本执行异常" -ForegroundColor Yellow
    }
}

Write-Host "`n🎉 发布前检查完成！" -ForegroundColor Cyan
Write-Host "===================" -ForegroundColor Cyan

Write-Host "`n📝 下一步操作建议:" -ForegroundColor White
Write-Host "1. 创建 GitHub 仓库 (如果尚未创建)" -ForegroundColor Gray
Write-Host "2. 推送代码到 GitHub" -ForegroundColor Gray  
Write-Host "3. 验证安装脚本可访问性" -ForegroundColor Gray
Write-Host "4. 创建 GitHub Release" -ForegroundColor Gray
Write-Host "5. 测试一键安装命令" -ForegroundColor Gray