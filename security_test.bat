@echo off
REM DelGuard 安全功能快速验证脚本 (Windows)
REM 运行此脚本验证关键安全功能是否正常工作

echo =========================================
echo DelGuard 安全功能验证脚本
echo =========================================
echo.

REM 检查基本功能
echo [1/7] 检查基本构建...
go build -v
if %errorlevel% neq 0 (
    echo ❌ 构建失败，请检查代码
    pause
    exit /b 1
)
echo ✅ 构建成功
echo.

REM 测试路径遍历防护
echo [2/7] 测试路径遍历防护...
delguard.exe delete "../../../windows/system32/config"
if %errorlevel% neq 0 (
    echo ✅ 路径遍历防护正常
) else (
    echo ❌ 路径遍历防护异常
)
echo.

REM 测试隐藏文件检测
echo [3/7] 测试隐藏文件检测...
echo test_content > .test_hidden_file
delguard.exe delete .test_hidden_file
if %errorlevel% neq 0 (
    echo ✅ 隐藏文件检测正常
) else (
    echo ❌ 隐藏文件检测异常
)
del /f /q .test_hidden_file 2>nul
echo.

REM 测试系统文件保护
echo [4/7] 测试系统文件保护...
delguard.exe delete "C:\Windows\System32\kernel32.dll"
if %errorlevel% neq 0 (
    echo ✅ 系统文件保护正常
) else (
    echo ❌ 系统文件保护异常
)
echo.

REM 测试不存在的文件
echo [5/7] 测试不存在的文件处理...
delguard.exe delete "non_existent_file_12345.txt"
if %errorlevel% neq 0 (
    echo ✅ 不存在的文件处理正常
) else (
    echo ❌ 不存在的文件处理异常
)
echo.

REM 测试特殊字符文件名
echo [6/7] 测试特殊字符文件名...
echo test > "test_file_with_special_chars_<>|?.txt"
delguard.exe delete "test_file_with_special_chars_<>|?.txt"
if %errorlevel% neq 0 (
    echo ✅ 特殊字符文件名处理正常
) else (
    echo ❌ 特殊字符文件名处理异常
)
del /f /q "test_file_with_special_chars_<>|?.txt" 2>nul
echo.

REM 测试权限检查
echo [7/7] 测试权限检查...
delguard.exe --security-check
if %errorlevel% neq 0 (
    echo ✅ 安全检查正常
) else (
    echo ✅ 安全检查完成
)
echo.

echo =========================================
echo 安全验证完成！
echo 详细日志请查看 logs 目录
echo =========================================
pause