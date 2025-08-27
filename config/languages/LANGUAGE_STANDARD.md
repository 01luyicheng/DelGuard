# DelGuard 国际化配置标准

## 语言代码规范

### 标准格式
- **主要格式**: `zh-CN`, `en-US`, `ja-JP` (遵循 IETF BCP 47 标准)
- **简化格式**: `zh`, `en`, `ja` (ISO 639-1 双字母代码)
- **兼容性格式**: 同时支持 `zh_cn`, `zh-cn`, `zh_CN` (自动规范化)

### 支持的语言
- `zh-CN` - 简体中文 (默认)
- `en-US` - 美式英语
- `ja-JP` - 日语
- `ko-KR` - 韩语
- `fr-FR` - 法语
- `de-DE` - 德语
- `es-ES` - 西班牙语
- `ru-RU` - 俄语
- `nl-NL` - 荷兰语
- `no-NO` - 挪威语

## 文件格式标准

### 支持的文件格式
1. **JSON** (推荐) - `.json`
2. **JSONC** (带注释) - `.jsonc`
3. **INI** - `.ini`, `.cfg`, `.conf`
4. **Properties** - `.properties`, `.env`

### 文件命名规范
```
<语言代码>.<扩展名>
```

示例:
- `zh-CN.json`
- `en-US.jsonc`
- `ja.ini`
- `de.properties`

### JSON 格式标准
```json
{
  "原文本": "翻译文本",
  "Error: file not found": "错误：文件未找到",
  "Confirm deletion": "确认删除"
}
```

### INI 格式标准
```ini
# 注释支持
原文本 = 翻译文本
Error: file not found = 错误：文件未找到
Confirm deletion = 确认删除

# 支持冒号分隔
原文本: 翻译文本
```

### Properties 格式标准
```properties
# 标准 properties 格式
original.text=翻译文本
error.message=错误消息
```

## 优先级规则

1. **CLI 参数** (`--lang`) - 最高优先级
2. **环境变量** (`DELGUARD_LANGUAGE`) 
3. **配置文件** (`config.json` 中的 language 字段)
4. **系统检测** - 自动检测系统语言
5. **默认值** - 简体中文 (`zh-CN`)

## 翻译规范

### 术语一致性
- **删除**: 统一使用"删除"而非"移除"、"清除"
- **文件**: 使用"文件"而非"文档"
- **目录**: 使用"目录"而非"文件夹"
- **确认**: 使用"确认"而非"确定"

### 格式占位符
- 保持与原文相同的占位符格式
- 支持 `%s`, `%d`, `%v` 等标准格式
- 支持 `{0}`, `{1}` 等索引格式

### 示例
```json
{
  "Delete %d files?": "删除 %d 个文件？",
  "File '%s' not found": "文件 '%s' 未找到",
  "Confirm deletion of {0} items": "确认删除 {0} 个项目？"
}
```

## 验证规则

### 必需字段
每个语言文件必须包含以下核心字段:
- `DelGuard v%s - Cross-platform secure deletion tool`
- `Usage:`
- `Options:`
- `Confirm deletion`
- `Error:`
- `Success:`

### 文件验证
- 文件必须是有效的 UTF-8 编码
- JSON 文件必须符合 JSON 语法
- 占位符数量必须与原文匹配
- 特殊字符需要正确转义

## 扩展指南

### 添加新语言
1. 创建新的语言文件: `config/languages/<lang-code>.json`
2. 复制 `en-US.json` 作为模板
3. 翻译所有文本内容
4. 运行验证: `go run verify_security.go`

### 测试国际化
```bash
# 测试特定语言
delguard --lang en-US --help
delguard --lang ja --version

# 测试自动检测
delguard --lang auto --help
```