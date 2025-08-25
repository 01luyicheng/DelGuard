# DelGuard Copilot Instructions

## 项目架构与核心组件
- DelGuard 是跨平台（Windows/macOS/Linux）安全删除工具，主入口为 `main.go`，核心逻辑分布于 `file_operations.go`、`security_check.go`、`protect.go`、`restore.go`、`safe_copy.go` 等文件。
- 语言包与配置文件支持多格式，分别位于 `config/languages/` 和用户/系统/当前目录，详见 `README.md` 和 `config/languages/README.md`。
- 主要功能包括安全删除、恢复、覆盖保护、安全复制、权限与路径校验，均有独立模块实现。

## 开发者工作流
- 构建：使用 `build.ps1`（Windows PowerShell）、`build_all.bat`、`build_all.sh` 或 `Makefile`，推荐直接运行 `go build` 或相关脚本。
- 测试：标准测试命令为 `go test ./...`，安全相关测试可用 `go test -run TestSecurity ./...`。
- 安装/部署：一键安装脚本见 `scripts/install.ps1`（Windows）和 `scripts/install.sh`（macOS/Linux），支持自动/手动安装和卸载。
- 交互/调试：可用 `-i` 参数启用交互模式，`--verbose` 查看详细信息，`--dry-run` 试运行。

## 项目约定与特殊模式
- 所有删除操作默认移动到系统回收站，彻底删除需加 `--force`。
- 文件覆盖保护默认启用，相关逻辑在 `overwrite_protect.go` 和 `safe_copy.go`，可用 `--protect`/`--disable-protect` 控制。
- 多语言支持自动检测系统语言，外部语言包优先级高于内置，缺失时回退英文。
- 配置文件优先级：`--config` 指定 > 用户目录 > 系统目录 > 当前目录。
- 交互模式和安全复制均有详细提示与确认，见 `interactive_ui.go` 和 `safe_copy.go`。

## 关键文件与目录
- `main.go`：程序入口，参数解析与主流程
- `file_operations.go`：文件删除/移动/恢复核心逻辑
- `security_check.go`、`protect.go`：安全校验与保护机制
- `safe_copy.go`：安全复制与覆盖保护
- `restore.go`：回收站恢复功能
- `config.go`、`i18n.go`：配置与多语言加载
- `config/languages/`：外部语言包
- `Makefile`、`build.ps1`、`build_all.bat`、`build_all.sh`：构建脚本
- `scripts/install.ps1`、`scripts/install.sh`：安装脚本

## 代码风格与模式
- 错误处理统一返回详细错误码与建议，见 `errors.go`。
- 所有交互提示均可国际化，原文为中文，翻译由语言包覆盖。
- 文件操作前均有权限与路径校验，防止误删系统/关键文件。
- 测试覆盖常见场景与安全边界，见 `*_test.go` 文件。

1. 请保持对话语言为中文
2. 当前系统为 Windows11 24H2 x64
3. 请永远不要在代码中使用伪代码、虚拟占位符、测试数据，我们的项目代码将会直接面向生产环境，请务必认真对待。请先阅读项目文档，确认需求，再进行开发。如果需要帮助，请告诉我，我可以为你提供测试用的API密钥、Linux服务器地址及ssh密钥、GitHub认证token、文档资料、图片等资源。
---
如需补充特殊约定、集成流程或有疑问，请在下方补充说明。
