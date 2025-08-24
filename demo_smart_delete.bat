@echo off
echo 🚀 DelGuard智能删除功能演示
echo ================================

echo.
echo 1. 创建测试文件...
echo test content > test_document.txt
echo test content > test_file.txt
echo test content > sample.log
echo test content > readme.md

echo.
echo 2. 测试智能搜索功能
echo 尝试删除不存在的文件 "test_doc"，应该会智能搜索相似文件
delguard.exe test_doc

echo.
echo 3. 测试正则表达式批量删除
echo 删除所有 .txt 文件
delguard.exe *.txt --force-confirm

echo.
echo 4. 清理剩余文件
del sample.log readme.md 2>nul

echo.
echo ✅ 演示完成！
pause