# 检查安装状态
function Get-InstallStatus {
    Write-Host "=== DelGuard 安装状态 ===" -ForegroundColor $ColorScheme.Header
    
    if (Test-Path $EXECUTABLE_PATH) {
        Write-Host "✓ 已安装" -ForegroundColor $ColorScheme.Success
        Write-Host "  位置: $EXECUTABLE_PATH" -ForegroundColor $ColorScheme.Normal
        
        try {
            $version = & $EXECUTABLE_PATH --version 2>$null
            Write-Host "  版本: $version" -ForegroundColor $ColorScheme.Normal
        } catch {
            Write-Host "  版本: 无法获取" -ForegroundColor $ColorScheme.Warning
        }
    } else {
        Write-Host "✗ 未安装" -ForegroundColor $ColorScheme.Error
    }
    
    # 检查 PATH
    $pathCheck = $env:PATH -split ';' | Where-Object { $_ -eq $INSTALL_DIR }
    if ($pathCheck) {
        Write-Host "✓ 已添加到 PATH" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "✗ 未添加到 PATH" -ForegroundColor $ColorScheme.Warning
    }
    
    # 检查别名
    if (Get-Alias delguard -ErrorAction SilentlyContinue) {
        Write-Host "✓ PowerShell 别名已配置" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "✗ PowerShell 别名未配置" -ForegroundColor $ColorScheme.Warning
    }
    
    # 检查配置目录
    if (Test-Path $CONFIG_DIR) {
        Write-Host "✓ 配置目录存在: $CONFIG_DIR" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "✗ 配置目录不存在" -ForegroundColor $ColorScheme.Warning
    }
    
    # 检查UTF-8设置
    $profilePath = $PROFILE.CurrentUserAllHosts
    if (Test-Path $profilePath) {
        $content = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
        if ($content -like "*[Console]::OutputEncoding = [System.Text.Encoding]::UTF8*") {
            Write-Host "✓ PowerShell UTF-8编码已配置" -ForegroundColor $ColorScheme.Success
        } else {
            Write-Host "✗ PowerShell UTF-8编码未配置" -ForegroundColor $ColorScheme.Warning
        }
    } else {
        Write-Host "✗ PowerShell配置文件不存在" -ForegroundColor $ColorScheme.Warning
    }
}