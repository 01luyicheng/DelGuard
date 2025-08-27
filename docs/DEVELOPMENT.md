# DelGuard 开发文档

## 项目结构

```
DelGuard/
├── .github/workflows/     # GitHub Actions 工作流
├── build/                 # 构建输出目录
├── config/               # 配置文件和语言包
│   └── languages/        # 多语言支持文件
├── docs/                 # 项目文档
├── tests/                # 测试文件
│   ├── benchmarks/       # 性能测试
│   ├── integration/      # 集成测试
│   ├── security/         # 安全测试
│   └── unit/            # 单元测试
├── utils/               # 工具函数
├── *.go                 # Go 源代码文件
├── go.mod               # Go 模块定义
├── go.sum               # Go 依赖校验
├── README.md            # 项目说明
├── LICENSE              # 许可证
├── build.ps1            # Windows 构建脚本
├── build.sh             # Unix 构建脚本
├── install.ps1          # Windows 安装脚本
└── install.sh           # Unix 安装脚本
```

## 核心模块

### 1. 配置管理 (config.go)
- 配置文件加载和验证
- 多格式支持 (JSON, JSONC, INI, ENV, Properties)
- 平台特定配置

### 2. 核心删除 (core_delete.go)
- 文件删除核心逻辑
- 回收站操作
- 安全检查集成

### 3. 安全检查 (security_check.go)
- 路径验证
- 系统文件保护
- 权限检查

### 4. 文件操作 (file_operations.go)
- 文件系统操作封装
- 跨平台兼容性
- 错误处理

### 5. 国际化 (i18n.go)
- 多语言支持
- 动态语言包加载
- 消息本地化

### 6. 日志系统 (logger.go)
- 结构化日志
- 多级别日志
- 日志轮转

## 开发环境设置

### 前置要求
- Go 1.19 或更高版本
- Git
- 支持的操作系统：Windows 10+, Linux, macOS

### 克隆项目
```bash
git clone https://github.com/01luyicheng/DelGuard.git
cd DelGuard
```

### 安装依赖
```bash
go mod download
```

### 运行测试
```bash
# 运行所有测试
go test -v ./...

# 运行特定测试
go test -v -run TestCoreDelete

# 运行基准测试
go test -bench=. -benchmem ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 构建项目
```bash
# 本地构建
go build -o delguard .

# 使用构建脚本（推荐）
# Windows
.\build.ps1 -Version v1.0.0 -Release

# Unix
./build.sh --version v1.0.0 --release
```

## 代码规范

### Go 代码风格
- 遵循 Go 官方代码风格指南
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 使用 `go vet` 检查潜在问题

### 命名规范
- 包名：小写，简短，有意义
- 函数名：驼峰命名，公开函数首字母大写
- 变量名：驼峰命名，简洁明了
- 常量名：全大写，下划线分隔

### 注释规范
- 所有公开函数必须有注释
- 复杂逻辑需要详细注释
- 使用 godoc 格式编写文档注释

### 错误处理
- 使用自定义错误类型
- 提供详细的错误信息
- 实现错误恢复机制

## 测试策略

### 单元测试
- 每个函数都应有对应的单元测试
- 测试覆盖率应达到 80% 以上
- 使用表驱动测试模式

### 集成测试
- 测试模块间的交互
- 验证完整的工作流程
- 模拟真实使用场景

### 性能测试
- 关键路径的性能基准测试
- 内存使用情况监控
- 并发安全性测试

### 安全测试
- 路径遍历攻击防护
- 权限提升防护
- 输入验证测试

## 发布流程

### 版本管理
- 使用语义化版本控制 (SemVer)
- 主版本号：不兼容的 API 修改
- 次版本号：向下兼容的功能性新增
- 修订号：向下兼容的问题修正

### 发布步骤
1. 更新版本号
2. 运行完整测试套件
3. 更新 CHANGELOG.md
4. 创建 Git 标签
5. 推送到 GitHub
6. GitHub Actions 自动构建和发布

### 构建产物
- 多平台二进制文件
- 安装脚本
- 文档和许可证
- 校验和文件

## 贡献指南

### 提交代码
1. Fork 项目
2. 创建特性分支
3. 编写代码和测试
4. 确保所有测试通过
5. 提交 Pull Request

### 代码审查
- 所有代码必须经过审查
- 确保符合代码规范
- 验证测试覆盖率
- 检查安全性问题

### 问题报告
- 使用 GitHub Issues
- 提供详细的重现步骤
- 包含系统环境信息
- 附上相关日志

## 调试技巧

### 本地调试
```bash
# 启用详细日志
delguard -v --log-level debug file.txt

# 试运行模式
delguard --dry-run file.txt

# 使用调试器
dlv debug . -- --help
```

### 性能分析
```bash
# CPU 性能分析
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### 日志分析
- 查看应用程序日志
- 分析错误模式
- 监控性能指标

## 常见问题

### 编译问题
- 确保 Go 版本兼容
- 检查依赖版本
- 清理模块缓存：`go clean -modcache`

### 测试失败
- 检查文件权限
- 验证测试环境
- 查看详细错误信息

### 性能问题
- 使用性能分析工具
- 检查内存泄漏
- 优化热点代码

## 参考资源

- [Go 官方文档](https://golang.org/doc/)
- [Go 代码审查指南](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go 测试最佳实践](https://golang.org/doc/tutorial/add-a-test)