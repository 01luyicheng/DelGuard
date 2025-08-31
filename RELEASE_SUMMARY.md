# DelGuard v1.6.1 发布摘要

DelGuard v1.6.1 是DelGuard的正式稳定版本，标志着项目进入成熟阶段。本版本在v1.5.9的基础上完成了全面的代码质量检查和发布准备，确保为用户提供最稳定可靠的安全删除体验。

## 🚀 版本亮点

### 🎉 里程碑版本
- **正式版发布**：v1.6.1作为DelGuard的正式稳定版本发布
- **生产就绪**：经过全面测试，可直接用于生产环境
- **向后兼容**：保持与之前版本的完全兼容性

### ✅ 质量保证
- **零缺陷发布**：完成全面的代码审查和质量检查
- **跨平台验证**：在Windows、Linux、macOS平台完整测试
- **功能完整性**：所有核心功能验证正常工作
- **安全审计**：通过安全性检查，无已知漏洞

### 🔧 技术改进
- **版本统一**：所有配置文件和脚本版本号统一为v1.6.1
- **构建优化**：更新所有构建脚本和CI/CD配置
- **安装体验**：优化安装脚本，提供更流畅的安装体验

## 📥 安装方法

### Windows (PowerShell) - 一行命令安装
```powershell
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; .\quick-install.ps1 }"
```

### Linux/macOS (Bash) - 一行命令安装
```bash
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash
```

### 手动安装
1. 从GitHub Releases下载对应平台的二进制文件
2. 解压到任意目录
3. 运行 `delguard install` 完成系统集成

## 📊 版本信息
- **版本号**：v1.6.1
- **发布日期**：2024-12-19
- **支持平台**：Windows 10/11、Linux (x64/ARM64/ARM)、macOS (Intel/Apple Silicon)
- **Go版本**：1.21+
- **许可证**：MIT

## 🎯 核心功能
- **安全删除**：将文件移动到回收站而非永久删除
- **跨平台支持**：Windows、Linux、macOS原生支持
- **系统集成**：可替换系统rm/del命令
- **批量操作**：支持通配符和批量处理
- **完整恢复**：支持从回收站恢复文件
- **元数据记录**：记录删除信息，支持可靠恢复

## 🔒 安全特性
- **权限验证**：安装和危险操作需要管理员权限
- **路径保护**：防止删除系统关键文件
- **二次确认**：危险操作前提供确认提示
- **备份机制**：自动备份原始系统命令
- **详细日志**：完整的操作记录和审计跟踪