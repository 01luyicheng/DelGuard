# 变更日志

本文档记录了 DelGuard 项目的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

### 计划中
- 图形用户界面 (GUI)
- 云存储集成
- 文件加密功能
- 多语言支持扩展

## [2.0.0] - 2024-01-XX

### 新增
- 🎉 **全新架构设计** - 采用模块化设计，提升可维护性
- 🔒 **增强安全功能** - 系统路径保护和权限验证
- 🔍 **智能搜索引擎** - 支持模式匹配、大小过滤和重复文件检测
- ⚡ **性能优化系统** - 内存管理和并发处理优化
- 🌐 **跨平台支持** - Windows、Linux、macOS 完整支持
- 📊 **监控和指标** - 实时性能监控和操作统计
- 🧪 **完整测试框架** - 单元测试、集成测试和性能测试
- 📖 **详细文档** - API文档、安装指南和用户手册
- 🔧 **配置管理** - 灵活的JSON配置文件支持
- 🗂️ **回收站集成** - 原生回收站支持，安全删除文件

### 改进
- **命令行界面** - 更直观的命令结构和参数
- **错误处理** - 详细的错误信息和恢复建议
- **日志系统** - 结构化日志和多级别输出
- **构建系统** - 自动化构建和发布流程

### 技术栈
- **语言**: Go 1.21+
- **架构**: 模块化设计，清晰的包结构
- **测试**: 完整的测试覆盖率
- **文档**: 全面的API和用户文档

### 项目结构
```
delguard/
├── cmd/delguard/           # 主程序入口
├── internal/               # 内部包
│   ├── core/              # 核心业务逻辑
│   ├── platform/          # 平台相关代码
│   ├── config/            # 配置管理
│   ├── monitor/           # 监控和指标
│   └── ui/                # 用户界面
├── pkg/delguard/          # 公共API
├── configs/               # 配置文件
├── docs/                  # 文档
├── scripts/               # 构建和部署脚本
└── tests/                 # 测试文件
```

## [1.x.x] - 历史版本

### [1.2.0] - 2023-12-XX
#### 新增
- 批量删除功能
- 基本配置文件支持
- Windows 回收站集成

#### 修复
- 修复了文件权限检查问题
- 解决了路径处理的跨平台兼容性问题

### [1.1.0] - 2023-11-XX
#### 新增
- 安全删除功能
- 基本的文件验证
- 简单的命令行界面

#### 改进
- 优化了删除性能
- 改进了错误消息

### [1.0.0] - 2023-10-XX
#### 新增
- 🎉 **首次发布**
- 基本的文件删除功能
- 简单的命令行工具
- Windows 平台支持

---

## 版本说明

### 语义化版本控制

DelGuard 遵循 [语义化版本控制](https://semver.org/lang/zh-CN/) 规范：

- **主版本号 (MAJOR)**: 当做了不兼容的 API 修改
- **次版本号 (MINOR)**: 当做了向下兼容的功能性新增
- **修订号 (PATCH)**: 当做了向下兼容的问题修正

### 发布周期

- **主版本**: 每年1-2次，包含重大功能更新
- **次版本**: 每季度1次，包含新功能和改进
- **修订版**: 根据需要，主要修复bug和安全问题

### 支持政策

- **当前版本**: 完整支持，包括新功能和bug修复
- **前一个主版本**: 安全更新和关键bug修复
- **更早版本**: 仅安全更新（如果有严重安全问题）

### 升级指南

#### 从 1.x 升级到 2.0

**重大变更:**
1. **命令行接口变更** - 部分命令参数已更改
2. **配置文件格式** - 新的JSON格式配置文件
3. **API变更** - 如果您使用DelGuard作为库，请查看API文档

**升级步骤:**
1. 备份现有配置和数据
2. 卸载旧版本
3. 安装新版本
4. 迁移配置文件
5. 测试功能

**配置迁移:**
```bash
# 备份旧配置
cp ~/.delguard/config ~/.delguard/config.backup

# 使用迁移工具
delguard config migrate --from 1.x --to 2.0

# 或手动创建新配置
delguard config init
```

### 已知问题

#### 2.0.0
- 在某些Linux发行版上，回收站功能可能需要额外配置
- macOS Catalina 及更早版本可能需要手动授权

#### 解决方案
- 查看 [故障排除指南](docs/INSTALL.md#故障排除)
- 提交 [Issue](https://github.com/your-username/delguard/issues)

### 贡献者

感谢所有为 DelGuard 做出贡献的开发者：

#### 2.0.0 版本贡献者
- [@username1](https://github.com/username1) - 核心架构设计
- [@username2](https://github.com/username2) - 跨平台支持
- [@username3](https://github.com/username3) - 测试框架
- [@username4](https://github.com/username4) - 文档编写

#### 历史贡献者
- [@founder](https://github.com/founder) - 项目创始人
- [@contributor1](https://github.com/contributor1) - 早期开发
- [@contributor2](https://github.com/contributor2) - Windows支持

### 致谢

特别感谢以下项目和社区：
- [Go 语言团队](https://golang.org/) - 优秀的编程语言
- [Cobra](https://github.com/spf13/cobra) - CLI框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- 所有提供反馈和建议的用户

---

## 获取更新

### 自动更新检查
```bash
delguard update check
```

### 手动更新
```bash
delguard update install
```

### 订阅发布通知
- 在 GitHub 上 Watch 本项目
- 关注我们的 [发布页面](https://github.com/your-username/delguard/releases)

---

**注意**: 在升级到新版本之前，请务必阅读相关的升级指南和已知问题。