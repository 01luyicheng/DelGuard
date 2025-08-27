# DelGuard - 智能文件删除保护工具

DelGuard 是一个跨平台的安全文件删除工具，通过将文件移动到系统回收站而非直接删除，为您的数据提供额外的保护层。

## ✨ 新功能

### 🔔 智能提示系统
- **删除提示**: 删除文件后显示 `DelGuard: [文件名]已被移动到回收站`
- **覆盖保护**: 覆盖文件前显示 `DelGuard: [文件名] 原文件已备份到回收站`
- **错误处理**: 智能识别错误类型并提供详细的解决建议

### 🛡️ 安全特性
- **回收站保护**: 文件被移动到系统回收站，而非直接删除
- **覆盖保护**: 自动备份将被覆盖的文件
- **安全检查**: 删除前进行多项安全检查
- **跨平台**: 支持 Windows、macOS 和 Linux

## 🚀 安装

### 从源码编译
```bash
git clone https://github.com/01luyicheng/DelGuard.git
cd DelGuard
go build -o delguard ./cmd/delguard
```

### 使用安装脚本
**Windows:**
```powershell
.\install.ps1
```

**Linux/macOS:**
```bash
./install.sh
```

## 📖 使用方法

### 基本命令
```bash
# 安全删除文件
delguard delete file.txt
delguard del file.txt

# 搜索文件
delguard search "*.txt"
delguard find "*.log"

# 查看版本
delguard version

# 查看帮助
delguard help
```

### 高级选项
```bash
# 递归删除目录
delguard delete -r directory/

# 详细输出
delguard delete -v file.txt

# 强制删除（跳过确认）
delguard delete -f file.txt

# 预览模式（不实际删除）
delguard delete --dry-run file.txt
```

### 配置管理
```bash
# 查看配置
delguard config show

# 设置配置项
delguard config set language zh-cn
```

## 🎯 错误处理

DelGuard 提供智能错误提示，常见错误包括：

- **权限不足**: 提示以管理员身份运行
- **文件不存在**: 检查路径是否正确
- **文件被占用**: 关闭相关程序后重试
- **磁盘空间不足**: 清理磁盘空间

## 🔧 配置

DelGuard 支持通过配置文件自定义行为。配置文件位于：
- Windows: `%APPDATA%\DelGuard\config.yaml`
- Linux: `~/.config/delguard/config.yaml`

### 配置选项
```yaml
# 删除行为配置
delete:
  confirm_before_delete: true    # 删除前确认
  use_recycle_bin: true         # 使用回收站
  backup_before_overwrite: true # 覆盖前备份

# 安全配置
security:
  max_file_size: "100MB"        # 最大文件大小限制
  forbidden_extensions: [".sys", ".dll"]  # 禁止删除的扩展名
  
# 界面配置
ui:
  language: "zh-cn"             # 界面语言
  show_progress: true           # 显示进度条

# 监控配置
monitor:
  enable_logging: true          # 启用操作日志
  log_level: "info"            # 日志级别
```

## 📄 技术栈

- **语言**: Go
- **平台**: 跨平台 (Windows/macOS/Linux)
- **依赖**: 标准库 + 系统API

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🌟 更新日志

### v1.1.0 (2024-12-19)
- ✨ 新增智能提示系统
- 🛡️ 增强错误处理机制
- 🔧 优化用户交互体验
- 📱 支持多语言提示

### v1.0.0 (2024-12)
- 🎉 初始版本发布
- ✨ 基本文件删除保护
- 🔄 回收站集成
- 🛡️ 覆盖保护功能