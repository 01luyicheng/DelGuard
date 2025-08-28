# 删除服务模块 (Delete Service Module)

## 概述

这是DelGuard项目的核心删除服务模块，提供安全、高效的文件删除功能。该模块经过全面优化，包含完整的错误处理、日志记录、性能监控和配置管理功能。

## 主要特性

### 🔒 安全删除
- 文件路径验证和保护
- 受保护系统路径检查
- 安全移动到回收站
- 防止误删重要文件

### ⚡ 高性能
- 支持并发批量删除
- 可配置的并发数限制
- 性能统计和监控
- 优化的资源使用

### 📊 完整监控
- 实时操作统计
- 错误分类和统计
- 性能指标监控
- 吞吐量和成功率跟踪

### 🔧 灵活配置
- JSON配置文件支持
- 运行时配置管理
- 跨平台配置路径
- 配置验证和默认值

### 📝 详细日志
- 多级别日志记录
- 文件和控制台输出
- 结构化日志格式
- 操作审计跟踪

## 文件结构

```
internal/core/delete/
├── service.go           # 主服务实现
├── service_test.go      # 单元测试
├── integration_test.go  # 集成测试
├── config.go           # 配置管理
├── errors.go           # 错误处理
├── logger.go           # 日志记录
├── metrics.go          # 性能统计
├── service_windows.go  # Windows平台实现
├── service_unix.go     # Unix/Linux平台实现
└── README.md           # 文档说明
```

## 核心组件

### 1. Service (服务)
主要的删除服务类，提供以下功能：
- `SafeDelete()` - 安全删除单个文件
- `BatchDelete()` - 批量删除文件
- `BatchDeleteWithContext()` - 支持上下文的批量删除
- `Execute()` - 命令行接口执行
- `ValidateFile()` - 文件路径验证

### 2. Config (配置)
配置管理系统：
```go
type Config struct {
    MaxConcurrency int      `json:"max_concurrency"`
    ProtectedPaths []string `json:"protected_paths"`
    EnableLogging  bool     `json:"enable_logging"`
}
```

### 3. Logger (日志)
多级别日志系统：
- DEBUG - 调试信息
- INFO - 一般信息
- WARN - 警告信息
- ERROR - 错误信息
- FATAL - 致命错误

### 4. Metrics (统计)
性能监控系统：
- 操作计数统计
- 时间性能统计
- 错误分类统计
- 并发使用统计

### 5. Errors (错误处理)
结构化错误处理：
- 错误分类和编码
- 可重试错误识别
- 详细错误信息
- 错误链追踪

## 使用示例

### 基本使用
```go
// 创建默认服务
service := NewService()

// 删除单个文件
err := service.SafeDelete("/path/to/file.txt")
if err != nil {
    log.Printf("删除失败: %v", err)
}
```

### 自定义配置
```go
// 创建自定义配置
config := &Config{
    MaxConcurrency: 10,
    ProtectedPaths: []string{"/system", "/usr/bin"},
    EnableLogging:  true,
}

// 使用自定义配置创建服务
service := NewService(config)
```

### 批量删除
```go
files := []string{
    "/path/to/file1.txt",
    "/path/to/file2.txt",
    "/path/to/file3.txt",
}

// 批量删除
results := service.BatchDelete(files)

// 检查结果
for _, result := range results {
    if !result.Success {
        log.Printf("删除失败 %s: %v", result.Path, result.Error)
    }
}
```

### 带上下文的批量删除
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

results := service.BatchDeleteWithContext(ctx, files)
```

### 获取统计信息
```go
metrics := service.GetMetrics()
fmt.Printf("成功率: %.2f%%\n", metrics.GetSuccessRate())
fmt.Printf("吞吐量: %.2f ops/s\n", metrics.GetThroughput())
fmt.Printf("平均耗时: %v\n", metrics.AverageDuration)
```

### 配置管理
```go
// 创建配置管理器
configPath := "/path/to/config.json"
cm := NewConfigManager(configPath)

// 加载配置
config, err := cm.LoadConfig()
if err != nil {
    log.Printf("加载配置失败: %v", err)
}

// 保存配置
err = cm.SaveConfig(config)
if err != nil {
    log.Printf("保存配置失败: %v", err)
}
```

### 自定义日志
```go
// 创建文件日志记录器
logger, err := NewFileLogger(LogLevelDebug, "/path/to/log.txt")
if err != nil {
    log.Fatal(err)
}
defer logger.Close()

// 使用自定义日志创建服务
service := NewServiceWithLogger(config, logger)
```

## 命令行接口

服务支持命令行参数：

```bash
# 基本删除
delguard delete file1.txt file2.txt

# 详细输出
delguard delete -v file1.txt

# 干运行模式
delguard delete -n file1.txt

# 强制删除（忽略错误）
delguard delete -f file1.txt

# 批量模式
delguard delete -b file1.txt file2.txt file3.txt

# 递归删除
delguard delete -r directory/
```

## 错误处理

模块提供详细的错误分类：

- `ErrFileNotFound` - 文件不存在
- `ErrPermissionDenied` - 权限被拒绝
- `ErrProtectedPath` - 受保护的路径
- `ErrInvalidPath` - 无效路径
- `ErrFileInUse` - 文件正在使用
- `ErrDiskFull` - 磁盘空间不足
- `ErrNetworkError` - 网络错误
- `ErrTimeout` - 操作超时
- `ErrCancelled` - 操作被取消

## 性能优化

### 并发控制
- 使用信号量控制并发数
- 避免资源竞争
- 优化内存使用

### 统计监控
- 原子操作保证线程安全
- 最小化锁竞争
- 高效的统计数据收集

### 错误处理
- 快速错误分类
- 避免重复错误检查
- 优化错误信息格式化

## 测试

模块包含完整的测试套件：

```bash
# 运行单元测试
go test ./internal/core/delete

# 运行集成测试
go test -tags=integration ./internal/core/delete

# 运行基准测试
go test -bench=. ./internal/core/delete

# 生成测试覆盖率报告
go test -cover ./internal/core/delete
```

## 平台支持

- ✅ Windows (回收站支持)
- ✅ macOS (废纸篓支持)
- ✅ Linux (Trash支持)
- ✅ 其他Unix系统

## 配置文件位置

默认配置文件位置：
- Windows: `%APPDATA%\delguard\config.json`
- macOS: `~/Library/Application Support/delguard/config.json`
- Linux: `~/.config/delguard/config.json`

## 日志文件位置

默认日志文件位置：
- Windows: `%APPDATA%\delguard\logs\delguard.log`
- macOS: `~/Library/Logs/delguard/delguard.log`
- Linux: `~/.local/share/delguard/logs/delguard.log`

## 贡献指南

1. 遵循Go代码规范
2. 添加适当的测试用例
3. 更新相关文档
4. 确保跨平台兼容性
5. 保持向后兼容性

## 许可证

本项目采用MIT许可证，详见LICENSE文件。