# DelGuard 安装指南

DelGuard 是一个跨平台的安全删除工具，提供比系统默认删除命令更安全的文件操作体验。

## 快速安装

### Windows 用户

使用 PowerShell（推荐）：
```powershell
# 下载并运行安装脚本
.\install.ps1

# 或者强制重新安装
.\install.ps1 -Force

# 自定义安装路径
.\install.ps1 -InstallPath "C:\tools"
```

### Linux/macOS 用户

使用终端：
```bash
# 下载并运行安装脚本
chmod +x install.sh
./install.sh

# 或者强制重新安装
./install.sh --force

# 自定义安装路径
./install.sh --install-path /usr/local/bin
```

## 安装后使用

安装完成后，重启终端或PowerShell，然后可以使用以下命令：

- `del <文件>` - 安全删除文件
- `rm <文件>` - 安全删除文件（Unix风格）
- `cp <源> <目标>` - 安全复制文件
- `delguard --help` - 查看完整帮助

## 卸载

### Windows
```powershell
.\install.ps1 -Uninstall
```

### Linux/macOS
```bash
./install.sh --uninstall
```

## 故障排除

如果遇到问题：

1. **权限错误**：确保有足够权限写入安装目录
2. **PATH问题**：重启终端或手动添加安装路径到PATH
3. **别名冲突**：使用 `-Force` 参数强制重新安装
4. **构建失败**：确保已安装Go语言环境

## 支持的系统

- Windows 10/11 (PowerShell 5.1+)
- Linux (bash, zsh, fish)
- macOS (bash, zsh, fish)