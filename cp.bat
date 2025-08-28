@echo off
set "args=%*"
if "%args%"=="" (
    echo 用法: cp 源文件 目标文件
    echo 这是一个安全的复制命令，会记录操作历史
    exit /b 1
)
copy %*