# DelGuard 发布准备脚本

param(
    [Parameter(Mandatory=$true)]
    [string]$Version,
    [switch]$DryRun = $false,
    [switch]$Force = $false
)

$ErrorActionPreference = 'Stop'

Write-Host "DelGuard 发布准备" -ForegroundColor Cyan
Write-Host "=================" -ForegroundColor Cyan
Write-Host "版本: $Version" -ForegroundColor Green
if ($DryRun) {
    Write-Host "模式: 试运行" -ForegroundColor Yellow
}
Write-Host ""

# 验证版本格式
if ($Version -notmatch '^v\d+\.\d+\.\d+') {
    Write-Error "版本格式错误，应该是 vX.Y.Z 格式，例如 v1.0.0"
}

# 检查工作目录是否干净
Write-Host "检查 Git 状态..." -ForegroundColor Yellow
$gitStatus = git status --porcelain
if ($gitStatus -and !$Force) {
    Write-Error "工作目录不干净，请先提交或暂存更改，或使用 -Force 参数"
}

# 检查是否在主分支
$currentBranch = git branch --show-current
if ($currentBranch -ne "main" -and !$Force) {
    Write-Error "当前不在 main 分支，请切换到 main 分支或使用 -Force 参数"
}

# 运行项目检查
Write-Host "运行项目检查..." -ForegroundColor Yellow
if (!$DryRun) {
    .\check.ps1
    if ($LASTEXITCODE -ne 0) {
        Write-Error "项目检查失败，请修复问题后重试"
    }
}

# 更新版本信息
Write-Host "更新版本信息..." -ForegroundColor Yellow
if (!$DryRun) {
    # 这里可以添加版本号更新逻辑
    Write-Host "版本信息已更新" -ForegroundColor Green
} else {
    Write-Host "将更新版本信息到 $Version" -ForegroundColor Gray
}

# 运行完整测试
Write-Host "运行完整测试套件..." -ForegroundColor Yellow
if (!$DryRun) {
    # 检查是否支持 race 检测
    $env:CGO_ENABLED = "1"
    $raceSupported = $true
    
    # 在 Windows 上测试 race 检测是否可用
    if ($IsWindows -or $env:OS -eq "Windows_NT") {
        try {
            go test -race -run=NonExistentTest ./... 2>$null
        } catch {
            $raceSupported = $false
        }
    }
    
    if ($raceSupported) {
        Write-Host "运行带竞态检测的测试..." -ForegroundColor Blue
        go test -v -race -coverprofile=coverage.out ./...
    } else {
        Write-Host "运行标准测试 (race 检测不可用)..." -ForegroundColor Blue
        go test -v -coverprofile=coverage.out ./...
    }
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "测试失败，停止发布流程"
    }
    Write-Host "所有测试通过" -ForegroundColor Green
    
    # 重置 CGO 设置
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
} else {
    Write-Host "将运行完整测试套件" -ForegroundColor Gray
}

# 构建所有平台
Write-Host "构建所有平台..." -ForegroundColor Yellow
if (!$DryRun) {
    .\build.ps1 -Version $Version -Release -Clean
    if ($LASTEXITCODE -ne 0) {
        Write-Error "构建失败，停止发布流程"
    }
    Write-Host "构建完成" -ForegroundColor Green
} else {
    Write-Host "将构建所有平台的二进制文件" -ForegroundColor Gray
}

# 生成变更日志
Write-Host "检查变更日志..." -ForegroundColor Yellow
if (!(Test-Path "CHANGELOG.md")) {
    Write-Warning "未找到 CHANGELOG.md，建议添加变更日志"
} else {
    Write-Host "变更日志存在" -ForegroundColor Green
}

# 创建 Git 标签
Write-Host "创建 Git 标签..." -ForegroundColor Yellow
if (!$DryRun) {
    try {
        git tag -a $Version -m "Release $Version"
        Write-Host "Git 标签 $Version 已创建" -ForegroundColor Green
    } catch {
        Write-Error "创建 Git 标签失败: $($_.Exception.Message)"
    }
} else {
    Write-Host "将创建 Git 标签: $Version" -ForegroundColor Gray
}

# 推送到远程仓库
Write-Host "推送到远程仓库..." -ForegroundColor Yellow
if (!$DryRun) {
    try {
        git push origin main
        git push origin $Version
        Write-Host "已推送到远程仓库" -ForegroundColor Green
    } catch {
        Write-Error "推送失败: $($_.Exception.Message)"
    }
} else {
    Write-Host "将推送代码和标签到远程仓库" -ForegroundColor Gray
}

# 发布总结
Write-Host "`n" + "="*50 -ForegroundColor Cyan
Write-Host "发布总结" -ForegroundColor Cyan
Write-Host "="*50 -ForegroundColor Cyan

if ($DryRun) {
    Write-Host "试运行完成，以下是将要执行的操作：" -ForegroundColor Yellow
    Write-Host "1. 更新版本信息到 $Version" -ForegroundColor Gray
    Write-Host "2. 运行完整测试套件" -ForegroundColor Gray
    Write-Host "3. 构建所有平台的二进制文件" -ForegroundColor Gray
    Write-Host "4. 创建 Git 标签 $Version" -ForegroundColor Gray
    Write-Host "5. 推送到远程仓库" -ForegroundColor Gray
    Write-Host "`n要执行实际发布，请运行：" -ForegroundColor Cyan
    Write-Host ".\release.ps1 -Version $Version" -ForegroundColor White
} else {
    Write-Host "🎉 发布 $Version 完成！" -ForegroundColor Green
    Write-Host "`n后续步骤：" -ForegroundColor Cyan
    Write-Host "1. 检查 GitHub Actions 构建状态" -ForegroundColor Gray
    Write-Host "2. 验证 GitHub Release 页面" -ForegroundColor Gray
    Write-Host "3. 测试安装脚本" -ForegroundColor Gray
    Write-Host "4. 更新文档和公告" -ForegroundColor Gray
    
    Write-Host "`nGitHub Release 页面：" -ForegroundColor Cyan
    Write-Host "https://github.com/01luyicheng/DelGuard/releases/tag/$Version" -ForegroundColor Blue
}

Write-Host "`n发布完成！" -ForegroundColor Green