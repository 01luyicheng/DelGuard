# DelGuard 跨平台构建脚本 (PowerShell)

Write-Host "🔨 DelGuard 跨平台构建开始..." -ForegroundColor Green

# 设置版本信息
$VERSION = "1.0.0"
$BUILD_TIME = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$GIT_COMMIT = try { (git rev-parse --short HEAD 2>$null) } catch { "unknown" }

# 构建标志
$LDFLAGS = "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"

# 创建构建目录
New-Item -ItemType Directory -Force -Path "build" | Out-Null

Write-Host "📦 构建 Windows 版本..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags $LDFLAGS -o "build/delguard-windows-amd64.exe" .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Windows 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ Windows 版本构建失败" -ForegroundColor Red
    exit 1
}

Write-Host "📦 构建 macOS 版本..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -ldflags $LDFLAGS -o "build/delguard-darwin-amd64" .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ macOS Intel 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ macOS Intel 版本构建失败" -ForegroundColor Red
    exit 1
}

$env:GOARCH = "arm64"
go build -ldflags $LDFLAGS -o "build/delguard-darwin-arm64" .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ macOS Apple Silicon 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ macOS Apple Silicon 版本构建失败" -ForegroundColor Red
    exit 1
}

Write-Host "📦 构建 Linux 版本..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags $LDFLAGS -o "build/delguard-linux-amd64" .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Linux 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ Linux 版本构建失败" -ForegroundColor Red
    exit 1
}

$env:GOARCH = "arm64"
go build -ldflags $LDFLAGS -o "build/delguard-linux-arm64" .
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Linux ARM64 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ Linux ARM64 版本构建失败" -ForegroundColor Red
    exit 1
}

# 重置环境变量
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "🎉 所有平台构建完成！" -ForegroundColor Green
Write-Host ""
Write-Host "构建文件列表:" -ForegroundColor Cyan
Get-ChildItem -Path "build" | Format-Table Name, Length, LastWriteTime

Write-Host ""
Write-Host "📋 构建信息:" -ForegroundColor Cyan
Write-Host "版本: $VERSION"
Write-Host "构建时间: $BUILD_TIME"
Write-Host "Git提交: $GIT_COMMIT"