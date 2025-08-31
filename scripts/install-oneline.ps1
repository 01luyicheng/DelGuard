# DelGuard Windows 一行安装脚本
# 使用方法: iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/your-repo/DelGuard/main/scripts/install-oneline.ps1'))

function Install-DelGuard {
    [CmdletBinding()]
    param(
        [string]$Version = "latest",
        [string]$InstallDir = "$env:LOCALAPPDATA\Programs\DelGuard",
        [string]$ConfigDir = "$env:APPDATA\DelGuard",
        [switch]$Force,
        [switch]$NoAliases,
        [switch]$NoConfig,
        [switch]$DryRun,
        [switch]$Uninstall,
        [switch]$Help
    )
    
    # 显示帮助
    if ($Help) {
        Write-Host @"
🛡️  DelGuard 一行安装脚本

使用方法: Install-DelGuard [选项]

选项:
    -Version <string>      指定要安装的版本 (默认: latest)
    -InstallDir <string>   指定安装目录 (默认: `$env:LOCALAPPDATA\Programs\DelGuard)
    -ConfigDir <string>    指定配置目录 (默认: `$env:APPDATA\DelGuard)
    -Force                 强制重新安装，覆盖现有版本
    -NoAliases             不配置PowerShell别名
    -NoConfig              不创建配置文件
    -DryRun                模拟安装，不实际执行
    -Uninstall             卸载 DelGuard
    -Help                  显示此帮助信息

示例:
    Install-DelGuard                        # 安装最新版本
    Install-DelGuard -Version v1.6.3       # 安装指定版本
    Install-DelGuard -InstallDir C:\Tools   # 安装到指定目录
    Install-DelGuard -Force -NoAliases     # 强制安装，不配置别名

一行命令安装:
    iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/your-repo/DelGuard/main/scripts/install-oneline.ps1'))

"@
        return
    }
    
    # 处理卸载
    if ($Uninstall) {
        Write-Host "🗑️  卸载 DelGuard" -ForegroundColor Yellow
        Write-Host "==================" -ForegroundColor Yellow
        Write-Host
        
        $filesToRemove = @(
            "$InstallDir\delguard.exe"
            "$InstallDir\delguard-uninstall.bat"
        )
        
        $dirsToRemove = @(
            "$ConfigDir"
        )
        
        # 显示将要删除的文件
        Write-Host "将要删除的文件:" -ForegroundColor Cyan
        foreach ($file in $filesToRemove) {
            if (Test-Path $file) {
                Write-Host "  📄 $file" -ForegroundColor White
            }
        }
        
        Write-Host "将要删除的目录:" -ForegroundColor Cyan
        foreach ($dir in $dirsToRemove) {
            if (Test-Path $dir) {
                Write-Host "  📁 $dir" -ForegroundColor White
            }
        }
        
        Write-Host
        $response = Read-Host "确定要继续卸载吗? (y/N)"
        
        if ($response -notmatch '^[Yy]$') {
            Write-Host "卸载已取消" -ForegroundColor Yellow
            return
        }
        
        # 执行卸载
        foreach ($file in $filesToRemove) {
            if (Test-Path $file) {
                Remove-Item $file -Force
                Write-Host "✅ 已删除: $file" -ForegroundColor Green
            }
        }
        
        foreach ($dir in $dirsToRemove) {
            if (Test-Path $dir) {
                Remove-Item $dir -Recurse -Force
                Write-Host "✅ 已删除: $dir" -ForegroundColor Green
            }
        }
        
        # 清理PowerShell配置
        $profilePaths = @($PROFILE.CurrentUserAllHosts, $PROFILE.CurrentUserCurrentHost)
        foreach ($profilePath in $profilePaths) {
            if (Test-Path $profilePath) {
                $content = Get-Content $profilePath -Raw
                $newContent = $content -replace '# DelGuard 别名配置.*?# DelGuard 配置完成\r?\n?', ""
                if ($newContent -ne $content) {
                    Set-Content -Path $profilePath -Value $newContent.TrimEnd()
                    Write-Host "✅ 已清理PowerShell配置: $profilePath" -ForegroundColor Green
                }
            }
        }
        
        # 清理PATH
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        $newPath = ($currentPath -split ';' | Where-Object { $_ -ne $InstallDir }) -join ';'
        if ($newPath -ne $currentPath) {
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-Host "✅ 已从用户PATH中移除 DelGuard" -ForegroundColor Green
        }
        
        Write-Host "✅ DelGuard 已成功卸载" -ForegroundColor Green
        return
    }
    
    # 检查管理员权限
    $isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
    if (-not $isAdmin) {
        Write-Host "⚠️  需要管理员权限运行此脚本" -ForegroundColor Yellow
        Write-Host "请以管理员身份运行 PowerShell 并执行:"
        Write-Host "  Set-ExecutionPolicy RemoteSigned -Scope CurrentUser"
        Write-Host "  iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/your-repo/DelGuard/main/scripts/install-oneline.ps1'))"
        return
    }
    
    # 设置安装参数
    $BINARY_NAME = "delguard.exe"
    
    # 检测系统架构
    $ARCH = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x64" } else { "x86" }
    
    # 获取版本信息
    $apiUrl = "https://api.github.com/repos/01luyicheng/DelGuard/releases"
    if ($Version -eq "latest") {
        $apiUrl = "$apiUrl/latest"
        $releaseInfo = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        $VERSION = $releaseInfo.tag_name
    } else {
        $VERSION = $Version -replace '^v?', 'v'
        $releases = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        $releaseInfo = $releases | Where-Object { $_.tag_name -eq $VERSION } | Select-Object -First 1
        if (-not $releaseInfo) {
            Write-Host "❌ 指定的版本不存在: $VERSION" -ForegroundColor Red
            return
        }
    }
    
    # 构建下载URL
    $downloadUrl = "https://github.com/01luyicheng/DelGuard/releases/download/$VERSION/delguard-windows-$ARCH.exe"
    
    Write-Host "🛡️  DelGuard 一行安装脚本" -ForegroundColor Cyan
    Write-Host "==========================" -ForegroundColor Cyan
    Write-Host
    Write-Host "系统架构: $ARCH" -ForegroundColor Green
    Write-Host "版本: $VERSION" -ForegroundColor Green
    Write-Host "安装目录: $InstallDir" -ForegroundColor Green
    Write-Host "配置目录: $ConfigDir" -ForegroundColor Green
    Write-Host "配置别名: $(if (-not $NoAliases) { '是' } else { '否' })" -ForegroundColor Green
    Write-Host "创建配置: $(if (-not $NoConfig) { '是' } else { '否' })" -ForegroundColor Green
    Write-Host
    
    # 检查现有安装
    $binaryPath = Join-Path $InstallDir $BINARY_NAME
    if (Test-Path $binaryPath -and -not $Force) {
        Write-Host "⚠️  DelGuard 已安装" -ForegroundColor Yellow
        Write-Host "使用 -Force 选项强制重新安装" -ForegroundColor Yellow
        $currentVersion = & $binaryPath --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "当前版本: $currentVersion" -ForegroundColor Cyan
        }
        return
    }
    
    # 干运行模式
    if ($DryRun) {
        Write-Host "🔍 干运行模式 - 模拟安装" -ForegroundColor Cyan
        Write-Host "将要执行的操作:"
        Write-Host "  - 下载: $downloadUrl"
        Write-Host "  - 安装到: $binaryPath"
        Write-Host "  - 添加到 PATH"
        return
    }
    
    # 创建安装目录
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    # 下载二进制文件
    Write-Host "📥 正在下载 DelGuard..." -ForegroundColor Yellow
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $binaryPath -UseBasicParsing
        Write-Host "✅ 下载完成" -ForegroundColor Green
    } catch {
        Write-Host "❌ 下载失败: $($_.Exception.Message)" -ForegroundColor Red
        return
    }
    
    # 添加到PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$InstallDir", "User")
        Write-Host "✅ 已将 DelGuard 添加到 PATH" -ForegroundColor Green
    }
    
    # 创建配置文件
    if (-not $NoConfig) {
        if (-not (Test-Path $ConfigDir)) {
            New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
        }
        
        $configPath = Join-Path $ConfigDir "config.json"
        $config = @{
            version = $VERSION
            install_dir = $InstallDir
            created_at = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
            settings = @{
                recycle_bin_enabled = $true
                log_level = "info"
                auto_cleanup = $true
                max_log_size = "100MB"
            }
        } | ConvertTo-Json -Depth 3
        
        Set-Content -Path $configPath -Value $config
        Write-Host "✅ 已创建配置文件: $configPath" -ForegroundColor Green
    }
    
    # 配置PowerShell别名
    if (-not $NoAliases) {
        $profilePath = $PROFILE.CurrentUserAllHosts
        if (-not (Test-Path $profilePath)) {
            New-Item -ItemType File -Path $profilePath -Force | Out-Null
        }
        
        $aliasConfig = @"
# DelGuard 别名配置
function safe-rm {
    param(
        [Parameter(Mandatory=`$true)]
        [string[]]`$Paths,
        [switch]`$Force,
        [switch]`$Recursive
    )
    
    foreach (`$path in `$Paths) {
        if (Test-Path `$path) {
            `$item = Get-Item `$path
            `$size = (Get-ChildItem `$path -Recurse | Measure-Object -Property Length -Sum).Sum
            `$sizeMB = [math]::Round(`$size / 1MB, 2)
            
            Write-Host "正在删除: `$path (`$(`$sizeMB)MB)" -ForegroundColor Yellow
            Remove-Item `$path -Recurse:`$Recursive -Force:`$Force
            Write-Host "✅ 已删除: `$path" -ForegroundColor Green
        } else {
            Write-Warning "文件不存在: `$path"
        }
    }
}

Set-Alias -Name dguard -Value delguard
Set-Alias -Name rm-safe -Value safe-rm
# DelGuard 配置完成
"@
        
        $profileContent = Get-Content $profilePath -Raw
        if ($profileContent -notlike "*DelGuard 配置完成*") {
            Add-Content -Path $profilePath -Value $aliasConfig
            Write-Host "✅ 已配置 PowerShell 别名" -ForegroundColor Green
        }
    }
    
    # 创建卸载脚本
    $uninstallScript = @"
@echo off
echo 正在卸载 DelGuard...
if exist "$InstallDir\delguard.exe" del /f "$InstallDir\delguard.exe"
if exist "$ConfigDir" rmdir /s /q "$ConfigDir"
echo DelGuard 已卸载
pause
"@
    $uninstallScriptPath = Join-Path $InstallDir "delguard-uninstall.bat"
    Set-Content -Path $uninstallScriptPath -Value $uninstallScript
    
    Write-Host
    Write-Host "🎉 安装完成！" -ForegroundColor Green
    Write-Host "运行 'delguard --help' 开始使用"
    Write-Host "卸载脚本: $uninstallScriptPath"
}

# 如果直接运行此脚本，则调用安装函数
if ($MyInvocation.InvocationName -ne '.') {
    Install-DelGuard @args
}