# DelGuard API 文档

## 概述

DelGuard 提供了完整的Go语言API，支持在其他Go项目中集成DelGuard的功能。

## 核心包

### delguard/internal/core/delete

删除服务包，提供安全文件删除功能。

#### 类型定义

```go
// DeleteResult 删除操作结果
type DeleteResult struct {
    Path    string  // 文件路径
    Success bool    // 是否成功
    Error   error   // 错误信息
}

// Service 删除服务
type Service struct {
    // 私有字段
}
```

#### 函数

##### NewService
```go
func NewService(config ...interface{}) *Service
```
创建新的删除服务实例。

**参数:**
- `config` - 可选的配置参数

**返回值:**
- `*Service` - 删除服务实例

##### ValidateFile
```go
func (s *Service) ValidateFile(filePath string) error
```
验证文件路径是否安全可删除。

**参数:**
- `filePath` - 要验证的文件路径

**返回值:**
- `error` - 验证错误，nil表示验证通过

##### SafeDelete
```go
func (s *Service) SafeDelete(filePath string) error
```
安全删除单个文件。

**参数:**
- `filePath` - 要删除的文件路径

**返回值:**
- `error` - 删除错误，nil表示删除成功

##### BatchDelete
```go
func (s *Service) BatchDelete(filePaths []string) []DeleteResult
```
批量删除多个文件。

**参数:**
- `filePaths` - 要删除的文件路径列表

**返回值:**
- `[]DeleteResult` - 每个文件的删除结果

##### MoveToRecycleBin
```go
func (s *Service) MoveToRecycleBin(filePath string) error
```
将文件移动到回收站。

**参数:**
- `filePath` - 要移动的文件路径

**返回值:**
- `error` - 操作错误，nil表示成功

### delguard/internal/core/search

搜索服务包，提供文件搜索和查找功能。

#### 类型定义

```go
// FileInfo 文件信息
type FileInfo struct {
    Path    string  // 文件路径
    Size    int64   // 文件大小
    ModTime int64   // 修改时间
    IsDir   bool    // 是否为目录
    Hash    string  // 文件哈希值
}

// DuplicateGroup 重复文件组
type DuplicateGroup struct {
    Hash  string     // 文件哈希值
    Files []FileInfo // 重复文件列表
}

// Service 搜索服务
type Service struct {
    // 私有字段
}
```

#### 函数

##### NewService
```go
func NewService(config ...interface{}) *Service
```
创建新的搜索服务实例。

##### FindFiles
```go
func (s *Service) FindFiles(rootPath, pattern string, recursive bool) ([]FileInfo, error)
```
按模式查找文件。

**参数:**
- `rootPath` - 搜索根目录
- `pattern` - 文件名模式（支持通配符）
- `recursive` - 是否递归搜索

**返回值:**
- `[]FileInfo` - 找到的文件列表
- `error` - 搜索错误

##### FindBySize
```go
func (s *Service) FindBySize(rootPath string, minSize int64, recursive bool) ([]FileInfo, error)
```
按大小查找文件。

**参数:**
- `rootPath` - 搜索根目录
- `minSize` - 最小文件大小（字节）
- `recursive` - 是否递归搜索

**返回值:**
- `[]FileInfo` - 找到的文件列表
- `error` - 搜索错误

##### FindDuplicates
```go
func (s *Service) FindDuplicates(rootPath string) ([]DuplicateGroup, error)
```
查找重复文件。

**参数:**
- `rootPath` - 搜索根目录

**返回值:**
- `[]DuplicateGroup` - 重复文件组列表
- `error` - 搜索错误

### delguard/internal/config

配置管理包，提供配置文件的读取、写入和验证功能。

#### 类型定义

```go
// Config 配置结构
type Config struct {
    Language         string `json:"language"`
    MaxFileSize      int64  `json:"max_file_size"`
    MaxBackupFiles   int    `json:"max_backup_files"`
    EnableRecycleBin bool   `json:"enable_recycle_bin"`
    EnableLogging    bool   `json:"enable_logging"`
    LogLevel         string `json:"log_level"`
    ConfigPath       string `json:"-"`
}
```

#### 函数

##### NewConfig
```go
func NewConfig() *Config
```
创建默认配置实例。

##### Load
```go
func Load() (*Config, error)
```
从默认位置加载配置文件。

##### LoadFromFile
```go
func (c *Config) LoadFromFile(filePath string) error
```
从指定文件加载配置。

##### SaveToFile
```go
func (c *Config) SaveToFile(filePath string) error
```
保存配置到指定文件。

##### Validate
```go
func (c *Config) Validate() error
```
验证配置的有效性。

## 使用示例

### 基本删除操作

```go
package main

import (
    "fmt"
    "delguard/internal/core/delete"
)

func main() {
    // 创建删除服务
    service := delete.NewService()
    
    // 验证文件
    if err := service.ValidateFile("test.txt"); err != nil {
        fmt.Printf("文件验证失败: %v\n", err)
        return
    }
    
    // 安全删除文件
    if err := service.SafeDelete("test.txt"); err != nil {
        fmt.Printf("删除失败: %v\n", err)
        return
    }
    
    fmt.Println("文件删除成功")
}
```

### 批量删除操作

```go
package main

import (
    "fmt"
    "delguard/internal/core/delete"
)

func main() {
    service := delete.NewService()
    
    files := []string{"file1.txt", "file2.txt", "file3.txt"}
    results := service.BatchDelete(files)
    
    for _, result := range results {
        if result.Success {
            fmt.Printf("✅ %s 删除成功\n", result.Path)
        } else {
            fmt.Printf("❌ %s 删除失败: %v\n", result.Path, result.Error)
        }
    }
}
```

### 文件搜索操作

```go
package main

import (
    "fmt"
    "delguard/internal/core/search"
)

func main() {
    service := search.NewService()
    
    // 搜索所有.txt文件
    files, err := service.FindFiles("/home/user", "*.txt", true)
    if err != nil {
        fmt.Printf("搜索失败: %v\n", err)
        return
    }
    
    fmt.Printf("找到 %d 个文件:\n", len(files))
    for _, file := range files {
        fmt.Printf("- %s (%d bytes)\n", file.Path, file.Size)
    }
}
```

### 重复文件检测

```go
package main

import (
    "fmt"
    "delguard/internal/core/search"
)

func main() {
    service := search.NewService()
    
    duplicates, err := service.FindDuplicates("/home/user/Documents")
    if err != nil {
        fmt.Printf("重复文件检测失败: %v\n", err)
        return
    }
    
    fmt.Printf("找到 %d 组重复文件:\n", len(duplicates))
    for i, group := range duplicates {
        fmt.Printf("组 %d (哈希: %s):\n", i+1, group.Hash[:8])
        for _, file := range group.Files {
            fmt.Printf("  - %s\n", file.Path)
        }
    }
}
```

### 配置管理

```go
package main

import (
    "fmt"
    "delguard/internal/config"
)

func main() {
    // 加载配置
    cfg, err := config.Load()
    if err != nil {
        fmt.Printf("加载配置失败: %v\n", err)
        return
    }
    
    // 修改配置
    cfg.Language = "en-us"
    cfg.MaxFileSize = 2 * 1024 * 1024 * 1024 // 2GB
    
    // 验证配置
    if err := cfg.Validate(); err != nil {
        fmt.Printf("配置验证失败: %v\n", err)
        return
    }
    
    // 保存配置
    if err := cfg.Save(); err != nil {
        fmt.Printf("保存配置失败: %v\n", err)
        return
    }
    
    fmt.Println("配置更新成功")
}
```

## 错误处理

DelGuard API 使用标准的Go错误处理模式。所有可能失败的操作都会返回error类型的值。

### 常见错误类型

- **文件不存在错误** - 当尝试操作不存在的文件时
- **权限错误** - 当没有足够权限执行操作时
- **系统保护错误** - 当尝试删除受保护的系统文件时
- **配置错误** - 当配置文件格式错误或包含无效值时

### 错误处理最佳实践

```go
if err := service.SafeDelete(filePath); err != nil {
    // 记录错误
    log.Printf("删除文件失败: %v", err)
    
    // 根据错误类型进行不同处理
    if os.IsNotExist(err) {
        fmt.Println("文件不存在")
    } else if os.IsPermission(err) {
        fmt.Println("权限不足")
    } else {
        fmt.Printf("未知错误: %v", err)
    }
    
    return err
}
```

## 性能考虑

### 批量操作
对于大量文件操作，建议使用批量API而不是循环调用单个文件API：

```go
// 推荐：使用批量API
results := service.BatchDelete(filePaths)

// 不推荐：循环调用单个API
for _, path := range filePaths {
    service.SafeDelete(path)
}
```

### 内存使用
在处理大量文件时，注意内存使用：

```go
// 分批处理大量文件
const batchSize = 1000
for i := 0; i < len(allFiles); i += batchSize {
    end := i + batchSize
    if end > len(allFiles) {
        end = len(allFiles)
    }
    
    batch := allFiles[i:end]
    results := service.BatchDelete(batch)
    // 处理结果...
}
```

## 线程安全

DelGuard的所有服务都是线程安全的，可以在多个goroutine中并发使用：

```go
var wg sync.WaitGroup
service := delete.NewService()

for _, file := range files {
    wg.Add(1)
    go func(filePath string) {
        defer wg.Done()
        if err := service.SafeDelete(filePath); err != nil {
            log.Printf("删除 %s 失败: %v", filePath, err)
        }
    }(file)
}

wg.Wait()