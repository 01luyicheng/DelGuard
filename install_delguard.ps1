# DelGuard 安装脚本

param(
    [string]$InstallPath = "$env:LOCALAPPDATA\DelGuard",
    [switch]$AddToPath = $false
)

Write-Host "=== DelGuard 安装程序 ===" -ForegroundColor Green

# 检查管理员权限
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")

if ($AddToPath -and -not $isAdmin) {
    Write-Host "警告: 添加到系统PATH需要管理员权限" -ForegroundColor Yellow
}

# 创建安装目录
Write-Host "`n创建安装目录: $InstallPath" -ForegroundColor Yellow
if (-not (Test-Path $InstallPath)) {
    New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
}

# 复制可执行文件
Write-Host "复制 DelGuard 可执行文件..." -ForegroundColor Yellow
$sourceExe = ".\delguard.exe"
$targetExe = "$InstallPath\delguard.exe"

if (Test-Path $sourceExe) {
    Copy-Item $sourceExe $targetExe -Force
    Write-Host "✅ 可执行文件复制成功" -ForegroundColor Green
} else {
    Write-Host "❌ 找不到源文件: $sourceExe" -ForegroundColor Red
    Write-Host "请先运行构建脚本生成可执行文件" -ForegroundColor Yellow
    exit 1
}

# 创建配置目录
$configDir = "$env:USERPROFILE\.delguard"
if (-not (Test-Path $configDir)) {
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
    Write-Host "✅ 配置目录创建成功: $configDir" -ForegroundColor Green
}

# 创建默认配置文件
$configFile = "$configDir\config.json"
if (-not (Test-Path $configFile)) {
    $defaultConfig = @{
        language = "zh-cn"
        max_file_size = 1073741824
        max_backup_files = 10
        enable_recycle_bin = $true
        enable_logging = $true
        log_level = "info"
    } | ConvertTo-Json -Depth 10

    Set-Content -Path $configFile -Value $defaultConfig -Encoding UTF8
    Write-Host "✅ 默认配置文件创建成功" -ForegroundColor Green
}

# 添加到PATH（可选）
if ($AddToPath) {
    if ($isAdmin) {
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
        if ($currentPath -notlike "*$InstallPath*") {
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$InstallPath", "Machine")
            Write-Host "✅ 已添加到系统PATH" -ForegroundColor Green
        } else {
            Write-Host "ℹ️ 已存在于系统PATH中" -ForegroundColor Blue
        }
    } else {
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($currentPath -notlike "*$InstallPath*") {
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$InstallPath", "User")
            Write-Host "✅ 已添加到用户PATH" -ForegroundColor Green
        } else {
            Write-Host "ℹ️ 已存在于用户PATH中" -ForegroundColor Blue
        }
    }
}

# 创建桌面快捷方式（可选）
$createShortcut = Read-Host "`n是否创建桌面快捷方式? (y/N)"
if ($createShortcut -eq 'y' -or $createShortcut -eq 'Y') {
    $WshShell = New-Object -comObject WScript.Shell
    $Shortcut = $WshShell.CreateShortcut("$env:USERPROFILE\Desktop\DelGuard.lnk")
    $Shortcut.TargetPath = $targetExe
    $Shortcut.WorkingDirectory = $InstallPath
    $Shortcut.Description = "DelGuard - 安全文件删除工具"
    $Shortcut.Save()
    Write-Host "✅ 桌面快捷方式创建成功" -ForegroundColor Green
}

# 测试安装
Write-Host "`n测试安装..." -ForegroundColor Yellow
try {
    $version = & $targetExe version 2>&1
    Write-Host "✅ 安装测试成功" -ForegroundColor Green
    Write-Host $version -ForegroundColor Cyan
} catch {
    Write-Host "❌ 安装测试失败: $_" -ForegroundColor Red
}

Write-Host "`n=== DelGuard 安装完成 ===" -ForegroundColor Green
Write-Host "安装路径: $InstallPath" -ForegroundColor Cyan
Write-Host "配置目录: $configDir" -ForegroundColor Cyan
Write-Host "`n使用方法:" -ForegroundColor Yellow
Write-Host "  $targetExe help" -ForegroundColor White
Write-Host "  $targetExe version" -ForegroundColor White
Write-Host "  $targetExe config show" -ForegroundColor White