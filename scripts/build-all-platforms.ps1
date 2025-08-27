# DelGuard 跨平台构建脚本
param([switch]$Verbose)

$ErrorActionPreference = 'Stop'

Write-Host "DelGuard 跨平台构建" -ForegroundColor Cyan
Write-Host "==================" -ForegroundColor Cyan

# 设置构建环境
$env:CGO_ENABLED = "0"
$env:GO111MODULE = "on"

# 创建构建目录
New-Item -ItemType Directory -Path "build" -Force | Out-Null

# 定义平台
$platforms = @(
    @{OS="windows"; Arch="amd64"; Ext=".exe"},
    @{OS="windows"; Arch="386"; Ext=".exe"},
    @{OS="linux"; Arch="amd64"; Ext=""},
    @{OS="linux"; Arch="386"; Ext=""},
    @{OS="linux"; Arch="arm64"; Ext=""},
    @{OS="darwin"; Arch="amd64"; Ext=""},
    @{OS="darwin"; Arch="arm64"; Ext=""}
)

$success = 0
$total = $platforms.Count

foreach ($platform in $platforms) {
    $os = $platform.OS
    $arch = $platform.Arch
    $ext = $platform.Ext
    
    $output = "build/delguard-$os-$arch$ext"
    
    Write-Host "构建 $os/$arch..." -NoNewline
    
    try {
        $env:GOOS = $os
        $env:GOARCH = $arch
        
        & go build -ldflags="-s -w" -o $output .
        
        if (Test-Path $output) {
            $size = (Get-Item $output).Length
            Write-Host " 成功 ($([math]::Round($size/1MB, 2)) MB)" -ForegroundColor Green
            $success++
        } else {
            Write-Host " 失败 (文件不存在)" -ForegroundColor Red
        }
    } catch {
        Write-Host " 失败 ($($_.Exception.Message))" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "构建结果: $success/$total 成功" -ForegroundColor $(if ($success -eq $total) { "Green" } else { "Yellow" })

if ($success -eq $total) {
    Write-Host "所有平台构建成功!" -ForegroundColor Green
} else {
    Write-Host "部分平台构建失败" -ForegroundColor Yellow
}

# 清理环境变量
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue