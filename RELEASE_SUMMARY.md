# DelGuard v1.5.3 发布总结

## 🚀 版本信息
- **版本号**: v1.5.3
- **发布日期**: 2024年12月19日
- **标签**: `v1.5.3`

## 🛡️ 主要改进

### 安全修复
1. **修复日志模块编译错误**
   - 修复了 `internal/logger/logger.go` 中的变量命名冲突
   - 解决了 `logFile` 变量名重复导致的编译错误

2. **增强PowerShell命令安全**
   - 改进了Windows平台的PowerShell命令执行方式
   - 使用更安全的参数传递机制，防止命令注入攻击

3. **加强路径验证**
   - 增强了对文件路径的验证逻辑
   - 防止潜在的路径遍历攻击

### Bug修复
- 修复了Windows平台下的编译错误
- 改进了错误处理的健壮性
- 优化了用户提示信息的清晰度

### 代码质量
- 移除了冗余代码和已弃用的方法
- 统一了错误消息的格式
- 增强了代码的可读性和维护性

## 📦 构建验证
- ✅ Windows (x64, ARM64) - 编译成功
- ✅ Linux (x64, ARM64, ARM) - 编译成功
- ✅ macOS (Intel, Apple Silicon) - 编译成功

## 🔧 安装方式

### Windows (PowerShell)
```powershell
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; .\quick-install.ps1 }"
```

### Linux/macOS (Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash
```

## 📋 系统支持
- **Windows**: Windows 10/11 (x64, ARM64)
- **Linux**: Ubuntu 18.04+, CentOS 7+, etc. (x64, ARM64, ARM)
- **macOS**: macOS 10.14+ (Intel, Apple Silicon)

## 🔗 快速开始
```bash
# 安全删除文件
rm file.txt  # 或 del file.txt (Windows)

# 查看回收站
delguard list

# 恢复文件
delguard restore file.txt

# 查看帮助
delguard --help
```

## 📖 文档更新
- 更新了CHANGELOG.md
- 完善了安全说明文档
- 优化了安装指南

## 🎯 后续计划
- 图形界面支持
- 网络同步功能
- 文件版本管理
- 更多平台支持

---

**发布地址**: https://github.com/01luyicheng/DelGuard/releases/tag/v1.5.3