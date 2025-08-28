# DelGuard 功能测试脚本

Write-Host "=== DelGuard 功能测试开始 ===" -ForegroundColor Green

# 设置测试环境
$testDir = "test_delguard_comprehensive"
$delguardExe = ".\delguard.exe"

# 清理之前的测试
if (Test-Path $testDir) {
    Remove-Item -Recurse -Force $testDir
}

# 创建测试目录和文件
Write-Host "`n1. 创建测试环境..." -ForegroundColor Yellow
New-Item -ItemType Directory -Path $testDir -Force | Out-Null
Set-Content -Path "$testDir\test1.txt" -Value "这是测试文件1"
Set-Content -Path "$testDir\test2.log" -Value "这是日志文件"
Set-Content -Path "$testDir\important.doc" -Value "重要文档内容"
New-Item -ItemType Directory -Path "$testDir\subdir" -Force | Out-Null
Set-Content -Path "$testDir\subdir\nested.txt" -Value "嵌套文件"

Write-Host "测试文件创建完成" -ForegroundColor Green

# 测试版本信息
Write-Host "`n2. 测试版本信息..." -ForegroundColor Yellow
& $delguardExe version

# 测试帮助信息
Write-Host "`n3. 测试帮助信息..." -ForegroundColor Yellow
& $delguardExe help

# 测试配置显示
Write-Host "`n4. 测试配置显示..." -ForegroundColor Yellow
& $delguardExe config show

# 测试搜索功能
Write-Host "`n5. 测试搜索功能..." -ForegroundColor Yellow
Write-Host "搜索 .txt 文件:"
& $delguardExe search $testDir

# 测试删除功能
Write-Host "`n6. 测试删除功能..." -ForegroundColor Yellow
Write-Host "删除测试文件:"
& $delguardExe delete "$testDir\test1.txt" -v

# 验证文件是否被删除
if (Test-Path "$testDir\test1.txt") {
    Write-Host "❌ 文件删除失败" -ForegroundColor Red
} else {
    Write-Host "✅ 文件删除成功" -ForegroundColor Green
}

# 测试恢复功能
Write-Host "`n7. 测试恢复功能..." -ForegroundColor Yellow
& $delguardExe restore --list

# 测试配置设置
Write-Host "`n8. 测试配置设置..." -ForegroundColor Yellow
& $delguardExe config set log_level debug

Write-Host "`n=== DelGuard 功能测试完成 ===" -ForegroundColor Green

# 清理测试环境
Write-Host "`n清理测试环境..." -ForegroundColor Yellow
if (Test-Path $testDir) {
    Remove-Item -Recurse -Force $testDir
}
Write-Host "测试环境清理完成" -ForegroundColor Green