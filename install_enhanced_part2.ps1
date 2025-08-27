# 颜色定义
$ColorScheme = @{
    Success = 'Green'
    Error = 'Red'
    Warning = 'Yellow'
    Info = 'Cyan'
    Header = 'Magenta'
    Normal = 'White'
}

# 日志函数
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] [$Level] $Message"
    
    # 根据日志级别选择颜色
    $color = switch ($Level) {
        "INFO" { $ColorScheme.Info }
        "ERROR" { $ColorScheme.Error }
        "WARNING" { $ColorScheme.Warning }
        "SUCCESS" { $ColorScheme.Success }
        default { $ColorScheme.Normal }
    }
    
    Write-Host $logMessage -ForegroundColor $color
    
    if (!(Test-Path (Split-Path $LOG_FILE))) {
        New-Item -ItemType Directory -Path (Split-Path $LOG_FILE) -Force | Out-Null
    }
    Add-Content -Path $LOG_FILE -Value $logMessage -Encoding UTF8
}

# 显示横幅
function Show-Banner {
    $banner = @"
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║                    🛡️  DelGuard $VERSION                    ║
║                   安全文件删除工具                           ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
"@
    Write-Host $banner -ForegroundColor $ColorScheme.Header
    Write-Host ""
}

# 检查管理员权限
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 获取系统架构
function Get-SystemArchitecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        "x86" { return "386" }
        default { return "amd64" }
    }
}

# 检查网络连接
function Test-NetworkConnection {
    try {
        $response = Invoke-WebRequest -Uri "https://api.github.com" -Method Head -TimeoutSec 10
        return $response.StatusCode -eq 200
    } catch {
        return $false
    }
}

# 获取最新版本信息
function Get-LatestRelease {
    try {
        Write-Log "获取最新版本信息..." "INFO"
        $response = Invoke-RestMethod -Uri $RELEASE_API -TimeoutSec 30
        return $response
    } catch {
        Write-Log "获取版本信息失败: $($_.Exception.Message)" "ERROR"
        throw "无法获取最新版本信息，请检查网络连接"
    }
}

# 下载文件
function Download-File {
    param([string]$Url, [string]$OutputPath)
    
    try {
        Write-Log "下载文件: $Url" "INFO"
        $webClient = New-Object System.Net.WebClient
        $webClient.DownloadFile($Url, $OutputPath)
        Write-Log "下载完成: $OutputPath" "SUCCESS"
    } catch {
        Write-Log "下载失败: $($_.Exception.Message)" "ERROR"
        throw "下载失败: $($_.Exception.Message)"
    }
}