/*
 * 文件用途：DelGuard项目快速开始指南
 * 作者：TREA
 * 功能说明：提供项目构建、安装和使用的快速参考
 * 输入：项目源码
 * 输出：构建好的安装包
 * 依赖：.NET 8.0 SDK, WiX Toolset v3
 * 交互方式：命令行执行构建和安装
 */

# DelGuard 快速开始指南

## 🚀 快速安装

### 方法1：PowerShell一键安装（推荐）
```powershell
# 以管理员身份运行PowerShell
Set-ExecutionPolicy Bypass -Scope Process -Force
iwr -useb https://raw.githubusercontent.com/01luyicheng/DelGuard/main/installer/install.ps1 | iex
```

### 方法2：手动下载安装
1. 下载 `DelGuard.msi` 安装包
2. 双击运行安装程序
3. 按照向导完成安装

### 方法3：命令行安装
```cmd
msiexec /i DelGuard.msi /quiet /norestart
```

## 🔧 开发者构建

### 环境要求
- .NET 8.0 SDK 或更高版本
- WiX Toolset v3.11.2
- Windows 10/11 (x64)

### 构建步骤
```cmd
# 1. 克隆项目
git clone https://github.com/01luyicheng/DelGuard.git
cd DelGuard

# 2. 构建主程序
dotnet build -c Release

# 3. 构建安装包
cd installer
build.bat

# 4. 输出文件在 installer\output\DelGuard.msi
```

## 📖 基本用法

```bash
# 删除文件
delguard file.txt

# 强制删除被占用的文件
delguard -f lockedfile.exe

# 递归删除文件夹
delguard -r C:\old_data

# 预览模式（安全确认）
delguard -p C:\temp\*

# 获取帮助
delguard --help
```

## 🛠️ 高级功能

### 批处理集成
```batch
@echo off
:: 静默删除临时文件
delguard -s -f C:\temp\*.tmp

:: 清理日志文件
delguard -r C:\logs\old_logs
```

### PowerShell脚本
```powershell
# 删除大文件
Get-ChildItem C:\temp -Recurse | Where-Object {$_.Length -gt 1GB} | ForEach-Object {
    delguard -f $_.FullName
}
```

## 🔍 故障排除

### 安装问题
- **权限错误**：以管理员身份运行安装程序
- **依赖缺失**：确保安装 .NET 8.0 Runtime
- **PATH问题**：重新打开命令提示符

### 使用问题
- **文件被占用**：使用 `-f` 参数强制删除
- **权限不足**：以管理员身份运行命令提示符
- **路径错误**：检查文件路径是否正确

## 📞 支持

- 📧 邮件支持：support@delguard.com
- 🐛 问题反馈：[GitHub Issues](https://github.com/01luyicheng/DelGuard/issues)
- 📚 详细文档：[Wiki](https://github.com/01luyicheng/DelGuard/wiki)

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

**⭐ 如果DelGuard帮助到了您，请给我们一个Star！**