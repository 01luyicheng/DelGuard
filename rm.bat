@echo off
set "args=%*"
if "%args%"=="" (
    echo 用法: rm [选项] 文件...
    echo 选项:
    echo   -f, --force     强制删除
    echo   -r, --recursive 递归删除目录
    echo   -v, --verbose   详细输出
    exit /b 1
)
"C:\Program Files\DelGuard\delguard.exe" delete %*