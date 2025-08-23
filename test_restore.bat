@echo off
echo 测试DelGuard文件恢复功能...
echo.

rem 创建测试文件
echo 这是一个测试文件 > test_file.txt
echo 这是另一个测试文件 > test_file2.txt

echo 创建测试文件完成
dir test_file*.txt

echo.
echo 删除测试文件...
delguard test_file.txt test_file2.txt

echo.
echo 列出可恢复文件...
delguard restore -l

echo.
echo 恢复测试文件...
delguard restore "test_file*.txt"

echo.
echo 验证文件已恢复...
dir test_file*.txt

echo.
echo 测试完成！
pause