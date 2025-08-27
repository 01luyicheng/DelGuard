#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 跨平台构建修复脚本
#>

[CmdletBinding()]
param(
    [switch]$Clean,
    [switch]$Verbose
)

$ErrorActionPreference = 'Stop'

function Write-LogInfo { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Cyan }
function Write-LogSuccess { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-LogWarning { param([string]$Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
function Write-LogError { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

Write-LogInfo "修复 DelGuard 跨平台构建问题..."

# 检查 Go 环境
try {
    $goVersion = go version
    Write-LogSuccess "Go 版本: $goVersion"
} catch {
    Write-LogError "Go 未安装，请先安装 Go"
    exit 1
}

# 清理构建缓存
if ($Clean) {
    Write-LogInfo "清理构建缓存..."
    go clean -cache
    try { go clean -modcache } catch { }
}

# 下载依赖
Write-LogInfo "下载依赖..."
go mod download
go mod tidy

# 设置构建环境
$env:CGO_ENABLED = "0"
$env:GO111MODULE = "on"

# 定义目标平台
$platforms = @(
    @{OS="windows"; Arch="amd64"},
    @{OS="windows"; Arch="386"},
    @{OS="linux"; Arch="amd64"},
    @{OS="linux"; Arch="386"},
    @{OS="linux"; Arch="arm64"},
    @{OS="darwin"; Arch="amd64"},
    @{OS="darwin"; Arch="arm64"}
)

# 创建构建目录
$buildDir = "build/cross-platform"
if (Test-Path $buildDir) {
    Remove-Item $buildDir -Recurse -Force
}
New-Item -ItemType Directory -Path $buildDir -Force | Out-Null

Write-LogInfo "开始跨平台构建..."

$successCount = 0
$totalCount = $platforms.Count

foreach ($platform in $platforms) {
    $os = $platform.OS
    $arch = $platform.Arch
    
    Write-Host "构建 $os/$arch..." -ForegroundColor Yellow
    
    $outputName = if ($os -eq "windows") { "delguard.exe" } else { "delguard" }
    $outputDir = "$buildDir/$os-$arch"
    $outputPath = "$outputDir/$outputName"
    
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    
    # 设置环境变量并构建
    $env:GOOS = $os
    $env:GOARCH = $arch
    
    try {
        $buildArgs = @(
            "build",
            "-ldflags=-s -w",
            "-o", $outputPath,
            "."
        )
        
        & go @buildArgs
        
        if (Test-Path $outputPath) {
            $fileSize = (Get-Item $outputPath).Length
            Write-LogSuccess "$os/$arch 构建成功 (大小: $fileSize bytes)"
            $successCount++
        } else {
            Write-LogError "$os/$arch 构建文件不存在"
        }
    } catch {
        Write-LogError "$os/$arch 构建失败: $($_.Exception.Message)"
    }
    
    Write-Host ""
}

Write-Host "构建结果: $successCount/$totalCount 成功" -ForegroundColor Cyan

if ($successCount -eq $totalCount) {
    Write-LogSuccess "所有平台构建成功！"
    
    # 创建发布包
    Write-LogInfo "创建发布包..."
    
    foreach ($platform in $platforms) {
        $os = $platform.OS
        $arch = $platform.Arch
        $platformDir = "$buildDir/$os-$arch"
        
        if (Test-Path $platformDir) {
            Write-Host "打包 $os-$arch..." -ForegroundColor Yellow
            
            # 复制安装脚本
            if ($os -eq "windows") {
                Copy-Item "scripts/safe-install.ps1" $platformDir -ErrorAction SilentlyContinue
                Copy-Item "scripts/install.ps1" $platformDir -ErrorAction SilentlyContinue
            } else {
                Copy-Item "scripts/safe-install.sh" $platformDir -ErrorAction SilentlyContinue
                Copy-Item "scripts/install.sh" $platformDir -ErrorAction SilentlyContinue
            }
            
            # 复制文档
            Copy-Item "README.md" $platformDir -ErrorAction SilentlyContinue
            Copy-Item "LICENSE" $platformDir -ErrorAction SilentlyContinue
            
            # 创建压缩包
            $archiveName = "delguard-$os-$arch.zip"
            $archivePath = "$buildDir/$archiveName"
            
            try {
                Compress-Archive -Path "$platformDir/*" -DestinationPath $archivePath -Force
                Write-LogSuccess "创建了 $archiveName"
            } catch {
                Write-LogWarning "无法创建压缩包 $archiveName : $($_.Exception.Message)"
            }
        }
    }
    
    Write-LogSuccess "跨平台构建和打包完成！"
    Write-LogInfo "构建文件位于: $buildDir"
    
} else {
    Write-LogWarning "部分平台构建失败，请检查错误信息"
    exit 1
}

# 恢复环境变量
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue