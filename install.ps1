# DelGuard 一键安装脚本 - Windows版本
# 跨平台文件安全删除工具

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:USERPROFILE\bin",
    [switch]$Help
)

# 颜色定义
$Red = "`e[31m"
$Green = "`e[32m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$NC = "`e[0m"  # No Color

function Show-Banner {
    Write-Host $Blue
    Write-Host "  ____       _       _     ____                      _     "
    Write-Host " |  _ \  ___| |_ ___| |_  / ___|  ___  _   _ _ __ | |_ "
    Write-Host " | | | |/ _ \ __/ _ \ __| \___ \ / _ \| | | | '_ \| __|"
    Write-Host " | |_| |  __/ ||  __/ |_   ___) | (_) | |_| | | | | |_ "
    Write-Host " |____/ \___|\__\___|\__| |____/ \___/ \__,_|_| |_|\__|"
    Write-Host $NC
    Write-Host "  一键安装脚本 - 跨平台文件安全删除工具"
    Write-Host ""
}

function Show-Help {
    Write-Host "使用方法: .\install.ps1 [选项]"
    Write-Host ""
    Write-Host "选项:"
    Write-Host "  -Version VERSION    指定版本 (默认: latest)"
    Write-Host "  -InstallDir PATH    安装目录 (默认: $env:USERPROFILE\bin)"
    Write-Host "  -Help               显示帮助信息"
    Write-Host ""
    Write-Host "示例:"
    Write-Host "  .\install.ps1                    # 安装最新版本"
    Write-Host "  .\install.ps1 -Version v1.0.0    # 安装指定版本"
    Write-Host "  .\install.ps1 -InstallDir C:\Tools  # 安装到指定目录"
}

function Write-Log {
    param([string]$Message)
    Write-Host "${Green}[INFO]${NC} $Message"
}

function Write-Warn {
    param([string]$Message)
    Write-Host "${Yellow}[WARN]${NC} $Message"
}

function Write-Error {
    param([string]$Message)
    Write-Host "${Red}[ERROR]${NC} $Message" -ForegroundColor Red
}

function Check-Requirements {
    Write-Log "检查系统要求..."
    
    # 检查操作系统
    if (-not $IsWindows -and -not $env:OS -like "Windows*") {
        Write-Error "此脚本仅适用于 Windows 系统"
        exit 1
    }
    
    # 检查 PowerShell 版本
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error "需要 PowerShell 5.0 或更高版本"
        exit 1
    }
    
    # 检查网络连接
    try {
        $response = Invoke-WebRequest -Uri "https://github.com" -UseBasicParsing -TimeoutSec 10
        if ($response.StatusCode -ne 200) {
            throw "无法连接到 GitHub"
        }
    }
    catch {
        Write-Error "无法连接到 GitHub，请检查网络连接"
        exit 1
    }
    
    # 检测系统架构
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { $script:Arch = "amd64" }
        "ARM64" { $script:Arch = "arm64" }
        "x86" { $script:Arch = "386" }
        default { 
            Write-Error "不支持的架构: $arch"
            exit 1
        }
    }
    
    Write-Log "检测到: windows-$($script:Arch)"
}

function Get-LatestRelease {
    try {
        $apiUrl = "https://api.github.com/repos/yourusername/DelGuard/releases/latest"
        $response = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        return $response.tag_name
    }
    catch {
        Write-Error "无法获取最新版本信息: $($_.Exception.Message)"
        exit 1
    }
}

function Download-Binary {
    param(
        [string]$Version,
        [string]$OS,
        [string]$Arch
    )
    
    $filename = "delguard_${OS}_${Arch}.exe"
    $downloadUrl = "https://github.com/yourusername/DelGuard/releases/download/${Version}/${filename}"
    
    Write-Log "下载 DelGuard ${Version}..."
    
    $tempDir = Join-Path $env:TEMP "delguard-install"
    if (-not (Test-Path $tempDir)) {
        New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    }
    
    $tempFile = Join-Path $tempDir "delguard.exe"
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
        if (-not (Test-Path $tempFile)) {
            throw "下载失败"
        }
        return $tempFile
    }
    catch {
        Write-Error "下载失败: $($_.Exception.Message)"
        Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
        exit 1
    }
}

function Install-Binary {
    param(
        [string]$SourceFile,
        [string]$InstallDir
    )
    
    Write-Log "安装到 ${InstallDir}..."
    
    # 创建安装目录
    if (-not (Test-Path $InstallDir)) {
        try {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        catch {
            Write-Error "无法创建安装目录: $($_.Exception.Message)"
            exit 1
        }
    }
    
    # 复制二进制文件
    $targetPath = Join-Path $InstallDir "delguard.exe"
    try {
        Copy-Item -Path $SourceFile -Destination $targetPath -Force
        Write-Log "安装完成！"
    }
    catch {
        Write-Error "安装失败: $($_.Exception.Message)"
        exit 1
    }
    finally {
        # 清理临时文件
        $tempDir = Split-Path $SourceFile -Parent
        Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
    }
}

function Setup-Path {
    param([string]$InstallDir)
    
    # 检查是否已在 PATH 中
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -split ";" -contains $InstallDir) {
        Write-Log "目录已在 PATH 中"
        return
    }
    
    # 添加到用户 PATH
    try {
        $newPath = $currentPath + ";" + $InstallDir
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Log "已将 ${InstallDir} 添加到用户 PATH"
        Write-Warn "请重新打开 PowerShell 或命令提示符以使用 delguard 命令"
    }
    catch {
        Write-Warn "无法自动添加到 PATH: $($_.Exception.Message)"
        Write-Warn "请手动将 ${InstallDir} 添加到 PATH 环境变量"
    }
}

function Main {
    Show-Banner
    
    if ($Help) {
        Show-Help
        return
    }
    
    Check-Requirements
    
    # 获取版本
    if ($Version -eq "latest") {
        $Version = Get-LatestRelease
    }
    
    Write-Log "准备安装版本: $Version"
    
    # 下载并安装
    $tempFile = Download-Binary -Version $Version -OS "windows" -Arch $script:Arch
    Install-Binary -SourceFile $tempFile -InstallDir $InstallDir
    Setup-Path -InstallDir $InstallDir
    
    Write-Host ""
    Write-Log "DelGuard 已成功安装！"
    Write-Host ""
    Write-Host "使用方法:"
    Write-Host "  delguard --help     # 查看帮助"
    Write-Host "  delguard file.txt   # 安全删除文件"
    Write-Host "  delguard --restore  # 恢复文件"
}

# 执行主函数
Main