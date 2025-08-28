# DelGuard Windows 安装脚本
# 版本: 2.0.0
# 描述: 自动安装DelGuard文件删除保护工具

param(
    [string]$InstallPath = "$env:ProgramFiles\DelGuard",
    [string]$ServiceName = "DelGuardService",
    [switch]$Silent = $false,
    [switch]$CreateDesktopShortcut = $true,
    [switch]$AddToPath = $true,
    [switch]$StartService = $true
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput $Message "Green" }
function Write-Warning { param([string]$Message) Write-ColorOutput $Message "Yellow" }
function Write-Error { param([string]$Message) Write-ColorOutput $Message "Red" }
function Write-Info { param([string]$Message) Write-ColorOutput $Message "Cyan" }

# 检查管理员权限
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 获取系统信息
function Get-SystemInfo {
    return @{
        OS = (Get-WmiObject -Class Win32_OperatingSystem).Caption
        Architecture = $env:PROCESSOR_ARCHITECTURE
        PowerShellVersion = $PSVersionTable.PSVersion.ToString()
        DotNetVersion = [System.Runtime.InteropServices.RuntimeInformation]::FrameworkDescription
    }
}

# 创建安装目录
function New-InstallDirectory {
    param([string]$Path)
    
    Write-Info "创建安装目录: $Path"
    
    if (Test-Path $Path) {
        Write-Warning "安装目录已存在，将进行覆盖安装"
        # 停止服务（如果正在运行）
        Stop-DelGuardService -Silent
    } else {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }
    
    Write-Success "安装目录创建成功"
}

# 构建可执行文件
function Build-DelGuard {
    param([string]$ProjectRoot)
    
    Write-Info "构建DelGuard可执行文件..."
    
    # 检查Go环境
    if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
        throw "未找到Go编译器，请先安装Go语言环境"
    }
    
    # 设置工作目录
    Push-Location $ProjectRoot
    
    try {
        # 下载依赖
        Write-Info "下载Go模块依赖..."
        & go mod download
        & go mod tidy
        
        # 构建Windows可执行文件
        Write-Info "编译Windows可执行文件..."
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        & go build -ldflags "-s -w" -o "delguard.exe" "./cmd/delguard"
        
        if ($LASTEXITCODE -ne 0) {
            throw "编译失败"
        }
        
        Write-Success "DelGuard编译成功"
    }
    finally {
        Pop-Location
    }
}

# 复制文件
function Copy-DelGuardFiles {
    param([string]$SourcePath, [string]$DestPath)
    
    Write-Info "复制DelGuard文件到安装目录..."
    
    # 获取当前脚本目录
    $ScriptDir = Split-Path -Parent $MyInvocation.ScriptName
    $ProjectRoot = Split-Path -Parent $ScriptDir
    
    # 检查并构建可执行文件
    $ExePath = Join-Path $ProjectRoot "delguard.exe"
    if (-not (Test-Path $ExePath)) {
        Write-Warning "未找到预编译的可执行文件，尝试构建..."
        Build-DelGuard -ProjectRoot $ProjectRoot
    }
    
    # 复制主程序
    if (Test-Path $ExePath) {
        Copy-Item $ExePath $DestPath -Force
        Write-Success "主程序复制成功"
    } else {
        throw "找不到DelGuard主程序: $ExePath"
    }
    
    # 复制配置文件
    $ConfigDir = Join-Path $ProjectRoot "configs"
    if (Test-Path $ConfigDir) {
        $DestConfigDir = Join-Path $DestPath "configs"
        Copy-Item $ConfigDir $DestConfigDir -Recurse -Force
        Write-Success "配置文件复制成功"
    }
    
    # 复制文档
    $DocsDir = Join-Path $ProjectRoot "docs"
    if (Test-Path $DocsDir) {
        $DestDocsDir = Join-Path $DestPath "docs"
        Copy-Item $DocsDir $DestDocsDir -Recurse -Force
        Write-Success "文档文件复制成功"
    }
}

# 创建配置文件
function New-ConfigFile {
    param([string]$InstallPath)
    
    Write-Info "创建配置文件..."
    
    $ConfigPath = Join-Path $InstallPath "config.yaml"
    $ConfigContent = @"
# DelGuard 配置文件
# 自动生成于: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

app:
  name: "DelGuard"
  version: "2.0.0"
  log_level: "info"
  data_dir: "$InstallPath\data"

monitor:
  enabled: true
  watch_paths:
    - "C:\Users"
    - "D:\"
  exclude_paths:
    - "C:\Windows"
    - "C:\Program Files"
  file_types:
    - ".doc"
    - ".docx"
    - ".xls"
    - ".xlsx"
    - ".ppt"
    - ".pptx"
    - ".pdf"
    - ".txt"
    - ".jpg"
    - ".png"
    - ".mp4"
    - ".mp3"

restore:
  backup_dir: "$InstallPath\backups"
  max_backup_size: "10GB"
  retention_days: 30

search:
  index_enabled: true
  index_update_interval: "1h"
  max_results: 1000

security:
  enable_encryption: true
  require_admin: false
  audit_log: true
"@
    
    $ConfigContent | Out-File -FilePath $ConfigPath -Encoding UTF8
    Write-Success "配置文件创建成功: $ConfigPath"
}

# 注册Windows服务
function Register-DelGuardService {
    param([string]$InstallPath, [string]$ServiceName)
    
    Write-Info "注册Windows服务..."
    
    $ExePath = Join-Path $InstallPath "delguard.exe"
    $ServiceDescription = "DelGuard文件删除保护服务"
    
    # 检查服务是否已存在
    $ExistingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($ExistingService) {
        Write-Warning "服务已存在，正在更新..."
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        & sc.exe delete $ServiceName
        Start-Sleep -Seconds 2
    }
    
    # 创建服务
    $CreateResult = & sc.exe create $ServiceName binPath= "`"$ExePath`" service" start= auto DisplayName= "DelGuard File Protection Service" depend= ""
    
    if ($LASTEXITCODE -eq 0) {
        # 设置服务描述
        & sc.exe description $ServiceName $ServiceDescription
        
        # 设置服务恢复选项
        & sc.exe failure $ServiceName reset= 86400 actions= restart/5000/restart/10000/restart/20000
        
        Write-Success "Windows服务注册成功"
    } else {
        throw "服务注册失败: $CreateResult"
    }
}

# 启动服务
function Start-DelGuardService {
    param([string]$ServiceName)
    
    Write-Info "启动DelGuard服务..."
    
    try {
        Start-Service -Name $ServiceName
        Write-Success "服务启动成功"
        
        # 验证服务状态
        $Service = Get-Service -Name $ServiceName
        if ($Service.Status -eq "Running") {
            Write-Success "服务运行状态正常"
        } else {
            Write-Warning "服务状态异常: $($Service.Status)"
        }
    } catch {
        Write-Error "服务启动失败: $($_.Exception.Message)"
        throw
    }
}

# 停止服务
function Stop-DelGuardService {
    param([string]$ServiceName = $ServiceName, [switch]$Silent)
    
    $Service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($Service -and $Service.Status -eq "Running") {
        if (-not $Silent) {
            Write-Info "停止DelGuard服务..."
        }
        Stop-Service -Name $ServiceName -Force
        if (-not $Silent) {
            Write-Success "服务已停止"
        }
    }
}

# 添加到系统PATH
function Add-ToSystemPath {
    param([string]$InstallPath)
    
    Write-Info "添加到系统PATH环境变量..."
    
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($CurrentPath -notlike "*$InstallPath*") {
        $NewPath = $CurrentPath + ";" + $InstallPath
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "Machine")
        Write-Success "已添加到系统PATH"
    } else {
        Write-Info "PATH中已存在安装路径"
    }
}

# 创建桌面快捷方式
function New-DesktopShortcut {
    param([string]$InstallPath)
    
    Write-Info "创建桌面快捷方式..."
    
    $WshShell = New-Object -ComObject WScript.Shell
    $DesktopPath = [Environment]::GetFolderPath("Desktop")
    $ShortcutPath = Join-Path $DesktopPath "DelGuard.lnk"
    $ExePath = Join-Path $InstallPath "delguard.exe"
    
    $Shortcut = $WshShell.CreateShortcut($ShortcutPath)
    $Shortcut.TargetPath = $ExePath
    $Shortcut.WorkingDirectory = $InstallPath
    $Shortcut.Description = "DelGuard文件删除保护工具"
    $Shortcut.Save()
    
    Write-Success "桌面快捷方式创建成功"
}

# 创建开始菜单项
function New-StartMenuShortcut {
    param([string]$InstallPath)
    
    Write-Info "创建开始菜单项..."
    
    $WshShell = New-Object -ComObject WScript.Shell
    $StartMenuPath = Join-Path $env:ProgramData "Microsoft\Windows\Start Menu\Programs"
    $DelGuardFolder = Join-Path $StartMenuPath "DelGuard"
    
    if (-not (Test-Path $DelGuardFolder)) {
        New-Item -ItemType Directory -Path $DelGuardFolder -Force | Out-Null
    }
    
    $ExePath = Join-Path $InstallPath "delguard.exe"
    
    # 主程序快捷方式
    $MainShortcut = $WshShell.CreateShortcut((Join-Path $DelGuardFolder "DelGuard.lnk"))
    $MainShortcut.TargetPath = $ExePath
    $MainShortcut.WorkingDirectory = $InstallPath
    $MainShortcut.Description = "DelGuard文件删除保护工具"
    $MainShortcut.Save()
    
    # 卸载快捷方式
    $UninstallShortcut = $WshShell.CreateShortcut((Join-Path $DelGuardFolder "卸载DelGuard.lnk"))
    $UninstallShortcut.TargetPath = "powershell.exe"
    $UninstallShortcut.Arguments = "-ExecutionPolicy Bypass -File `"$(Join-Path $InstallPath 'uninstall.ps1')`""
    $UninstallShortcut.WorkingDirectory = $InstallPath
    $UninstallShortcut.Description = "卸载DelGuard"
    $UninstallShortcut.Save()
    
    Write-Success "开始菜单项创建成功"
}

# 创建卸载脚本
function New-UninstallScript {
    param([string]$InstallPath, [string]$ServiceName)
    
    Write-Info "创建卸载脚本..."
    
    $UninstallScript = @"
# DelGuard 卸载脚本
param([switch]`$Silent = `$false)

if (-not `$Silent) {
    `$Confirm = Read-Host "确定要卸载DelGuard吗？(y/N)"
    if (`$Confirm -ne "y" -and `$Confirm -ne "Y") {
        Write-Host "取消卸载"
        exit 0
    }
}

Write-Host "正在卸载DelGuard..." -ForegroundColor Yellow

# 停止并删除服务
try {
    Stop-Service -Name "$ServiceName" -Force -ErrorAction SilentlyContinue
    & sc.exe delete "$ServiceName"
    Write-Host "服务已删除" -ForegroundColor Green
} catch {
    Write-Host "删除服务时出错: `$(`$_.Exception.Message)" -ForegroundColor Red
}

# 从PATH中移除
try {
    `$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    `$NewPath = `$CurrentPath -replace ";$InstallPath", "" -replace "$InstallPath;", "" -replace "$InstallPath", ""
    [Environment]::SetEnvironmentVariable("Path", `$NewPath, "Machine")
    Write-Host "已从PATH中移除" -ForegroundColor Green
} catch {
    Write-Host "从PATH移除时出错: `$(`$_.Exception.Message)" -ForegroundColor Red
}

# 删除快捷方式
try {
    Remove-Item "`$env:PUBLIC\Desktop\DelGuard.lnk" -ErrorAction SilentlyContinue
    Remove-Item "`$env:ProgramData\Microsoft\Windows\Start Menu\Programs\DelGuard" -Recurse -ErrorAction SilentlyContinue
    Write-Host "快捷方式已删除" -ForegroundColor Green
} catch {
    Write-Host "删除快捷方式时出错: `$(`$_.Exception.Message)" -ForegroundColor Red
}

# 删除安装目录
try {
    Set-Location `$env:TEMP
    Remove-Item "$InstallPath" -Recurse -Force
    Write-Host "安装目录已删除" -ForegroundColor Green
} catch {
    Write-Host "删除安装目录时出错: `$(`$_.Exception.Message)" -ForegroundColor Red
}

Write-Host "DelGuard卸载完成" -ForegroundColor Green
if (-not `$Silent) {
    Read-Host "按回车键退出"
}
"@
    
    $UninstallPath = Join-Path $InstallPath "uninstall.ps1"
    $UninstallScript | Out-File -FilePath $UninstallPath -Encoding UTF8
    Write-Success "卸载脚本创建成功"
}

# 验证安装
function Test-Installation {
    param([string]$InstallPath, [string]$ServiceName)
    
    Write-Info "验证安装..."
    
    $Issues = @()
    
    # 检查主程序
    $ExePath = Join-Path $InstallPath "delguard.exe"
    if (-not (Test-Path $ExePath)) {
        $Issues += "主程序文件不存在"
    }
    
    # 检查配置文件
    $ConfigPath = Join-Path $InstallPath "config.yaml"
    if (-not (Test-Path $ConfigPath)) {
        $Issues += "配置文件不存在"
    }
    
    # 检查服务
    $Service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if (-not $Service) {
        $Issues += "Windows服务未注册"
    } elseif ($Service.Status -ne "Running") {
        $Issues += "服务未运行"
    }
    
    if ($Issues.Count -eq 0) {
        Write-Success "安装验证通过"
        return $true
    } else {
        Write-Error "安装验证失败:"
        foreach ($Issue in $Issues) {
            Write-Error "  - $Issue"
        }
        return $false
    }
}

# 主安装流程
function Install-DelGuard {
    try {
        Write-ColorOutput "
╔══════════════════════════════════════════════════════════════╗
║                    DelGuard 安装程序                         ║
║                     版本: 2.0.0                             ║
╚══════════════════════════════════════════════════════════════╝
" "Cyan"

        # 检查管理员权限
        if (-not (Test-Administrator)) {
            throw "需要管理员权限才能安装DelGuard。请以管理员身份运行PowerShell。"
        }

        # 显示系统信息
        if (-not $Silent) {
            $SystemInfo = Get-SystemInfo
            Write-Info "系统信息:"
            Write-Info "  操作系统: $($SystemInfo.OS)"
            Write-Info "  架构: $($SystemInfo.Architecture)"
            Write-Info "  PowerShell版本: $($SystemInfo.PowerShellVersion)"
            Write-Info "  .NET版本: $($SystemInfo.DotNetVersion)"
            Write-Info ""
            Write-Info "安装路径: $InstallPath"
            Write-Info "服务名称: $ServiceName"
            Write-Info ""
        }

        # 确认安装
        if (-not $Silent) {
            $Confirm = Read-Host "是否继续安装？(Y/n)"
            if ($Confirm -eq "n" -or $Confirm -eq "N") {
                Write-Info "安装已取消"
                return
            }
        }

        Write-Info "开始安装DelGuard..."

        # 1. 创建安装目录
        New-InstallDirectory -Path $InstallPath

        # 2. 复制文件
        Copy-DelGuardFiles -SourcePath "." -DestPath $InstallPath

        # 3. 创建配置文件
        New-ConfigFile -InstallPath $InstallPath

        # 4. 注册Windows服务
        Register-DelGuardService -InstallPath $InstallPath -ServiceName $ServiceName

        # 5. 添加到PATH
        if ($AddToPath) {
            Add-ToSystemPath -InstallPath $InstallPath
        }

        # 6. 创建快捷方式
        if ($CreateDesktopShortcut) {
            New-DesktopShortcut -InstallPath $InstallPath
        }
        New-StartMenuShortcut -InstallPath $InstallPath

        # 7. 创建卸载脚本
        New-UninstallScript -InstallPath $InstallPath -ServiceName $ServiceName

        # 8. 启动服务
        if ($StartService) {
            Start-DelGuardService -ServiceName $ServiceName
        }

        # 9. 验证安装
        $InstallSuccess = Test-Installation -InstallPath $InstallPath -ServiceName $ServiceName

        if ($InstallSuccess) {
            Write-Success "
╔══════════════════════════════════════════════════════════════╗
║                   DelGuard 安装成功！                        ║
╚══════════════════════════════════════════════════════════════╝"
            Write-Success "安装路径: $InstallPath"
            Write-Success "服务状态: 运行中"
            Write-Success "配置文件: $(Join-Path $InstallPath 'config.yaml')"
            Write-Info ""
            Write-Info "使用方法:"
            Write-Info "  命令行: delguard --help"
            Write-Info "  服务管理: services.msc"
            Write-Info "  卸载: 运行 $(Join-Path $InstallPath 'uninstall.ps1')"
        } else {
            throw "安装验证失败"
        }

    } catch {
        Write-Error "安装失败: $($_.Exception.Message)"
        Write-Error "请检查错误信息并重试"
        exit 1
    }
}

# 执行安装
Install-DelGuard