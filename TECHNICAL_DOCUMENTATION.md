# DelGuard 技术文档（精简版）

本技术文档已更新为简洁版本，详细技术指南请参考 [TECHNICAL_GUIDE.md](TECHNICAL_GUIDE.md)。

## 快速技术概览

### 核心实现
- **语言**: Go 1.19+
- **架构**: 跨平台命令行工具
- **功能**: 安全文件删除到系统回收站

### 平台技术方案
| 平台 | 技术实现 | 关键API |
|------|----------|---------|
| Windows | Win32 API | SHFileOperationW |
| macOS | AppleScript | osascript + Finder |
| Linux | freedesktop规范 | ~/.local/share/Trash/ |

### 安全机制
- 关键路径保护（防止删除系统文件）
- 权限检查
- 文件类型验证

### 项目结构
```
├── main.go      # 主程序入口
├── platform.go  # 平台分发
├── [platform].go # 各平台实现
├── restore.go   # 文件恢复
├── protect.go   # 安全保护
└── config.go    # 配置管理
```

### 快速开始
```bash
git clone https://github.com/01luyicheng/DelGuard.git
cd DelGuard
go build
```

完整技术细节请查看 [TECHNICAL_GUIDE.md](TECHNICAL_GUIDE.md) 文件。