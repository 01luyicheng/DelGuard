# DelGuard Windows 卸载脚本

param(
    [switch]$Force = $false,
    [switch]$KeepConfig = $false
)

# 颜色定义
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
}

# 配置
$BinaryName = "delguard.exe"
$PossibleDirs = @(
    "$env:LOCALAPPDATA\Programs\DelGuard",
    "$env:ProgramFiles\DelGuard",
    "$env:ProgramFiles(x86)\DelGuard"
)
$ConfigDir = Join-Path $env:APPDATA "delguard"

# 日志函数
function Write-Log {
    param(
        [string]$Message,
        [string]$Level = "INFO"
    )
    
    $color = switch ($Level) {
        "INFO" { $Colors.Blue }
        "SUCCESS" { $Colors.Green }
        "WARNING" { $Colors.Yellow }
        "ERROR" { $Colors.Red }
        default { "White" }
    }
    
    Write-Host "[$Level] $Message" -ForegroundColor $color
}

# 查找已安装的二进制文件
function Find-InstalledBinary {
    foreach ($dir in $PossibleDirs) {
        $binaryPath = Join-Path $dir $BinaryName
        if (Test-Path $binaryPath) {
            return $binaryPath
        }
    }
    
    # 检查 PATH 中的位置
    $pathDirs = $env:PATH -split ";"
    foreach ($dir in $pathDirs) {
        if ($dir) {
            $binaryPath = Join-Path $dir $BinaryName
            if (Test-Path $binaryPath) {
                return $binaryPath
            }
        }
    }
    
    return $null
}

# 移除二进制文件
function Remove-Binary {
    param([string]$BinaryPath)
    
    $installDir = Split-Path $BinaryPath -Parent
    Write-Log "移除安装目录: $installDir"
    
    try {
        # 停止可能正在运行的进程
        Get-Process -Name "delguard" -ErrorAction SilentlyContinue | Stop-Process -Force
        
        # 移除整个安装目录
        if (Test-Path $installDir) {
            Remove-Item $installDir -Recurse -Force
            Write-Log "安装目录已移除" "SUCCESS"
        }
    }
    catch {
        Write-Log "移除安装目录失败: $($_.Exception.Message)" "ERROR"
        return $false
    }
    
    return $true
}

# 从 PATH 中移除
function Remove-FromPath {
    param([string]$Directory)
    
    Write-Log "从 PATH 环境变量中移除..."
    
    try {
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        
        if ($currentPath -like "*$Directory*") {
            $newPath = ($currentPath -split ";" | Where-Object { $_ -ne $Directory }) -join ";"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-Log "已从用户 PATH 中移除: $Directory" "SUCCESS"
        }
        
        # 检查系统 PATH
        $systemPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
        if ($systemPath -like "*$Directory*") {
            Write-Log "检测到系统 PATH 中包含 DelGuard，需要管理员权限移除" "WARNING"
        }
    }
    catch {
        Write-Log "从 PATH 移除失败: $($_.Exception.Message)" "WARNING"
    }
}

# 移除配置文件
function Remove-Config {
    if (-not $KeepConfig -and (Test-Path $ConfigDir)) {
        Write-Log "移除配置目录: $ConfigDir"
        
        if (-not $Force) {
            $confirmation = Read-Host "是否保留配置文件和日志? [y/N]"
            if ($confirmation -match "^[Yy]") {
                Write-Log "配置目录已保留" "INFO"
                return
            }
        }
        
        try {
            Remove-Item $ConfigDir -Recurse -Force
            Write-Log "配置目录已移除" "SUCCESS"
        }
        catch {
            Write-Log "移除配置目录失败: $($_.Exception.Message)" "WARNING"
        }
    }
    else {
        Write-Log "配置目录已保留或不存在" "INFO"
    }
}

# 移除 PowerShell 别名
function Remove-PowerShellAliases {
    Write-Log "移除 PowerShell 别名..."
    
    try {
        $profilePath = $PROFILE.CurrentUserAllHosts
        
        if (Test-Path $profilePath) {
            $profileContent = Get-Content $profilePath -Raw
            
            if ($profileContent -match "DelGuard aliases") {
                # 移除 DelGuard 相关的别名
                $lines = Get-Content $profilePath
                $newLines = @()
                $skipSection = $false
                
                foreach ($line in $lines) {
                    if ($line -match "# DelGuard aliases") {
                        $skipSection = $true
                        continue
                    }
                    
                    if ($skipSection) {
                        if ($line -match "^Set-Alias.*delguard" -or 
                            $line -match "^function delguard-" -or
                            $line -match "^Set-Alias.*(del|rm-safe|trash|restore|empty-trash)") {
                            continue
                        }
                        elseif ($line.Trim() -eq "") {
                            $skipSection = $false
                            continue
                        }
                    }
                    
                    if (-not $skipSection) {
                        $newLines += $line
                    }
                }
                
                Set-Content -Path $profilePath -Value $newLines -Encoding UTF8
                Write-Log "已从 PowerShell 配置文件移除别名" "SUCCESS"
            }
            else {
                Write-Log "PowerShell 配置文件中未找到 DelGuard 别名" "INFO"
            }
        }
        else {
            Write-Log "PowerShell 配置文件不存在" "INFO"
        }
    }
    catch {
        Write-Log "移除 PowerShell 别名失败: $($_.Exception.Message)" "WARNING"
    }
}

# 清理回收站
function Clear-Trash {
    Write-Log "检查回收站..."
    
    $binaryPath = Find-InstalledBinary
    if ($binaryPath -and (Test-Path $binaryPath)) {
        if (-not $Force) {
            $confirmation = Read-Host "是否清空回收站? [y/N]"
            if ($confirmation -notmatch "^[Yy]") {
                return
            }
        }
        
        try {
            & $binaryPath empty --force 2>$null
            Write-Log "回收站已清空" "SUCCESS"
        }
        catch {
            Write-Log "清空回收站失败，可能回收站已为空" "WARNING"
        }
    }
}

# 主函数
function Main {
    Write-Host "🗑️  DelGuard Windows 卸载脚本" -ForegroundColor Green
    Write-Host "==============================" -ForegroundColor Green
    Write-Host ""
    
    # 查找已安装的二进制文件
    $binaryPath = Find-InstalledBinary
    
    if ($binaryPath) {
        Write-Log "找到已安装的 DelGuard: $binaryPath"
        
        # 确认卸载
        if (-not $Force) {
            Write-Host ""
            $confirmation = Read-Host "确认卸载 DelGuard? [y/N]"
            if ($confirmation -notmatch "^[Yy]") {
                Write-Log "卸载已取消"
                exit 0
            }
        }
        
        # 清理回收站
        Clear-Trash
        
        # 执行卸载步骤
        $installDir = Split-Path $binaryPath -Parent
        
        if (Remove-Binary -BinaryPath $binaryPath) {
            Remove-FromPath -Directory $installDir
            Remove-PowerShellAliases
            Remove-Config
            
            Write-Log "DelGuard 已完全卸载" "SUCCESS"
            Write-Log "感谢使用 DelGuard！" "SUCCESS"
            
            Write-Host ""
            Write-Host "注意: 请重新启动 PowerShell 以使环境变量更改生效" -ForegroundColor Yellow
        }
        else {
            Write-Log "卸载过程中发生错误" "ERROR"
            exit 1
        }
    }
    else {
        Write-Log "未找到已安装的 DelGuard" "WARNING"
        Write-Log "可能的安装位置:" "INFO"
        foreach ($dir in $PossibleDirs) {
            Write-Host "  - $(Join-Path $dir $BinaryName)" -ForegroundColor Cyan
        }
    }
}

# 运行主函数
Main