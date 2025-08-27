#Requires -Version 5.1
<#
.SYNOPSIS
    修复 DelGuard 跨平台构建问题

.DESCRIPTION
    解决 Linux 和 macOS 平台的构建失败问题，确保所有平台都能正常构建
#>

[CmdletBinding()]
param(
    [switch]$Verbose,
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'

function Write-LogInfo { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Cyan }
function Write-LogSuccess { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-LogWarning { param([string]$Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
function Write-LogError { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

# 检查 Go 环境
function Test-GoEnvironment {
    Write-LogInfo "检查 Go 环境..."
    
    try {
        $goVersion = go version
        Write-LogSuccess "Go 版本: $goVersion"
        
        $goEnv = go env GOOS GOARCH
        Write-LogInfo "当前 Go 环境: $goEnv"
        
        return $true
    } catch {
        Write-LogError "Go 未安装或不在 PATH 中"
        return $false
    }
}

# 检查 CGO 依赖
function Test-CGODependencies {
    Write-LogInfo "检查 CGO 依赖..."
    
    # 检查是否有 CGO 代码
    $cgoFiles = Get-ChildItem -Path "." -Filter "*.go" -Recurse | 
                ForEach-Object { 
                    $content = Get-Content $_.FullName -Raw
                    if ($content -match 'import\s+"C"' -or $content -match '#include') {
                        $_.FullName
                    }
                }
    
    if ($cgoFiles) {
        Write-LogWarning "发现 CGO 代码文件:"
        $cgoFiles | ForEach-Object { Write-Host "  $_" -ForegroundColor Yellow }
        return $false
    } else {
        Write-LogSuccess "未发现 CGO 依赖"
        return $true
    }
}

# 修复构建标签
function Fix-BuildTags {
    Write-LogInfo "修复构建标签..."
    
    # 检查所有 Go 文件的构建标签
    $goFiles = Get-ChildItem -Path "." -Filter "*.go" -Recurse
    
    foreach ($file in $goFiles) {
        $content = Get-Content $file.FullName -Raw
        $modified = $false
        
        # 修复 Windows 构建标签
        if ($file.Name -like "*windows*" -and $content -notmatch "//go:build windows") {
            if ($content -match "// \+build windows") {
                $content = $content -replace "// \+build windows", "//go:build windows`n// +build windows"
                $modified = $true
            } elseif ($content -notmatch "//go:build" -and $file.Name -like "*windows*") {
                $content = "//go:build windows`n// +build windows`n`n" + $content
                $modified = $true
            }
        }
        
        # 修复 Unix 构建标签
        if ($file.Name -like "*unix*" -and $content -notmatch "//go:build !windows") {
            if ($content -match "// \+build !windows") {
                $content = $content -replace "// \+build !windows", "//go:build !windows`n// +build !windows"
                $modified = $true
            } elseif ($content -notmatch "//go:build" -and $file.Name -like "*unix*") {
                $content = "//go:build !windows`n// +build !windows`n`n" + $content
                $modified = $true
            }
        }
        
        # 修复 Darwin 构建标签
        if ($file.Name -like "*darwin*" -and $content -notmatch "//go:build darwin") {
            if ($content -match "// \+build darwin") {
                $content = $content -replace "// \+build darwin", "//go:build darwin`n// +build darwin"
                $modified = $true
            } elseif ($content -notmatch "//go:build" -and $file.Name -like "*darwin*") {
                $content = "//go:build darwin`n// +build darwin`n`n" + $content
                $modified = $true
            }
        }
        
        # 修复 Linux 构建标签
        if ($file.Name -like "*linux*" -and $content -notmatch "//go:build linux") {
            if ($content -match "// \+build linux") {
                $content = $content -replace "// \+build linux", "//go:build linux`n// +build linux"
                $modified = $true
            } elseif ($content -notmatch "//go:build" -and $file.Name -like "*linux*") {
                $content = "//go:build linux`n// +build linux`n`n" + $content
                $modified = $true
            }
        }
        
        if ($modified -and -not $DryRun) {
            Set-Content $file.FullName $content -Encoding UTF8
            Write-LogSuccess "已修复构建标签: $($file.Name)"
        } elseif ($modified -and $DryRun) {
            Write-LogInfo "需要修复构建标签: $($file.Name)"
        }
    }
}

# 创建缺失的平台特定文件
function Create-MissingPlatformFiles {
    Write-LogInfo "检查并创建缺失的平台特定文件..."
    
    # 检查是否需要创建 Linux 特定文件
    if (-not (Test-Path "restore_linux.go")) {
        Write-LogWarning "缺少 restore_linux.go 文件"
        if (-not $DryRun) {
            # 这里可以创建基本的 Linux 实现文件
        }
    }
    
    # 检查是否需要创建 macOS 特定文件
    if (-not (Test-Path "restore_darwin.go")) {
        Write-LogWarning "缺少 restore_darwin.go 文件"
        if (-not $DryRun) {
            # 这里可以创建基本的 macOS 实现文件
        }
    }
}

# 测试跨平台构建
function Test-CrossPlatformBuild {
    Write-LogInfo "测试跨平台构建..."
    
    $platforms = @(
        @{GOOS="windows"; GOARCH="amd64"},
        @{GOOS="windows"; GOARCH="386"},
        @{GOOS="linux"; GOARCH="amd64"},
        @{GOOS="linux"; GOARCH="arm64"},
        @{GOOS="darwin"; GOARCH="amd64"},
        @{GOOS="darwin"; GOARCH="arm64"}
    )
    
    $results = @()
    
    foreach ($platform in $platforms) {
        $env:GOOS = $platform.GOOS
        $env:GOARCH = $platform.GOARCH
        $env:CGO_ENABLED = "0"  # 禁用 CGO 以避免交叉编译问题
        
        $platformName = "$($platform.GOOS)-$($platform.GOARCH)"
        Write-LogInfo "测试构建: $platformName"
        
        try {
            if ($DryRun) {
                Write-LogInfo "试运行: go build -o delguard-$platformName ."
                $results += @{Platform=$platformName; Status="DryRun"; Error=$null}
            } else {
                $output = go build -o "delguard-$platformName" . 2>&1
                if ($LASTEXITCODE -eq 0) {
                    Write-LogSuccess "构建成功: $platformName"
                    $results += @{Platform=$platformName; Status="Success"; Error=$null}
                    
                    # 清理测试构建文件
                    if (Test-Path "delguard-$platformName") {
                        Remove-Item "delguard-$platformName" -Force
                    }
                    if (Test-Path "delguard-$platformName.exe") {
                        Remove-Item "delguard-$platformName.exe" -Force
                    }
                } else {
                    Write-LogError "构建失败: $platformName"
                    Write-LogError "错误信息: $output"
                    $results += @{Platform=$platformName; Status="Failed"; Error=$output}
                }
            }
        } catch {
            Write-LogError "构建异常: $platformName - $_"
            $results += @{Platform=$platformName; Status="Exception"; Error=$_.Exception.Message}
        }
    }
    
    # 恢复原始环境变量
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    
    return $results
}

# 生成构建报告
function Generate-BuildReport {
    param([array]$Results)
    
    Write-Host "`n构建测试报告:" -ForegroundColor Cyan
    Write-Host "===============" -ForegroundColor Cyan
    
    $successCount = 0
    $failCount = 0
    
    foreach ($result in $Results) {
        $status = switch ($result.Status) {
            "Success" { 
                $successCount++
                "✓ 成功" 
            }
            "Failed" { 
                $failCount++
                "✗ 失败" 
            }
            "Exception" { 
                $failCount++
                "✗ 异常" 
            }
            "DryRun" { "○ 试运行" }
        }
        
        $color = switch ($result.Status) {
            "Success" { "Green" }
            "Failed" { "Red" }
            "Exception" { "Red" }
            "DryRun" { "Yellow" }
        }
        
        Write-Host "$($result.Platform.PadRight(15)) $status" -ForegroundColor $color
        
        if ($result.Error -and $Verbose) {
            Write-Host "  错误详情: $($result.Error)" -ForegroundColor Gray
        }
    }
    
    Write-Host "`n总结:" -ForegroundColor Cyan
    Write-Host "成功: $successCount" -ForegroundColor Green
    Write-Host "失败: $failCount" -ForegroundColor Red
    
    if ($failCount -eq 0 -and $successCount -gt 0) {
        Write-LogSuccess "所有平台构建测试通过！"
        return $true
    } else {
        Write-LogWarning "存在构建失败的平台，需要进一步修复"
        return $false
    }
}

# 主函数
function Main {
    Write-Host "DelGuard 跨平台构建修复工具" -ForegroundColor Cyan
    Write-Host "============================" -ForegroundColor Cyan
    Write-Host ""
    
    if (-not (Test-GoEnvironment)) {
        Write-LogError "Go 环境检查失败，请先安装 Go"
        exit 1
    }
    
    Test-CGODependencies
    Fix-BuildTags
    Create-MissingPlatformFiles
    
    Write-Host ""
    $results = Test-CrossPlatformBuild
    $success = Generate-BuildReport $results
    
    if ($success) {
        Write-Host "`n跨平台构建修复完成！所有平台都可以正常构建。" -ForegroundColor Green
    } else {
        Write-Host "`n仍有部分平台构建失败，请检查错误信息并手动修复。" -ForegroundColor Yellow
    }
}

# 执行主函数
Main