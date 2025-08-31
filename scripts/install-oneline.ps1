# DelGuard 一行命令安装脚本 (Windows)
# 使用方法：复制粘贴以下命令到PowerShell即可
# powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install-oneline.ps1' -UseBasicParsing | Invoke-Expression }"

# 检查管理员权限
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "❌ 需要管理员权限运行" -ForegroundColor Red
    Write-Host "请右键点击PowerShell并选择'以管理员身份运行'" -ForegroundColor Yellow
    exit 1
}

# 设置参数
$Owner = "01luyicheng"  # GitHub用户名
$Repo = "DelGuard"
$Version = "v1.5.8"

# 检测系统架构
$arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
}

Write-Host "🚀 正在安装 DelGuard $Version..." -ForegroundColor Green

# 下载二进制文件
$downloadUrl = "https://github.com/$Owner/$Repo/releases/download/$Version/delguard-windows-$arch.exe"
$installDir = "$env:ProgramFiles\DelGuard"
$tempDir = [System.IO.Path]::GetTempPath()
$downloadPath = Join-Path $tempDir "delguard.exe"

try {
    # 下载文件
    Write-Host "📥 正在下载..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $downloadUrl -OutFile $downloadPath -UseBasicParsing
    
    # 创建安装目录
    if (!(Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }
    
    # 复制文件
    Copy-Item $downloadPath "$installDir\delguard.exe" -Force
    
    # 添加到PATH
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($currentPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "Machine")
    }
    
    # 创建快捷删除脚本
    $delScript = @"
@echo off
"$installDir\delguard.exe" delete %*
"@
    $delScript | Out-File -FilePath "$installDir\del.bat" -Encoding ASCII
    
    Write-Host "✅ DelGuard 安装完成！" -ForegroundColor Green
    Write-Host "📖 使用说明:" -ForegroundColor Yellow
    Write-Host "  delguard --help    - 查看帮助" -ForegroundColor White
    Write-Host "  delguard list      - 查看回收站" -ForegroundColor White
    Write-Host "  delguard restore   - 恢复文件" -ForegroundColor White
    Write-Host "⚠️  请重新打开终端" -ForegroundColor Yellow
    
} catch {
    Write-Host "❌ 安装失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
} finally {
    # 清理临时文件
    if (Test-Path $downloadPath) {
        Remove-Item $downloadPath -Force -ErrorAction SilentlyContinue
    }
}