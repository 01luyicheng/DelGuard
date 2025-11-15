# DelGuard v1.4.1 快速开始

## 🚀 一行命令安装

### Windows
```powershell
# 一行命令安装（最简单）
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install-oneline.ps1' -UseBasicParsing | Invoke-Expression }"

# 或者使用完整脚本（可自定义参数）
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; .\\quick-install.ps1 }"
```

## ✅ 安装验证

安装完成后，运行以下命令验证：

```bash
# 查看帮助信息
delguard --help

# 查看版本信息
delguard --version
# 应该显示：delguard version 1.4.1

# 查看系统状态
delguard status
# 应该显示系统信息和回收站状态
```

## 📖 基本使用

### 1. 安全删除文件
```bash
# 删除文件（移动到回收站）
del important_file.txt     # Windows

# 删除多个文件
del *.tmp *.log

# 使用通配符删除
del temp_*
```

### 2. 删除文件夹
```bash
# 删除空文件夹
del empty_folder

# 递归删除文件夹及其内容
del -r my_folder/

# 使用通配符删除多个文件夹
del -r temp_*/
```

### 3. 强制删除被占用的文件
```bash
# 强制删除文件（绕过系统限制）
del -f locked_file.exe

# 强制删除文件夹
del -f -r persistent_folder
```

### 4. 预览模式（安全删除）
```bash
# 预览将要删除的文件
del -p *.tmp

# 预览递归删除
del -p -r backup_folder/

# 预览模式不会实际删除，只是显示将要删除的内容
```

## 🎯 常用场景

### 清理临时文件
```bash
# 清理Windows临时文件夹
del -r -s "C:\Windows\Temp\*"

# 清理用户临时文件夹
del -r -s "%TEMP%\*"
```

### 清理日志文件
```bash
# 删除应用程序日志
del -f "C:\ProgramData\App\Logs\*.log"

# 清理事件日志备份
del -f -r "C:\Windows\System32\winevt\Logs\Archive-*.evtx"
```

### 清理浏览器缓存
```bash
# Chrome缓存
del -r -s "%LOCALAPPDATA%\Google\Chrome\User Data\Default\Cache\*"

# Edge缓存
del -r -s "%LOCALAPPDATA%\Microsoft\Edge\User Data\Default\Cache\*"
```

### 卸载软件后清理
```bash
# 删除应用程序残留
del -r -f "C:\Program Files\UninstalledApp\"

# 删除用户配置
del -r -f "%APPDATA%\UninstalledApp\"
```

## 🛡️ 安全特性

### 回收站保护
- 所有删除的文件/文件夹默认移动到回收站
- 可从回收站恢复误删的文件
- 支持回收站大小限制和自动清理

### 系统保护
```bash
# DelGuard会阻止删除以下关键系统目录
# C:\Windows
# C:\Program Files
# C:\Program Files (x86)
# C:\ProgramData
```

### 错误处理
```bash
# 详细的错误报告
del -v problematic_file
# 会显示具体的错误信息和解决建议

# 静默模式（适合脚本）
del -s unwanted_files
# 减少输出，适合批处理脚本
```

## ⚙️ 高级配置

### 1. 配置文件
DelGuard使用JSON配置文件，默认位置：
- Windows: `C:\Program Files\DelGuard\config\windows.json`

### 2. 配置示例
```json
{
  "uiOption": "OnlyErrorDialogs",
  "logLevel": "info"
}
```

### 3. 环境变量
```bash
# 设置默认删除模式
set DELGUARD_DEFAULT_MODE=recycle

# 设置默认日志级别
set DELGUARD_LOG_LEVEL=warning
```

## 🔧 故障排除

### 常见问题
```bash
# Q: 删除操作失败，提示"访问被拒绝"
# A: 使用管理员权限运行命令提示符

# Q: 文件被其他程序占用，无法删除
# A: 使用 -f 参数强制删除，或先关闭占用该文件的程序

# Q: 删除大文件夹时程序无响应
# A: 使用 -v 参数查看详细进度，耐心等待操作完成
```

### 日志文件位置
```bash
# Windows
C:\Program Files\DelGuard\logs\delguard_YYYY-MM-DD.log
```

## 🎉 最佳实践

1. **预览模式优先**：对不熟悉的删除操作，先使用 `-p` 参数预览
2. **定期清理回收站**：定期清空回收站以释放磁盘空间
3. **使用详细模式**：对于复杂操作，使用 `-v` 参数查看详细日志
4. **备份重要数据**：删除重要文件前先备份
5. **批量处理**：结合通配符和递归参数高效处理大量文件

## 📚 更多资源

- [完整文档](README.md)
- [安装指南](INSTALL.md)
- [GitHub仓库](https://github.com/01luyicheng/DelGuard)
- [问题反馈](https://github.com/01luyicheng/DelGuard/issues)

---

💡 **提示**: DelGuard将文件移动到回收站而非永久删除，提供了额外的安全保障。但请注意，回收站空间有限，建议定期清理。