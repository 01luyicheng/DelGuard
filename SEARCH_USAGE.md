# DelGuard 智能搜索功能使用指南

DelGuard 现在包含一个强大的独立搜索工具，支持智能文件搜索、内容搜索和正则表达式搜索。

## 基本使用

### 启动搜索工具
```bash
delguard search [选项] <搜索目标>
```

### 搜索模式

#### 1. 智能文件名搜索（默认）
```bash
# 搜索文件名包含"config"的文件
delguard search config

# 搜索精确文件名
delguard search "config.json"
```

#### 2. 文件内容搜索
```bash
# 搜索文件内容包含"API_KEY"的文件
delguard search -content "API_KEY"

# 结合递归搜索
delguard search -content -recursive "database"
```

#### 3. 正则表达式搜索
```bash
# 使用正则表达式搜索所有.log文件
delguard search ".*\.log$"

# 搜索日期格式的文件名
delguard search "\d{4}-\d{2}-\d{2}"
```

## 高级选项

### 搜索范围
```bash
# 搜索指定目录
delguard search -dir /path/to/project "main.go"

# 递归搜索子目录
delguard search -recursive "*.go"

# 搜索父目录
delguard search -parent "config"
```

### 过滤选项
```bash
# 按文件扩展名过滤
delguard search -ext .go,.py "utils"

# 按文件大小过滤
delguard search -min-size 1024 -max-size 1048576 "large"

# 按修改日期过滤
delguard search -after 2024-01-01 -before 2024-12-31 "backup"
```

### 输出格式
```bash
# 表格输出（默认）
delguard search -output table "config"

# JSON输出
delguard search -output json "*.json" > results.json

# CSV输出
delguard search -output csv "*.csv" > results.csv

# 列表输出
delguard search -output list "README"
```

### 搜索配置
```bash
# 调整相似度阈值
delguard search -threshold 80 "config"

# 限制结果数量
delguard search -max-results 10 "*.txt"

# 区分大小写
delguard search -case-sensitive "Config"

# 显示详细信息
delguard search -verbose "main.go"
```

## 实际用例

### 1. 查找配置文件
```bash
delguard search -ext .json,.yaml,.yml,.toml -dir /project "config"
```

### 2. 查找大日志文件
```bash
delguard search -ext .log -min-size 10485760 -recursive "error"
```

### 3. 查找最近修改的脚本
```bash
delguard search -ext .sh,.bat,.ps1 -after 2024-01-01 -recursive "backup"
```

### 4. 查找包含敏感信息的文件
```bash
delguard search -content -recursive "password|secret|key"
```

### 5. 结合删除功能使用
```bash
# 先搜索确认，然后删除
delguard search -ext .tmp -recursive "temp"
# 确认后删除找到的文件
delguard *.tmp --recursive
```

## 缓存和历史记录

搜索工具自动启用缓存和历史记录功能：

- **缓存**: 搜索结果缓存30分钟，提高重复搜索速度
- **历史记录**: 保存最近100次搜索，支持搜索建议

### 查看统计信息
```bash
# 查看缓存统计（需要集成到主程序）
delguard --search-stats
```

## 集成使用

搜索功能已集成到主程序的删除操作中：

```bash
# 当文件不存在时，自动搜索相似文件
delguard myfile.txt

# 启用内容搜索
delguard --search-content "API_KEY" --delete

# 智能搜索并交互式删除
delguard --smart-search -interactive "*.log"
```

## 性能提示

1. **大目录搜索**: 使用 `-max-results` 限制结果数量
2. **内容搜索**: 首次搜索可能较慢，后续搜索会利用缓存
3. **正则搜索**: 复杂的正则表达式可能影响性能
4. **网络驱动器**: 考虑使用 `-dir` 限制搜索范围

## 安全特性

- 不会搜索系统关键目录
- 内容搜索时自动跳过二进制文件
- 搜索历史本地存储，不发送网络请求
- 支持取消长时间运行的搜索操作