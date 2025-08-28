# DelGuard 跨平台构建脚本

Write-Host "=== DelGuard 跨平台构建开始 ===" -ForegroundColor Green

# 设置Go环境
$env:PATH = "C:\Program Files\Go\bin;" + $env:PATH

# 创建构建目录
$buildDir = "build"
if (Test-Path $buildDir) {
    Remove-Item -Recurse -Force $buildDir
}
New-Item -ItemType Directory -Path $buildDir -Force | Out-Null

Write-Host "`n构建 Windows 版本..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o "$buildDir/delguard_windows_amd64.exe" ./cmd/delguard
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Windows 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ Windows 版本构建失败" -ForegroundColor Red
}

Write-Host "`n构建 Linux 版本..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o "$buildDir/delguard_linux_amd64" ./cmd/delguard
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Linux 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ Linux 版本构建失败" -ForegroundColor Red
}

Write-Host "`n构建 macOS 版本..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o "$buildDir/delguard_darwin_amd64" ./cmd/delguard
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ macOS 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ macOS 版本构建失败" -ForegroundColor Red
}

Write-Host "`n构建 ARM64 版本..." -ForegroundColor Yellow

# Linux ARM64
$env:GOOS = "linux"
$env:GOARCH = "arm64"
go build -o "$buildDir/delguard_linux_arm64" ./cmd/delguard
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Linux ARM64 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ Linux ARM64 版本构建失败" -ForegroundColor Red
}

# macOS ARM64 (Apple Silicon)
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o "$buildDir/delguard_darwin_arm64" ./cmd/delguard
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ macOS ARM64 版本构建成功" -ForegroundColor Green
} else {
    Write-Host "❌ macOS ARM64 版本构建失败" -ForegroundColor Red
}

# 重置环境变量
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "`n=== 构建结果 ===" -ForegroundColor Green
Get-ChildItem $buildDir | Format-Table Name, Length, LastWriteTime

Write-Host "`n=== DelGuard 跨平台构建完成 ===" -ForegroundColor Green