@echo off
echo === DelGuard 文件覆盖保护功能测试 ===

set TEST_DIR=%TEMP%\delguard_test_%RANDOM%
mkdir %TEST_DIR%
cd %TEST_DIR%

echo 创建测试文件...
echo 这是原始文件内容 > original.txt
echo 这是新文件内容 > new.txt

echo.
echo 测试1: 启用覆盖保护
delguard --protect
if %ERRORLEVEL% NEQ 0 goto error

echo.
echo 测试2: 执行安全复制操作
:: 这里可以添加实际的安全复制测试
echo 测试文件创建完成

pause
goto cleanup

:error
echo 测试失败！
pause

:cleanup
cd ..
rmdir /s /q %TEST_DIR%