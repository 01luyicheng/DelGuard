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
go build -o DelGuard.exe .
```

### Windows 用户
下载最新的 `DelGuard.exe` 文件，将其添加到系统 PATH 中即可使用。

## 📖 使用方法

### 基本删除
```bash
DelGuard 文件名
DelGuard test.txt
```

### 批量删除
```bash
DelGuard *.tmp
DelGuard folder/
```

### 覆盖文件保护
当目标文件已存在时，DelGuard 会自动创建备份：
```bash
DelGuard newfile.txt existingfile.txt
```

## 🎯 错误处理

DelGuard 提供智能错误提示，常见错误包括：

- **权限不足**: 提示以管理员身份运行
- **文件不存在**: 检查路径是否正确
- **文件被占用**: 关闭相关程序后重试
- **磁盘空间不足**: 清理磁盘空间

## 🔧 配置

配置文件位于 `~/.delguard/config.json`，支持自定义：
- 回收站行为
- 安全检查级别
- 提示信息显示

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