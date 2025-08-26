# DelGuard 安装指南

## 概述

DelGuard 是一个跨平台的安全删除工具，提供了完善的安装脚本来确保在各种操作系统上的可靠安装。

## 支持的平台

- **Windows**: Windows 10/11 (x64, ARM64)
- **Linux**: Ubuntu, Debian, CentOS, RHEL, Arch Linux, 等
- **macOS**: Intel 和 Apple Silicon (M1/M2)
- **FreeBSD**: 最新版本

## 安装方法

### Windows 安装

#### 方法 1: PowerShell 脚本 (推荐)
```powershell
# 以管理员身份运行 PowerShell
.\install.ps1

# 或指定安装路径
.\install.ps1 --install-path "C:\Program Files\DelGuard"

# 强制重新安装
.\install.ps1 --force
```

#### 方法 2: 批处理脚本
```cmd
# 以管理员身份运行命令提示符
install.bat

# 或双击 install.bat 文件
```

### Linux/macOS 安装

```bash
# 基本安装
./install.sh

# 系统级安装 (需要 sudo)
./install.sh --system-wide

# 指定安装路径
./install.sh --install-path /usr/local/bin

# 强制重新安装
./install.sh --force

# 跳过 Shell 别名配置
./install.sh --skip-aliases
```

## 安装选项

### 通用选项

| 选项 | 描述 | 示例 |
|------|------|------|
| `--install-path` | 指定安装目录 | `--install-path /usr/local/bin` |
| `--force` | 强制覆盖现有安装 | `--force` |
| `--quiet` | 静默安装模式 | `--quiet` |
| `--help` | 显示帮助信息 | `--help` |
| `--version` | 安装指定版本 | `--version v1.2.0` |
| `--debug` | 启用调试输出 | `--debug` |

### Unix 特有选项

| 选项 | 描述 | 示例 |
|------|------|------|
| `--system-wide` | 系统级安装 | `--system-wide` |
| `--skip-aliases` | 跳过 Shell 别名配置 | `--skip-aliases` |
| `--check-only` | 仅检查安装状态 | `--check-only` |
| `--uninstall` | 卸载 DelGuard | `--uninstall` |

### Windows 特有选项

| 选项 | 描述 | 示例 |
|------|------|------|
| `--user-install` | 用户级安装 | `--user-install` |
| `--no-path` | 不添加到 PATH | `--no-path` |
| `--no-aliases` | 不配置别名 | `--no-aliases` |

## 安装后配置

### 1. 重启 Shell 或重新加载配置

**Linux/macOS:**
```bash
# Bash
source ~/.bashrc

# Zsh
source ~/.zshrc

# Fish
source ~/.config/fish/config.fish
```

**Windows PowerShell:**
```powershell
# 重新加载 PowerShell 配置文件
. $PROFILE
```

### 2. 验证安装

```bash
# 检查版本
delguard --version

# 查看帮助
delguard --help

# 测试功能
delguard --check
```

## 使用方法

安装完成后，DelGuard 提供以下命令：

### 基本命令

```bash
# 安全删除文件
del file.txt
delguard file.txt

# 安全删除目录
del -r directory/
delguard -r directory/

# 安全复制 (替换系统 cp)
cp source.txt destination.txt

# 查看回收站
delguard --list

# 恢复文件
delguard --restore file.txt
```

### 高级选项

```bash
# 永久删除 (跳过回收站)
delguard --permanent file.txt

# 安全擦除 (多次覆写)
delguard --secure-erase file.txt

# 批量操作
delguard --batch *.tmp

# 定时清理
delguard --schedule-cleanup 7d
```

## 故障排除

### 常见问题

#### 1. 权限错误

**Windows:**
```powershell
# 以管理员身份运行 PowerShell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Linux/macOS:**
```bash
# 使用 sudo 进行系统级安装
sudo ./install.sh --system-wide

# 或安装到用户目录
./install.sh --install-path ~/bin
```

#### 2. 依赖缺失

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install curl tar golang-go
```

**CentOS/RHEL:**
```bash
sudo yum install curl tar golang
```

**macOS:**
```bash
# 安装 Homebrew (如果未安装)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装依赖
brew install go curl
```

#### 3. 构建失败

```bash
# 手动构建
go mod tidy
go build -o delguard

# 检查 Go 环境
go version
go env
```

#### 4. PATH 配置问题

**手动添加到 PATH:**

**Linux/macOS (.bashrc/.zshrc):**
```bash
export PATH="$HOME/bin:$PATH"
```

**Windows (PowerShell Profile):**
```powershell
$env:PATH += ";C:\Program Files\DelGuard"
```

### 卸载

#### Windows
```powershell
# 使用 PowerShell 脚本卸载
.\install.ps1 --uninstall

# 或手动删除
Remove-Item -Path "C:\Program Files\DelGuard" -Recurse -Force
```

#### Linux/macOS
```bash
# 使用安装脚本卸载
./install.sh --uninstall

# 或手动删除
rm -f ~/bin/delguard
rm -f /usr/local/bin/delguard
```

### 日志和调试

安装过程中的日志文件位置：

- **Windows**: `%TEMP%\delguard-install-*.log`
- **Linux/macOS**: `/tmp/delguard-install-*.log`

启用调试模式：
```bash
# Unix
./install.sh --debug

# Windows
.\install.ps1 -Debug
```

## 高级配置

### 自定义配置文件

DelGuard 支持配置文件自定义：

```json
{
  "trash_dir": "~/.delguard/trash",
  "max_file_size": "1GB",
  "auto_cleanup_days": 30,
  "secure_erase_passes": 3,
  "confirm_delete": true,
  "log_operations": true
}
```

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `DELGUARD_CONFIG` | 配置文件路径 | `~/.delguard/config.json` |
| `DELGUARD_TRASH` | 回收站目录 | `~/.delguard/trash` |
| `DELGUARD_LOG_LEVEL` | 日志级别 | `INFO` |

## 安全注意事项

1. **管理员权限**: Windows 系统级安装需要管理员权限
2. **执行策略**: Windows 可能需要调整 PowerShell 执行策略
3. **防火墙**: 某些防火墙可能阻止下载依赖
4. **杀毒软件**: 可能需要将 DelGuard 添加到白名单

## 支持和反馈

如果遇到安装问题，请：

1. 检查系统要求和依赖
2. 查看安装日志文件
3. 尝试使用 `--debug` 选项
4. 提交 Issue 并附上日志信息

## 更新日志

### v2.1.1 (Enhanced Fixed)
- 修复了所有已知的安装脚本问题
- 增强了跨平台兼容性
- 改进了错误处理和用户体验
- 添加了完整的回滚功能
- 统一了多语言支持

### v2.1.0
- 添加了 PowerShell 安装脚本
- 改进了 Unix Shell 脚本兼容性
- 增加了安装验证功能

### v2.0.0
- 重写了安装系统
- 添加了多平台支持
- 实现了自动依赖检查