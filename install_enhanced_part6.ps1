# 主程序
try {
    # 显示横幅
    Show-Banner
    
    # 根据参数执行相应操作
    if ($Status) {
        Get-InstallStatus
    } elseif ($Uninstall) {
        Uninstall-DelGuard
    } else {
        Install-DelGuard
    }
} catch {
    Write-Log "错误: $($_.Exception.Message)" "ERROR"
    exit 1
}