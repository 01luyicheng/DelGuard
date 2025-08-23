# DelGuard - 跨平台安全删除工具

DelGuard 是一个现代化的跨平台安全删除工具，支持 Windows、macOS 和 Linux 系统。它通过将文件移动到系统回收站而非直接删除，为您的数据提供额外的安全保障。

## 🚀 特性

- **跨平台支持**: 完美支持 Windows、macOS、Linux
- **安全删除**: 文件移动到回收站，可随时恢复
- **智能检测**: 自动识别系统语言和配置
- **别名支持**: 兼容传统的 `del` 和 `rm` 命令
- **路径保护**: 防止意外删除关键系统目录
- **交互模式**: 删除前确认，避免误操作
- **多语言**: 支持中文、英文界面
- **长路径支持**: 处理深层嵌套的文件结构
- **符号链接**: 正确处理符号链接，不删除目标文件
- **强制删除**: 支持 `-force` 参数直接彻底删除文件
- **权限管理**: 管理员权限操作需要二次确认
- **错误处理**: 详细的错误代码和建议信息
- **配置管理**: 用户可配置默认行为和语言设置

## 📦 安装

### 一行命令安装（推荐）

#### Windows (PowerShell 7+)
```powershell
# 一键安装
iwr -useb https://raw.githubusercontent.com/your-username/DelGuard/main/install | iex

# 或安装并设置默认交互模式
iwr -useb https://raw.githubusercontent.com/your-username/DelGuard/main/install | iex -- --default-interactive
```

#### macOS / Linux
```bash
# 一键安装
bash -c "$(curl -fsSL https://raw.githubusercontent.com/your-username/DelGuard/main/install)"

# 或安装并设置默认交互模式
bash -c "$(curl -fsSL https://raw.githubusercontent.com/your-username/DelGuard/main/install)" -- --default-interactive

# 或者使用 curl 管道方式
curl -fsSL https://raw.githubusercontent.com/your-username/DelGuard/main/install | bash
```

### 手动安装（备用方案）

#### Windows

##### 方法一：自动安装
```powershell
# 下载后运行
.\install.ps1

# 或卸载
.\install.ps1 -Uninstall
```

##### 方法二：手动安装
1. 下载 `DelGuard.exe`
2. 添加到系统 PATH
3. 创建别名（可选）

#### macOS / Linux

##### 方法一：自动安装
```bash
# 下载后运行
chmod +x install.sh
./install.sh

# 或卸载（手动删除）
rm ~/.local/bin/delguard  # 用户安装
sudo rm /usr/local/bin/delguard  # 系统安装
# 编辑 ~/.zshrc 或 ~/.bashrc 删除别名配置
```

#### 方法二：Homebrew（即将支持）
```bash
brew install delguard
```

### Linux

#### 方法一：自动安装（推荐）
```bash
# 下载后运行
chmod +x install.sh
./install.sh

# 或卸载
rm ~/.local/bin/delguard  # 用户安装
sudo rm /usr/local/bin/delguard  # 系统安装
```

#### 方法二：包管理器（即将支持）
```bash
# Ubuntu/Debian
sudo apt install delguard

# CentOS/RHEL
sudo yum install delguard

# Arch Linux
sudo pacman -S delguard
```

## 🎯 使用

### 基本用法

```bash
# 删除单个文件
del document.txt
rm photo.jpg

# 删除多个文件
del file1.txt file2.txt file3.txt

# 删除目录（需要递归参数）
del -r project_folder
rm --recursive old_data/

# 强制删除（跳过确认）
del -f important.doc
rm --force cache.tmp
```

### 高级用法

```bash
# 交互模式（删除前确认）
del -i sensitive_data.xlsx

# 详细输出
del -v large_folder/

# 组合使用
del -r -f -v temp_build/

# 强制删除（不经过回收站，直接彻底删除）
del --force confidential.doc
rm -force secret_data/

# 跳过关键路径保护确认（危险！）
del --skip-protection system_file.tmp

# 显示帮助信息
del --help
rm -help

# 显示版本信息
del --version
rm -version
```

### 恢复文件

#### Windows
1. 打开回收站
2. 右键点击文件 → 还原

#### macOS
1. 打开废纸篓（Dock右侧）
2. 右键点击文件 → 放回原处

#### Linux
```bash
# 使用DelGuard恢复
delguard restore 文件名

# 或手动从 ~/.local/share/Trash/files/ 恢复
```

## ⚙️ 配置

### 环境变量

| 变量名 | 描述 | 示例 |
|--------|------|------|
| `DELGUARD_INTERACTIVE` | 强制交互模式 | `true` |
| `DELGUARD_LANG` | 设置语言 | `zh-CN` 或 `en-US` |
| `DELGUARD_VERBOSE` | 详细输出 | `true` |

### 配置文件

配置文件路径：
- Windows: `%APPDATA%\DelGuard\config.json`
- macOS: `~/Library/Application Support/DelGuard/config.json`
- Linux: `~/.config/DelGuard/config.json`

示例配置：
```json
{
  "interactive": false,
  "language": "auto",
  "verbose": false
}
```

## 🛡️ 安全特性

### 路径保护
DelGuard 会自动保护以下关键路径：
- 系统根目录（`/` 或 `C:\`）
- 用户主目录
- 系统文件夹（Windows、Program Files 等）
- 重要配置目录

### 恢复机制
所有删除的文件都可以通过系统回收站恢复：
- **Windows**: 资源管理器回收站
- **macOS**: Finder 废纸篓
- **Linux**: `~/.local/share/Trash`

## 🔧 构建

### 环境要求
- Go 1.19 或更高版本
- Git

### 从源码构建
```bash
# 克隆仓库
git clone https://github.com/your-username/DelGuard.git
cd DelGuard

# 构建所有平台
./build.sh  # macOS/Linux
# 或
build.bat   # Windows

# 构建特定平台
GOOS=windows GOARCH=amd64 go build -o delguard-windows.exe
GOOS=darwin GOARCH=amd64 go build -o delguard-macos
GOOS=linux GOARCH=amd64 go build -o delguard-linux
```

## 🧪 测试

### 运行测试
```bash
# Windows
.\scripts\tests\test_delguard.ps1

# macOS/Linux
./scripts/tests/test_delguard.sh
```

### 测试覆盖
- ✅ 基础文件删除
- ✅ 同名文件处理
- ✅ 符号链接支持
- ✅ 长路径处理
- ✅ 目录递归删除
- ✅ 关键路径保护
- ✅ 交互模式
- ✅ 多语言支持

## 📋 常见问题

### Q: 如何恢复误删的文件？
A: 所有删除的文件都会进入系统回收站，可以通过系统回收站界面恢复。

### Q: 支持网络驱动器吗？
A: 支持，但跨设备删除会使用复制+删除的方式，可能较慢。

### Q: 如何完全卸载？
A: 运行安装脚本的卸载命令，或手动删除：
- 可执行文件
- 配置文件
- shell别名配置

### Q: 支持哪些语言？
A: 目前支持中文（简体）和英文，根据系统语言自动切换。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 开发环境
```bash
# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 代码格式化
go fmt ./...
```

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- 感谢所有贡献者的支持
- 特别感谢测试用户的反馈和建议

---

## 📞 支持

- 📧 邮箱: support@delguard.dev
- 🐛 Issue: [GitHub Issues](https://github.com/your-repo/DelGuard/issues)
- 💬 讨论: [GitHub Discussions](https://github.com/your-repo/DelGuard/discussions)

**让删除更安全，让数据有保障！**