# DelGuard GitHub 发布指南

## 🚀 发布步骤

### 1. 创建GitHub仓库

访问 https://github.com/new 创建新仓库：
- **Repository name**: `DelGuard`
- **Description**: `跨平台安全删除工具 - 智能回收站管理，文件恢复，系统命令替换`
- **Public**: ✅
- **Add README**: ❌ (已有)
- **Add .gitignore**: ❌ (已有)
- **Add license**: ❌ (已有)

### 2. 推送代码到GitHub

```bash
# 添加远程仓库
git remote add origin https://github.com/YOUR_USERNAME/DelGuard.git

# 推送代码
git push -u origin main

# 推送标签
git push origin v1.0.0
```

### 3. 创建GitHub Release

访问 https://github.com/YOUR_USERNAME/DelGuard/releases/new

#### Release 信息：
- **Tag**: v1.4.1
- **Release title**: DelGuard v1.4.1 - 一键安装，跨平台安全删除工具
- **Description**: 

```markdown
# 🎉 DelGuard v1.0.0 正式发布

DelGuard 是一款跨平台的命令行安全删除工具，通过拦截系统原生删除命令，将文件移动到回收站而非直接删除，为用户提供文件误删保护。

## ✨ 核心功能

### 🛡️ 安全删除
- 替换系统 rm/del 命令，自动将删除文件移动到对应系统回收站
- 支持删除前确认，防止误操作
- 支持预览模式，查看将要删除的文件

### 🌍 跨平台支持
- **Windows**: 支持 Windows 10/11，x64架构
- **macOS**: 支持 Intel 和 Apple Silicon 芯片
- **Linux**: 支持 x64 和 ARM64 架构

### 📁 统一回收站管理
- 统一处理 Windows 回收站、macOS 废纸篓、Linux Trash 目录
- 支持查看回收站内容、文件大小、删除时间等信息

### 🔄 文件恢复功能
- 通过命令行从回收站恢复指定文件
- 支持按名称或索引恢复文件
- 支持恢复到指定位置

### 🇨🇳 中文友好界面
- 完整的中文操作提示和错误信息
- 彩色终端输出
- Unicode 图标支持

## 📦 下载

### 预编译二进制
- [Windows x64](https://github.com/YOUR_USERNAME/DelGuard/releases/download/v1.0.0/delguard-windows-amd64.exe)
- [macOS Intel](https://github.com/YOUR_USERNAME/DelGuard/releases/download/v1.0.0/delguard-darwin-amd64)
- [macOS Apple Silicon](https://github.com/YOUR_USERNAME/DelGuard/releases/download/v1.0.0/delguard-darwin-arm64)
- [Linux x64](https://github.com/YOUR_USERNAME/DelGuard/releases/download/v1.0.0/delguard-linux-amd64)
- [Linux ARM64](https://github.com/YOUR_USERNAME/DelGuard/releases/download/v1.0.0/delguard-linux-arm64)

### 安装方式

#### Windows
```powershell
# 以管理员身份运行PowerShell
iwr -useb https://raw.githubusercontent.com/YOUR_USERNAME/DelGuard/main/scripts/install.ps1 | iex
```

#### Linux/macOS
```bash
# 使用sudo权限运行
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/DelGuard/main/scripts/install.sh | sudo bash
```

## 🚀 快速开始

### 基本使用
```bash
# 安全删除文件
delguard delete file.txt
rm file.txt  # 安装后自动替换

# 查看回收站内容
delguard list

# 恢复文件
delguard restore file.txt

# 清空回收站
delguard empty
```

### 高级功能
```bash
# 预览删除的文件
delguard delete -n *.log

# 批量恢复
delguard restore --all --filter="*.txt"

# 查看系统状态
delguard status
```

## 🛡️ 安全特性
- **权限检查**: 安装脚本需要管理员权限
- **备份机制**: 安装前自动备份原始命令
- **路径验证**: 严格验证文件路径，防止路径遍历攻击
- **确认机制**: 危险操作前提供二次确认
- **日志记录**: 完整的操作日志，便于审计

## 🔧 技术栈
- **语言**: Go 1.21+
- **CLI框架**: Cobra
- **配置管理**: Viper
- **跨平台**: 原生支持 Windows/macOS/Linux

## 📖 文档
- [完整文档](https://github.com/YOUR_USERNAME/DelGuard#readme)
- [安装指南](https://github.com/YOUR_USERNAME/DelGuard#%EF%B8%8F-快速开始)
- [命令详解](https://github.com/YOUR_USERNAME/DelGuard#-%E5%91%BD%E4%BB%A4%E8%AF%A6%E8%A7%A3)

## 🤝 贡献
欢迎提交 Issue 和 Pull Request！

## 📄 许可证
MIT License - 详见 [LICENSE](LICENSE) 文件

---

**⚠️ 重要提醒**: DelGuard 会替换系统原生的删除命令，请在充分测试后再在生产环境中使用。
```

### 4. 上传构建文件

在 Release 页面中，上传以下文件：

从 `build/` 目录上传：
- `delguard-windows-amd64.exe`
- `delguard-darwin-amd64`
- `delguard-darwin-arm64`
- `delguard-linux-amd64`
- `delguard-linux-arm64`

### 5. 发布设置

- **Set as the latest release**: ✅
- **Create a discussion**: ✅ (选择 "Announcements")
- **Prerelease**: ❌

### 6. 发布后验证

```bash
# 测试下载链接
curl -L https://github.com/YOUR_USERNAME/DelGuard/releases/download/v1.0.0/delguard-windows-amd64.exe -o delguard-test.exe

# 测试功能
./delguard-test.exe --version
./delguard-test.exe status
```

## 📋 发布检查清单

### 功能测试 ✅
- [x] 跨平台编译成功
- [x] Windows版本测试通过
- [x] 删除功能正常
- [x] 恢复功能正常
- [x] 查看回收站功能正常
- [x] 清空回收站功能正常
- [x] 状态查看功能正常

### 文档完整 ✅
- [x] README.md 完整
- [x] CHANGELOG.md 更新
- [x] LICENSE 文件
- [x] 安装脚本测试
- [x] 构建脚本可用

### 构建文件 ✅
- [x] Windows x64
- [x] macOS Intel
- [x] macOS Apple Silicon
- [x] Linux x64
- [x] Linux ARM64

### 版本管理 ✅
- [x] Git标签 v1.0.0
- [x] 版本号统一
- [x] 提交信息完整

## 🎯 后续发布计划

### v1.1.0 (计划中)
- 图形界面支持
- 网络同步功能
- 文件版本管理
- 更多平台支持

### v1.2.0 (规划中)
- 插件系统
- 批量操作优化
- 高级过滤功能
- 国际化支持