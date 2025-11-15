/*
 * 文件用途：DelGuard PowerShell安装脚本
 * 作者：TREA
 * 功能说明：自动化安装DelGuard的PowerShell脚本，支持在线安装
 * 输入：GitHub发布版本或本地安装包
 * 输出：安装成功状态
 * 依赖：PowerShell 5.1+, .NET 8.0 Runtime
 * 交互方式：PowerShell命令行执行
 */

#Requires -Version 5.1
#Requires -RunAsAdministrator

<#
.SYNOPSIS
    DelGuard自动化安装脚本

.DESCRIPTION
    该脚本用于自动下载和安装DelGuard文件删除工具
    支持从GitHub Releases下载最新版本或安装本地MSI包

.PARAMETER LocalMsi
    本地MSI安装包路径，如果提供则直接安装本地包

.PARAMETER Version
    要安装的版本号，默认安装最新版本

.PARAMETER InstallPath
    自定义安装路径，默认使用程序文件目录

.EXAMPLE
    .\install.ps1
    安装最新版本的DelGuard

.EXAMPLE
    .\install.ps1 -LocalMsi .\DelGuard.msi
    安装本地的DelGuard.msi包

.EXAMPLE
    .\install.ps1 -Version "1.0.0"
    安装指定版本的DelGuard
#>

param(
    [string]$LocalMsi = "",
    [string]$Version = "latest",
    [string]$InstallPath = ""
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# 配置变量
$GitHubRepo = "01luyicheng/DelGuard"
$ProductName = "DelGuard"
$TempDir = $env:TEMP
$LogFile = Join-Path $TempDir "DelGuard_Install_$(Get-Date -Format 'yyyyMMdd_HHmmss').log"

# 日志函数
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logEntry = "[$timestamp] [$Level] $Message"
    Write-Host $logEntry
    Add-Content -Path $LogFile -Value $logEntry
}

function Test-Administrator {
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    return $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-LatestVersion {
    try {
        $apiUrl = "https://api.github.com/repos/$GitHubRepo/releases/latest"
        $response = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        return $response.tag_name
    }
    catch {
        Write-Log "获取最新版本失败: $($_.Exception.Message)" "ERROR"
        return $null
    }
}

function Download-Installer {
    param([string]$Version)
    
    if ($Version -eq "latest") {
        $Version = Get-LatestVersion
        if (-not $Version) {
            throw "无法获取最新版本信息"
        }
    }
    
    $fileName = "DelGuard.$Version.msi"
    $downloadUrl = "https://github.com/$GitHubRepo/releases/download/$Version/$fileName"
    $localPath = Join-Path $TempDir $fileName
    
    Write-Log "正在下载 DelGuard $Version..."
    Write-Log "下载地址: $downloadUrl"
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $localPath -UseBasicParsing
        Write-Log "下载完成: $localPath"
        return $localPath
    }
    catch {
        Write-Log "下载失败: $($_.Exception.Message)" "ERROR"
        throw
    }
}

function Install-DelGuard {
    param([string]$MsiPath)
    
    Write-Log "开始安装 DelGuard..."
    Write-Log "安装包路径: $MsiPath"
    
    try {
        # 检查安装包是否存在
        if (-not (Test-Path $MsiPath)) {
            throw "安装包不存在: $MsiPath"
        }
        
        # 执行安装
        $arguments = @(
            "/i", $MsiPath
            "/qn"  # 静默安装
            "/norestart"
            "/l*v", (Join-Path $TempDir "DelGuard_MSI_Install.log")
        )
        
        Write-Log "执行安装命令: msiexec $arguments"
        $process = Start-Process -FilePath "msiexec.exe" -ArgumentList $arguments -Wait -PassThru
        
        if ($process.ExitCode -eq 0) {
            Write-Log "DelGuard安装成功！"
            return $true
        }
        else {
            Write-Log "安装失败，退出码: $($process.ExitCode)" "ERROR"
            return $false
        }
    }
    catch {
        Write-Log "安装过程中发生错误: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

function Test-Installation {
    try {
        # 检查安装目录
        $installDir = Join-Path $env:ProgramFiles "DelGuard"
        if (Test-Path $installDir) {
            Write-Log "安装目录存在: $installDir"
            
            # 检查可执行文件
            $exePath = Join-Path $installDir "delguard.exe"
            if (Test-Path $exePath) {
                Write-Log "可执行文件存在: $exePath"
                
                # 测试命令
                $testResult = & "$exePath" --help 2>&1
                if ($LASTEXITCODE -eq 0) {
                    Write-Log "功能测试通过"
                    return $true
                }
                else {
                    Write-Log "功能测试失败" "WARNING"
                }
            }
            else {
                Write-Log "可执行文件不存在" "WARNING"
            }
        }
        else {
            Write-Log "安装目录不存在" "WARNING"
        }
        
        return $false
    }
    catch {
        Write-Log "安装验证失败: $($_.Exception.Message)" "ERROR"
        return $false
    }
}

function Add-ToPath {
    $installDir = Join-Path $env:ProgramFiles "DelGuard"
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    
    if ($currentPath -notlike "*$installDir*") {
        Write-Log "正在添加到系统PATH..."
        $newPath = $currentPath + ";" + $installDir
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
        Write-Log "已添加到系统PATH"
    }
    else {
        Write-Log "已在PATH中，跳过添加"
    }
}

# 主安装流程
function Main {
    Write-Log "开始 DelGuard 安装流程"
    Write-Log "PowerShell版本: $($PSVersionTable.PSVersion)"
    Write-Log "操作系统: $([Environment]::OSVersion)"
    
    # 检查管理员权限
    if (-not (Test-Administrator)) {
        Write-Log "请以管理员身份运行此脚本" "ERROR"
        Write-Host "请右键点击脚本，选择\"以管理员身份运行\""
        exit 1
    }
    
    # 检查.NET 8.0运行时
    try {
        $dotnetVersion = & dotnet --version 2>$null
        Write-Log ".NET运行时版本: $dotnetVersion"
    }
    catch {
        Write-Log "未检测到.NET运行时，请确保已安装.NET 8.0或更高版本" "WARNING"
    }
    
    try {
        $msiPath = $LocalMsi
        
        # 如果没有提供本地MSI，则下载
        if ([string]::IsNullOrEmpty($msiPath)) {
            $msiPath = Download-Installer -Version $Version
        }
        elseif (-not (Test-Path $msiPath)) {
            throw "指定的本地安装包不存在: $msiPath"
        }
        
        # 执行安装
        $installSuccess = Install-DelGuard -MsiPath $msiPath
        
        if ($installSuccess) {
            # 验证安装
            $testSuccess = Test-Installation
            
            if ($testSuccess) {
                Write-Log "安装验证通过！"
                
                # 添加到PATH
                Add-ToPath
                
                Write-Host "`n================================" -ForegroundColor Green
                Write-Host "DelGuard 安装成功！" -ForegroundColor Green
                Write-Host "================================" -ForegroundColor Green
                Write-Host "`n使用方法:" -ForegroundColor Yellow
                Write-Host "  delguard --help          # 查看帮助" -ForegroundColor White
                Write-Host "  delguard file.txt        # 删除文件" -ForegroundColor White
                Write-Host "  delguard -f locked.exe   # 强制删除" -ForegroundColor White
                Write-Host "`n详细说明请查看: https://github.com/$GitHubRepo" -ForegroundColor Cyan
                Write-Host "日志文件: $LogFile" -ForegroundColor Gray
                
                exit 0
            }
            else {
                Write-Log "安装验证失败" "ERROR"
                exit 1
            }
        }
        else {
            Write-Log "安装失败" "ERROR"
            exit 1
        }
    }
    catch {
        Write-Log "安装脚本执行失败: $($_.Exception.Message)" "ERROR"
        Write-Host "`n安装失败，请查看日志文件获取详细信息: $LogFile" -ForegroundColor Red
        exit 1
    }
    finally {
        # 清理临时文件
        if ([string]::IsNullOrEmpty($LocalMsi) -and (Test-Path $msiPath)) {
            Write-Log "清理临时文件..."
            Remove-Item $msiPath -Force -ErrorAction SilentlyContinue
        }
    }
}

# 执行主函数
Main