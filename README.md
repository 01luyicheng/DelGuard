# DelGuard - 跨平台安全删除工具

DelGuard 是一款跨平台的命令行安全删除工具，通过拦截系统原生删除命令（rm/del），将文件移动到回收站而非直接删除，为用户提供文件误删保护。

## 🌟 特性

- **🛡️ 安全删除拦截**：替换系统rm/del命令，自动将删除文件移动到对应系统回收站
- **🌍 跨平台支持**：支持Windows、macOS、Linux三大主流操作系统
- **📁 统一回收站管理**：统一处理Windows回收站、macOS废纸篓、Linux Trash目录
- **🔄 文件恢复功能**：通过命令行从回收站恢复指定文件
- **🇨🇳 中文友好界面**：提供友好的中文操作提示和错误信息
- **⚡ 一键安装部署**：自动安装脚本，无缝替换系统删除命令
- **📊 文件管理操作**：支持查看回收站内容、批量恢复、清空回收站等管理功能

## 🚀 快速开始

### 一行命令安装 (推荐)

#### Windows
```powershell
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; .\quick-install.ps1 }"
```

#### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash
```

### 手动安装

#### 编译项目

```bash
# 克隆项目
git clone <repository-url>
cd DelGuard

# 安装依赖
go mod tidy

# 编译
go build -o delguard .
```

#### 手动安装DelGuard

#### Windows
```powershell
# 以管理员身份运行PowerShell
.\scripts\install.ps1
```

#### Linux/macOS
```bash
# 使用sudo权限运行
sudo ./scripts/install.sh
```

### 基本使用

安装完成后，原有的删除命令将被安全地替换：

```bash
# 安全删除文件（移动到回收站）
rm file.txt
del file.txt  # Windows

# 查看回收站内容
delguard list
delguard ls

# 恢复文件
delguard restore file.txt
delguard restore --index 1

# 清空回收站
delguard empty

# 查看状态
delguard status
```

## 📖 命令详解

### 删除命令
```bash
# 删除单个文件
delguard delete file.txt

# 删除多个文件
delguard delete file1.txt file2.txt

# 删除目录
delguard delete folder/

# 强制删除（跳过确认）
delguard delete -f file.txt
```

### 列表命令
```bash
# 查看回收站内容
delguard list

# 详细列表格式
delguard list -l

# 按大小排序
delguard list --sort=size

# 反向排序
delguard list --sort=time --reverse

# 过滤文件
delguard list --filter="*.txt"

# 限制显示数量
delguard list --limit=10
```

### 恢复命令
```bash
# 按名称恢复文件
delguard restore file.txt

# 按索引恢复文件
delguard restore --index 1

# 恢复到指定位置
delguard restore file.txt --target /path/to/restore/

# 批量恢复
delguard restore --all --filter="*.txt"
```

### 管理命令
```bash
# 清空回收站
delguard empty

# 清空前确认
delguard empty --confirm

# 查看系统状态
delguard status

# 安装系统集成
delguard install

# 卸载系统集成
delguard uninstall
```

## 🔧 配置

DelGuard 支持通过配置文件自定义行为：

```yaml
# ~/.delguard/config.yaml
trash:
  auto_empty_days: 30  # 自动清理天数
  confirm_delete: true # 删除前确认
  
display:
  color: true          # 彩色输出
  unicode: true        # Unicode图标
  
logging:
  level: info          # 日志级别
  file: ~/.delguard/delguard.log
```

## 🏗️ 技术架构

### 核心组件
- **Go语言**：主要开发语言，提供优秀的跨平台支持
- **Cobra框架**：命令行界面框架
- **Viper**：配置文件管理
- **跨平台回收站API**：
  - Windows: Shell32 API (SHFileOperation)
  - macOS: ~/.Trash目录操作
  - Linux: XDG Trash规范实现

### 项目结构
```
DelGuard/
├── cmd/                 # 命令行命令实现
│   ├── root.go         # 根命令
│   ├── delete.go       # 删除命令
│   ├── list.go         # 列表命令
│   ├── restore.go      # 恢复命令
│   ├── empty.go        # 清空命令
│   ├── status.go       # 状态命令
│   ├── install.go      # 安装命令
│   └── uninstall.go    # 卸载命令
├── internal/
│   └── filesystem/     # 文件系统操作
│       ├── trash.go    # 回收站接口
│       ├── windows.go  # Windows实现
│       ├── macos.go    # macOS实现
│       └── linux.go    # Linux实现
├── scripts/            # 安装脚本
│   ├── install.ps1     # Windows安装脚本
│   └── install.sh      # Linux/macOS安装脚本
├── main.go             # 程序入口
├── go.mod              # Go模块定义
└── README.md           # 项目说明
```

## 🔒 安全性

DelGuard 在设计时充分考虑了安全性：

1. **权限检查**：安装脚本需要管理员权限，确保系统级操作的安全性
2. **备份机制**：安装前自动备份原始命令，支持完整卸载恢复
3. **路径验证**：严格验证文件路径，防止路径遍历攻击
4. **确认机制**：危险操作前提供二次确认，防止误操作
5. **日志记录**：完整的操作日志，便于审计和问题排查

## 🤝 贡献

欢迎提交Issue和Pull Request来帮助改进DelGuard！

### 开发环境设置
```bash
# 克隆项目
git clone <repository-url>
cd DelGuard

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 本地构建
go build -o delguard .
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

如果您遇到问题或有任何建议，请：

1. 查看 [FAQ](docs/FAQ.md)
2. 搜索现有的 [Issues](../../issues)
3. 创建新的 [Issue](../../issues/new)

## 🙏 致谢

感谢所有为DelGuard项目做出贡献的开发者和用户！

---

**⚠️ 重要提醒**：DelGuard会替换系统原生的删除命令，请在充分测试后再在生产环境中使用。建议先在测试环境中验证功能的正确性。