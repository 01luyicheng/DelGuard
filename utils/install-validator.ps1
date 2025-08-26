# DelGuard Installation Validator and Rollback System (PowerShell)
# Version: 2.1.1 Enhanced
# Provides installation verification and rollback capabilities

# Global variables for rollback
$Global:RollbackLog = ""
$Global:BackupDir = ""
$Global:RollbackActions = @()

# Initialize rollback system
function Initialize-RollbackSystem {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $Global:BackupDir = Join-Path $env:TEMP "delguard-backup-$timestamp"
    $Global:RollbackLog = Join-Path $env:TEMP "delguard-rollback-$timestamp.log"
    
    New-Item -ItemType Directory -Path $Global:BackupDir -Force | Out-Null
    
    $logHeader = @"
DelGuard Installation Rollback Log
Backup Directory: $Global:BackupDir
Started: $(Get-Date)

"@
    
    Set-Content -Path $Global:RollbackLog -Value $logHeader -Encoding UTF8
    Write-Host "✓ 回滚系统已初始化: $Global:BackupDir" -ForegroundColor Green
}

# Record rollback action
function Record-RollbackAction {
    param(
        [string]$ActionType,
        [string]$Target,
        [string]$BackupPath = ""
    )
    
    $action = "$ActionType`:$Target`:$BackupPath"
    $Global:RollbackActions += $action
    
    $logEntry = "$(Get-Date -Format 'HH:mm:ss') RECORD: $action"
    Add-Content -Path $Global:RollbackLog -Value $logEntry -Encoding UTF8
    Write-Debug "记录回滚操作: $ActionType for $Target"
}

# Backup file before modification
function Backup-File {
    param(
        [string]$FilePath,
        [string]$BackupName = ""
    )
    
    if (-not (Test-Path $FilePath)) {
        Write-Debug "文件不存在，无需备份: $FilePath"
        return $true
    }
    
    if ([string]::IsNullOrEmpty($BackupName)) {
        $BackupName = Split-Path $FilePath -Leaf
    }
    
    $backupPath = Join-Path $Global:BackupDir $BackupName
    
    try {
        Copy-Item -Path $FilePath -Destination $backupPath -Force
        Record-RollbackAction -ActionType "RESTORE_FILE" -Target $FilePath -BackupPath $backupPath
        Write-Debug "已备份文件: $FilePath -> $backupPath"
        return $true
    }
    catch {
        Write-Error "备份文件失败: $FilePath - $($_.Exception.Message)"
        return $false
    }
}

# Record file creation for rollback
function Record-FileCreation {
    param([string]$FilePath)
    
    Record-RollbackAction -ActionType "DELETE_FILE" -Target $FilePath
    Write-Debug "记录文件创建用于回滚: $FilePath"
}

# Record directory creation for rollback
function Record-DirectoryCreation {
    param([string]$DirectoryPath)
    
    Record-RollbackAction -ActionType "DELETE_DIR" -Target $DirectoryPath
    Write-Debug "记录目录创建用于回滚: $DirectoryPath"
}

# Verify DelGuard installation
function Test-DelGuardInstallation {
    param(
        [string]$InstallPath,
        [string]$ExpectedVersion = ""
    )
    
    Write-Host "ℹ 正在验证 DelGuard 安装..." -ForegroundColor Cyan
    
    $verificationErrors = @()
    
    # Check if executable exists
    $delguardExe = Join-Path $InstallPath "delguard.exe"
    if (-not (Test-Path $delguardExe)) {
        $verificationErrors += "可执行文件未找到: $delguardExe"
    }
    
    # Check if executable works
    if (Test-Path $delguardExe) {
        try {
            $versionOutput = & $delguardExe --version 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✓ DelGuard 版本: $versionOutput" -ForegroundColor Green
                
                # Check version if specified
                if (-not [string]::IsNullOrEmpty($ExpectedVersion) -and $ExpectedVersion -ne "latest") {
                    if ($versionOutput -notlike "*$ExpectedVersion*") {
                        $verificationErrors += "版本不匹配: 期望 $ExpectedVersion, 实际 $versionOutput"
                    }
                }
            }
            else {
                $verificationErrors += "可执行文件运行失败: $delguardExe --version"
            }
        }
        catch {
            $verificationErrors += "可执行文件测试失败: $($_.Exception.Message)"
        }
    }
    
    # Check PATH configuration
    try {
        $pathDelguard = Get-Command delguard -ErrorAction SilentlyContinue
        if ($null -eq $pathDelguard) {
            $verificationErrors += "DelGuard 未在 PATH 中找到"
        }
        else {
            $pathDelguardPath = $pathDelguard.Source
            if ($pathDelguardPath -ne $delguardExe) {
                Write-Warning "PATH 中的 DelGuard ($pathDelguardPath) 与安装位置 ($delguardExe) 不同"
            }
        }
    }
    catch {
        $verificationErrors += "PATH 检查失败: $($_.Exception.Message)"
    }
    
    # Check PowerShell profile configuration
    $profileConfigured = $false
    $profiles = @($PROFILE.AllUsersAllHosts, $PROFILE.AllUsersCurrentHost, $PROFILE.CurrentUserAllHosts, $PROFILE.CurrentUserCurrentHost)
    
    foreach ($profile in $profiles) {
        if ((Test-Path $profile) -and (Get-Content $profile -Raw -ErrorAction SilentlyContinue) -like "*DelGuard*") {
            Write-Host "✓ PowerShell 配置文件中找到配置: $profile" -ForegroundColor Green
            $profileConfigured = $true
            break
        }
    }
    
    if (-not $profileConfigured) {
        $verificationErrors += "PowerShell 配置文件中未找到 DelGuard 配置"
    }
    
    # Test basic functionality
    if (Test-Path $delguardExe) {
        try {
            $helpOutput = & $delguardExe --help 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✓ 帮助命令正常工作" -ForegroundColor Green
            }
            else {
                $verificationErrors += "帮助命令失败"
            }
        }
        catch {
            $verificationErrors += "功能测试失败: $($_.Exception.Message)"
        }
    }
    
    # Report verification results
    if ($verificationErrors.Count -eq 0) {
        Write-Host "✓ 安装验证通过" -ForegroundColor Green
        return $true
    }
    else {
        Write-Host "✗ 安装验证失败:" -ForegroundColor Red
        foreach ($error in $verificationErrors) {
            Write-Host "  • $error" -ForegroundColor Red
        }
        return $false
    }
}

# Perform rollback
function Invoke-Rollback {
    Write-Host "ℹ 开始安装回滚..." -ForegroundColor Cyan
    
    $rollbackErrors = @()
    $actionsCount = $Global:RollbackActions.Count
    
    if ($actionsCount -eq 0) {
        Write-Host "ℹ 没有记录的回滚操作" -ForegroundColor Cyan
        return $true
    }
    
    Write-Host "ℹ 执行 $actionsCount 个回滚操作..." -ForegroundColor Cyan
    
    # Process rollback actions in reverse order
    for ($i = $actionsCount - 1; $i -ge 0; $i--) {
        $action = $Global:RollbackActions[$i]
        $parts = $action -split ':'
        $actionType = $parts[0]
        $target = $parts[1]
        $backupPath = if ($parts.Length -gt 2) { $parts[2] } else { "" }
        
        $logEntry = "$(Get-Date -Format 'HH:mm:ss') ROLLBACK: $action"
        Add-Content -Path $Global:RollbackLog -Value $logEntry -Encoding UTF8
        
        switch ($actionType) {
            "DELETE_FILE" {
                if (Test-Path $target) {
                    try {
                        Remove-Item -Path $target -Force
                        Write-Host "✓ 已删除文件: $target" -ForegroundColor Green
                    }
                    catch {
                        $rollbackErrors += "删除文件失败: $target - $($_.Exception.Message)"
                    }
                }
            }
            "DELETE_DIR" {
                if (Test-Path $target -PathType Container) {
                    try {
                        Remove-Item -Path $target -Force -Recurse -ErrorAction SilentlyContinue
                        Write-Host "✓ 已删除目录: $target" -ForegroundColor Green
                    }
                    catch {
                        Write-Warning "目录非空或删除失败: $target"
                    }
                }
            }
            "RESTORE_FILE" {
                if (Test-Path $backupPath) {
                    try {
                        Copy-Item -Path $backupPath -Destination $target -Force
                        Write-Host "✓ 已恢复文件: $target" -ForegroundColor Green
                    }
                    catch {
                        $rollbackErrors += "恢复文件失败: $target - $($_.Exception.Message)"
                    }
                }
                else {
                    $rollbackErrors += "备份文件未找到: $backupPath"
                }
            }
            default {
                Write-Warning "未知的回滚操作类型: $actionType"
            }
        }
    }
    
    # Report rollback results
    if ($rollbackErrors.Count -eq 0) {
        Write-Host "✓ 回滚成功完成" -ForegroundColor Green
        
        # Clean up backup directory if empty
        try {
            $items = Get-ChildItem -Path $Global:BackupDir -ErrorAction SilentlyContinue
            if ($items.Count -eq 0) {
                Remove-Item -Path $Global:BackupDir -Force
                Write-Host "ℹ 已清理备份目录" -ForegroundColor Cyan
            }
        }
        catch {
            # Ignore cleanup errors
        }
        
        return $true
    }
    else {
        Write-Host "✗ 回滚完成但有错误:" -ForegroundColor Red
        foreach ($error in $rollbackErrors) {
            Write-Host "  • $error" -ForegroundColor Red
        }
        Write-Host "ℹ 备份文件保留在: $Global:BackupDir" -ForegroundColor Cyan
        return $false
    }
}

# Check system health before installation
function Test-SystemHealth {
    Write-Host "ℹ 检查系统健康状况..." -ForegroundColor Cyan
    
    $healthIssues = @()
    
    # Check disk space
    try {
        $drive = Get-PSDrive -Name C -ErrorAction SilentlyContinue
        if ($drive -and $drive.Free -lt 100MB) {
            $freeSpaceMB = [math]::Round($drive.Free / 1MB, 2)
            $healthIssues += "磁盘空间不足: ${freeSpaceMB}MB 可用"
        }
    }
    catch {
        $healthIssues += "无法检查磁盘空间"
    }
    
    # Check memory
    try {
        $memory = Get-CimInstance -ClassName Win32_OperatingSystem -ErrorAction SilentlyContinue
        if ($memory -and $memory.FreePhysicalMemory -lt 102400) {  # Less than 100MB
            $freeMemoryMB = [math]::Round($memory.FreePhysicalMemory / 1024, 2)
            $healthIssues += "内存不足: ${freeMemoryMB}MB 可用"
        }
    }
    catch {
        # Memory check is optional
    }
    
    # Check for conflicting installations
    try {
        $existingDelguard = Get-Command delguard -ErrorAction SilentlyContinue
        if ($existingDelguard) {
            Write-Warning "发现现有的 DelGuard 安装: $($existingDelguard.Source)"
        }
    }
    catch {
        # This is expected if DelGuard is not installed
    }
    
    # Check write permissions
    $testFile = Join-Path $env:TEMP "delguard-write-test-$(Get-Random).txt"
    try {
        "test" | Out-File -FilePath $testFile -Encoding UTF8
        Remove-Item -Path $testFile -Force -ErrorAction SilentlyContinue
    }
    catch {
        $healthIssues += "无法写入临时目录"
    }
    
    # Check PowerShell execution policy
    $executionPolicy = Get-ExecutionPolicy
    if ($executionPolicy -eq "Restricted") {
        $healthIssues += "PowerShell 执行策略过于严格: $executionPolicy"
    }
    
    # Report health check results
    if ($healthIssues.Count -eq 0) {
        Write-Host "✓ 系统健康检查通过" -ForegroundColor Green
        return $true
    }
    else {
        Write-Host "⚠ 检测到系统健康问题:" -ForegroundColor Yellow
        foreach ($issue in $healthIssues) {
            Write-Host "  • $issue" -ForegroundColor Yellow
        }
        return $false
    }
}

# Generate installation report
function New-InstallationReport {
    param(
        [string]$InstallPath,
        [string]$Status
    )
    
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $reportFile = Join-Path $env:TEMP "delguard-install-report-$timestamp.txt"
    
    $reportContent = @"
DelGuard Installation Report
===========================
Date: $(Get-Date)
Status: $Status
Install Path: $InstallPath
Platform: Windows $([Environment]::OSVersion.Version)
User: $env:USERNAME
PowerShell: $($PSVersionTable.PSVersion)

Installation Details:
"@
    
    $delguardExe = Join-Path $InstallPath "delguard.exe"
    if (Test-Path $delguardExe) {
        $reportContent += "`n✓ Executable installed: $delguardExe"
        try {
            $versionOutput = & $delguardExe --version 2>$null
            if ($LASTEXITCODE -eq 0) {
                $reportContent += "`n✓ Version: $versionOutput"
            }
            else {
                $reportContent += "`n✗ Version check failed"
            }
        }
        catch {
            $reportContent += "`n✗ Version check failed: $($_.Exception.Message)"
        }
    }
    else {
        $reportContent += "`n✗ Executable not found"
    }
    
    try {
        $pathDelguard = Get-Command delguard -ErrorAction SilentlyContinue
        if ($pathDelguard) {
            $reportContent += "`n✓ DelGuard found in PATH: $($pathDelguard.Source)"
        }
        else {
            $reportContent += "`n✗ DelGuard not found in PATH"
        }
    }
    catch {
        $reportContent += "`n✗ PATH check failed"
    }
    
    $reportContent += "`n`nLog Files:"
    $reportContent += "`n• Rollback Log: $Global:RollbackLog"
    $reportContent += "`n• This Report: $reportFile"
    
    Set-Content -Path $reportFile -Value $reportContent -Encoding UTF8
    Write-Host "ℹ 安装报告已生成: $reportFile" -ForegroundColor Cyan
    
    return $reportFile
}

# Export functions
Export-ModuleMember -Function Initialize-RollbackSystem, Record-RollbackAction, Backup-File
Export-ModuleMember -Function Record-FileCreation, Record-DirectoryCreation, Test-DelGuardInstallation
Export-ModuleMember -Function Invoke-Rollback, Test-SystemHealth, New-InstallationReport