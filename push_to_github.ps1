# DelGuard GitHub 发布脚本
# 用于将 v1.4.1 版本推送到GitHub

param(
    [string]$Owner = "01luyicheng",  # GitHub用户名
    [string]$Repo = "DelGuard",
    [string]$Version = "v1.5.5",
    [switch]$Force
)

$ErrorActionPreference = "Stop"

Write-Host "🚀 DelGuard GitHub 发布脚本" -ForegroundColor Green
Write-Host "版本: $Version" -ForegroundColor Cyan
Write-Host "仓库: $Owner/$Repo" -ForegroundColor Cyan
Write-Host ""

# 检查Git是否已初始化
if (!(Test-Path ".git")) {
    Write-Host "📁 初始化Git仓库..." -ForegroundColor Yellow
    git init
    git remote add origin "https://github.com/$Owner/$Repo.git"
} else {
    Write-Host "✅ Git仓库已存在" -ForegroundColor Green
}

# 检查远程仓库
$remotes = git remote -v
if ($remotes -notlike "*origin*") {
    git remote add origin "https://github.com/$Owner/$Repo.git"
}

# 检查是否有未提交的更改
$status = git status --porcelain
if ($status) {
    Write-Host "📋 检测到未提交的更改:" -ForegroundColor Yellow
    Write-Host $status -ForegroundColor White
    
    if (!$Force) {
        $response = Read-Host "是否继续提交更改？(y/N)"
        if ($response -ne "y" -and $response -ne "Y") {
            Write-Host "❌ 操作已取消" -ForegroundColor Red
            exit 1
        }
    }
}

# 添加所有文件
git add .

# 提交更改
git commit -m "release: 发布 DelGuard $Version

- ✨ 新增一键安装功能
- 🔧 支持Windows、Linux、macOS一行命令安装
- 📦 提供完整安装脚本和一行命令脚本
- 🛡️ 智能平台检测和权限验证
- 📖 更新安装文档和使用指南
- 🚀 版本号更新至 v1.4.1"

# 创建标签
git tag -a $Version -m "DelGuard $Version - 一键安装功能发布"

# 推送到GitHub
try {
    Write-Host "📤 推送到GitHub..." -ForegroundColor Yellow
    git push -u origin main
    git push origin $Version
    
    Write-Host "✅ 推送成功！" -ForegroundColor Green
    Write-Host ""
    Write-Host "🔗 GitHub仓库: https://github.com/$Owner/$Repo" -ForegroundColor Cyan
    Write-Host "🏷️  发布标签: https://github.com/$Owner/$Repo/releases/tag/$Version" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "📖 下一步:" -ForegroundColor Yellow
    Write-Host "1. 访问GitHub仓库创建Release" -ForegroundColor White
    Write-Host "2. 上传构建好的二进制文件" -ForegroundColor White
    Write-Host "3. 发布新版本通知用户" -ForegroundColor White
    
} catch {
    Write-Host "❌ 推送失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "请检查网络连接和GitHub权限" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "🎉 发布准备完成！" -ForegroundColor Green