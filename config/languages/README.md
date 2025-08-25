# DelGuard Language Packs

在此目录放置语言包文件，可在不修改代码的情况下扩展或覆盖内置翻译。

- 支持文件名：`<lang>.(json|jsonc|ini|cfg|conf|env|properties)`，例如 `en-US.json`、`fr-FR.ini`、`de-DE.properties`、`ja.jsonc`
- 语义：所有格式均表示一个 `map<string,string>`，键是代码中的中文原文，值是目标语言译文
- 加载顺序：外部语言包优先级高于内置翻译，外部同键会覆盖内置
- 语言选择：应用会自动检测系统语言；如缺少对应语言包，则回退到英文 `en-US`；中文 `zh-CN` 为源语言无需语言包
- 提示：单行消息显示时会自动添加前缀 `DelGuard: `

## 可用语言包

目前支持以下语言：

- **ar-SA.json** - العربية (阿拉伯语)
- **de-DE.properties** - Deutsch (德语)
- **es-ES.json** - Español (西班牙语)
- **fr-FR.ini** - Français (法语)
- **hi-IN.json** - हिन्दी (印地语)
- **it-IT.properties** - Italiano (意大利语)
- **ja.json** - 日本語 (日语)
- **ko-KR.json** - 한국어 (韩语)
- **nl-NL.jsonc** - Nederlands (荷兰语)
- **no-NO.json** - Norsk (挪威语)
- **pt-BR.ini** - Português (葡萄牙语)
- **ru-RU.json** - Русский (俄语)
- **th-TH.json** - ไทย (泰语)
- **vi-VN.json** - Tiếng Việt (越南语)

## JSON 示例

```json
{
  "删除 %s ? [y/N/a/q]: ": "Delete %s ? [y/N/a/q]: ",
  "已将 %s 移动到回收站\n": "Moved %s to Trash\n"
}
```

## JSONC 示例（支持注释）

```jsonc
{
  // 行注释
  "错误：无法删除 %s: %v\n": "Error: failed to delete %s: %v\n"
}
```

## INI/CFG/CONF 示例（键值对，忽略段落名）

```ini
[general]
删除 %s ? [y/N/a/q]: = Delete %s ? [y/N/a/q]: 
已将 %s 移动到回收站\n = Moved %s to Trash\n
```

## .env / .properties 示例

```properties
删除 %s ? [y/N/a/q]: = Delete %s ? [y/N/a/q]: 
已将 %s 移动到回收站\n = Moved %s to Trash\n
```
