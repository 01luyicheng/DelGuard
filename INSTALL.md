# DelGuard 一键安装指南

## 🚀 一行命令安装

### Windows (PowerShell)
```powershell
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; .\quick-install.ps1 }"
```

### Linux/macOS (Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash
```

### 备用安装方法 (使用wget)
```bash
wget -qO- https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash
```

## 📦 安装选项

### 指定版本
```powershell
# Windows
.\quick-install.ps1 -Version v1.4.1

# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash -s -- --version v1.4.1
```

### 强制重新安装
```powershell
# Windows
.\quick-install.ps1 -Force

# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash -s -- --force
```

## 🔧 手动安装

### 1. 从GitHub下载
访问 [GitHub Releases](https://github.com/01luyicheng/DelGuard/releases) 下载对应平台的二进制文件。

### 2. 手动安装
```bash
# Linux/macOS
chmod +x delguard-linux-amd64
sudo mv delguard-linux-amd64 /usr/local/bin/delguard

# Windows
# 将 delguard-windows-amd64.exe 重命名为 delguard.exe 并添加到 PATH
```

### 3. 运行安装程序
```bash
delguard install
```

## ✅ 验证安装

安装完成后，运行以下命令验证：

```bash
delguard --version
delguard status
```

## 🗑️ 卸载

### Windows
```powershell
# 如果已添加到PATH
delguard-uninstall

# 或者运行卸载脚本
"$env:ProgramFiles\DelGuard\uninstall.bat"
```

### Linux/macOS
```bash
sudo delguard-uninstall
```

## 📋 系统要求

- **Windows**: Windows 10/11, PowerShell 5.1+
- **Linux**: Ubuntu 18.04+, CentOS 7+, 或其他现代Linux发行版
- **macOS**: macOS 10.14+

## 🔗 相关链接

- [GitHub仓库](https://github.com/01luyicheng/DelGuard)
- [问题反馈](https://github.com/01luyicheng/DelGuard/issues)
- [使用文档](README.md)

## ⚠️ 注意事项

1. **管理员权限**: 安装需要管理员/root权限
2. **防病毒软件**: 某些防病毒软件可能会误报，请添加信任
3. **PATH更新**: 安装后可能需要重新打开终端以使PATH生效
4. **备份**: 安装前建议备份重要数据