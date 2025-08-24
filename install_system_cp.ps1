# DelGuard 系统级安装脚本
# 以管理员身份运行此脚本以安装系统级cp命令

param(
    [switch]$Force
)

# 检查管理员权限
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "需要管理员权限来安装系统级命令..." -ForegroundColor Yellow
    Write-Host "正在以管理员身份重新启动..." -ForegroundColor Yellow
    
    # 以管理员身份重新启动
    Start-Process powershell -ArgumentList "-ExecutionPolicy Bypass -File `"$($MyInvocation.MyCommand.Path)`" -Force:$Force" -Verb RunAs
    exit
}

$InstallDir = "C:\Program Files\DelGuard"
$ExeName = "delguard.exe"
$CurrentPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$ExePath = Join-Path $CurrentPath $ExeName

Write-Host "正在安装DelGuard系统级命令..." -ForegroundColor Green

# 创建安装目录
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-Host "创建目录: $InstallDir" -ForegroundColor Green
}

# 复制可执行文件
Copy-Item -Path $ExePath -Destination (Join-Path $InstallDir $ExeName) -Force
Write-Host "复制可执行文件完成" -ForegroundColor Green

# 创建系统级命令脚本
$Commands = @{
    "cp" = "--cp"
    "del" = ""
    "rm" = ""
}

foreach ($cmd in $Commands.Keys) {
    $cmdPath = Join-Path $InstallDir "$cmd.bat"
    $args = $Commands[$cmd]
    $content = "@`"$InstallDir\$ExeName`" $args %*"
    Set-Content -Path $cmdPath -Value $content -Encoding ASCII
    Write-Host "创建命令: $cmd" -ForegroundColor Green
}

# 添加到系统PATH
$CurrentSystemPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
if (-not $CurrentSystemPath.Contains($InstallDir)) {
    $NewPath = $InstallDir + ";" + $CurrentSystemPath
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "Machine")
    Write-Host "已添加到系统PATH" -ForegroundColor Green
} else {
    Write-Host "路径已存在于系统PATH中" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "安装完成！" -ForegroundColor Green
Write-Host ""
Write-Host "现在你可以在任何位置使用以下命令：" -ForegroundColor Cyan
Write-Host "  cp source.txt dest.txt    - 安全复制文件" -ForegroundColor White
Write-Host "  del filename              - 安全删除文件" -ForegroundColor White
Write-Host "  rm filename               - 安全删除文件" -ForegroundColor White
Write-Host ""
Write-Host "请重新打开命令提示符或PowerShell窗口以使用新命令。" -ForegroundColor Yellow

# 测试安装
Write-Host ""
Write-Host "正在测试安装..." -ForegroundColor Green
Start-Sleep -Seconds 2

# 测试cp命令
$testFile = Join-Path $env:TEMP "delguard_test.txt"
$testDest = Join-Path $env:TEMP "delguard_test_copy.txt"
"测试内容" | Set-Content -Path $testFile

$cmdTest = "cmd /c `"cp `"$testFile`" `"$testDest`"`""
Invoke-Expression $cmdTest

if (Test-Path $testDest) {
    Write-Host "✅ 系统级cp命令测试成功！" -ForegroundColor Green
    Remove-Item $testFile, $testDest -Force
} else {
    Write-Host "⚠️ 系统级cp命令可能需要重启后生效" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "安装过程完成！" -ForegroundColor Green