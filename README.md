## 🌐 语言包（i18n）

外部语言包放置于 `config/languages/` 目录，支持多格式：`<lang>.(json|jsonc|ini|cfg|conf|env|properties)`。

- 示例：`en-US.json`、`fr-FR.ini`、`de-DE.properties`、`ja.jsonc`
- 语义：键为中文原文，值为目标语言译文
- 外部语言包会覆盖内置翻译；缺少目标语言时回退到英文 `en-US`；中文 `zh-CN` 为源语言无需语言包
- 详细格式与示例见 `config/languages/README.md`
# DelGuard - 跨平台安全删除工具

DelGuard 是一个现代化的跨平台安全删除工具，支持 Windows、macOS 和 Linux 系统。它通过将文件移动到系统回收站而非直接删除，为您的数据提供额外的安全保障。

## 🚀 特性

- **跨平台支持**: 完美支持 Windows、macOS、Linux
- **安全删除**: 文件移动到回收站，可随时恢复
- **智能检测**: 自动识别系统语言和配置
- **别名支持**: 兼容传统的 `del` 和 `rm` 、‘cp’命令
- **路径保护**: 防止意外删除关键系统目录
- **交互模式**: 删除前确认，避免误操作
- **多语言**: 支持中文、英文界面
- **长路径支持**: 处理深层嵌套的文件结构
- **符号链接**: 正确处理符号链接，不删除目标文件
- **强制删除**: 支持 `-force` 参数直接彻底删除文件
- **权限管理**: 管理员权限操作需要二次确认
- **错误处理**: 详细的错误代码和建议信息
- **配置管理**: 用户可配置默认行为和语言设置
- **覆盖保护**: 文件被覆盖时自动备份到回收站
- **操作审计**: 完整记录文件操作历史
- **恢复功能**: 从回收站恢复被覆盖或删除的文件
- **安全复制**: 复制文件前检查目标文件，避免意外覆盖

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

### 覆盖保护
- **文件覆盖检测**: 自动检测即将被覆盖的文件
- **备份机制**: 被覆盖文件先移动到回收站
- **恢复支持**: 可从回收站恢复被覆盖的文件
- **配置开关**: 可启用/禁用覆盖保护功能

### 操作确认
- **批量操作确认**: 删除多个文件时要求用户确认
- **隐藏文件确认**: 删除隐藏文件时额外确认
- **系统文件确认**: 删除系统相关文件时警告

### 恢复功能
- **恢复路径验证**: 验证恢复目标路径的合法性
- **系统目录保护**: 禁止恢复到系统关键目录
- **文件冲突检测**: 检测恢复目标是否已存在文件

### 安全复制功能
- **文件一致性检查**: 复制前计算并比较源文件和目标文件的SHA256哈希值，精确判断文件内容是否相同
- **智能覆盖保护**: 文件不一致时提示用户确认是否覆盖，文件相同时自动跳过复制操作
- **自动备份**: 覆盖前自动将原文件移动到系统回收站，确保数据可恢复
- **详细信息展示**: 向用户展示源文件和目标文件的完整信息，包括文件路径、大小、最后修改时间和哈希值
- **交互式确认**: 提供 `-i` 参数支持交互式确认，`-f` 参数支持强制覆盖
- **智能跳过**: 当源文件和目标文件内容完全一致时，自动跳过复制操作，避免无意义的覆盖

**安全复制使用场景：**
- 备份重要文件前确认不会意外覆盖
- 同步文件时避免重复复制相同内容
- 团队协作中确保文件版本一致性
- 自动化脚本中增加文件操作安全性

## 📦 安装

### 本地一键安装（推荐）

#### Windows
```bash
# 方法一：批处理安装（最简单）
install_one_click.bat

# 方法二：PowerShell安装
powershell -ExecutionPolicy Bypass -File install_one_click.ps1

# 方法三：管理员安装（系统范围）
# 以管理员身份运行PowerShell
powershell -ExecutionPolicy Bypass -File install_one_click.ps1 -SystemInstall
```

#### 一行命令在线安装

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

### 安全复制功能
```bash
# 安全复制文件（如果目标文件存在且内容不同，会提示确认并备份原文件）
delguard --safe-copy source.txt destination.txt

# 强制安全复制（跳过确认，直接备份原文件并覆盖）
delguard --safe-copy --force source.txt destination.txt

# 交互式安全复制（总是提示确认）
delguard --safe-copy -i source.txt destination.txt

# 复制多个文件到目录
delguard --safe-copy file1.txt file2.txt directory/

# 详细模式显示更多信息
delguard --safe-copy --verbose source.txt destination.txt

# 实际使用示例
# 场景1：文件内容相同，自动跳过
$ delguard --safe-copy config.json backup/config.json
文件内容相同，跳过复制

# 场景2：文件内容不同，提示用户确认
$ delguard --safe-copy -i new_config.json config.json
目标文件已存在且内容不同:
  源文件: new_config.json (大小: 2048字节, 修改时间: 2024-01-15 14:30:00, SHA256: a1b2c3d4e5f6...)
  目标文件: config.json (大小: 1024字节, 修改时间: 2024-01-14 10:20:00, SHA256: f6e5d4c3b2a1...)
是否覆盖目标文件？[y/N]: y
已将 config.json 移动到回收站
成功复制 new_config.json -> config.json

# 场景3：强制覆盖模式
$ delguard --safe-copy --force important.txt backup/important.txt
已将 backup/important.txt 移动到回收站
成功复制 important.txt -> backup/important.txt
```

### 常用选项
```
-v, --verbose           详细模式
-q, --quiet             安静模式，减少输出
-r, --recursive         递归删除目录
-n, --dry-run           试运行，不实际删除
--force                 强制彻底删除，不经过回收站
-i, --interactive       交互模式，逐项确认
--protect               启用文件覆盖保护
--disable-protect       禁用文件覆盖保护
--safe-copy             安全复制模式
--install               安装 shell 别名（默认启用交互模式）
--version               显示版本信息
--help                  显示帮助信息
```

## ⚙️ 配置

DelGuard 支持多格式配置文件与外部覆盖：

- 支持扩展名：`.json`、`.jsonc`（支持注释）、`.ini`、`.cfg`、`.conf`、`.env`、`.properties`
- 默认查找顺序（按先后优先级）：
  - 用户目录：`~/.delguard/config.(json|jsonc|ini|cfg|conf)`、`~/.delguard/.env`、`~/.delguard/delguard.properties`
  - 系统目录：
    - Windows: `%SystemRoot%\delguard\`
    - macOS/Linux: `/etc/delguard/`
  - 当前目录：`config.(json|jsonc|ini|cfg|conf)`、`.env`、`delguard.properties`
- 指定外部配置路径：
  - 使用 `--config` 明确指定文件路径（优先级最高）
  - 示例：
    - Windows: `delguard --config C:\\Users\\User\\.delguard\\config.jsonc`
    - Linux/macOS: `delguard --config ~/.delguard/config.ini`

配置示例（JSON）：
```json
{
  "use_recycle_bin": true,
  "interactive_mode": "confirm",
  "language": "zh-CN",
  "log_level": "info",
  "safe_mode": "normal",
  "max_file_size": 10737418240,
  "enable_security_checks": true,
  "enable_overwrite_protection": true
}
```

配置示例（.env/.properties）：
```properties
use_recycle_bin=true
interactive_mode=confirm
language=zh-CN
log_level=info
safe_mode=normal
max_file_size=10737418240
enable_security_checks=true
enable_overwrite_protection=true
```

说明：
- `.jsonc` 会自动移除注释后再解析
- `.ini/.cfg/.conf` 支持 `key=value` 或 `key: value` 格式，忽略 `[section]` 名称
- `.env/.properties` 使用简单的 `key=value` 键值对格式

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