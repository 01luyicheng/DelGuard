# DelGuard 测试用例文档

## 概述
本文档描述了DelGuard项目的测试用例，用于验证所有功能是否正常工作。

## 编译项目
```bash
# 编译DelGuard
go build -o delguard.exe ./cmd/delguard  # Windows
go build -o delguard ./cmd/delguard      # Linux/macOS
```

## 1. 文件恢复功能测试

### 1.1 基本恢复测试
```bash
# 创建测试文件
echo "测试文件内容" > test_file.txt

# 删除文件到回收站
./delguard delete test_file.txt

# 列出回收站中的文件
./delguard restore --list

# 恢复文件
./delguard restore test_file.txt

# 验证文件是否恢复
cat test_file.txt
```

### 1.2 批量恢复测试
```bash
# 创建多个测试文件
for i in {1..5}; do echo "测试文件$i" > "test_file_$i.txt"; done

# 批量删除
./delguard delete test_file_*.txt

# 批量恢复
./delguard restore --pattern "test_file_*.txt"

# 验证所有文件都已恢复
ls test_file_*.txt
```

### 1.3 交互式恢复测试
```bash
# 创建测试文件
echo "交互测试文件" > interactive_test.txt

# 删除文件
./delguard delete interactive_test.txt

# 交互式恢复
./delguard restore --interactive

# 按提示选择要恢复的文件
```

## 2. 智能搜索功能测试

### 2.1 基本搜索测试
```bash
# 创建不同类型的测试文件
echo "文档内容" > document.txt
echo "图片描述" > image.jpg
echo "视频信息" > video.mp4

# 搜索所有文件
./delguard search "*"

# 搜索特定类型文件
./delguard search "*.txt"

# 搜索包含特定内容的文件
./delguard search --content "文档"
```

### 2.2 模糊搜索测试
```bash
# 创建相似名称的文件
echo "测试1" > test_document.txt
echo "测试2" > test_doc.txt
echo "测试3" > testing_file.txt

# 模糊搜索
./delguard search --fuzzy "test"

# 验证搜索结果包含相关文件
```

### 2.3 高级搜索测试
```bash
# 按大小搜索
./delguard search --size ">1KB"

# 按修改时间搜索
./delguard search --modified "today"

# 组合搜索条件
./delguard search --type "text" --size "<10KB" --modified "last week"
```

## 3. 配置文件格式支持测试

### 3.1 JSON配置测试
```bash
# 创建JSON配置文件
cat > config.json << EOF
{
  "language": "zh-CN",
  "max_files": 100,
  "interactive": true
}
EOF

# 使用JSON配置
./delguard --config config.json search "*"
```

### 3.2 YAML配置测试
```bash
# 创建YAML配置文件
cat > config.yaml << EOF
language: zh-CN
max_files: 100
interactive: true
features:
  - smart_search
  - auto_restore
EOF

# 使用YAML配置
./delguard --config config.yaml search "*"
```

### 3.3 TOML配置测试
```bash
# 创建TOML配置文件
cat > config.toml << EOF
language = "zh-CN"
max_files = 100
interactive = true

[features]
smart_search = true
auto_restore = true
EOF

# 使用TOML配置
./delguard --config config.toml search "*"
```

### 3.4 INI配置测试
```bash
# 创建INI配置文件
cat > config.ini << EOF
[general]
language=zh-CN
max_files=100
interactive=true

[features]
smart_search=true
auto_restore=true
EOF

# 使用INI配置
./delguard --config config.ini search "*"
```

## 4. 国际化功能测试

### 4.1 中文界面测试
```bash
# 设置中文环境
export LANG=zh_CN.UTF-8

# 测试中文错误消息
./delguard delete non_existent_file.txt

# 测试中文帮助信息
./delguard --help

# 测试中文交互提示
./delguard restore --interactive
```

### 4.2 英文界面测试
```bash
# 设置英文环境
export LANG=en_US.UTF-8

# 测试英文错误消息
./delguard delete non_existent_file.txt

# 测试英文帮助信息
./delguard --help
```

## 5. 跨平台功能测试

### 5.1 Windows平台测试
```powershell
# 测试Windows回收站功能
.\delguard.exe delete test_file.txt

# 验证文件在回收站中
.\delguard.exe restore --list

# 测试Windows路径处理
.\delguard.exe search "C:\Users\*\Documents\*.txt"
```

### 5.2 Linux平台测试
```bash
# 测试Linux回收站功能
./delguard delete test_file.txt

# 验证文件在~/.local/share/Trash中
./delguard restore --list

# 测试Linux权限处理
./delguard delete /tmp/test_file.txt
```

### 5.3 macOS平台测试
```bash
# 测试macOS废纸篓功能
./delguard delete test_file.txt

# 验证文件在~/.Trash中
./delguard restore --list

# 测试macOS特殊路径
./delguard search "~/Documents/*.txt"
```

## 6. 安装脚本测试

### 6.1 PowerShell安装脚本测试
```powershell
# 测试PowerShell安装脚本
.\scripts\install.ps1

# 验证中文显示
.\scripts\install.ps1 --help

# 测试强制安装
.\scripts\install.ps1 -Force

# 测试卸载
.\scripts\install.ps1 -Uninstall
```

### 6.2 Bash安装脚本测试
```bash
# 测试Bash安装脚本
./scripts/install.sh

# 验证中文显示
./scripts/install.sh --help

# 测试强制安装
./scripts/install.sh --force

# 测试卸载
./scripts/install.sh --uninstall
```

## 7. 性能测试

### 7.1 大文件处理测试
```bash
# 创建大文件
dd if=/dev/zero of=large_file.bin bs=1M count=100

# 测试大文件删除
time ./delguard delete large_file.bin

# 测试大文件恢复
time ./delguard restore large_file.bin
```

### 7.2 批量文件处理测试
```bash
# 创建大量小文件
for i in {1..1000}; do echo "文件$i" > "file_$i.txt"; done

# 测试批量删除性能
time ./delguard delete file_*.txt

# 测试批量恢复性能
time ./delguard restore --pattern "file_*.txt"
```

## 8. 错误处理测试

### 8.1 文件不存在错误测试
```bash
# 测试删除不存在的文件
./delguard delete non_existent_file.txt

# 验证错误消息是否正确显示
```

### 8.2 权限错误测试
```bash
# 创建只读文件
echo "只读文件" > readonly_file.txt
chmod 444 readonly_file.txt

# 测试删除只读文件
./delguard delete readonly_file.txt

# 验证权限错误处理
```

### 8.3 磁盘空间错误测试
```bash
# 在空间不足的情况下测试恢复
# (需要手动创建空间不足的环境)
./delguard restore large_file.bin
```

## 9. 边界条件测试

### 9.1 空文件名测试
```bash
# 测试空文件名处理
./delguard delete ""

# 验证错误处理
```

### 9.2 特殊字符文件名测试
```bash
# 创建包含特殊字符的文件
touch "文件 with spaces.txt"
touch "文件@#$%^&*().txt"
touch "文件中文名称.txt"

# 测试特殊字符文件处理
./delguard delete "文件 with spaces.txt"
./delguard delete "文件@#$%^&*().txt"
./delguard delete "文件中文名称.txt"

# 测试恢复
./delguard restore --list
./delguard restore "文件中文名称.txt"
```

### 9.3 长路径测试
```bash
# 创建深层目录结构
mkdir -p "very/deep/directory/structure/for/testing/long/paths"
echo "深层文件" > "very/deep/directory/structure/for/testing/long/paths/deep_file.txt"

# 测试长路径处理
./delguard delete "very/deep/directory/structure/for/testing/long/paths/deep_file.txt"
./delguard restore "deep_file.txt"
```

## 10. 集成测试

### 10.1 完整工作流测试
```bash
# 1. 创建测试环境
mkdir test_workspace
cd test_workspace

# 2. 创建各种类型的文件
echo "文档内容" > document.txt
echo "配置信息" > config.json
echo "脚本内容" > script.sh

# 3. 执行删除操作
../delguard delete *.txt *.json

# 4. 验证文件已删除
ls -la

# 5. 列出回收站内容
../delguard restore --list

# 6. 选择性恢复
../delguard restore document.txt

# 7. 批量恢复剩余文件
../delguard restore --all

# 8. 验证所有文件已恢复
ls -la

# 9. 清理测试环境
cd ..
rm -rf test_workspace
```

## 测试结果记录

### 测试环境
- 操作系统: [记录测试时的操作系统]
- Go版本: [记录Go版本]
- PowerShell版本: [记录PowerShell版本]
- 测试日期: [记录测试日期]

### 测试结果
- [ ] 文件恢复功能
- [ ] 智能搜索功能
- [ ] 配置文件格式支持
- [ ] 国际化功能
- [ ] 跨平台功能
- [ ] 安装脚本功能
- [ ] 性能测试
- [ ] 错误处理
- [ ] 边界条件测试
- [ ] 集成测试

### 发现的问题
[记录测试过程中发现的问题和解决方案]

### 改进建议
[记录测试后的改进建议]