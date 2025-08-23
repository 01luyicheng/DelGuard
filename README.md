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

## 🔒 安全功能

### 文件类型检测
- **隐藏文件检测**: 自动识别Windows隐藏属性文件
- **特殊文件保护**: 防止删除符号链接、设备文件、套接字文件
- **系统文件保护**: 阻止删除Windows系统文件和关键目录

### 权限验证
- **文件所有权检查**: 验证用户是否有权限删除指定文件
- **目录权限验证**: 检查目标目录的写权限
- **只读文件保护**: 阻止删除只读属性文件

### 资源限制
- **文件大小限制**: 限制单个文件最大为10GB
- **磁盘空间检查**: 确保有足够空间进行删除操作

### 路径验证
- **路径遍历攻击防护**: 防止 `../../../` 等路径攻击
- **非法字符检测**: 检测并阻止包含 `< > : " | ? *` 的文件名
- **路径长度限制**: 限制最大路径长度为4096字符

### 操作确认
- **批量操作确认**: 删除多个文件时要求用户确认
- **隐藏文件确认**: 删除隐藏文件时额外确认
- **系统文件确认**: 删除系统相关文件时警告

### 恢复功能
- **恢复路径验证**: 验证恢复目标路径的合法性
- **系统目录保护**: 禁止恢复到系统关键目录
- **文件冲突检测**: 检测恢复目标是否已存在文件

## 📦 安装

### 一行命令安装（推荐）

#### Windows (PowerShell 7+)
```powershell
# 一键安装
iwr -useb https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.ps1 | iex

# 或安装并设置默认交互模式
iwr -useb https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.ps1 | iex -- --default-interactive
```

#### macOS / Linux
```bash
# 一键安装
bash -c "$(curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.sh)"

# 或安装并设置默认交互模式
bash -c "$(curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.sh)" -- --default-interactive

# 或者使用 curl 管道方式
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install.sh | bash
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

# 或卸载
./install.sh --uninstall
```

##### 方法二：手动安装
```bash
# 下载二进制文件
# 添加到 PATH 或创建符号链接
sudo ln -s /path/to/delguard /usr/local/bin/delguard
```

## 🛠 使用方法

### 基本用法
```bash
# 删除文件
delguard file.txt

# 交互式删除（推荐）
delguard -i *.txt

# 递归删除目录
delguard -r directory

# 强制删除（不进入回收站）
delguard --force sensitive_file.txt
```

### 文件恢复功能
```bash
# 列出回收站中所有可恢复的文件
delguard restore -l

# 恢复所有文件
delguard restore

# 按名称模式恢复文件（支持通配符）
delguard restore "*.txt"
delguard restore "document*"

# 限制恢复文件数量
delguard restore "*.jpg" -max 10

# 交互模式确认每个文件
delguard restore -i

# 列出匹配模式的文件
delguard restore -l "*.pdf"
```

### 常用选项
```
-v, --verbose           详细模式
-q, --quiet             安静模式，减少输出
-r, --recursive         递归删除目录
-n, --dry-run           试运行，不实际删除
--force                 强制彻底删除，不经过回收站
-i, --interactive       交互模式，逐项确认
--install               安装shell别名（默认启用交互模式）
--version               显示版本信息
--help                  显示帮助信息
```

## ⚙️ 配置

DelGuard 支持通过配置文件进行自定义，配置文件位置：
- Windows: `%USERPROFILE%\.delguard\config.json`
- macOS/Linux: `~/.delguard/config.json`

配置示例：
```json
{
  "use_recycle_bin": true,
  "interactive_mode": "confirm",
  "language": "zh-CN",
  "log_level": "info",
  "safe_mode": "normal",
  "max_file_size": 10737418240,
  "enable_security_checks": true
}
```

## 🧪 测试

运行测试：
```bash
go test ./...
```

运行安全测试：
```bash
go test -run TestSecurity ./...
```

## 📄 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。