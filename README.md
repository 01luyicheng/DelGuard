# DelGuard - 安全文件删除工具

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

DelGuard 是一个功能强大的跨平台安全文件删除工具，提供智能文件管理、安全删除和系统保护功能。

## ✨ 主要特性

### 🔒 安全删除
- **智能回收站支持** - 自动将文件移动到系统回收站
- **系统路径保护** - 防止误删重要系统文件
- **权限验证** - 删除前进行安全权限检查
- **批量操作** - 支持批量文件删除

### 🔍 智能搜索
- **模式匹配** - 支持通配符和正则表达式搜索
- **大小过滤** - 按文件大小范围查找文件
- **重复文件检测** - 基于MD5哈希的重复文件识别
- **递归搜索** - 深度目录结构搜索

### ⚡ 性能优化
- **内存管理** - 智能内存使用和垃圾回收优化
- **并发处理** - 多线程文件操作提升性能
- **进度监控** - 实时操作进度和性能指标

### 🌐 跨平台支持
- **Windows** - 完整的Windows API集成
- **Linux/Unix** - 原生Unix系统支持
- **macOS** - macOS系统优化

## 🚀 快速开始

### 安装

#### 从源码构建
```bash
git clone https://github.com/your-username/delguard.git
cd delguard
go build -o delguard ./cmd/delguard
```

#### 使用预编译二进制文件
从 [Releases](https://github.com/your-username/delguard/releases) 页面下载适合您系统的二进制文件。

### 基本使用

#### 安全删除文件
```bash
# 删除单个文件
delguard delete file.txt

# 批量删除文件
delguard delete file1.txt file2.txt file3.txt

# 安全删除（移动到回收站）
delguard delete --safe important.doc
```

#### 搜索文件
```bash
# 按模式搜索
delguard search --pattern "*.log" /var/log

# 按大小搜索
delguard search --size ">100MB" /home/user

# 查找重复文件
delguard search --duplicates /home/user/Documents
```

#### 配置管理
```bash
# 查看当前配置
delguard config show

# 设置配置项
delguard config set language zh-cn
delguard config set max_file_size 1073741824

# 重置配置
delguard config reset
```

## 📖 详细文档

### 命令行参数

#### 全局选项
- `--config <file>` - 指定配置文件路径
- `--verbose` - 启用详细输出
- `--help` - 显示帮助信息
- `--version` - 显示版本信息

#### delete 命令
```bash
delguard delete [选项] <文件路径...>

选项:
  --safe              移动到回收站而不是永久删除
  --force             强制删除，跳过确认
  --recursive         递归删除目录
  --batch             批量模式，从文件读取路径列表
```

#### search 命令
```bash
delguard search [选项] <搜索路径>

选项:
  --pattern <模式>    文件名模式匹配
  --size <大小>       按文件大小过滤
  --duplicates        查找重复文件
  --recursive         递归搜索子目录
  --output <格式>     输出格式 (text|json|csv)
```

### 配置文件

DelGuard 使用JSON格式的配置文件，默认位置：
- Windows: `%USERPROFILE%\.delguard\config.json`
- Linux/macOS: `~/.delguard/config.json`

#### 配置示例
```json
{
  "language": "zh-cn",
  "max_file_size": 1073741824,
  "max_backup_files": 10,
  "enable_recycle_bin": true,
  "enable_logging": true,
  "log_level": "info",
  "security": {
    "enable_path_validation": true,
    "enable_malware_detection": true,
    "enable_system_protection": true
  },
  "performance": {
    "enable_performance_monitoring": true,
    "enable_memory_optimization": true,
    "gc_percent": 100,
    "memory_limit_mb": 1024
  }
}
```

## 🔧 开发

### 项目结构
```
delguard/
├── cmd/delguard/           # 主程序入口
├── internal/               # 内部包
│   ├── core/              # 核心业务逻辑
│   │   ├── delete/        # 删除服务
│   │   ├── search/        # 搜索服务
│   │   └── restore/       # 恢复服务
│   ├── platform/          # 平台相关代码
│   │   ├── windows/       # Windows实现
│   │   ├── linux/         # Linux实现
│   │   └── common/        # 通用实现
│   ├── config/            # 配置管理
│   ├── monitor/           # 监控和指标
│   └── ui/                # 用户界面
├── pkg/delguard/          # 公共API
├── configs/               # 配置文件
├── docs/                  # 文档
├── scripts/               # 构建和部署脚本
└── tests/                 # 测试文件
```

### 构建

#### 开发构建
```bash
go build -o build/delguard ./cmd/delguard
```

#### 发布构建
```bash
# Windows
powershell -ExecutionPolicy Bypass -File scripts/build_new.ps1 -Release

# Linux/macOS
./scripts/build.sh --release
```

### 测试

#### 运行所有测试
```bash
# 使用脚本
powershell -ExecutionPolicy Bypass -File scripts/run_tests.ps1 -TestType all -Coverage

# 直接使用go test
go test ./... -v -cover
```

#### 性能测试
```bash
go test -bench=. -benchmem ./tests/benchmarks/
```

### 质量保证
```bash
# 运行质量检查
powershell -ExecutionPolicy Bypass -File scripts/qa_check.ps1

# 自动修复格式问题
powershell -ExecutionPolicy Bypass -File scripts/qa_check.ps1 -Fix
```

## 🤝 贡献

我们欢迎所有形式的贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

### 开发流程
1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

- 📖 [文档](docs/)
- 🐛 [问题报告](https://github.com/your-username/delguard/issues)
- 💬 [讨论区](https://github.com/your-username/delguard/discussions)

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者和用户！

---

**DelGuard** - 让文件删除更安全、更智能！