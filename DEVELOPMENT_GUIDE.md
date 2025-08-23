# DelGuard 二次开发指南

## 项目概述

DelGuard 是一个跨平台的安全文件删除工具，它将文件移动到系统回收站而不是直接删除，从而提供了一层安全保障。它支持 Windows、macOS 和 Linux 系统，并提供了丰富的安全检查机制。

## 项目架构

```
DelGuard/
├── main.go              # 主程序入口和核心逻辑
├── platform.go          # 平台检测和分发
├── windows.go           # Windows 平台实现
├── macos.go             # macOS 平台实现
├── linux.go             # Linux 平台实现
├── restore.go           # 文件恢复功能
├── protect.go           # 关键路径保护
├── errors.go            # 错误处理
├── config.go            # 配置管理
├── i18n.go              # 国际化支持
├── file_validator.go    # 文件验证
├── installer.go         # 安装器
└── scripts/             # 安装脚本
```

## 核心功能说明

### 1. 文件删除机制

DelGuard 不会直接删除文件，而是将文件移动到系统回收站/废纸篓中：

- **Windows**: 使用 `SHFileOperationW` API 将文件移动到回收站
- **macOS**: 使用 `osascript` 调用 Finder 将文件移动到废纸篓
- **Linux**: 遵循 freedesktop.org 规范，将文件移动到 `~/.local/share/Trash` 目录

### 2. 文件恢复机制

DelGuard 提供了从回收站恢复文件的功能：

- **Windows**: 暂不支持命令行恢复，建议使用资源管理器
- **macOS**: 从 `~/.Trash` 目录恢复文件
- **Linux**: 从 `~/.local/share/Trash` 目录恢复文件，并解析 `.trashinfo` 文件获取原始路径

## 二次开发指南

### 添加新平台支持

要添加新平台支持，您需要:

1. 创建平台特定的文件（如 [platform].go）
2. 实现 `moveToTrash[Platform]` 函数
3. 在 [platform].go 中添加存根函数以满足编译要求
4. 在 `platform.go` 中添加新平台的检测和调用逻辑

示例:
```go
// 在新平台文件中
func moveToTrashNewPlatform(filePath string) error {
    // 实现将文件移动到该平台回收站的逻辑
    return nil
}

// 为其他平台添加存根函数
func moveToTrashWindows(filePath string) error {
    return ErrUnsupportedPlatform
}

func moveToTrashMacOS(filePath string) error {
    return ErrUnsupportedPlatform
}

func moveToTrashLinux(filePath string) error {
    return ErrUnsupportedPlatform
}
```

### 扩展安全检查

DelGuard 提供了多种安全检查机制，您可以通过修改以下文件来扩展：

- `protect.go` - 关键路径保护
- `file_validator.go` - 文件验证
- `errors.go` - 错误处理

### 添加新的命令行选项

命令行参数在 `main.go` 中定义，您可以添加新的选项来扩展功能：

```go
// 在 main.go 中添加新的标志
var newOption bool
flag.BoolVar(&newOption, "new-option", false, "新选项的描述")
```

### 国际化支持

DelGuard 支持多语言，翻译文件在 `i18n.go` 中定义。要添加新语言：

1. 在 `i18n.go` 中的 `translations` 映射中添加新语言
2. 为每种语言提供相应的翻译

## API 参考

### 核心函数

#### moveToTrash(filePath string) error
将文件移动到系统回收站

#### restoreFromTrash(pattern string, opts RestoreOptions) error
从回收站恢复文件

#### IsCriticalPath(path string) bool
检查路径是否为关键系统路径

## 构建和测试

### 构建项目
```bash
go build
```

### 运行测试
```bash
go test ./...
```

## 最佳实践

1. 始终进行充分的错误处理
2. 保持平台特定代码的独立性
3. 遵循各平台的回收站规范
4. 确保安全检查机制的完整性
5. 提供清晰的错误信息和用户反馈