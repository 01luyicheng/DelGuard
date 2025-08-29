# DelGuard 安装脚本

本目录包含 DelGuard 的智能安装脚本，支持跨平台自动检测和安装。

## 🚀 快速安装

### 一键安装（推荐）

```bash
# Unix/Linux/macOS
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | bash

# Windows PowerShell
iwr -useb https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.ps1 | iex
```

### 手动安装

#### Linux/macOS

```bash
# 下载安装脚本
curl -O https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.sh

# 添加执行权限
chmod +x install.sh

# 运行安装
sudo ./install.sh
```

#### Windows

```powershell
# 下载安装脚本
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.ps1" -OutFile "install.ps1"

# 以管理员身份运行
.\install.ps1
```

## 📋 脚本功能

### install.sh (Linux/macOS)

- ✅ 自动检测操作系统和架构
- ✅ 检查系统依赖和权限
- ✅ 从 GitHub Releases 下载最新版本
- ✅ 安装到系统目录 (`/usr/local/bin`)
- ✅ 创建配置目录和默认配置
- ✅ 自动配置 Shell 别名 (bash/zsh)
- ✅ 验证安装完整性

### install.ps1 (Windows)

- ✅ 自动检测系统架构
- ✅ 检查 PowerShell 版本和网络连接
- ✅ 从 GitHub Releases 下载最新版本
- ✅ 安装到程序目录
- ✅ 自动添加到系统 PATH
- ✅ 创建配置目录和默认配置
- ✅ 配置 PowerShell 别名
- ✅ 验证安装完整性

### quick-install.sh (跨平台)

- ✅ 自动检测系统类型
- ✅ 选择合适的安装脚本
- ✅ 一键完成安装

## 🔧 安装选项

### Linux/macOS 选项

```bash
# 指定版本
./install.sh v1.0.0

# 自定义安装目录
INSTALL_DIR="/opt/delguard/bin" ./install.sh

# 自定义配置目录
CONFIG_DIR="$HOME/.delguard" ./install.sh
```

### Windows 选项

```powershell
# 指定版本
.\install.ps1 -Version "v1.0.0"

# 自定义安装目录
.\install.ps1 -InstallDir "C:\Tools\DelGuard"

# 跳过别名配置
.\install.ps1 -NoAlias

# 强制重新安装
.\install.ps1 -Force
```

## 📁 安装位置

### Linux/macOS

- **二进制文件**: `/usr/local/bin/delguard`
- **配置目录**: `~/.config/delguard/`
- **配置文件**: `~/.config/delguard/config.yaml`
- **日志文件**: `~/.config/delguard/delguard.log`

### Windows

- **二进制文件**: `C:\Program Files\DelGuard\delguard.exe`
- **配置目录**: `%APPDATA%\DelGuard\`
- **配置文件**: `%APPDATA%\DelGuard\config.yaml`
- **日志文件**: `%APPDATA%\DelGuard\delguard.log`

## 🎯 别名配置

安装完成后，以下别名将自动配置：

```bash
# 通用别名
del <file>        # 等同于 delguard delete
rm <file>         # 安全替代系统 rm 命令
trash <file>      # 等同于 delguard delete
restore <file>    # 等同于 delguard restore
empty-trash       # 等同于 delguard empty
```

## 🔍 系统要求

### Linux

- **操作系统**: Linux (任何现代发行版)
- **架构**: x86_64, ARM64, ARM
- **依赖**: curl, tar
- **权限**: sudo (用于安装到系统目录)

### macOS

- **操作系统**: macOS 10.12 或更高版本
- **架构**: x86_64, ARM64 (Apple Silicon)
- **依赖**: curl, tar
- **权限**: sudo (用于安装到系统目录)

### Windows

- **操作系统**: Windows 10 或更高版本
- **架构**: x86_64, ARM64, x86
- **依赖**: PowerShell 5.0 或更高版本
- **权限**: 管理员权限

## 🛠️ 故障排除

### 常见问题

1. **权限不足**
   ```bash
   # Linux/macOS: 使用 sudo
   sudo ./install.sh
   
   # Windows: 以管理员身份运行 PowerShell
   ```

2. **网络连接问题**
   ```bash
   # 检查网络连接
   curl -I https://github.com
   
   # 使用代理
   export https_proxy=http://proxy:port
   ./install.sh
   ```

3. **架构不支持**
   ```bash
   # 检查系统架构
   uname -m  # Linux/macOS
   echo $env:PROCESSOR_ARCHITECTURE  # Windows
   ```

4. **下载失败**
   ```bash
   # 手动下载并安装
   wget https://github.com/01luyicheng/DelGuard/releases/latest/download/delguard-linux-amd64.tar.gz
   tar -xzf delguard-linux-amd64.tar.gz
   sudo cp delguard /usr/local/bin/
   ```

### 卸载

#### Linux/macOS

```bash
# 删除二进制文件
sudo rm -f /usr/local/bin/delguard

# 删除配置目录
rm -rf ~/.config/delguard

# 手动删除别名配置
# 编辑 ~/.bashrc, ~/.zshrc 等文件，删除 DelGuard 相关行
```

#### Windows

```powershell
# 删除安装目录
Remove-Item "C:\Program Files\DelGuard" -Recurse -Force

# 删除配置目录
Remove-Item "$env:APPDATA\DelGuard" -Recurse -Force

# 从 PATH 中移除
# 手动编辑系统环境变量

# 删除 PowerShell 别名
# 编辑 PowerShell 配置文件，删除 DelGuard 相关行
```

## 📞 支持

如果遇到安装问题，请：

1. 查看 [GitHub Issues](https://github.com/01luyicheng/DelGuard/issues)
2. 提交新的 Issue 并包含：
   - 操作系统和版本
   - 系统架构
   - 错误信息
   - 安装日志

## 🔄 更新

要更新到最新版本，只需重新运行安装脚本：

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.sh | bash

# Windows
iwr -useb https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.ps1 | iex
```

安装脚本会自动检测并更新到最新版本。