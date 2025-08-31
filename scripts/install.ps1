# DelGuard Windows 智能安装脚本
# 支持 Windows 系统自动检测和安装

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:ProgramFiles\DelGuard",
    [switch]$Force,
    [switch]$NoAlias
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 全局变量
$GitHubRepo = "01luyicheng/DelGuard"
$ConfigDir = "$env:APPDATA\DelGuard"
$TempDir = "$env:TEMP\delguard-install"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $colors = @{
        "Red" = [ConsoleColor]::Red
        "Green" = [ConsoleColor]::Green
        "Yellow" = [ConsoleColor]::Yellow
        "Blue" = [ConsoleColor]::Blue
        "White" = [ConsoleColor]::White
    }
    
    Write-Host $Message -ForegroundColor $colors[$Color]
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

# 检测系统架构
function Get-SystemArchitecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        "x86" { return "386" }
        default {
            Write-Error "不支持的架构: $arch"
            exit 1
        }
    }
}

# 检查管理员权限
function Test-AdminRights {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查系统依赖..."
    
    # 检查PowerShell版本
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error "需要 PowerShell 5.0 或更高版本"
        exit 1
    }
    
    # 检查网络连接
    try {
        $null = Invoke-WebRequest -Uri "https://api.github.com" -UseBasicParsing -TimeoutSec 10
    }
    catch {
        Write-Error "无法连接到 GitHub，请检查网络连接"
        exit 1
    }
    
    Write-Success "依赖检查通过"
}

# 检查权限
function Test-Permissions {
    Write-Info "检查安装权限..."
    
    if (-not (Test-Path $InstallDir)) {
        try {
            New-Item -Path $InstallDir -ItemType Directory -Force | Out-Null
        }
        catch {
            Write-Error "无法创建安装目录: $InstallDir"
            Write-Info "请以管理员身份运行此脚本"
            exit 1
        }
    }
    
    # 测试写入权限
    $testFile = Join-Path $InstallDir "test.tmp"
    try {
        "test" | Out-File -FilePath $testFile -Force
        Remove-Item $testFile -Force
    }
    catch {
        Write-Error "没有写入权限: $InstallDir"
        Write-Info "请以管理员身份运行此脚本"
        exit 1
    }
    
    Write-Success "权限检查通过"
}

# 获取最新版本
function Get-LatestVersion {
    Write-Info "获取最新版本信息..."
    
    if ($Version -eq "latest") {
        try {
            $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$GitHubRepo/releases/latest"
            $script:Version = $response.tag_name
        }
        catch {
            Write-Error "无法获取最新版本信息: $_"
            exit 1
        }
    }
    
    Write-Info "目标版本: $Version"
}

# 下载二进制文件
function Get-Binary {
    Write-Info "下载 DelGuard 二进制文件..."
    
    $arch = Get-SystemArchitecture
    $binaryName = "delguard-windows-$arch.exe"
    $downloadUrl = "https://github.com/$GitHubRepo/releases/download/$Version/$binaryName"
    
    # 创建临时目录
    if (Test-Path $TempDir) {
        Remove-Item $TempDir -Recurse -Force
    }
    New-Item -Path $TempDir -ItemType Directory -Force | Out-Null
    
    $binaryPath = Join-Path $TempDir "delguard.exe"
    
    Write-Info "下载地址: $downloadUrl"
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $binaryPath -UseBasicParsing
    }
    catch {
        Write-Error "下载失败: $_"
        Remove-TempFiles
        exit 1
    }
    
    Write-Success "下载完成"
}

# 安装二进制文件
function Install-Binary {
    Write-Info "安装 DelGuard..."
    
    # 直接使用下载的二进制文件
    $binaryPath = Join-Path $TempDir "delguard.exe"
    
    if (-not (Test-Path $binaryPath)) {
        Write-Error "找不到二进制文件"
        Remove-TempFiles
        exit 1
    }
    
    # 复制到安装目录
    $targetPath = Join-Path $InstallDir "delguard.exe"
    try {
        Copy-Item $binaryPath $targetPath -Force
    }
    catch {
        Write-Error "安装失败: $_"
        Remove-TempFiles
        exit 1
    }
    
    Write-Success "DelGuard 已安装到 $targetPath"
}

# 添加到系统PATH
function Add-ToPath {
    Write-Info "添加到系统 PATH..."
    
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($currentPath -notlike "*$InstallDir*") {
        try {
            $newPath = "$currentPath;$InstallDir"
            [Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")
            Write-Success "已添加到系统 PATH"
        }
        catch {
            Write-Warning "无法添加到系统 PATH，请手动添加: $InstallDir"
        }
    }
    else {
        Write-Info "已存在于系统 PATH 中"
    }
}

# 创建配置目录
function New-Config {
    Write-Info "创建配置目录..."
    
    if (-not (Test-Path $ConfigDir)) {
        New-Item -Path $ConfigDir -ItemType Directory -Force | Out-Null
    }
    
    # 创建默认配置文件
    $configContent = @"
# DelGuard 配置文件
verbose: false
force: false
quiet: false

# 回收站设置
trash:
  auto_clean: false
  max_days: 30
  max_size: "1GB"

# 日志设置
log:
  level: "info"
  file: "$($ConfigDir -replace '\\', '/')/delguard.log"
"@
    
    $configPath = Join-Path $ConfigDir "config.yaml"
    $configContent | Out-File -FilePath $configPath -Encoding UTF8 -Force
    
    Write-Success "配置目录已创建: $ConfigDir"
}

# 配置PowerShell别名
function Set-PowerShellAliases {
    if ($NoAlias) {
        Write-Info "跳过别名配置"
        return
    }
    
    Write-Info "配置 PowerShell 别名..."
    
    $profilePath = $PROFILE.CurrentUserAllHosts
    $profileDir = Split-Path $profilePath -Parent
    
    # 创建配置文件目录
    if (-not (Test-Path $profileDir)) {
        New-Item -Path $profileDir -ItemType Directory -Force | Out-Null
    }
    
    # 别名配置内容
    $aliasContent = @"

# DelGuard 别名配置
Set-Alias -Name del -Value delguard
Set-Alias -Name trash -Value delguard
function rm { delguard delete @args }
function restore { delguard restore @args }
function empty-trash { delguard empty @args }
"@
    
    # 检查是否已经配置过
    if (Test-Path $profilePath) {
        $currentContent = Get-Content $profilePath -Raw
        if ($currentContent -notlike "*DelGuard 别名配置*") {
            Add-Content -Path $profilePath -Value $aliasContent
            Write-Success "已添加别名到 PowerShell 配置文件"
        }
        else {
            Write-Info "别名已存在于 PowerShell 配置文件中"
        }
    }
    else {
        $aliasContent | Out-File -FilePath $profilePath -Encoding UTF8 -Force
        Write-Success "已创建 PowerShell 配置文件并添加别名"
    }
}

# 验证安装
function Test-Installation {
    Write-Info "验证安装..."
    
    $binaryPath = Join-Path $InstallDir "delguard.exe"
    
    # 检查二进制文件
    if (-not (Test-Path $binaryPath)) {
        Write-Error "二进制文件不存在"
        return $false
    }
    
    # 测试运行
    try {
        $null = & $binaryPath --version 2>$null
        Write-Success "安装验证通过"
        return $true
    }
    catch {
        Write-Warning "无法运行 delguard --version，但文件已安装"
        return $true
    }
}

# 清理临时文件
function Remove-TempFiles {
    if (Test-Path $TempDir) {
        Remove-Item $TempDir -Recurse -Force
        Write-Info "已清理临时文件"
    }
}

# 显示安装完成信息
function Show-CompletionInfo {
    Write-Success "🎉 DelGuard 安装完成！"
    Write-Host ""
    Write-Host "📍 安装位置: $InstallDir\delguard.exe"
    Write-Host "📁 配置目录: $ConfigDir"
    Write-Host ""
    Write-Host "🚀 快速开始:"
    Write-Host "  delguard --help          # 查看帮助"
    Write-Host "  delguard delete <file>   # 删除文件到回收站"
    Write-Host "  delguard list           # 查看回收站内容"
    Write-Host "  delguard restore <file> # 恢复文件"
    Write-Host "  delguard empty          # 清空回收站"
    Write-Host ""
    
    if (-not $NoAlias) {
        Write-Host "💡 别名已配置 (重新打开 PowerShell 后生效):"
        Write-Host "  del <file>     # 等同于 delguard delete"
        Write-Host "  rm <file>      # 等同于 delguard delete (安全替代)"
        Write-Host "  restore <file> # 等同于 delguard restore"
        Write-Host "  empty-trash    # 等同于 delguard empty"
        Write-Host ""
    }
    
    Write-Host "📖 更多信息: https://github.com/$GitHubRepo"
    Write-Host ""
    Write-Warning "请重新打开 PowerShell 以使 PATH 和别名生效"
}

# 主函数
function Main {
    Write-Host "🛡️  DelGuard Windows 智能安装脚本" -ForegroundColor Cyan
    Write-Host "====================================" -ForegroundColor Cyan
    Write-Host ""
    
    try {
        # 检查系统环境
        Test-Dependencies
        Test-Permissions
        
        # 下载和安装
        Get-LatestVersion
        Get-Binary
        Install-Binary
        
        # 配置
        Add-ToPath
        New-Config
        Set-PowerShellAliases
        
        # 验证和清理
        Test-Installation
        Remove-TempFiles
        
        # 显示完成信息
        Show-CompletionInfo
    }
    catch {
        Write-Error "安装失败: $_"
        Remove-TempFiles
        exit 1
    }
}

# 运行主函数
Main