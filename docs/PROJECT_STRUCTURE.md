# DelGuard 项目结构

## 核心文件

### 主程序
- `main.go` - 程序入口点
- `main_commands.go` - 命令行处理
- `types.go` - 核心数据类型定义

### 核心功能
- `core_delete.go` - 核心删除逻辑
- `smart_delete.go` - 智能删除功能
- `file_operations.go` - 文件操作工具
- `restore.go` - 文件恢复功能

### 平台支持
- `windows.go` - Windows平台特定功能
- `linux.go` - Linux平台特定功能  
- `macos.go` - macOS平台特定功能
- `platform.go` - 跨平台抽象层

### 安全与验证
- `security_check.go` - 安全检查
- `input_validator.go` - 输入验证
- `file_validator.go` - 文件验证
- `protect.go` - 文件保护机制
- `verify_security.go` - 安全验证

### 配置管理
- `config.go` - 基础配置
- `config_advanced.go` - 高级配置
- `smart_config.go` - 智能配置
- `config_validator_impl.go` - 配置验证

### 用户界面
- `interactive_ui.go` - 交互式界面
- `help_system.go` - 帮助系统
- `enhanced_feedback.go` - 增强反馈

### 工具与实用程序
- `path_utils.go` - 路径工具
- `string_utils.go` - 字符串工具
- `search_tool.go` - 搜索工具
- `smart_search.go` - 智能搜索
- `similarity.go` - 相似度计算

### 日志与错误处理
- `logger.go` - 日志系统
- `log_sanitizer.go` - 日志清理
- `errors.go` - 错误处理
- `health_check.go` - 健康检查

### 资源管理
- `resource_manager.go` - 资源管理
- `concurrency_manager.go` - 并发管理
- `trash_monitor.go` - 回收站监控

## 配置文件

### 语言支持
- `config/languages/` - 多语言配置文件
  - `zh-cn.json` - 简体中文
  - `en-US.json` - 英文
  - 其他语言文件

### 安装配置
- `config/install-config.json` - 安装配置
- `config/install-messages.json` - 安装消息
- `config/security_template.json` - 安全模板

## 构建与部署

### 构建脚本
- `build.sh` - Unix构建脚本
- `build.bat` - Windows批处理构建
- `build.ps1` - PowerShell构建脚本
- `build_crossplatform.sh` - 跨平台构建

### 安装脚本
- `install.sh` - Unix安装脚本
- `install.bat` - Windows批处理安装
- `install.ps1` - PowerShell安装脚本

### 验证脚本
- `verify_crossplatform.ps1` - 跨平台验证
- `check.ps1` - 检查脚本

## 文档

### 用户文档
- `README.md` - 项目说明
- `docs/QUICK_START.md` - 快速开始
- `docs/SECURITY.md` - 安全说明
- `docs/DEVELOPMENT.md` - 开发指南

### 发布文档
- `CHANGELOG.md` - 更新日志
- `RELEASE_CHECKLIST.md` - 发布检查清单
- `PRODUCTION_READINESS_REPORT.md` - 生产就绪报告

## 测试

### 单元测试
- `tests/unit/` - 单元测试

### 集成测试
- `tests/integration/` - 集成测试

### 性能测试
- `tests/benchmarks/` - 性能基准测试

### 安全测试
- `tests/security/` - 安全测试

## 开发工具

### 脚本
- `dev-tools/scripts/` - 开发脚本
- `scripts/` - 通用脚本

### 实用工具
- `utils/` - 实用工具

## 构建产物

- `build/` - 构建输出目录
- `dist/` - 分发包目录
- `delguard.exe` - Windows可执行文件

## 项目特点

1. **跨平台支持** - 支持Windows、Linux、macOS
2. **安全第一** - 多重安全检查和验证
3. **智能删除** - 智能识别和处理不同类型文件
4. **多语言支持** - 支持多种界面语言
5. **可配置** - 丰富的配置选项
6. **可恢复** - 支持从回收站恢复文件
7. **高性能** - 优化的删除算法