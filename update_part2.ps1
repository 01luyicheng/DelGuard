# 查找已安装的DelGuard
function Find-InstalledDelGuard {
    # 检查常见安装位置
    $possibleLocations = @(
        "$env:LOCALAPPDATA\$APP_NAME\$EXECUTABLE_NAME",
        "$env:ProgramFiles\$APP_NAME\$EXECUTABLE_NAME",
        "$env:USERPROFILE\bin\$EXECUTABLE_NAME",
        "$env:USERPROFILE\.local\bin\$EXECUTABLE_NAME"
    )
    
    foreach ($location in $possibleLocations) {
        if (Test-Path $location) {
            return $location
        }
    }
    
    # 尝试从PATH中查找
    $fromPath = Get-Command $EXECUTABLE_NAME -ErrorAction SilentlyContinue
    if ($fromPath) {
        return $fromPath.Source
    }
    
    return $null
}

# 获取已安装版本
function Get-InstalledVersion {
    param([string]$ExecutablePath)
    
    try {
        $output = & $ExecutablePath --version 2>$null
        if ($output) {
            # 提取版本号（假设格式为 "DelGuard v1.2.3" 或类似）
            if ($output -match '(\d+\.\d+\.\d+)') {
                return $Matches[1]
            }
        }
    } catch {
        # 忽略错误
    }
    
    return "未知"
}

# 获取最新版本信息
function Get-LatestRelease {
    try {
        Write-Host "获取最新版本信息..." -ForegroundColor $ColorScheme.Info
        $response = Invoke-RestMethod -Uri $RELEASE_API -TimeoutSec 30
        return $response
    } catch {
        Write-Host "获取版本信息失败: $($_.Exception.Message)" -ForegroundColor $ColorScheme.Error
        throw "无法获取最新版本信息，请检查网络连接"
    }
}

# 下载文件
function Download-File {
    param([string]$Url, [string]$OutputPath)
    
    try {
        Write-Host "下载文件: $Url" -ForegroundColor $ColorScheme.Info
        $webClient = New-Object System.Net.WebClient
        $webClient.DownloadFile($Url, $OutputPath)
        Write-Host "下载完成: $OutputPath" -ForegroundColor $ColorScheme.Success
    } catch {
        Write-Host "下载失败: $($_.Exception.Message)" -ForegroundColor $ColorScheme.Error
        throw "下载失败: $($_.Exception.Message)"
    }
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

# 显示横幅
function Show-Banner {
    $banner = @"
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║                🔄 DelGuard 一键更新工具                      ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
"@
    Write-Host $banner -ForegroundColor $ColorScheme.Header
    Write-Host ""
}