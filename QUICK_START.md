# DelGuard v1.4.1 快速开始

## 🚀 一行命令安装

### Windows
```powershell
# 一行命令安装（最简单）
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install-oneline.ps1' -UseBasicParsing | Invoke-Expression }"

# 或者使用完整脚本（可自定义参数）
powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.ps1' -OutFile 'quick-install.ps1'; .\quick-install.ps1 }"
```

### Linux/macOS
```bash
# 一行命令安装（最简单）
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/install-oneline.sh | sudo bash

# 或者使用完整脚本（可自定义参数）
curl -fsSL https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash

# 备用wget命令
wget -qO- https://raw.githubusercontent.com/01luyicheng/DelGuard/main/scripts/quick-install.sh | sudo bash
```

## ✅ 安装验证

安装完成后，运行以下命令验证：

```bash
# 查看版本信息
delguard --version
# 应该显示：delguard version 1.4.1

# 查看系统状态
delguard status
# 应该显示系统信息和回收站状态

# 查看帮助
delguard --help
```

## 📖 基本使用

### 1. 安全删除文件
```bash
# 删除文件（移动到回收站）
rm important_file.txt      # Linux/macOS
del important_file.txt     # Windows

# 删除目录
rm -r my_folder/         # Linux/macOS
rmdir my_folder          # Windows

# 删除多个文件
rm file1.txt file2.txt
```

### 2. 查看回收站
```bash
# 查看回收站内容
delguard list

# 详细查看
delguard list -l

# 按时间排序
delguard list --sort=time

# 按大小排序
delguard list --sort=size

# 限制显示数量
delguard list --limit=10
```

### 3. 恢复文件
```bash
# 按名称恢复
delguard restore important_file.txt

# 按索引恢复（查看list中的索引号）
delguard restore --index 1

# 恢复到指定位置
delguard restore important_file.txt --target /path/to/restore/

# 批量恢复
delguard restore --all --filter="*.txt"
```

### 4. 清空回收站
```bash
# 清空回收站
delguard empty

# 清空前确认
delguard empty --confirm
```

## 🔧 高级功能

### 预览删除
```bash
# 预览将要删除的文件（不实际删除）
delguard delete -n *.log
```

### 强制删除
```bash
# 跳过确认直接删除
delguard delete -f large_file.zip
```

### 交互式删除
```bash
# 逐个确认删除
delguard delete -i *.tmp
```

## 🛠️ 配置管理

### 查看配置
```bash
# 查看当前配置
delguard config

# 编辑配置文件
# 配置文件位置：
# Windows: %USERPROFILE%\.delguard\config.yaml
# Linux/macOS: ~/.delguard/config.yaml
```

### 示例配置
```yaml
# ~/.delguard/config.yaml
trash:
  auto_empty_days: 30    # 30天后自动清理
  confirm_delete: true # 删除前确认
  
display:
  color: true          # 彩色输出
  unicode: true        # Unicode图标
  
logging:
  level: info          # 日志级别
  file: ~/.delguard/delguard.log
```

## 🗑️ 卸载

### Windows
```powershell
# 运行卸载脚本
delguard-uninstall

# 或者
c:\Program Files\DelGuard\uninstall.bat
```

### Linux/macOS
```bash
# 运行卸载脚本
sudo delguard-uninstall
```

## 🐛 常见问题

### 安装问题
**Q: 安装失败怎么办？**
A: 检查网络连接，确保有管理员权限，查看错误提示。

**Q: 安装后命令不可用？**
A: 重启终端或运行 `refreshenv` (Windows) / `source ~/.bashrc` (Linux/macOS)。

### 使用问题
**Q: 删除的文件在哪里？**
A: 文件被移动到系统回收站：
- Windows: 回收站
- macOS: 废纸篓
- Linux: ~/.local/share/Trash

**Q: 如何永久删除文件？**
A: 使用 `--permanent` 参数：
```bash
delguard delete --permanent file.txt
```

**Q: 如何恢复误删的文件？**
A: 使用恢复命令：
```bash
delguard list                    # 查看回收站
delguard restore filename.txt    # 恢复文件
```

## 📞 获取帮助

### 文档资源
- 📖 [完整文档](README.md)
- 🔧 [安装指南](INSTALL.md)
- 🐛 [问题反馈](https://github.com/01luyicheng/DelGuard/issues)
- 💬 [GitHub Discussions](https://github.com/01luyicheng/DelGuard/discussions)
- 📧 邮件支持：等待设置

## 🎯 下一步

1. **立即安装**：使用上方的一行命令安装
2. **测试功能**：删除和恢复几个测试文件
3. **配置优化**：根据需求调整配置文件
4. **日常使用**：开始在日常工作中使用DelGuard保护您的文件

---

**享受安全删除的便捷体验！** 🛡️✨