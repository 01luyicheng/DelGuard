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