# DelGuard 系统级安装指南

## 目标：在任何位置直接使用 `cp` 命令

### 方法1：添加到用户PATH（推荐）

#### 手动步骤：
1. **复制当前目录路径**：
   ```
   C:\Users\21601\Documents\project\DelGuard
   ```

2. **添加到用户PATH**：
   - 按 `Win + R`，输入 `sysdm.cpl`，回车
   - 点击"环境变量"
   - 在"用户变量"中找到 `PATH`，点击"编辑"
   - 点击"新建"，粘贴上面的路径
   - 点击"确定"保存所有窗口

3. **验证安装**：
   重新打开命令提示符或PowerShell，输入：
   ```bash
   cp --help
   ```

### 方法2：创建用户目录快捷方式

#### 创建批处理文件：
在当前目录运行：
```batch
@echo off
echo 正在创建用户级命令...

:: 创建用户目录命令
set "USER_BIN=%USERPROFILE%\bin"
if not exist "%USER_BIN%" mkdir "%USER_BIN%"

:: 复制可执行文件
copy /y delguard.exe "%USER_BIN%\delguard.exe"

:: 创建批处理命令
echo @"%USER_BIN%\delguard.exe" --cp %%* > "%USER_BIN%\cp.bat"
echo @"%USER_BIN%\delguard.exe" %%* > "%USER_BIN%\del.bat"
echo @"%USER_BIN%\delguard.exe" %%* > "%USER_BIN%\rm.bat"

:: 添加到用户PATH
setx PATH "%PATH%;%USER_BIN%"

echo 安装完成！请重新打开终端。
pause
```

### 方法3：管理员权限安装（系统级）

#### 以管理员身份运行：
1. 右键点击命令提示符，选择"以管理员身份运行"
2. 运行以下命令：

```batch
@echo off
echo 正在安装系统级命令...

:: 创建系统目录
set "SYSTEM_DIR=C:\Windows\System32"

:: 复制可执行文件（需要管理员权限）
copy /y delguard.exe "%SYSTEM_DIR%\delguard.exe"

:: 创建系统级批处理命令
echo @"%SYSTEM_DIR%\delguard.exe" --cp %%* > "%SYSTEM_DIR%\cp.bat"
echo @"%SYSTEM_DIR%\delguard.exe" %%* > "%SYSTEM_DIR%\del.bat"
echo @"%SYSTEM_DIR%\delguard.exe" %%* > "%SYSTEM_DIR%\rm.bat"

echo 系统级安装完成！
echo 现在你可以在任何位置使用：
echo   cp source.txt dest.txt
echo   del filename
echo   rm filename
pause
```

### 方法4：PowerShell一键安装

#### 运行PowerShell脚本：
以管理员身份打开PowerShell，运行：

```powershell
# 一键安装脚本
$CurrentDir = Get-Location
$ExePath = Join-Path $CurrentDir "delguard.exe"
$UserBin = Join-Path $env:USERPROFILE "bin"

# 创建用户bin目录
if (!(Test-Path $UserBin)) {
    New-Item -ItemType Directory -Path $UserBin -Force
}

# 复制文件
Copy-Item $ExePath $UserBin -Force

# 创建命令
$Commands = @{
    "cp" = "--cp"
    "del" = ""
    "rm" = ""
}

foreach ($cmd in $Commands.Keys) {
    $cmdPath = Join-Path $UserBin "$cmd.bat"
    $content = "@`"$UserBin\delguard.exe`" $($Commands[$cmd]) %*"
    Set-Content $cmdPath $content -Encoding ASCII
}

# 添加到用户PATH
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$UserBin*") {
    $NewPath = $UserBin + ";" + $UserPath
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "已添加到用户PATH" -ForegroundColor Green
}

Write-Host "安装完成！请重新打开PowerShell或CMD窗口。" -ForegroundColor Green
```

### 验证安装

安装完成后，重新打开终端，测试：

```bash
# 测试cp命令
cp test.txt test_copy.txt

# 测试del命令
del test_copy.txt

# 测试帮助
cp --help
```

### 常见问题解决

#### PowerShell配置错误修复：
如果PowerShell配置文件有语法错误，运行：

```powershell
# 修复PowerShell配置
$ProfilePath = $PROFILE
if (Test-Path $ProfilePath) {
    $Content = Get-Content $ProfilePath -Raw
    $Content = $Content -replace '.*DelGuard.*', ''
    Set-Content $ProfilePath $Content.Trim()
    Write-Host "PowerShell配置已修复" -ForegroundColor Green
}
```

#### 快速测试（无需安装）
在当前目录直接测试：
```bash
# 使用完整路径
C:\Users\21601\Documents\project\DelGuard\delguard.exe --cp test.txt test_copy.txt

# 或使用相对路径
.\delguard.exe --cp test.txt test_copy.txt
```

选择最适合你的安装方法，完成后即可在任何位置使用 `cp` 命令！