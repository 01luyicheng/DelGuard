# DelGuard 快速开始指南

## 安装

### 方法一：一键安装（推荐）

**Windows (PowerShell)**
```powershell
iwr -useb https://raw.githubusercontent.com/01luyicheng/DelGuard/main/install.ps1 | iex
```

**Linux/macOS (Bash)**
```bash
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/install.sh | bash
```

### 方法二：手动安装

1. 下载最新版本的可执行文件
2. 将文件放置到系统PATH目录
3. 运行 `delguard --version` 验证安装

## 基本使用

### 删除单个文件
```bash
delguard file.txt
```

### 删除多个文件
```bash
delguard file1.txt file2.txt folder/
```

### 永久删除（跳过回收站）
```bash
delguard -p file.txt
```

### 交互式删除
```bash
delguard -i file.txt
```

### 递归删除目录
```bash
delguard -r folder/
```

## 常用选项

| 选项 | 说明 |
|------|------|
| `-p, --permanent` | 永久删除，不使用回收站 |
| `-i, --interactive` | 交互式模式，每个文件都询问 |
| `-f, --force` | 强制删除，跳过安全检查 |
| `-r, --recursive` | 递归删除目录 |
| `-v, --verbose` | 详细输出 |
| `-q, --quiet` | 静默模式 |
| `--config` | 指定配置文件路径 |

## 配置文件

DelGuard 会在以下位置查找配置文件：
- Windows: `%APPDATA%\DelGuard\config.json`
- Linux/macOS: `~/.config/delguard/config.json`

### 示例配置
```json
{
  "use_recycle_bin": true,
  "interactive_mode": "auto",
  "language": "zh-CN",
  "safe_mode": true,
  "log_level": "info"
}
```

## 安全特性

### 自动保护
- ✅ 系统文件自动保护
- ✅ 重要目录检查
- ✅ 大文件删除确认
- ✅ 权限验证

### 手动确认
使用 `-i` 选项可以对每个文件进行确认：
```bash
delguard -i *.tmp
```

## 恢复文件

### 从回收站恢复
```bash
delguard --restore file.txt
```

### 查看回收站内容
```bash
delguard --list-trash
```

## 高级功能

### 搜索并删除
```bash
delguard --search "*.tmp" --delete
```

### 按大小过滤
```bash
delguard --size ">100MB" folder/
```

### 按时间过滤
```bash
delguard --older-than "30d" logs/
```

## 故障排除

### 权限问题
如果遇到权限错误，尝试以管理员身份运行：

**Windows**
```powershell
# 以管理员身份运行PowerShell
delguard file.txt
```

**Linux/macOS**
```bash
sudo delguard file.txt
```

### 配置问题
重置配置到默认值：
```bash
delguard --reset-config
```

### 查看日志
```bash
delguard --show-logs
```

## 获取帮助

### 命令行帮助
```bash
delguard --help
```

### 查看版本
```bash
delguard --version
```

### 在线文档
- [完整文档](docs/README.md)
- [安全指南](docs/SECURITY.md)
- [开发文档](docs/DEVELOPMENT.md)

## 常见问题

**Q: 如何确保文件被安全删除？**
A: 使用 `-p` 选项进行永久删除，DelGuard 会进行多次覆写。

**Q: 可以恢复永久删除的文件吗？**
A: 永久删除的文件无法通过 DelGuard 恢复，请谨慎使用。

**Q: 如何批量删除特定类型的文件？**
A: 使用通配符：`delguard *.tmp` 或搜索功能。

**Q: 程序运行很慢怎么办？**
A: 检查是否在处理大量文件，可以使用 `-v` 查看进度。

## 支持

如果遇到问题：
1. 查看 [FAQ](docs/FAQ.md)
2. 搜索 [Issues](https://github.com/01luyicheng/DelGuard/issues)
3. 创建新的 Issue
4. 发送邮件到支持团队