# DelGuard 功能测试脚本
Write-Host "=== DelGuard 功能测试 ===" -ForegroundColor Cyan

# 测试版本信息
Write-Host "`n1. 测试版本信息:" -ForegroundColor Yellow
& "C:\Program Files\DelGuard\delguard.exe" version

# 测试配置显示
Write-Host "`n2. 测试配置显示:" -ForegroundColor Yellow
& "C:\Program Files\DelGuard\delguard.exe" config show

# 测试智能搜索
Write-Host "`n3. 测试智能搜索功能:" -ForegroundColor Yellow
& "C:\Program Files\DelGuard\delguard.exe" search "*.txt" --verbose

# 创建测试文件并测试删除恢复
Write-Host "`n4. 测试文件删除和恢复:" -ForegroundColor Yellow
"测试内容 - $(Get-Date)" | Out-File -FilePath "delguard_test.txt" -Encoding UTF8
Write-Host "创建测试文件: delguard_test.txt"

Write-Host "删除测试文件..."
& "C:\Program Files\DelGuard\delguard.exe" delete "delguard_test.txt" --verbose

Write-Host "搜索已删除的文件..."
& "C:\Program Files\DelGuard\delguard.exe" search "delguard_test" --verbose

Write-Host "尝试恢复文件..."
& "C:\Program Files\DelGuard\delguard.exe" restore "delguard_test.txt" --verbose

# 检查文件是否恢复
if (Test-Path "delguard_test.txt") {
    Write-Host "✓ 文件恢复成功!" -ForegroundColor Green
    Remove-Item "delguard_test.txt" -Force
} else {
    Write-Host "✗ 文件恢复失败" -ForegroundColor Red
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Cyan