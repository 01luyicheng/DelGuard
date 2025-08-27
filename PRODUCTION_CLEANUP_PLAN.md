# DelGuard 生产环境清理计划

## 🎯 清理目标
确保项目完全准备好用于生产环境，移除所有测试代码、调试输出和开发工具。

## 📋 需要清理的内容

### 1. 测试文件清理
以下文件应该移动到 `tests/` 目录或删除：
- `delguard_test.go` - 主测试文件
- `main_test.go` - 主程序测试
- `path_utils_test.go` - 路径工具测试
- `verify_crossplatform.go` - 跨平台验证测试
- `test_linux_compatibility.go` - Linux兼容性测试

### 2. 调试代码清理
需要移除或条件化的调试输出：
- 所有 `log_debug()` 调用
- 所有 `print_debug()` 调用
- 测试相关的 `fmt.Println()` 输出
- 开发阶段的详细日志输出

### 3. 开发脚本清理
以下脚本仅用于开发，应移动到 `dev-tools/` 目录：
- `scripts/test-*.ps1` - 测试脚本
- `scripts/test-*.sh` - 测试脚本
- `scripts/comprehensive-compatibility-check*.ps1` - 兼容性检查
- `scripts/fix-cross-platform*.ps1` - 跨平台修复脚本
- `scripts/verify-cross-platform.ps1` - 验证脚本

### 4. 配置文件优化
- 移除测试用的配置模板
- 确保默认配置适合生产环境
- 移除调试级别的默认日志设置

## 🚀 推荐的清理步骤

### 步骤1: 创建开发工具目录
```bash
mkdir -p dev-tools/tests
mkdir -p dev-tools/scripts
```

### 步骤2: 移动测试文件
```bash
mv *_test.go dev-tools/tests/
mv verify_crossplatform.go dev-tools/tests/
mv test_linux_compatibility.go dev-tools/tests/
```

### 步骤3: 移动开发脚本
```bash
mv scripts/test-* dev-tools/scripts/
mv scripts/comprehensive-compatibility-check* dev-tools/scripts/
mv scripts/fix-cross-platform* dev-tools/scripts/
mv scripts/verify-cross-platform.ps1 dev-tools/scripts/
```

### 步骤4: 清理调试代码
- 将所有 `log_debug` 调用包装在条件判断中
- 移除测试相关的 `fmt.Println` 输出
- 设置生产环境默认日志级别为 "info"

### 步骤5: 创建缺失的文件
- 创建 `scripts/uninstall.ps1`
- 确保所有引用的文件都存在

## 📁 建议的最终目录结构

```
DelGuard/
├── build/                    # 构建输出
├── config/                   # 配置文件
├── docs/                     # 文档
├── scripts/                  # 生产环境脚本
│   ├── install.ps1          # PowerShell安装脚本
│   ├── install.sh           # Bash安装脚本
│   ├── uninstall.ps1        # PowerShell卸载脚本
│   └── build-all-platforms.ps1  # 构建脚本
├── utils/                    # 工具函数
├── dev-tools/               # 开发工具（新增）
│   ├── tests/               # 测试文件
│   └── scripts/             # 开发脚本
├── *.go                     # 核心Go源码
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── CHANGELOG.md
```

## ✅ 验证清单

- [ ] 所有测试文件已移动到 `dev-tools/tests/`
- [ ] 所有开发脚本已移动到 `dev-tools/scripts/`
- [ ] 调试输出已清理或条件化
- [ ] 默认配置适合生产环境
- [ ] 所有引用的文件都存在
- [ ] 安装脚本可以正常运行
- [ ] 构建脚本可以正常运行
- [ ] 文档已更新反映最终结构

## 🔒 安全检查

- [ ] 没有硬编码的测试路径
- [ ] 没有调试信息泄露
- [ ] 所有用户输入都经过验证
- [ ] 错误处理不暴露系统信息
- [ ] 日志记录不包含敏感信息

## 📝 注意事项

1. **保留测试能力**: 测试文件移动到 `dev-tools/` 而不是删除，便于后续维护
2. **条件化调试**: 调试代码通过环境变量或配置控制，而不是完全删除
3. **文档更新**: 确保README和其他文档反映最终的项目结构
4. **版本标记**: 清理完成后应该标记为正式发布版本

## 🎉 完成标准

项目清理完成的标准：
- 可以直接用于生产环境
- 没有测试代码影响性能
- 安装过程简洁可靠
- 用户体验专业流畅
- 代码结构清晰易维护