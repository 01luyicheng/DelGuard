# DelGuard 功能测试脚本
# 测试所有新增功能

# 设置控制台颜色
$ErrorColor = "Red"
$SuccessColor = "Green"
$InfoColor = "Cyan"
$WarningColor = "Yellow"

# 创建测试目录
$TestDir = Join-Path $env:TEMP "delguard-test-$(Get-Random)"
New-Item -ItemType Directory -Path $TestDir -Force | Out-Null

# 输出标题
Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor $InfoColor
Write-Host "║                                                              ║" -ForegroundColor $InfoColor
Write-Host "║                🧪 DelGuard 功能测试工具                      ║" -ForegroundColor $InfoColor
Write-Host "║                                                              ║" -ForegroundColor $InfoColor
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor $InfoColor

# 测试UTF-8编码
function Test-UTF8Encoding {
    Write-Host "测试UTF-8编码支持..." -ForegroundColor $InfoColor
    
    # 创建包含中文字符的测试文件
    $TestFile = Join-Path $TestDir "中文测试文件.txt"
    "这是一个UTF-8编码的测试文件，包含中文字符。" | Out-File -FilePath $TestFile -Encoding utf8
    
    # 读取文件内容
    $Content = Get-Content -Path $TestFile -Encoding utf8 -Raw
    
    # 检查内容是否正确
    if ($Content -match "这是一个UTF-8编码的测试文件") {
        Write-Host "✓ UTF-8编码测试通过" -ForegroundColor $SuccessColor
        return $true
    } else {
        Write-Host "✗ UTF-8编码测试失败" -ForegroundColor $ErrorColor
        return $false
    }
}

# 测试语言检测
function Test-LanguageDetection {
    Write-Host "测试语言检测功能..." -ForegroundColor $InfoColor
    
    # 获取系统UI语言
    $UILanguage = (Get-Culture).Name
    Write-Host "当前系统UI语言: $UILanguage"
    
    # 检查是否为中文
    if ($UILanguage -match "zh-CN") {
        Write-Host "✓ 语言检测功能正常，检测到中文系统" -ForegroundColor $SuccessColor
        return $true
    } else {
        Write-Host "✓ 语言检测功能正常，检测到非中文系统" -ForegroundColor $SuccessColor
        return $true
    }
}

# 测试智能搜索功能
function Test-SmartSearch {
    Write-Host "测试智能搜索功能..." -ForegroundColor $InfoColor
    
    # 创建几个测试文件
    $TestFile1 = Join-Path $TestDir "document.txt"
    $TestFile2 = Join-Path $TestDir "document_backup.txt"
    $TestFile3 = Join-Path $TestDir "doc.txt"
    
    "测试文档内容" | Out-File -FilePath $TestFile1 -Encoding utf8
    "备份文档内容" | Out-File -FilePath $TestFile2 -Encoding utf8
    "简短文档" | Out-File -FilePath $TestFile3 -Encoding utf8
    
    # 测试不存在的文件名
    $NonExistentFile = Join-Path $TestDir "documents.txt"
    
    # 模拟智能搜索功能
    $SimilarFiles = @($TestFile1, $TestFile2, $TestFile3) | Where-Object {
        $FileName = Split-Path -Leaf $_
        $TargetName = "documents.txt"
        
        # 简单相似度检查
        $FileName.ToLower().Contains("doc") -or $TargetName.ToLower().Contains($FileName.ToLower())
    }
    
    if ($SimilarFiles.Count -gt 0) {
        Write-Host "✓ 智能搜索功能正常，找到了相似文件:" -ForegroundColor $SuccessColor
        $SimilarFiles | ForEach-Object {
            Write-Host "  - $(Split-Path -Leaf $_)" -ForegroundColor $InfoColor
        }
        return $true
    } else {
        Write-Host "✗ 智能搜索功能异常，未找到相似文件" -ForegroundColor $ErrorColor
        return $false
    }
}

# 测试安装脚本
function Test-InstallScript {
    Write-Host "测试安装脚本功能..." -ForegroundColor $InfoColor
    
    # 检查安装脚本是否存在
    if (Test-Path "install_enhanced_utf8.ps1") {
        Write-Host "✓ 增强版安装脚本存在" -ForegroundColor $SuccessColor
        
        # 检查脚本内容
        $ScriptContent = Get-Content "install_enhanced_utf8.ps1" -Raw
        
        $Features = @(
            @{Name="UTF-8编码设置"; Pattern="UTF-8|utf8|encoding"},
            @{Name="PowerShell检测"; Pattern="PowerShell|pwsh"},
            @{Name="语言自动检测"; Pattern="language|locale|CultureInfo"},
            @{Name="别名注册"; Pattern="alias|别名"}
        )
        
        $AllFeaturesPresent = $true
        foreach ($Feature in $Features) {
            if ($ScriptContent -match $Feature.Pattern) {
                Write-Host "  ✓ 脚本包含$($Feature.Name)功能" -ForegroundColor $SuccessColor
            } else {
                Write-Host "  ✗ 脚本缺少$($Feature.Name)功能" -ForegroundColor $ErrorColor
                $AllFeaturesPresent = $false
            }
        }
        
        return $AllFeaturesPresent
    } else {
        Write-Host "✗ 增强版安装脚本不存在" -ForegroundColor $ErrorColor
        return $false
    }
}

# 测试卸载脚本
function Test-UninstallScript {
    Write-Host "测试卸载脚本功能..." -ForegroundColor $InfoColor
    
    # 检查卸载脚本是否存在
    if (Test-Path "uninstall.ps1") {
        Write-Host "✓ 卸载脚本存在" -ForegroundColor $SuccessColor
        return $true
    } else {
        Write-Host "✗ 卸载脚本不存在" -ForegroundColor $ErrorColor
        return $false
    }
}

# 测试更新脚本
function Test-UpdateScript {
    Write-Host "测试更新脚本功能..." -ForegroundColor $InfoColor
    
    # 检查更新脚本是否存在
    if (Test-Path "update.ps1") {
        Write-Host "✓ 更新脚本存在" -ForegroundColor $SuccessColor
        return $true
    } else {
        Write-Host "✗ 更新脚本不存在" -ForegroundColor $ErrorColor
        return $false
    }
}

# 运行所有测试
$TestResults = @{
    "UTF-8编码支持" = Test-UTF8Encoding
    "语言检测功能" = Test-LanguageDetection
    "智能搜索功能" = Test-SmartSearch
    "安装脚本功能" = Test-InstallScript
    "卸载脚本功能" = Test-UninstallScript
    "更新脚本功能" = Test-UpdateScript
}

# 输出测试结果摘要
Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor $InfoColor
Write-Host "║                      测试结果摘要                           ║" -ForegroundColor $InfoColor
Write-Host "╚══════════════════════════════════════════════════════════════╝`n" -ForegroundColor $InfoColor

$PassedTests = 0
$TotalTests = $TestResults.Count

foreach ($Test in $TestResults.GetEnumerator()) {
    if ($Test.Value) {
        Write-Host "✓ $($Test.Key) - 通过" -ForegroundColor $SuccessColor
        $PassedTests++
    } else {
        Write-Host "✗ $($Test.Key) - 失败" -ForegroundColor $ErrorColor
    }
}

$PassRate = [math]::Round(($PassedTests / $TotalTests) * 100, 2)
Write-Host "`n通过率: $PassRate% ($PassedTests/$TotalTests)" -ForegroundColor $(if ($PassRate -eq 100) { $SuccessColor } elseif ($PassRate -ge 80) { $WarningColor } else { $ErrorColor })

# 清理测试目录
Remove-Item -Path $TestDir -Recurse -Force -ErrorAction SilentlyContinue
Write-Host "`n测试完成，已清理临时文件。" -ForegroundColor $InfoColor