#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 增强安装脚本 - Windows版本

.DESCRIPTION
    自动下载并安装 DelGuard 安全删除工具到系统中。
    支持 PowerShell 5.1+ 和 PowerShell 7+。
    增强功能：自动设置UTF-8编码、智能语言检测、环境兼容性检查。

.PARAMETER Force
    强制重新安装，即使已经安装

.PARAMETER SystemWide
    系统级安装（需要管理员权限）

.PARAMETER Uninstall
    卸载 DelGuard

.PARAMETER Status
    检查安装状态

.PARAMETER SetUtf8
    设置PowerShell为UTF-8编码（默认启用）

.PARAMETER NoSetUtf8
    不设置PowerShell为UTF-8编码

.EXAMPLE
    .\install_enhanced.ps1
    标准安装

.EXAMPLE
    .\install_enhanced.ps1 -Force
    强制重新安装

.EXAMPLE
    .\install_enhanced.ps1 -SystemWide
    系统级安装

.EXAMPLE
    .\install_enhanced.ps1 -Uninstall
    卸载 DelGuard
#>

[CmdletBinding()]
param(
    [switch]$Force,
    [switch]$SystemWide,
    [switch]$Uninstall,
    [switch]$Status,
    [switch]$SetUtf8 = $true,
    [switch]$NoSetUtf8
)

# 如果指定了NoSetUtf8，则覆盖SetUtf8的默认值
if ($NoSetUtf8) {
    $SetUtf8 = $false
}

# 设置错误处理
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# 常量定义
$REPO_URL = "https://github.com/01luyicheng/DelGuard"
$RELEASE_API = "https://api.github.com/repos/01luyicheng/DelGuard/releases/latest"
$APP_NAME = "DelGuard"
$EXECUTABLE_NAME = "delguard.exe"
$VERSION = "2.1.0"

# 路径配置
if ($SystemWide) {
    $INSTALL_DIR = "$env:ProgramFiles\$APP_NAME"
    $CONFIG_DIR = "$env:ProgramData\$APP_NAME"
} else {
    $INSTALL_DIR = "$env:LOCALAPPDATA\$APP_NAME"
    $CONFIG_DIR = "$env:APPDATA\$APP_NAME"
}

$EXECUTABLE_PATH = Join-Path $INSTALL_DIR $EXECUTABLE_NAME
$LOG_FILE = Join-Path $CONFIG_DIR "install.log"
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
# 检查系统环境
function Test-SystemEnvironment {
    Write-Log "检查系统环境..." "INFO"
    
    # 检查操作系统版本
    $osInfo = Get-CimInstance -ClassName Win32_OperatingSystem
    $osVersion = [Version]$osInfo.Version
    $osName = $osInfo.Caption
    
    Write-Log "操作系统: $osName ($osVersion)" "INFO"
    
    # 检查PowerShell版本
    $psVersion = $PSVersionTable.PSVersion
    Write-Log "PowerShell版本: $psVersion" "INFO"
    
    # 检查.NET版本
    $dotNetVersion = Get-ChildItem 'HKLM:\SOFTWARE\Microsoft\NET Framework Setup\NDP' -Recurse | 
                    Get-ItemProperty -Name Version -ErrorAction SilentlyContinue | 
                    Where-Object { $_.PSChildName -match '^(?!S)\p{L}'} | 
                    Select-Object -ExpandProperty Version -First 1
    
    if ($dotNetVersion) {
        Write-Log ".NET版本: $dotNetVersion" "INFO"
    } else {
        Write-Log "无法检测.NET版本" "WARNING"
    }
    
    # 检查磁盘空间
    $systemDrive = $env:SystemDrive
    $driveInfo = Get-PSDrive $systemDrive.TrimEnd(':')
    $freeSpaceMB = [math]::Round($driveInfo.Free / 1MB, 2)
    
    Write-Log "系统盘 $systemDrive 可用空间: $freeSpaceMB MB" "INFO"
    
    if ($freeSpaceMB -lt 100) {
        Write-Log "系统盘空间不足，建议至少保留100MB空间" "WARNING"
    }
    
    # 检查是否有防病毒软件可能阻止安装
    $avProducts = Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntiVirusProduct -ErrorAction SilentlyContinue
    
    if ($avProducts) {
        foreach ($av in $avProducts) {
            Write-Log "检测到防病毒软件: $($av.displayName)" "INFO"
        }
        Write-Log "如果安装过程被阻止，请考虑暂时禁用防病毒软件" "INFO"
    }
    
    # 检查是否有其他程序占用端口
    $requiredPorts = @(8080, 8081) # 假设DelGuard使用这些端口
    foreach ($port in $requiredPorts) {
        $portInUse = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
        if ($portInUse) {
            Write-Log "端口 $port 已被占用，可能会影响DelGuard的某些功能" "WARNING"
        }
    }
    
    # 检查系统语言
    $currentCulture = [System.Globalization.CultureInfo]::CurrentCulture
    $currentUICulture = [System.Globalization.CultureInfo]::CurrentUICulture
    
    Write-Log "系统区域设置: $($currentCulture.Name)" "INFO"
    Write-Log "系统UI语言: $($currentUICulture.Name)" "INFO"
    
    # 检查是否支持UTF-8
    $utf8Support = [System.Text.Encoding]::UTF8.GetString([System.Text.Encoding]::UTF8.GetBytes("测试UTF-8支持")) -eq "测试UTF-8支持"
    if ($utf8Support) {
        Write-Log "系统支持UTF-8编码" "SUCCESS"
    } else {
        Write-Log "系统可能不完全支持UTF-8编码，可能导致中文显示问题" "WARNING"
    }
    
    Write-Log "系统环境检查完成" "SUCCESS"
}

# 设置PowerShell为UTF-8编码
function Set-PowerShellUtf8Encoding {
    Write-Log "配置PowerShell UTF-8编码..." "INFO"
    
    # 检查PowerShell版本
    $psVersion = $PSVersionTable.PSVersion
    
    if ($psVersion.Major -ge 7) {
        Write-Log "PowerShell 7+ 默认支持UTF-8，无需额外配置" "SUCCESS"
        return
    }
    
    # 为PowerShell 5.1配置UTF-8
    try {
        # 检查是否已经配置了UTF-8
        $profilePath = $PROFILE.CurrentUserAllHosts
        $profileExists = Test-Path $profilePath
        
        if ($profileExists) {
            $profileContent = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
            if ($profileContent -like "*[Console]::OutputEncoding = [System.Text.Encoding]::UTF8*") {
                Write-Log "PowerShell UTF-8编码已配置" "SUCCESS"
                return
            }
        }
        
        # 创建或更新配置文件
        if (-not $profileExists) {
            $profileDir = Split-Path $profilePath -Parent
            if (-not (Test-Path $profileDir)) {
                New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
            }
            New-Item -ItemType File -Path $profilePath -Force | Out-Null
        }
        
        # 添加UTF-8配置
        $utf8Config = @"

# DelGuard UTF-8编码配置
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::InputEncoding = [System.Text.Encoding]::UTF8
`$OutputEncoding = [System.Text.Encoding]::UTF8
# 设置默认代码页为UTF-8
chcp 65001 > `$null
"@
        
        Add-Content -Path $profilePath -Value $utf8Config -Encoding UTF8
        
        # 立即应用UTF-8设置
        [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
        [Console]::InputEncoding = [System.Text.Encoding]::UTF8
        $OutputEncoding = [System.Text.Encoding]::UTF8
        
        # 尝试设置代码页
        try {
            chcp 65001 > $null
            Write-Log "已设置当前会话的代码页为UTF-8 (65001)" "SUCCESS"
        } catch {
            Write-Log "无法设置代码页，但UTF-8编码已配置" "WARNING"
        }
        
        Write-Log "PowerShell UTF-8编码配置成功" "SUCCESS"
        Write-Log "请重新启动PowerShell以完全应用UTF-8设置" "INFO"
        
    } catch {
        Write-Log "配置UTF-8编码失败: $($_.Exception.Message)" "ERROR"
        Write-Log "请手动编辑 $profilePath 添加UTF-8配置" "INFO"
    }
}

# 检测系统语言并设置DelGuard语言
function Set-DelGuardLanguage {
    Write-Log "检测系统语言..." "INFO"
    
    # 获取系统UI语言
    $currentUICulture = [System.Globalization.CultureInfo]::CurrentUICulture
    $languageCode = $currentUICulture.Name
    
    Write-Log "检测到系统UI语言: $languageCode" "INFO"
    
    # 确定DelGuard使用的语言
    $delguardLang = "en-US" # 默认英语
    
    if ($languageCode -like "zh*") {
        $delguardLang = "zh-CN"
        Write-Log "将使用中文(简体)作为DelGuard界面语言" "INFO"
    } elseif ($languageCode -like "ja*") {
        $delguardLang = "ja"
        Write-Log "将使用日语作为DelGuard界面语言" "INFO"
    } else {
        Write-Log "将使用英语作为DelGuard界面语言" "INFO"
    }
    
    # 创建或更新DelGuard语言配置
    $configFile = Join-Path $CONFIG_DIR "config.json"
    
    try {
        # 确保配置目录存在
        if (!(Test-Path $CONFIG_DIR)) {
            New-Item -ItemType Directory -Path $CONFIG_DIR -Force | Out-Null
        }
        
        # 读取现有配置（如果存在）
        $config = @{}
        if (Test-Path $configFile) {
            $configContent = Get-Content $configFile -Raw -ErrorAction SilentlyContinue
            if ($configContent) {
                try {
                    $config = $configContent | ConvertFrom-Json -AsHashtable
                } catch {
                    Write-Log "现有配置文件格式无效，将创建新配置" "WARNING"
                    $config = @{}
                }
            }
        }
        
        # 更新语言设置
        $config["language"] = $delguardLang
        
        # 保存配置
        $config | ConvertTo-Json -Depth 10 | Set-Content -Path $configFile -Encoding UTF8
        
        Write-Log "DelGuard语言配置已更新为: $delguardLang" "SUCCESS"
        
    } catch {
        Write-Log "配置DelGuard语言失败: $($_.Exception.Message)" "ERROR"
        Write-Log "DelGuard将使用默认语言设置" "INFO"
    }
}
# 安装 DelGuard
function Install-DelGuard {
    Write-Log "开始安装 $APP_NAME..." "INFO"
    
    # 检查管理员权限（系统级安装时）
    if ($SystemWide -and !(Test-Administrator)) {
        Write-Log "系统级安装需要管理员权限" "ERROR"
        throw "请以管理员身份运行 PowerShell"
    }
    
    # 检查系统环境
    Test-SystemEnvironment
    
    # 设置UTF-8编码（如果启用）
    if ($SetUtf8) {
        Set-PowerShellUtf8Encoding
    }
    
    # 检查网络连接
    if (!(Test-NetworkConnection)) {
        Write-Log "网络连接检查失败" "ERROR"
        throw "无法连接到 GitHub，请检查网络连接"
    }
    
    # 检查现有安装
    if ((Test-Path $EXECUTABLE_PATH) -and !$Force) {
        Write-Log "$APP_NAME 已经安装在 $EXECUTABLE_PATH" "WARNING"
        Write-Log "使用 -Force 参数强制重新安装" "INFO"
        return
    }
    
    try {
        # 获取最新版本
        $release = Get-LatestRelease
        $version = $release.tag_name
        Write-Log "最新版本: $version" "SUCCESS"
        
        # 确定下载URL
        $arch = Get-SystemArchitecture
        $assetName = "$APP_NAME-windows-$arch.zip"
        $asset = $release.assets | Where-Object { $_.name -eq $assetName }
        
        if (!$asset) {
            Write-Log "未找到适合的安装包: $assetName" "ERROR"
            throw "未找到适合当前系统的安装包"
        }
        
        $downloadUrl = $asset.browser_download_url
        Write-Log "下载URL: $downloadUrl" "INFO"
        
        # 创建临时目录
        $tempDir = Join-Path $env:TEMP "delguard-install"
        if (Test-Path $tempDir) {
            Remove-Item $tempDir -Recurse -Force
        }
        New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
        
        # 下载文件
        $zipPath = Join-Path $tempDir "$assetName"
        Download-File -Url $downloadUrl -OutputPath $zipPath
        
        # 解压文件
        Write-Log "解压安装包..." "INFO"
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
        
        # 创建安装目录
        if (!(Test-Path $INSTALL_DIR)) {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        }
        
        # 复制文件
        $extractedExe = Get-ChildItem -Path $tempDir -Filter $EXECUTABLE_NAME -Recurse | Select-Object -First 1
        if ($extractedExe) {
            Copy-Item -Path $extractedExe.FullName -Destination $EXECUTABLE_PATH -Force
            Write-Log "已安装到: $EXECUTABLE_PATH" "SUCCESS"
        } else {
            throw "在安装包中未找到可执行文件"
        }
        
        # 添加到 PATH
        Add-ToPath -Path $INSTALL_DIR
        
        # 安装 PowerShell 别名
        Install-PowerShellAliases
        
        # 创建配置目录
        if (!(Test-Path $CONFIG_DIR)) {
            New-Item -ItemType Directory -Path $CONFIG_DIR -Force | Out-Null
        }
        
        # 设置DelGuard语言
        Set-DelGuardLanguage
        
        # 清理临时文件
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        
        Write-Log "$APP_NAME $version 安装成功！" "SUCCESS"
        Write-Log "可执行文件位置: $EXECUTABLE_PATH" "INFO"
        Write-Log "配置目录: $CONFIG_DIR" "INFO"
        Write-Log "" "INFO"
        Write-Log "使用方法:" "INFO"
        Write-Log "  delguard file.txt          # 删除文件到回收站" "INFO"
        Write-Log "  delguard -p file.txt       # 永久删除文件" "INFO"
        Write-Log "  delguard --help            # 查看帮助" "INFO"
        Write-Log "" "INFO"
        Write-Log "请重新启动 PowerShell 以使用 delguard 命令" "INFO"
        
    } catch {
        Write-Log "安装失败: $($_.Exception.Message)" "ERROR"
        throw
    }
}

# 添加到 PATH
function Add-ToPath {
    param([string]$Path)
    
    try {
        if ($SystemWide) {
            $envPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
            $target = "Machine"
        } else {
            $envPath = [Environment]::GetEnvironmentVariable("PATH", "User")
            $target = "User"
        }
        
        if ($envPath -notlike "*$Path*") {
            $newPath = "$envPath;$Path"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, $target)
            Write-Log "已添加到 PATH: $Path" "SUCCESS"
            
            # 更新当前会话的 PATH
            $env:PATH = "$env:PATH;$Path"
        } else {
            Write-Log "PATH 中已存在: $Path" "INFO"
        }
    } catch {
        Write-Log "添加到 PATH 失败: $($_.Exception.Message)" "WARNING"
    }
}

# 安装 PowerShell 别名
function Install-PowerShellAliases {
    try {
        $profilePath = $PROFILE.CurrentUserAllHosts
        
        if (!(Test-Path $profilePath)) {
            New-Item -ItemType File -Path $profilePath -Force | Out-Null
        }
        
        $aliasContent = @"

# DelGuard 别名配置
if (Test-Path "$EXECUTABLE_PATH") {
    Set-Alias -Name delguard -Value "$EXECUTABLE_PATH" -Scope Global
    Set-Alias -Name dg -Value "$EXECUTABLE_PATH" -Scope Global
    # 兼容Unix命令
    Set-Alias -Name rm -Value "$EXECUTABLE_PATH" -Scope Global
    Set-Alias -Name del -Value "$EXECUTABLE_PATH" -Scope Global
}
"@
        
        $currentContent = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
        if ($currentContent -notlike "*DelGuard 别名配置*") {
            Add-Content -Path $profilePath -Value $aliasContent -Encoding UTF8
            Write-Log "已添加 PowerShell 别名配置" "SUCCESS"
        } else {
            Write-Log "PowerShell 别名已存在" "INFO"
        }
    } catch {
        Write-Log "配置 PowerShell 别名失败: $($_.Exception.Message)" "WARNING"
    }
}

# 卸载 DelGuard
function Uninstall-DelGuard {
    Write-Log "开始卸载 $APP_NAME..." "INFO"
    
    try {
        # 删除可执行文件
        if (Test-Path $EXECUTABLE_PATH) {
            Remove-Item $EXECUTABLE_PATH -Force
            Write-Log "已删除: $EXECUTABLE_PATH" "SUCCESS"
        }
        
        # 删除安装目录（如果为空）
        if ((Test-Path $INSTALL_DIR) -and ((Get-ChildItem $INSTALL_DIR).Count -eq 0)) {
            Remove-Item $INSTALL_DIR -Force
            Write-Log "已删除安装目录: $INSTALL_DIR" "SUCCESS"
        }
        
        # 从 PATH 中移除
        Remove-FromPath -Path $INSTALL_DIR
        
        # 移除 PowerShell 别名
        Remove-PowerShellAliases
        
        Write-Log "$APP_NAME 卸载完成" "SUCCESS"
        Write-Log "配置文件保留在: $CONFIG_DIR" "INFO"
        Write-Log "如需完全清理，请手动删除配置目录" "INFO"
        
    } catch {
        Write-Log "卸载失败: $($_.Exception.Message)" "ERROR"
        throw
    }
}

# 从 PATH 中移除
function Remove-FromPath {
    param([string]$Path)
    
    try {
        if ($SystemWide) {
            $envPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
            $target = "Machine"
        } else {
            $envPath = [Environment]::GetEnvironmentVariable("PATH", "User")
            $target = "User"
        }
        
        if ($envPath -like "*$Path*") {
            $newPath = $envPath -replace [regex]::Escape(";$Path"), ""
            $newPath = $newPath -replace [regex]::Escape("$Path;"), ""
            $newPath = $newPath -replace [regex]::Escape($Path), ""
            [Environment]::SetEnvironmentVariable("PATH", $newPath, $target)
            Write-Log "已从 PATH 中移除: $Path" "SUCCESS"
        }
    } catch {
        Write-Log "从 PATH 中移除失败: $($_.Exception.Message)" "WARNING"
    }
}

# 移除 PowerShell 别名
function Remove-PowerShellAliases {
    try {
        $profilePath = $PROFILE.CurrentUserAllHosts
        
        if (Test-Path $profilePath) {
            $content = Get-Content $profilePath -Raw
            $newContent = $content -replace "(?s)# DelGuard 别名配置.*?(?=\r?\n\r?\n|\r?\n$|$)", ""
            $newContent = $newContent.Trim()
            
            if ($newContent) {
                Set-Content -Path $profilePath -Value $newContent -Encoding UTF8
            } else {
                Remove-Item $profilePath -Force
            }
            Write-Log "已移除 PowerShell 别名配置" "SUCCESS"
        }
    } catch {
        Write-Log "移除 PowerShell 别名失败: $($_.Exception.Message)" "WARNING"
    }
}
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
