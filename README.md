# DelGuard - Windows文件删除工具

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/01luyicheng/DelGuard)](https://github.com/01luyicheng/DelGuard/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub stars](https://img.shields.io/github/stars/01luyicheng/DelGuard?style=social)](https://github.com/01luyicheng/DelGuard)

## 项目简介

DelGuard是一个强大的Windows文件删除工具，用于安全的删除文件和文件夹，将它们移动到回收站而不是直接彻底删除。它支持强制删除、预览模式、递归删除等多种功能，是系统维护和文件管理的得力助手。

## 功能特性

- **安全删除**：将文件和文件夹移动到回收站，而不是永久删除
- **强制删除**：绕过系统限制删除被占用的文件
- **预览模式**：先显示将要删除的文件，确认后再执行删除
- **递归删除**：支持删除文件夹及其所有子内容
- **静默模式**：无提示删除，适合批处理脚本
- **详细日志**：显示详细的删除过程和结果
- **通配符支持**：支持使用通配符批量删除文件
- **安全保护**：防止删除系统关键目录

## 快速安装

### 方法一：下载安装包（推荐）

1. 访问 [GitHub Releases](https://github.com/01luyicheng/DelGuard/releases/latest) 下载最新版本
2. 以管理员身份运行 `install.bat`
3. 按照提示完成安装
4. 重新打开命令提示符即可使用 `delguard` 命令

### 方法二：手动安装

1. 下载 `delguard.exe` 可执行文件
2. 将文件放置在合适的目录（如 `C:\Program Files\DelGuard`）
3. 手动将目录添加到系统PATH环境变量

## 使用方法

### 基本用法

```bash
# 删除单个文件
delguard file.txt

# 删除文件夹
delguard C:\path\to\folder

# 强制删除被占用的文件
delguard -f lockedfile.exe

# 预览模式（显示将要删除的文件）
delguard -p C:\temp\*

# 递归删除文件夹及其内容
delguard -r C:\old_data

# 静默模式（无提示）
delguard -s temp_files

# 详细模式（显示详细日志）
delguard -v large_folder
```

### 命令行参数

- `路径`：要删除的文件或文件夹路径（支持通配符）
- `-f, --force`：强制删除模式
- `-p, --preview`：预览模式，显示将要删除的文件
- `-r, --recursive`：递归删除文件夹及其所有内容
- `-s, --silent`：静默模式，最小化输出
- `-v, --verbose`：详细模式，显示详细日志
- `-h, --help`：显示帮助信息

### 使用示例

```bash
# 删除临时文件夹中的所有文件
delguard -r C:\Windows\Temp\*

# 强制删除被系统占用的日志文件
delguard -f C:\Windows\Logs\error.log

# 预览删除操作，确认安全后再执行
delguard -p -r C:\Users\%USERNAME%\Downloads\old_files

# 静默删除临时文件（适合批处理脚本）
delguard -s -f C:\temp\*.tmp
```

## 配置文件

DelGuard使用JSON配置文件来管理设置，配置文件位于安装目录的 `config` 文件夹中。

### 配置选项

```json
{
  "uiOption": "OnlyErrorDialogs",  // UI选项：OnlyErrorDialogs, AllDialogs, NoDialogs
  "logLevel": "info"              // 日志级别：debug, info, warning, error
}
```

### 配置说明

- **uiOption**：控制用户界面显示
  - `OnlyErrorDialogs`：仅显示错误对话框（默认）
  - `AllDialogs`：显示所有对话框
  - `NoDialogs`：不显示任何对话框

- **logLevel**：控制日志输出详细程度
  - `debug`：调试级别，最详细
  - `info`：信息级别（默认）
  - `warning`：警告级别
  - `error`：错误级别，最少输出

## 系统要求

- 操作系统：Windows 10/11 (x64)
- 运行时：.NET 8.0 或更高版本
- 权限：普通用户权限（某些操作需要管理员权限）

## 安全注意事项

1. **谨慎使用强制删除**：强制删除可能会影响正在运行的程序
2. **系统目录保护**：DelGuard会阻止删除关键的系统目录
3. **预览模式建议**：对于不熟悉的删除操作，建议先使用预览模式
4. **权限要求**：某些删除操作可能需要管理员权限

## 卸载方法

### 通过控制面板卸载

1. 打开"控制面板" → "程序" → "程序和功能"
2. 找到 "DelGuard"
3. 点击"卸载"并按照提示完成卸载

### 通过命令行卸载

```bash
# 使用安装提供的卸载脚本
uninstall.bat

# 或手动删除
rmdir /S /Q "%ProgramFiles%\DelGuard"
```

## 故障排除

### 常见问题

**Q: 删除操作失败，提示"访问被拒绝"**
A: 请以管理员身份运行命令提示符，然后执行删除操作

**Q: 文件被其他程序占用，无法删除**
A: 使用 `-f` 参数强制删除，或先关闭占用该文件的程序

**Q: 安装后无法在命令行使用delguard命令**
A: 请重新启动命令提示符，或手动检查PATH环境变量设置

**Q: 删除大文件夹时程序无响应**
A: 使用 `-v` 参数查看详细进度，耐心等待操作完成

### 日志文件

DelGuard在安装目录下生成日志文件，可用于故障排除：
- 日志位置：`C:\Program Files\DelGuard\logs\`
- 日志文件名：`delguard_YYYY-MM-DD.log`

## 开发信息

- 项目地址：[GitHub Repository](https://github.com/01luyicheng/DelGuard)
- 许可证：MIT License
- 版本：1.0.0

## 技术支持

如遇到问题或有功能建议，请通过以下方式联系：
- 提交Issue：[GitHub Issues](https://github.com/01luyicheng/DelGuard/issues)

---

**注意**：使用DelGuard删除文件前，请确保您真的需要删除这些文件。重要文件请先备份。