# DelGuard 构建脚本 - Windows版本
# 构建所有支持平台的二进制文件

param(
    [string]$Version = "v1.3.0",
    [switch]$Release = $false,
    [switch]$Clean = $false
)

# 设置错误处理
$ErrorActionPreference = 'Stop'

# 常量定义
$APP_NAME = "DelGuard"
$BUILD_DIR = "build"
$DIST_DIR = "dist"

# 支持的平台
$PLATFORMS = @(
    @{OS="windows"; ARCH="amd64"; EXT=".exe"},
    @{OS="windows"; ARCH="arm64"; EXT=".exe"},
    @{OS="windows"; ARCH="386"; EXT=".exe"},
    @{OS="linux"; ARCH="amd64"; EXT=""},
    @{OS="linux"; ARCH="arm64"; EXT=""},
    @{OS="linux"; ARCH="arm"; EXT=""},
    @{OS="linux"; ARCH="386"; EXT=""},
    @{OS="darwin"; ARCH="amd64"; EXT=""},
    @{OS="darwin"; ARCH="arm64"; EXT=""}
)

Write-Host "DelGuard 构建脚本" -ForegroundColor Cyan
Write-Host "=================" -ForegroundColor Cyan
Write-Host "版本: $Version" -ForegroundColor Green
Write-Host ""

# 清理构建目录
if ($Clean) {
    Write-Host "清理构建目录..." -ForegroundColor Yellow
    if (Test-Path $BUILD_DIR) {
        Remove-Item $BUILD_DIR -Recurse -Force
    }
    if (Test-Path $DIST_DIR) {
        Remove-Item $DIST_DIR -Recurse -Force
    }
}

# 创建构建目录
New-Item -ItemType Directory -Path $BUILD_DIR -Force | Out-Null
New-Item -ItemType Directory -Path $DIST_DIR -Force | Out-Null

# 运行测试
Write-Host "运行测试..." -ForegroundColor Yellow
go test -v ./...
if ($LASTEXITCODE -ne 0) {
    Write-Error "测试失败，停止构建"
    exit 1
}

Write-Host "测试通过！" -ForegroundColor Green
Write-Host ""

# 构建所有平台
foreach ($platform in $PLATFORMS) {
    $os = $platform.OS
    $arch = $platform.ARCH
    $ext = $platform.EXT
    
    $outputName = "$APP_NAME-$os-$arch$ext"
    $outputPath = Join-Path $BUILD_DIR $outputName
    
    Write-Host "构建 $os/$arch..." -ForegroundColor Blue
    
    # 设置环境变量
    $env:GOOS = $os
    $env:GOARCH = $arch
    $env:CGO_ENABLED = "0"
    
    # 构建命令
    $ldflags = "-s -w -X main.Version=$Version -X main.BuildTime=$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    
    try {
        go build -ldflags $ldflags -o $outputPath .
        
        if (Test-Path $outputPath) {
            $size = (Get-Item $outputPath).Length
            $sizeKB = [math]::Round($size / 1KB, 2)
            Write-Host "  ✓ $outputName ($sizeKB KB)" -ForegroundColor Green
            
            # 创建发布包
            if ($Release) {
                $archiveName = "$APP_NAME-$os-$arch"
                $archivePath = Join-Path $DIST_DIR $archiveName
                
                if ($os -eq "windows") {
                    # Windows 使用 ZIP
                    $zipPath = "$archivePath.zip"
                    Compress-Archive -Path $outputPath, "README.md", "LICENSE" -DestinationPath $zipPath -Force
                    Write-Host "  ✓ 创建发布包: $zipPath" -ForegroundColor Cyan
                } else {
                    # Unix 使用 tar.gz
                    $tarPath = "$archivePath.tar.gz"
                    tar -czf $tarPath -C $BUILD_DIR $outputName -C .. README.md LICENSE
                    Write-Host "  ✓ 创建发布包: $tarPath" -ForegroundColor Cyan
                }
            }
        } else {
            Write-Host "  ✗ 构建失败" -ForegroundColor Red
        }
    } catch {
        Write-Host "  ✗ 构建失败: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# 重置环境变量
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "构建完成！" -ForegroundColor Green
Write-Host "构建文件位于: $BUILD_DIR" -ForegroundColor Gray

if ($Release) {
    Write-Host "发布包位于: $DIST_DIR" -ForegroundColor Gray
}

# 显示构建统计
$buildFiles = Get-ChildItem $BUILD_DIR -File
$totalSize = ($buildFiles | Measure-Object -Property Length -Sum).Sum
$totalSizeMB = [math]::Round($totalSize / 1MB, 2)

Write-Host ""
Write-Host "构建统计:" -ForegroundColor Cyan
Write-Host "  文件数量: $($buildFiles.Count)" -ForegroundColor Gray
Write-Host "  总大小: $totalSizeMB MB" -ForegroundColor Gray

if ($Release) {
    $distFiles = Get-ChildItem $DIST_DIR -File
    Write-Host "  发布包数量: $($distFiles.Count)" -ForegroundColor Gray
}