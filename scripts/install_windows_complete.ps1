param(
    [string]$InstallPath = "C:\Program Files\DelGuard",
    [switch]$Silent
)

# 检查管理员权限
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    if (-NOT $Silent) {
        Write-Host "需要管理员权限来安装DelGuard。正在重新启动为管理员..." -ForegroundColor Yellow
    }
    Start-Process PowerShell -Verb RunAs "-File `"$PSCommandPath`" -InstallPath `"$InstallPath`" $(if($Silent){'-Silent'})"
    exit
}

try {
    if (-NOT $Silent) {
        Write-Host "开始安装DelGuard到 $InstallPath..." -ForegroundColor Green
    }

    # 创建安装目录
    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }

    # 复制主程序
    if (Test-Path "delguard.exe") {
        Copy-Item "delguard.exe" "$InstallPath\" -Force
        if (-NOT $Silent) {
            Write-Host "✓ 已复制主程序" -ForegroundColor Green
        }
    } else {
        throw "找不到delguard.exe文件"
    }

    # 创建批处理文件用于命令别名
    $batFiles = @{
        "rm.bat" = "@echo off`nset `"args=%*`"`nif `"%args%`"==`"`" (`n    echo 用法: rm [选项] 文件...`n    echo 选项:`n    echo   -f, --force     强制删除`n    echo   -r, --recursive 递归删除目录`n    echo   -v, --verbose   详细输出`n    exit /b 1`n)`n`"$InstallPath\delguard.exe`" delete %*"
        "del.bat" = "@echo off`nset `"args=%*`"`nif `"%args%`"==`"`" (`n    echo 用法: del [选项] 文件...`n    echo 选项:`n    echo   -f, --force     强制删除`n    echo   -r, --recursive 递归删除目录`n    echo   -v, --verbose   详细输出`n    exit /b 1`n)`n`"$InstallPath\delguard.exe`" delete %*"
        "cp.bat" = "@echo off`nset `"args=%*`"`nif `"%args%`"==`"`" (`n    echo 用法: cp 源文件 目标文件`n    echo 这是一个安全的复制命令，会记录操作历史`n    exit /b 1`n)`ncopy %*"
        "delguard.bat" = "@echo off`n`"$InstallPath\delguard.exe`" %*"
    }

    foreach ($file in $batFiles.Keys) {
        $content = $batFiles[$file]
        Set-Content -Path "$InstallPath\$file" -Value $content -Encoding ASCII
        if (-NOT $Silent) {
            Write-Host "✓ 已创建 $file" -ForegroundColor Green
        }
    }

    # 创建PowerShell函数文件
    $psProfile = @"
# DelGuard PowerShell 函数
function rm {
    param([Parameter(ValueFromRemainingArguments=`$true)]`$args)
    if (`$args.Count -eq 0) {
        Write-Host "用法: rm [选项] 文件..." -ForegroundColor Yellow
        Write-Host "选项:" -ForegroundColor Yellow
        Write-Host "  -f, --force     强制删除" -ForegroundColor Yellow
        Write-Host "  -r, --recursive 递归删除目录" -ForegroundColor Yellow
        Write-Host "  -v, --verbose   详细输出" -ForegroundColor Yellow
        return
    }
    & "$InstallPath\delguard.exe" delete @args
}

function del {
    param([Parameter(ValueFromRemainingArguments=`$true)]`$args)
    if (`$args.Count -eq 0) {
        Write-Host "用法: del [选项] 文件..." -ForegroundColor Yellow
        Write-Host "选项:" -ForegroundColor Yellow
        Write-Host "  -f, --force     强制删除" -ForegroundColor Yellow
        Write-Host "  -r, --recursive 递归删除目录" -ForegroundColor Yellow
        Write-Host "  -v, --verbose   详细输出" -ForegroundColor Yellow
        return
    }
    & "$InstallPath\delguard.exe" delete @args
}

function delguard {
    param([Parameter(ValueFromRemainingArguments=`$true)]`$args)
    & "$InstallPath\delguard.exe" @args
}

# 设置别名
Set-Alias -Name "cp" -Value "Copy-Item" -Force
"@

    Set-Content -Path "$InstallPath\DelGuard-Profile.ps1" -Value $psProfile -Encoding UTF8
    if (-NOT $Silent) {
        Write-Host "✓ 已创建PowerShell配置文件" -ForegroundColor Green
    }

    # 添加到系统PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    if ($currentPath -notlike "*$InstallPath*") {
        $newPath = "$currentPath;$InstallPath"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
        if (-NOT $Silent) {
            Write-Host "✓ 已添加到系统PATH" -ForegroundColor Green
        }
    }

    # 更新当前会话的PATH
    $env:PATH = "$env:PATH;$InstallPath"

    # 创建用户启动脚本
    $startupScript = @"
# 加载DelGuard PowerShell函数
if (Test-Path "$InstallPath\DelGuard-Profile.ps1") {
    . "$InstallPath\DelGuard-Profile.ps1"
}
"@

    $profilePath = $PROFILE.AllUsersAllHosts
    $profileDir = Split-Path $profilePath -Parent
    if (-not (Test-Path $profileDir)) {
        New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
    }

    if (Test-Path $profilePath) {
        $existingContent = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
        if ($existingContent -notlike "*DelGuard*") {
            Add-Content -Path $profilePath -Value "`n$startupScript" -Encoding UTF8
        }
    } else {
        Set-Content -Path $profilePath -Value $startupScript -Encoding UTF8
    }

    if (-NOT $Silent) {
        Write-Host "✓ 已配置PowerShell启动脚本" -ForegroundColor Green
    }

    # 验证安装
    $testCommands = @("rm.bat", "del.bat", "cp.bat", "delguard.bat", "delguard.exe")
    $allSuccess = $true
    
    foreach ($cmd in $testCommands) {
        if (-not (Test-Path "$InstallPath\$cmd")) {
            if (-NOT $Silent) {
                Write-Host "✗ 缺少文件: $cmd" -ForegroundColor Red
            }
            $allSuccess = $false
        }
    }

    if ($allSuccess) {
        if (-NOT $Silent) {
            Write-Host "`n✓ DelGuard 安装成功!" -ForegroundColor Green
            Write-Host "安装路径: $InstallPath" -ForegroundColor Cyan
            Write-Host "`n可用命令:" -ForegroundColor Cyan
            Write-Host "  rm <文件>     - 安全删除文件" -ForegroundColor White
            Write-Host "  del <文件>    - 安全删除文件" -ForegroundColor White
            Write-Host "  cp <源> <目标> - 复制文件" -ForegroundColor White
            Write-Host "  delguard      - DelGuard主程序" -ForegroundColor White
            Write-Host "`n注意: 请重新启动PowerShell或CMD以使用新命令" -ForegroundColor Yellow
        }
        return $true
    } else {
        throw "安装验证失败"
    }

} catch {
    if (-NOT $Silent) {
        Write-Host "安装失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $false
}