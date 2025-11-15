# DelGuard 一键安装指南

## 🚀 一行命令安装

### Windows (PowerShell)
```powershell
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; .\\quick-install.ps1 }"
```

### Windows (管理员CMD)
```cmd
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; quick-install.ps1 }"
```

## 📦 安装选项

### 指定版本
```powershell
# Windows
.\quick-install.ps1 -Version v1.4.1
```

### 强制重新安装
```powershell
# Windows
.\quick-install.ps1 -Force
```

## 🔧 手动安装

### 1. 从GitHub下载
访问 [GitHub Releases](https://github.com/01luyicheng/DelGuard/releases) 下载Windows版本的安装包。

### 2. 手动安装
```cmd
# 将下载的delguard.exe放到目标目录
# 例如：C:\Program Files\DelGuard\
mkdir "%ProgramFiles%\DelGuard"
copy delguard.exe "%ProgramFiles%\DelGuard\"
```

### 3. 运行安装程序
```cmd
# 使用提供的MSI安装包
DelGuard.msi
```

## ✅ 验证安装

安装完成后，运行以下命令验证：

```cmd
delguard --help
```

## 🗑️ 卸载

### Windows
```powershell
# 使用提供的卸载脚本
"$env:ProgramFiles\DelGuard\uninstall.bat"

# 或者通过控制面板
# 控制面板 -> 程序 -> 程序和功能 -> DelGuard -> 卸载
```

## 📋 系统要求

- **Windows**: Windows 10/11, PowerShell 5.1+
- **运行时**: .NET 8.0 Runtime 或更高版本
- **权限**: 安装需要管理员权限

## 🔗 相关链接

- [GitHub仓库](https://github.com/01luyicheng/DelGuard)
- [问题反馈](https://github.com/01luyicheng/DelGuard/issues)
- [使用文档](README.md)

## ⚠️ 注意事项

1. **管理员权限**: 安装需要管理员权限
2. **防病毒软件**: 某些防病毒软件可能会误报，请添加信任
3. **PATH更新**: 安装后可能需要重新打开终端以使PATH生效
4. **.NET运行时**: 确保系统已安装.NET 8.0 Runtime