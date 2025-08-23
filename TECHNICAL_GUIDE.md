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
├── main.go              # 主程序
├── platform.go          # 平台抽象
├── windows.go           # Windows实现
├── macos.go             # macOS实现
├── linux.go             # Linux实现
├── restore.go           # 文件恢复
├── protect.go           # 安全保护
├── config.go            # 配置管理
├── logger.go            # 日志系统
├── errors.go            # 错误处理
├── go.mod               # 依赖管理
└── Makefile             # 构建脚本
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
- `IsProtectedPath(path string) bool` - 检查路径是否受保护
- `LoadConfig() (*Config, error)` - 加载配置文件

### 错误类型
- `ErrFileNotFound` - 文件不存在
- `ErrPermissionDenied` - 权限不足
- `ErrProtectedPath` - 受保护路径
- `ErrInvalidPath` - 无效路径

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