# DelGuard 技术指南

## 项目简介
DelGuard 是一个用 Go 编写的跨平台安全文件删除工具，核心功能是将文件移动到系统回收站而非直接删除，提供误删恢复能力。

## 技术架构

### 核心模块
- **main.go** - 程序入口，命令行解析
- **platform.go** - 平台检测和分发
- **windows.go/macos.go/linux.go** - 各平台具体实现
- **restore.go** - 文件恢复功能
- **protect.go** - 路径保护机制
- **config.go** - 配置管理
- **safe_copy.go** - 安全复制功能

## 核心功能说明

### 1. 文件删除机制

DelGuard 不会直接删除文件，而是将文件移动到系统回收站/废纸篓中：

- **Windows**: 使用 `SHFileOperationW` API 将文件移动到回收站
- **macOS**: 使用 `osascript` 调用 Finder 将文件移动到废纸篓
- **Linux**: 遵循 freedesktop.org Trash 规范，将文件移动到 `~/.local/share/Trash` 目录

### 2. 文件覆盖保护机制

DelGuard 提供文件覆盖保护功能，在以下场景自动激活：

- **文件复制**: 当目标文件已存在时，先将现有文件移动到回收站
- **文件移动**: 当目标位置已有同名文件时，先备份现有文件
- **文件写入**: 当写入操作会覆盖现有文件时，先创建备份

覆盖保护的工作流程：
1. 检查目标文件是否存在
2. 执行安全检查（权限、系统文件等）
3. 将现有文件移动到回收站
4. 记录操作日志
5. 执行正常的文件操作

### 3. 安全复制机制

DelGuard 提供安全复制功能，可在复制文件时防止意外覆盖：

- **文件一致性检查**: 复制前计算并比较源文件和目标文件的SHA256哈希值
- **智能覆盖保护**: 文件不一致时提示用户确认是否覆盖
- **自动备份**: 覆盖前自动将原文件移动到回收站
- **详细信息展示**: 显示文件哈希值帮助用户判断文件是否相同

安全复制的工作流程：
1. 检查源文件和目标文件是否存在
2. 如果目标文件存在，计算两个文件的哈希值
3. 如果哈希值相同，跳过复制
4. 如果哈希值不同，提示用户确认
5. 用户确认后，将现有文件移动到回收站
6. 执行文件复制操作

### 1. 文件删除机制

DelGuard 不会直接删除文件，而是将文件移动到系统回收站/废纸篓中：

- **Windows**: 使用 `SHFileOperationW` API 将文件移动到回收站
- **macOS**: 使用 `osascript` 调用 Finder 将文件移动到废纸篓
- **Linux**: 遵循 freedesktop.org 规范，将文件移动到 `~/.local/share/Trash` 目录

### 2. 文件覆盖保护机制

DelGuard 提供文件覆盖保护功能，在以下场景自动激活：

- **文件复制**: 当目标文件已存在时，先将现有文件移动到回收站
- **文件移动**: 当目标位置已有同名文件时，先备份现有文件
- **文件写入**: 当写入操作会覆盖现有文件时，先创建备份

覆盖保护的工作流程：
1. 检查目标文件是否存在
2. 执行安全检查（权限、系统文件等）
3. 将现有文件移动到回收站
4. 记录操作日志
5. 执行正常的文件操作

#### 安全复制功能（Safe Copy）

DelGuard 的安全复制功能提供了比传统 `cp` 命令更智能的文件保护机制：

**功能特性：**
- **哈希值比较**：使用SHA256算法计算文件指纹，精确判断文件内容是否相同
- **智能交互**：当文件内容相同时，自动跳过复制操作，避免无意义的覆盖
- **详细对比**：向用户展示源文件和目标文件的详细信息，包括文件大小、修改时间、哈希值
- **自动备份**：覆盖操作前自动将目标文件移动到系统回收站
- **交互确认**：提供 `-i` 参数支持交互式确认，`-f` 参数支持强制覆盖

**使用示例：**
```bash
# 交互式安全复制（推荐）
delguard cp -i source.txt destination.txt

# 强制覆盖模式（自动备份）
delguard cp -f source.txt destination.txt

# 详细输出模式
delguard cp -v source.txt destination.txt
```

**文件对比信息展示：**
```
目标文件已存在且内容不同:
  源文件: /path/to/source.txt (大小: 1024字节, 修改时间: 2024-01-15 10:30:00, SHA256: a1b2c3d4e5f6...)
  目标文件: /path/to/destination.txt (大小: 2048字节, 修改时间: 2024-01-14 15:20:00, SHA256: f6e5d4c3b2a1...)
是否覆盖目标文件？[y/N]: 
```

**相同文件智能跳过：**
```
文件 /path/to/source.txt 和 /path/to/destination.txt 内容相同，跳过复制
```

### 3. 文件恢复机制

DelGuard 提供了从回收站恢复文件的功能：

## 平台实现细节

### Windows 实现
使用 Win32 API 的 `SHFileOperationW` 函数：
```go
// 关键API调用
shell32.NewProc("SHFileOperationW")
FO_DELETE | FOF_ALLOWUNDO // 允许撤销的删除
```

### macOS 实现
通过 AppleScript 调用 Finder：
```go
// AppleScript命令
tell application "Finder" to delete (POSIX file "路径")
```

### Linux 实现
遵循 freedesktop.org Trash 规范：
```
~/.local/share/Trash/
├── files/     # 被删除的文件
└── info/      # 删除信息(.trashinfo)
```

## 安全机制

### 路径保护
```go
// 阻止删除的关键路径
protectedPaths := []string{
    "/", "C:\\", "/usr", "/bin", "C:\\Windows",
}
```

### 文件检查流程
1. 检查文件是否存在
2. 验证路径合法性
3. 检查文件权限
4. 确认是否为特殊文件类型

## 配置系统

### 配置文件位置
- **Windows**: `%USERPROFILE%\.delguard\config.json`
- **macOS**: `~/.delguard/config.json`
- **Linux**: `~/.delguard/config.json`

### 配置示例
```json
{
  "interactive": true,
  "language": "zh-CN",
  "max_file_size": 104857600
}
```

## 构建方法

### 环境要求
- Go 1.19+
- Git

### 构建命令
```bash
# 构建所有平台
make build-all

# 构建特定平台
GOOS=windows GOARCH=amd64 go build -o delguard.exe
GOOS=darwin GOARCH=amd64 go build -o delguard-mac
GOOS=linux GOARCH=amd64 go build -o delguard-linux
```

## 代码结构

```
DelGuard/
├── main.go                 # 主程序
├── platform.go             # 平台抽象
├── windows.go              # Windows实现
├── macos.go                # macOS实现
├── linux.go                # Linux实现
├── restore.go              # 文件恢复
├── protect.go              # 安全保护
├── overwrite_protect.go    # 文件覆盖保护
├── safe_copy.go            # 安全复制功能
├── file_operations.go      # 安全文件操作
├── config.go               # 配置管理
├── logger.go               # 日志系统
├── errors.go               # 错误处理
├── go.mod                  # 依赖管理
└── Makefile                # 构建脚本
```

## 开发指南

### 添加新平台支持
1. 在 `platform.go` 中添加平台检测
2. 创建新的平台实现文件
3. 实现 `moveToTrash` 和 `restoreFromTrash` 函数

### 扩展配置
1. 在 `config.go` 的 `Config` 结构体中添加新字段
2. 在 `validate()` 函数中添加验证逻辑
3. 更新默认配置值

### 调试技巧
```bash
# 开启调试日志
go run . --debug

# 查看详细输出
go run . --verbose
```

## API 参考

### 核心函数
- `MoveToTrash(path string) error` - 移动文件到回收站
- `RestoreFromTrash(filename string) error` - 从回收站恢复文件
- `SafeCopy(src, dst string, opts SafeCopyOptions) error` - 安全复制文件
- `IsProtectedPath(path string) bool` - 检查路径是否受保护
- `LoadConfig() (*Config, error)` - 加载配置文件

### 安全复制API
- `SafeCopy(src, dst string, opts SafeCopyOptions) error` - 安全复制文件
- `calculateFileHash(filePath string) (string, error)` - 计算文件SHA256哈希值
- `backupExistingFile(filePath string) error` - 将现有文件备份到回收站

### SafeCopyOptions 结构体
```go
type SafeCopyOptions struct {
    Interactive bool  // 交互模式，询问用户确认
    Force       bool  // 强制模式，自动备份并覆盖
    Verbose     bool  // 详细输出模式
}
```

### 错误类型
- `ErrFileNotFound` - 文件不存在
- `ErrPermissionDenied` - 权限不足
- `ErrProtectedPath` - 受保护路径
- `ErrInvalidPath` - 无效路径
- `ErrHashMismatch` - 文件哈希值不匹配

## 测试

### 运行测试
```bash
go test ./...
```

### 手动测试
```bash
# 测试基本功能
echo "test" > test.txt
delguard test.txt

# 测试恢复功能
delguard restore test.txt
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码变更
4. 创建 Pull Request

## 许可证

MIT License - 详见 LICENSE 文件