# DelGuard 一键安装指南

DelGuard 是一个跨平台的文件安全删除工具，提供简单易用的一键安装方式。

## 🚀 一键安装

### Linux/macOS

使用 curl 安装：
```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/DelGuard/main/install.sh | bash
```

使用 wget 安装：
```bash
wget -qO- https://raw.githubusercontent.com/yourusername/DelGuard/main/install.sh | bash
```

### Windows

在 PowerShell 中运行：
```powershell
iwr -useb https://raw.githubusercontent.com/yourusername/DelGuard/main/install.ps1 | iex
```

或者：
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/yourusername/DelGuard/main/install.ps1'))
```

### 手动下载安装

1. 访问 [GitHub Releases](https://github.com/yourusername/DelGuard/releases)
2. 下载对应平台的二进制文件
3. 解压并移动到系统 PATH 目录

## 📋 系统要求

### 支持的操作系统
- **Windows**: Windows 7/8/10/11 (x64, ARM64)
- **macOS**: macOS 10.12+ (Intel, Apple Silicon)
- **Linux**: Ubuntu, Debian, CentOS, Fedora, Alpine 等主流发行版 (x64, ARM64, ARM)

### 依赖要求
- **Windows**: PowerShell 5.0+ 或命令提示符
- **Linux/macOS**: Bash, curl 或 wget

## ⚙️ 安装选项

### 自定义安装目录

#### Linux/macOS
```bash
# 安装到用户目录
curl -fsSL https://raw.githubusercontent.com/yourusername/DelGuard/main/install.sh | bash -s -- -d ~/.local/bin

# 安装到系统目录 (需要 sudo)
curl -fsSL https://raw.githubusercontent.com/yourusername/DelGuard/main/install.sh | sudo bash -s -- -d /usr/local/bin
```

#### Windows
```powershell
# 安装到指定目录
. { iwr -useb https://raw.githubusercontent.com/yourusername/DelGuard/main/install.ps1 } | iex; install -InstallDir "C:\Tools"
```

### 安装特定版本

#### Linux/macOS
```bash
# 安装 v1.0.0 版本
curl -fsSL https://raw.githubusercontent.com/yourusername/DelGuard/main/install.sh | bash -s -- -v v1.0.0
```

#### Windows
```powershell
# 安装 v1.0.0 版本
. { iwr -useb https://raw.githubusercontent.com/yourusername/DelGuard/main/install.ps1 } | iex; install -Version v1.0.0
```

## 🎯 使用方法

安装完成后，您可以使用以下命令：

```bash
# 查看帮助
delguard --help

# 安全删除文件
delguard file.txt

# 删除多个文件
delguard file1.txt file2.txt

# 永久删除（不经过回收站）
delguard -p file.txt

# 恢复文件
delguard --restore

# 查看删除历史
delguard --history
```

## 🔧 验证安装

### Linux/macOS
```bash
# 检查版本
delguard --version

# 检查安装位置
which delguard
```

### Windows
```powershell
# 检查版本
delguard --version

# 检查安装位置
Get-Command delguard
```

## 🔄 更新

### 自动更新
DelGuard 支持自动更新：
```bash
delguard --update
```

### 手动更新
重新运行一键安装脚本即可更新到最新版本。

## 🗑️ 卸载

### Linux/macOS
```bash
# 使用安装脚本卸载
curl -fsSL https://raw.githubusercontent.com/yourusername/DelGuard/main/install.sh | bash -s -- --uninstall

# 或手动删除
rm -f $(which delguard)
```

### Windows
```powershell
# 使用安装脚本卸载
. { iwr -useb https://raw.githubusercontent.com/yourusername/DelGuard/main/install.ps1 } | iex; install -Uninstall

# 或手动删除
Remove-Item -Path "$env:USERPROFILE\bin\delguard.exe" -Force
```

## 📦 包管理器安装 (即将支持)

### Homebrew (macOS/Linux)
```bash
brew install delguard
```

### Chocolatey (Windows)
```powershell
choco install delguard
```

### Scoop (Windows)
```powershell
scoop install delguard
```

## 🐛 故障排除

### 常见问题

#### 1. 权限错误 (Linux/macOS)
```bash
# 使用 sudo 安装到系统目录
curl -fsSL https://raw.githubusercontent.com/yourusername/DelGuard/main/install.sh | sudo bash

# 或安装到用户目录
curl -fsSL https://raw.githubusercontent.com/yourusername/DelGuard/main/install.sh | bash -s -- -d ~/.local/bin
```

#### 2. 执行策略错误 (Windows)
```powershell
# 临时允许脚本执行
Set-ExecutionPolicy Bypass -Scope Process -Force
. { iwr -useb https://raw.githubusercontent.com/yourusername/DelGuard/main/install.ps1 } | iex
```

#### 3. 网络连接问题
- 检查网络连接
- 尝试使用代理
- 手动下载安装包

#### 4. 找不到命令
安装后可能需要重新打开终端，或手动添加安装目录到 PATH。

### 获取帮助

- 📖 [完整文档](https://github.com/yourusername/DelGuard/wiki)
- 🐛 [报告问题](https://github.com/yourusername/DelGuard/issues)
- 💬 [讨论区](https://github.com/yourusername/DelGuard/discussions)

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。