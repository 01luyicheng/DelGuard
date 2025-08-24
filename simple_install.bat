@echo off
echo 正在设置DelGuard系统级命令...

:: 获取当前目录
set "CURRENT_DIR=%~dp0"
set "EXE_PATH=%CURRENT_DIR%delguard.exe"

:: 检查文件是否存在
if not exist "%EXE_PATH%" (
    echo 错误: 未找到 delguard.exe
    pause
    exit /b 1
)

echo.
echo 你可以通过以下方式在任何位置使用cp命令：
echo.
echo 方法1: 将当前目录添加到PATH
echo   setx PATH "%PATH%;%CURRENT_DIR%" /M
echo.
echo 方法2: 复制到系统目录
echo   复制 delguard.exe 到 C:\Windows\System32\ 或 C:\Windows\
echo.
echo 方法3: 使用完整路径
echo   "%CURRENT_DIR%delguard.exe" --cp source.txt dest.txt
echo.
echo 方法4: 创建快捷方式
echo   在当前目录创建 cp.bat 文件，内容为：
echo   @"%CURRENT_DIR%delguard.exe" --cp %%*
echo.

:: 创建本地批处理文件
echo @"%CURRENT_DIR%delguard.exe" --cp %%* > cp.bat
echo @"%CURRENT_DIR%delguard.exe" %%* > del.bat
echo @"%CURRENT_DIR%delguard.exe" %%* > rm.bat

echo.
echo 已创建本地批处理文件: cp.bat, del.bat, rm.bat
echo 你可以将这些文件所在目录添加到PATH，或直接复制到系统目录
echo.
pause