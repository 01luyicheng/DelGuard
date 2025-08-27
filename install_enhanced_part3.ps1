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