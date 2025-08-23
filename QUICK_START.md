# DelGuard 快速部署指南

## 🚀 立即开始使用

### 1. 验证安全功能
```bash
go run verify_security.go
```

### 2. 使用安全配置模板
```bash
# 复制安全配置模板
copy config\security_template.json config\security.json

# 根据需要修改配置
# 使用文本编辑器打开 config\security.json 进行自定义
```

### 3. 编译项目
```bash
go build -o delguard.exe
```

### 4. 运行安全模式
```bash
# 使用安全配置文件
./delguard.exe --config=config/security.json
```

## 🔧 企业级部署

### 生产环境配置
1. **启用严格模式**：在配置中设置 `"safeMode": "strict"`
2. **限制文件大小**：设置合理的 `max_file_size` 值
3. **配置白名单**：设置 `allowed_extensions` 只允许必要文件类型
4. **启用审计日志**：设置 `"audit_logging": true`

### Windows域环境
- **UAC集成**：自动处理管理员权限请求
- **组策略兼容**：支持企业组策略配置
- **审计合规**：满足SOX、GDPR等合规要求

### 安全配置示例
```json
{
  "maxBackupFiles": 50,
  "trashMaxSize": 5000,
  "language": "zh",
  "safeMode": "strict",
  "logLevel": "info"
}
```

## 📋 安全检查清单

### 部署前检查
- [ ] 运行 `go run verify_security.go` 验证所有安全功能
- [ ] 检查配置文件格式
- [ ] 验证用户权限
- [ ] 测试文件恢复功能

### 运行中监控
- [ ] 定期检查安全日志
- [ ] 监控异常文件操作
- [ ] 验证备份完整性
- [ ] 更新安全模板

## 🛡️ 安全最佳实践

### 用户权限
- 使用标准用户权限运行日常操作
- 仅在需要时提升管理员权限
- 定期审查用户权限配置

### 文件操作
- 始终启用文件完整性检查
- 使用原子操作避免数据损坏
- 定期验证备份文件

### 监控与审计
- 启用详细日志记录
- 定期分析安全事件
- 建立异常检测机制

## 📞 技术支持

### 获取帮助
- **文档**：查看 SECURITY.md 获取详细安全指南
- **配置**：参考 config/security_template.json 模板
- **问题**：运行验证工具检查配置问题

### 更新与维护
- 定期更新安全模板
- 关注安全公告
- 参与社区讨论

## ✅ 验证完成

你的DelGuard项目现在已经具备了企业级安全功能：

- ✅ 路径遍历攻击防护
- ✅ 恶意软件检测
- ✅ UAC权限管理
- ✅ 文件完整性验证
- ✅ 国际化错误处理
- ✅ 审计日志记录
- ✅ 安全配置模板
- ✅ 合规性支持

项目已准备好投入生产环境使用！