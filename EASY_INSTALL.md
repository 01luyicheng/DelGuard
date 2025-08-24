# 快速安装指南：在任何位置使用cp命令

## 方法1：直接运行（无需安装）
在任何目录下，使用完整路径运行：
```
c:\Users\21601\Documents\project\DelGuard\delguard.exe --cp source.txt dest.txt
```

## 方法2：添加到PATH（推荐）

### 步骤1：打开PowerShell
右键开始菜单 → "Windows PowerShell"

### 步骤2：运行以下命令
```powershell
# 添加项目目录到用户PATH
$projectDir = "c:\Users\21601\Documents\project\DelGuard"
$currentPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
if ($currentPath -notlike "*$projectDir*") {
    [Environment]::SetEnvironmentVariable('PATH', "$projectDir;$currentPath", 'User')
    Write-Host "已添加到PATH，请重启终端" -ForegroundColor Green
} else {
    Write-Host "已存在于PATH中" -ForegroundColor Yellow
}
```

### 步骤3：重启终端后测试
```
cp --help
cp source.txt dest.txt
```

## 方法3：创建快捷方式

### 创建cp.bat文件
在任意文本编辑器中创建 `cp.bat` 文件，内容：
```batch
@c:\Users\21601\Documents\project\DelGuard\delguard.exe --cp %*
```

将文件保存到：`%USERPROFILE%\bin\cp.bat`

### 添加用户bin目录到PATH
```cmd
setx PATH "%PATH%;%USERPROFILE%\bin"
```

## 验证安装
重启命令提示符或PowerShell，然后运行：
```
cp --help
cp test.txt test2.txt
```

## 修复PowerShell配置错误
如果看到PowerShell错误，修复方法：
1. 打开记事本：
   ```
   notepad $PROFILE
   ```
2. 找到第16行，修复字符串引号
3. 保存文件并重启PowerShell