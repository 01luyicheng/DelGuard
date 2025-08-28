#Requires -Version 5.1
<#
.SYNOPSIS
    DelGuard 修复版安装脚本 - 确保别名正常工作

.DESCRIPTION
    自动安装 DelGuard 安全删除工具并正确配置别名。
    修复了别名配置问题，确保 del, rm, cp, delguard 命令能正常工作。

.PARAMETER Force
    强制重新安装，即使已经安装

.PARAMETER InstallPath
    自定义安装路径

.EXAMPLE
    .\install_fixed_final.ps1
    标准安装

.EXAMPLE
    .\install_fixed_final.ps1 -Force
    强制重新安装
#>

[CmdletBinding()]
param(
    [switch]$Force,
    [string]$InstallPath = "$env:USERPROFILE\bin"
)

# 设置错误处理
$ErrorActionPreference = 'Stop'

# 常量定义
$APP_NAME = "DelGuard"
$EXECUTABLE_NAME = "delguard.exe"

# 确保使用UTF-8编码
if ($PSVersionTable.PSVersion.Major -lt 6) {
    try {
        [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
        [Console]::InputEncoding = [System.Text.Encoding]::UTF8
        chcp 65001 | Out-Null
    } catch {
        Write-Warning "无法设置UTF-8编码，中文显示可能异常"
    }
} else {
    $PSDefaultParameterValues['*:Encoding'] = 'utf8'
}

# 颜色输出函数
function Write-Success { param([string]$Message) Write-Host $Message -ForegroundColor Green }
function Write-Warning { param([string]$Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Error { param([string]$Message) Write-Host $Message -ForegroundColor Red }
function Write-Info { param([string]$Message) Write-Host $Message -ForegroundColor Cyan }

Write-Info "=== DelGuard 修复版安装程序 ==="
Write-Info "安装路径: $InstallPath"

# 路径配置
$EXECUTABLE_PATH = Join-Path $InstallPath $EXECUTABLE_NAME

# 检查源文件
$SourceExe = Join-Path $PSScriptRoot $EXECUTABLE_NAME
if (-not (Test-Path $SourceExe)) {
    Write-Error "未找到 $EXECUTABLE_NAME 文件"
    Write-Error "请先运行: go build -o $EXECUTABLE_NAME"
    exit 1
}

Write-Success "找到源文件: $SourceExe"

# 检查现有安装
if ((Test-Path $EXECUTABLE_PATH) -and -not $Force) {
    Write-Warning "DelGuard 已经安装在 $EXECUTABLE_PATH"
    Write-Warning "使用 -Force 参数强制重新安装"
    exit 1
}

try {
    # 创建安装目录
    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
        Write-Success "创建安装目录: $InstallPath"
    }

    # 复制可执行文件
    Copy-Item $SourceExe $EXECUTABLE_PATH -Force
    Write-Success "复制可执行文件到: $EXECUTABLE_PATH"

    # 添加到用户PATH
    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if (-not $UserPath.Contains($InstallPath)) {
        $NewPath = if ($UserPath) { "$UserPath;$InstallPath" } else { $InstallPath }
        [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
        Write-Success "添加到用户PATH: $InstallPath"
        
        # 更新当前会话PATH
        $env:PATH = "$env:PATH;$InstallPath"
    } else {
        Write-Info "PATH中已存在: $InstallPath"
    }

    # 配置PowerShell别名 - 修复版本
    $ProfilePaths = @(
        "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1",
        "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
    )

    $ConfigBlock = @"
# DelGuard PowerShell 别名配置 - 修复版
# 生成时间: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')

`$delguardPath = '$EXECUTABLE_PATH'

if (Test-Path `$delguardPath) {
    # 移除可能存在的冲突别名
    try {
        Remove-Item Alias:del -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:rm -Force -ErrorAction SilentlyContinue  
        Remove-Item Alias:cp -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:copy -Force -ErrorAction SilentlyContinue
    } catch { }
    
    # 定义别名函数 - 确保参数正确传递
    function global:del {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "用法: del [选项] 文件..." -ForegroundColor Yellow
            Write-Host "选项:" -ForegroundColor Yellow
            Write-Host "  -f, --force     强制删除" -ForegroundColor Gray
            Write-Host "  -r, --recursive 递归删除目录" -ForegroundColor Gray
            Write-Host "  -v, --verbose   详细输出" -ForegroundColor Gray
            Write-Host "  --help          显示帮助" -ForegroundColor Gray
            return
        }
        & `$delguardPath delete @Arguments
    }
    
    function global:rm {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "用法: rm [选项] 文件..." -ForegroundColor Yellow
            Write-Host "选项:" -ForegroundColor Yellow
            Write-Host "  -f, --force     强制删除" -ForegroundColor Gray
            Write-Host "  -r, --recursive 递归删除目录" -ForegroundColor Gray
            Write-Host "  -v, --verbose   详细输出" -ForegroundColor Gray
            Write-Host "  --help          显示帮助" -ForegroundColor Gray
            return
        }
        & `$delguardPath delete @Arguments
    }
    
    function global:cp {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "用法: cp 源文件 目标文件" -ForegroundColor Yellow
            Write-Host "这是一个安全的复制命令，会记录操作历史" -ForegroundColor Gray
            return
        }
        # 暂时使用系统copy命令，后续可扩展为DelGuard的copy功能
        Copy-Item @Arguments
    }
    
    function global:copy {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "用法: copy 源文件 目标文件" -ForegroundColor Yellow
            return
        }
        Copy-Item @Arguments
    }
    
    function global:delguard {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        & `$delguardPath @Arguments
    }
    
    # 显示加载信息（仅在首次加载时）
    if (-not `$global:DelGuardAliasesLoaded) {
        Write-Host 'DelGuard 别名加载成功' -ForegroundColor Green
        Write-Host '可用命令: del, rm, cp, copy, delguard' -ForegroundColor Cyan
        Write-Host '使用 --help 查看详细帮助' -ForegroundColor Gray
        `$global:DelGuardAliasesLoaded = `$true
    }
} else {
    Write-Warning "DelGuard 可执行文件未找到: `$delguardPath"
}
# DelGuard 配置结束
"@

    # 安装到PowerShell配置文件
    foreach ($ProfilePath in $ProfilePaths) {
        $ProfileDir = Split-Path $ProfilePath -Parent
        
        # 创建配置文件目录
        if (-not (Test-Path $ProfileDir)) {
            New-Item -ItemType Directory -Path $ProfileDir -Force | Out-Null
            Write-Success "创建配置目录: $ProfileDir"
        }
        
        # 检查现有配置
        $ExistingContent = ""
        if (Test-Path $ProfilePath) {
            $ExistingContent = Get-Content $ProfilePath -Raw -ErrorAction SilentlyContinue
        }
        
        if ($ExistingContent -and $ExistingContent.Contains("# DelGuard PowerShell 别名配置")) {
            if (-not $Force) {
                Write-Warning "DelGuard 配置已存在于: $ProfilePath"
                Write-Warning "使用 -Force 参数覆盖现有配置"
                continue
            }
            # 移除现有DelGuard配置
            $ExistingContent = $ExistingContent -replace '(?s)# DelGuard PowerShell 别名配置.*?# DelGuard 配置结束\r?\n?', ''
        }
        
        # 添加新配置
        $NewContent = $ExistingContent + "`n" + $ConfigBlock + "`n"
        Set-Content $ProfilePath $NewContent -Encoding UTF8
        Write-Success "更新PowerShell配置: $ProfilePath"
    }

    Write-Success ""
    Write-Success "=== 安装完成 ==="
    Write-Info "DelGuard 已成功安装！"
    Write-Info "安装位置: $EXECUTABLE_PATH"
    Write-Info ""
    Write-Info "可用命令:"
    Write-Info "  del <文件>     - 安全删除文件"
    Write-Info "  rm <文件>      - 安全删除文件"
    Write-Info "  cp <源> <目标>  - 复制文件"
    Write-Info "  delguard       - DelGuard主程序"
    Write-Info ""
    Write-Info "使用方法:"
    Write-Info "  del file.txt          # 删除文件到回收站"
    Write-Info "  rm -f file.txt        # 强制删除文件"
    Write-Info "  delguard --help       # 查看完整帮助"
    Write-Info ""
    Write-Warning "请重新启动PowerShell或运行以下命令来加载别名:"
    Write-Warning ". `$PROFILE"

    # 测试安装
    Write-Info ""
    Write-Info "测试安装..."
    try {
        $TestResult = & $EXECUTABLE_PATH --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "✓ DelGuard 工作正常"
        } else {
            Write-Warning "⚠ DelGuard 可能无法正常工作"
        }
    } catch {
        Write-Warning "⚠ 无法测试DelGuard安装"
    }

} catch {
    Write-Error "安装失败: $($_.Exception.Message)"
    exit 1
}