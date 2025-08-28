# DelGuard v2.0.0 使用说明

## 安装状态
✅ **DelGuard 已成功安装并优化完成**

- **安装路径**: `C:\Program Files\DelGuard\`
- **主程序**: `delguard.exe`
- **配置文件**: `config.yaml`
- **版本**: v2.0.0

## 核心功能测试结果

### ✅ 智能搜索功能
- 支持模糊匹配和多条件搜索
- 支持通配符模式 (*.txt, *.log等)
- 搜索索引优化，提升性能

### ✅ 文件恢复功能  
- 安全删除文件到回收站
- 支持文件完整性验证
- 恢复历史记录跟踪

### ✅ 性能优化
- 内存管理和缓存系统
- 实时性能监控
- 长期运行稳定性保障

## 使用方法

### 基本命令
```powershell
# 使用完整路径（推荐）
& "C:\Program Files\DelGuard\delguard.exe" <命令> [选项] [文件...]

# 如果PATH已配置
delguard <命令> [选项] [文件...]
```

### 主要功能

#### 1. 安全删除文件
```powershell
& "C:\Program Files\DelGuard\delguard.exe" delete file.txt
& "C:\Program Files\DelGuard\delguard.exe" delete folder/ --recursive
& "C:\Program Files\DelGuard\delguard.exe" delete file.txt --force --verbose
```

#### 2. 智能搜索
```powershell
# 搜索特定文件
& "C:\Program Files\DelGuard\delguard.exe" search "filename"

# 通配符搜索
& "C:\Program Files\DelGuard\delguard.exe" search "*.txt"

# 详细搜索
& "C:\Program Files\DelGuard\delguard.exe" search "pattern" --verbose
```

#### 3. 文件恢复
```powershell
# 恢复特定文件
& "C:\Program Files\DelGuard\delguard.exe" restore filename.txt

# 详细恢复信息
& "C:\Program Files\DelGuard\delguard.exe" restore filename.txt --verbose
```

#### 4. 配置管理
```powershell
# 查看配置
& "C:\Program Files\DelGuard\delguard.exe" config show

# 查看版本
& "C:\Program Files\DelGuard\delguard.exe" version

# 查看帮助
& "C:\Program Files\DelGuard\delguard.exe" --help
```

## 选项说明

- `-f, --force`: 强制删除，跳过确认
- `-r, --recursive`: 递归删除目录
- `-v, --verbose`: 详细输出
- `--dry-run`: 预览模式，不实际删除

## 配置信息

当前配置：
- **语言**: 中文 (zh-cn)
- **最大文件大小**: 1GB
- **最大备份文件数**: 10
- **回收站**: 已启用
- **日志记录**: 已启用 (debug级别)

## 性能监控

DelGuard 包含实时性能监控：
- CPU使用率监控
- 内存使用统计
- 文件监控数量
- 删除操作统计
- 运行时间跟踪

## 注意事项

1. **权限要求**: 某些系统文件可能需要管理员权限
2. **PATH配置**: 如需全局使用命令别名，需要管理员权限配置PATH
3. **文件恢复**: 恢复功能依赖于Windows回收站机制
4. **性能优化**: 长期运行时会自动进行内存优化和垃圾回收

## 故障排除

### 命令不识别
如果出现"无法将delguard项识别为cmdlet"错误：
```powershell
# 使用完整路径
& "C:\Program Files\DelGuard\delguard.exe" --help
```

### 权限问题
如果遇到权限拒绝：
```powershell
# 以管理员身份运行PowerShell
Start-Process PowerShell -Verb RunAs
```

## 测试脚本

使用 `test_delguard.ps1` 脚本可以快速测试所有功能：
```powershell
.\test_delguard.ps1
```

---

**DelGuard v2.0.0** - 安全、智能的文件删除保护工具