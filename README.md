# DelGuard - 安全删除工具

DelGuard 是一款专业的安全删除工具，提供可靠的文件删除功能，支持回收站删除和永久删除两种模式。

## 功能特性

- 🗑️ **智能删除**: 支持回收站删除和永久删除
- 🛡️ **安全保护**: 内置安全检查，防止误删重要文件
- 🌍 **多语言支持**: 支持中文、英文等多种语言
- ⚡ **高性能**: 优化的删除算法，支持大文件处理
- 📊 **详细日志**: 完整的操作日志记录
- 🎯 **交互模式**: 支持自动、询问、从不三种交互模式

## 快速安装

### Windows (PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/01luyicheng/DelGuard/main/install.ps1 | iex
```

### Linux/macOS (Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/install.sh | bash
```

## 使用方法

### 基本用法
```bash
# 删除单个文件（默认移动到回收站）
delguard file.txt

# 删除多个文件
delguard file1.txt file2.txt dir1/

# 永久删除文件
delguard -p file.txt

# 交互式删除
delguard -i file.txt
```

### 命令行选项
```
-p, --permanent     永久删除文件（不使用回收站）
-i, --interactive   交互式模式，每个文件都询问
-f, --force         强制删除，跳过安全检查
-v, --verbose       详细输出
-q, --quiet         静默模式
-r, --recursive     递归删除目录
--config            指定配置文件路径
--version           显示版本信息
--help              显示帮助信息
```

### 配置文件

DelGuard 支持通过配置文件自定义行为。配置文件位置：
- Windows: `%APPDATA%\DelGuard\config.json`
- Linux/macOS: `~/.config/delguard/config.json`

示例配置：
```json
{
  "use_recycle_bin": true,
  "interactive_mode": "auto",
  "language": "zh-CN",
  "max_file_size": 1073741824,
  "log_level": "info",
  "log_file": "",
  "safe_mode": true,
  "backup_before_delete": false
}
```

## 安全特性

DelGuard 内置多重安全保护机制：

1. **系统文件保护**: 自动识别并保护系统关键文件
2. **大文件警告**: 删除大文件时提供额外确认
3. **路径验证**: 验证文件路径的合法性和安全性
4. **权限检查**: 确保有足够权限执行删除操作
5. **回滚机制**: 支持从回收站恢复已删除文件

## 开发

### 构建要求
- Go 1.19 或更高版本
- 支持的操作系统：Windows, Linux, macOS

### 从源码构建
```bash
git clone https://github.com/01luyicheng/DelGuard.git
cd DelGuard
go build -o delguard
```

### 运行测试
```bash
go test -v ./...
```

### 性能测试
```bash
go test -bench=. ./tests/benchmarks/
```

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 支持

如果您遇到问题或有建议，请：
1. 查看 [FAQ](docs/FAQ.md)
2. 搜索现有的 [Issues](https://github.com/01luyicheng/DelGuard/issues)
3. 创建新的 Issue

---

**注意**: DelGuard 是一个安全删除工具，不包含任何恶意软件检测功能。请谨慎使用永久删除功能。