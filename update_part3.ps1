# 主程序
try {
    Show-Banner
    
    # 查找已安装的DelGuard
    $installedPath = Find-InstalledDelGuard
    if (-not $installedPath) {
        Write-Host "未找到已安装的DelGuard。请先安装DelGuard。" -ForegroundColor $ColorScheme.Error
        exit 1
    }
    
    $installDir = Split-Path $installedPath -Parent
    Write-Host "已找到DelGuard: $installedPath" -ForegroundColor $ColorScheme.Success
    
    # 获取已安装版本
    $installedVersion = Get-InstalledVersion -ExecutablePath $installedPath
    Write-Host "当前版本: $installedVersion" -ForegroundColor $ColorScheme.Info
    
    # 获取最新版本
    $release = Get-LatestRelease
    $latestVersion = $release.tag_name -replace 'v', ''
    Write-Host "最新版本: $latestVersion" -ForegroundColor $ColorScheme.Info
    
    # 比较版本
    $updateAvailable = $Force -or ($installedVersion -ne $latestVersion -and $installedVersion -ne "未知")
    
    if (-not $updateAvailable) {
        Write-Host "DelGuard已经是最新版本。" -ForegroundColor $ColorScheme.Success
        exit 0
    }
    
    Write-Host "发现新版本！" -ForegroundColor $ColorScheme.Warning
    
    # 如果只是检查更新，则退出
    if ($CheckOnly) {
        Write-Host "有可用更新。使用不带 -CheckOnly 参数的命令来执行更新。" -ForegroundColor $ColorScheme.Info
        exit 0
    }
    
    # 确认更新
    $confirmation = Read-Host "是否更新到最新版本？(Y/N)"
    if ($confirmation -ne "Y" -and $confirmation -ne "y") {
        Write-Host "更新已取消。" -ForegroundColor $ColorScheme.Warning
        exit 0
    }
    
    # 确定下载URL
    $arch = Get-SystemArchitecture
    $assetName = "$APP_NAME-windows-$arch.zip"
    $asset = $release.assets | Where-Object { $_.name -eq $assetName }
    
    if (-not $asset) {
        Write-Host "未找到适合的安装包: $assetName" -ForegroundColor $ColorScheme.Error
        exit 1
    }
    
    $downloadUrl = $asset.browser_download_url
    
    # 创建临时目录
    $tempDir = Join-Path $env:TEMP "delguard-update"
    if (Test-Path $tempDir) {
        Remove-Item $tempDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    
    # 下载文件
    $zipPath = Join-Path $tempDir "$assetName"
    Download-File -Url $downloadUrl -OutputPath $zipPath
    
    # 解压文件
    Write-Host "解压安装包..." -ForegroundColor $ColorScheme.Info
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
    
    # 备份当前可执行文件
    $backupPath = "$installedPath.backup"
    Copy-Item -Path $installedPath -Destination $backupPath -Force
    Write-Host "已备份当前版本到: $backupPath" -ForegroundColor $ColorScheme.Info
    
    # 停止可能正在运行的DelGuard进程
    $processes = Get-Process | Where-Object { $_.Path -eq $installedPath }
    if ($processes) {
        Write-Host "正在停止DelGuard进程..." -ForegroundColor $ColorScheme.Warning
        $processes | Stop-Process -Force
        Start-Sleep -Seconds 1
    }
    
    # 复制新文件
    $extractedExe = Get-ChildItem -Path $tempDir -Filter $EXECUTABLE_NAME -Recurse | Select-Object -First 1
    if ($extractedExe) {
        Copy-Item -Path $extractedExe.FullName -Destination $installedPath -Force
        Write-Host "已更新到: $installedPath" -ForegroundColor $ColorScheme.Success
    } else {
        Write-Host "在安装包中未找到可执行文件，恢复备份..." -ForegroundColor $ColorScheme.Error
        Copy-Item -Path $backupPath -Destination $installedPath -Force
        throw "更新失败：在安装包中未找到可执行文件"
    }
    
    # 清理临时文件
    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    
    # 验证更新
    $newVersion = Get-InstalledVersion -ExecutablePath $installedPath
    Write-Host "DelGuard已成功更新到版本: $newVersion" -ForegroundColor $ColorScheme.Success
    
} catch {
    Write-Host "更新失败: $($_.Exception.Message)" -ForegroundColor $ColorScheme.Error
    exit 1
}