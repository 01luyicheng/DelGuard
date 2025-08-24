# DelGuard 智能删除功能

## 🚀 新功能概览

DelGuard 现在具备了强大的智能删除功能，让文件删除操作更加智能、安全和人性化。

## ✨ 核心特性

### 1. 智能文件搜索
当您输入的文件名不存在时，DelGuard 会自动搜索相似的文件名并提供选择：

```bash
# 输入不存在的文件名
delguard test_doc

# 输出示例：
🔍 未找到文件 'test_doc'，找到以下相似文件：

[1] 📄 test_document.txt (相似度: 85.2%, 类型: filename)
[2] 📄 test_data.txt (相似度: 72.1%, 类型: filename)
[3] 📁 test_docs/ (相似度: 90.5%, 类型: filename)

请选择文件编号，或输入 'n' 取消操作: 
```

### 2. 正则表达式批量操作
支持通配符和正则表达式进行批量文件操作：

```bash
# 删除所有 .txt 文件
delguard *.txt

# 删除所有以 test_ 开头的文件
delguard test_*

# 使用正则表达式
delguard "^backup_\d{4}\.log$"
```

### 3. 扩展搜索范围
- **文件内容搜索**：在文件内容中搜索关键词
- **父目录搜索**：向上搜索父目录
- **子目录搜索**：递归搜索子目录

```bash
# 搜索文件内容
delguard --search-content "重要文档"

# 搜索父目录
delguard --search-parent config.txt

# 递归搜索子目录
delguard -r temp_file
```

### 4. 二次确认机制
批量操作时提供详细的确认界面：

```bash
# 批量删除确认示例
⚠️  找到 15 个匹配文件

[1] test1.txt (2.5 KB)
[2] test2.txt (1.8 KB)
[3] test3.txt (3.2 KB)
...

⚠️  确认删除以上文件？(y/N): 
```

### 5. 人性化错误提示
提供清晰、友好的错误信息和解决建议：

```bash
❌ 无法删除文件 'important.txt'
原因：文件被其他程序占用
建议：关闭相关程序后重试，或使用 --force 参数强制删除
```

## 🛠️ 命令行参数

### 智能搜索选项
```bash
--smart-search      # 启用智能搜索（默认开启）
--search-content    # 搜索文件内容
--search-parent     # 搜索父目录
--similarity=N      # 相似度阈值（0-100，默认60）
--max-results=N     # 最大搜索结果数（默认10）
--force-confirm     # 跳过二次确认
```

## 📝 使用示例

### 基础智能搜索
```bash
# 智能搜索相似文件名
delguard test_fil

# 输出：
🔍 自动选择高相似度文件: test_file.txt (95.2%)
✅ 成功删除: test_file.txt
```

### 批量操作
```bash
# 批量删除跳过确认
delguard *.tmp --force-confirm

# 交互式批量删除
delguard *.log
# 会显示所有匹配文件并要求确认
```

### 内容搜索
```bash
# 在文件内容中搜索
delguard --search-content "配置文件"

# 输出：
🔍 在文件内容中找到匹配：
[1] 📄 app.config (匹配: "配置文件设置")
[2] 📄 settings.ini (匹配: "用户配置文件")
```

## 🔧 高级功能

### 相似度算法
使用 Levenshtein 距离算法计算文件名相似度，支持：
- 字符插入、删除、替换
- 大小写不敏感匹配
- 路径分隔符标准化

### 搜索优先级
1. 精确匹配
2. 正则表达式匹配
3. 文件名相似度匹配
4. 文件内容匹配

### 安全保护
- 系统文件保护
- 隐藏文件确认
- 大文件警告
- 批量操作二次确认

## 🎯 最佳实践

1. **使用智能搜索**：让 DelGuard 帮您找到想要删除的文件
2. **设置合适的相似度阈值**：根据需要调整 `--similarity` 参数
3. **谨慎使用 `--force-confirm`**：仅在确定操作安全时跳过确认
4. **利用内容搜索**：通过 `--search-content` 找到包含特定内容的文件

## 🚨 注意事项

- 智能搜索功能默认开启，可通过 `--smart-search=false` 禁用
- 批量操作会显示详细的文件列表供确认
- 所有删除操作都会将文件移动到回收站，可通过 `restore` 命令恢复
- 系统关键文件会有特殊保护和警告提示

---

**DelGuard** - 让文件删除更智能、更安全！