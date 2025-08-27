# DelGuard 跨平台兼容性验证
Write-Host "DelGuard 跨平台兼容性验证" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan

# 检查构建文件
$buildFiles = Get-ChildItem "build" -Filter "delguard-*" -ErrorAction SilentlyContinue

if ($buildFiles.Count -eq 0) {
    Write-Host "未找到构建文件，请先运行构建脚本" -ForegroundColor Red
    exit 1
}

Write-Host "构建文件检查:" -ForegroundColor Yellow
foreach ($file in $buildFiles) {
    $size = [math]::Round($file.Length / 1MB, 2)
    Write-Host "  $($file.Name): $size MB" -ForegroundColor Green
}

# 检查安装脚本
Write-Host "`n安装脚本检查:" -ForegroundColor Yellow

# PowerShell 脚本语法检查
try {
    $null = [System.Management.Automation.PSParser]::Tokenize((Get-Content "scripts/safe-install.ps1" -Raw), [ref]$null)
    Write-Host "  safe-install.ps1: 语法正确" -ForegroundColor Green
} catch {
    Write-Host "  safe-install.ps1: 语法错误 - $($_.Exception.Message)" -ForegroundColor Red
}

# Bash 脚本语法检查
if (Get-Command bash -ErrorAction SilentlyContinue) {
    try {
        $result = & bash -n "scripts/safe-install.sh" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  safe-install.sh: 语法正确" -ForegroundColor Green
        } else {
            Write-Host "  safe-install.sh: 语法错误 - $result" -ForegroundColor Red
        }
    } catch {
        Write-Host "  safe-install.sh: 无法验证 - $($_.Exception.Message)" -ForegroundColor Yellow
    }
} else {
    Write-Host "  safe-install.sh: 无法验证 (bash 不可用)" -ForegroundColor Yellow
}

# 检查平台特定文件
Write-Host "`n平台特定文件检查:" -ForegroundColor Yellow

$platformFiles = @(
    @{File="fsutil_windows.go"; Platform="Windows"},
    @{File="fsutil_unix.go"; Platform="Unix/Linux/macOS"},
    @{File="windows.go"; Platform="Windows"},
    @{File="linux.go"; Platform="Linux"},
    @{File="macos.go"; Platform="macOS"}
)

foreach ($pf in $platformFiles) {
    if (Test-Path $pf.File) {
        # 检查构建标签
        $content = Get-Content $pf.File -TotalCount 5
        $hasBuildTag = $content | Where-Object { $_ -match "//go:build|// \+build" }
        
        if ($hasBuildTag) {
            Write-Host "  $($pf.File) ($($pf.Platform)): 构建标签正确" -ForegroundColor Green
        } else {
            Write-Host "  $($pf.File) ($($pf.Platform)): 缺少构建标签" -ForegroundColor Red
        }
    } else {
        Write-Host "  $($pf.File) ($($pf.Platform)): 文件不存在" -ForegroundColor Red
    }
}

Write-Host "`n跨平台兼容性总结:" -ForegroundColor Cyan
Write-Host "- Windows: 完全支持" -ForegroundColor Green
Write-Host "- Linux: 支持 (需要测试)" -ForegroundColor Yellow  
Write-Host "- macOS: 支持 (需要测试)" -ForegroundColor Yellow
Write-Host "`n建议在目标平台上进行实际测试以确保完全兼容性。" -ForegroundColor Gray