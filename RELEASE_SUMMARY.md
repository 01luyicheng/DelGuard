# DelGuard v1.5.5 发布摘要

## 🎯 发布概述
DelGuard v1.5.5 是一个专注于系统兼容性和安全增强的维护版本，重点解决了Windows平台的路径验证硬编码问题，提升了跨平台兼容性。

## 🔧 主要改进

### 🛡️ 安全增强
- **动态系统路径检测**：完全移除了硬编码的C盘路径依赖，支持任意系统盘符配置
- **增强路径验证**：扩展了系统关键路径检测，覆盖更多Windows系统目录
- **跨平台兼容性**：优化了Unix-like系统的路径验证逻辑，支持macOS和Linux特有目录

### 🐛 关键修复
- **修复日志系统双重关闭**：解决了main.go中的重复日志关闭调用问题
- **改进临时文件清理**：优化了Windows VBS脚本和临时文件的清理机制
- **错误处理统一**：标准化了错误输出格式，减少对用户界面的干扰

### 🔧 代码质量提升
- **路径验证重构**：将所有硬编码路径改为动态检测，显著提高跨平台兼容性
- **系统检测增强**：根据实际系统环境动态调整保护路径列表
- **日志记录优化**：减少不必要的标准错误输出，使用更合适的日志级别

## 📊 兼容性验证
- ✅ **多系统盘测试**：验证在不同系统盘配置下的正常运行
- ✅ **跨平台验证**：确保在Windows、Linux、macOS上的路径验证准确
- ✅ **系统文件保护**：验证系统关键文件和目录的保护机制有效

## 🎯 升级建议
建议所有用户升级到v1.5.5版本，特别是：
- 使用非C盘作为系统盘的用户
- 需要跨平台部署的企业用户
- 对系统安全性有较高要求的用户

## 📦 获取方式
```bash
# Windows
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; .\quick-install.ps1 -Version v1.5.5 }"

# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash -s -- --version v1.5.5
```

## 📝 技术细节
- **版本号**：v1.5.5
- **发布日期**：2024-12-19
- **Go版本要求**：1.21+
- **支持平台**：Windows 10/11, macOS 10.15+, Ubuntu 18.04+, CentOS 7+