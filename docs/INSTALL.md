# DelGuard 安装指南

本指南将帮助您在不同平台上安装和配置 DelGuard。

## 系统要求

### 最低要求
- **操作系统**: Windows 10+, Linux (Ubuntu 18.04+, CentOS 7+), macOS 10.14+
- **内存**: 512 MB RAM
- **存储**: 50 MB 可用磁盘空间
- **网络**: 下载安装包时需要互联网连接

### 推荐配置
- **操作系统**: Windows 11, Ubuntu 22.04 LTS, macOS 12+
- **内存**: 2 GB RAM
- **存储**: 200 MB 可用磁盘空间
- **处理器**: 双核 2.0 GHz 或更高

## 安装方法

### 方法 1: 预编译二进制文件（推荐）

#### Windows

1. **下载安装包**
   - 访问 [Releases 页面](https://github.com/your-username/delguard/releases)
   - 下载 `delguard-windows-amd64.zip`

2. **解压安装**
   ```powershell
   # 解压到程序目录
   Expand-Archive -Path delguard-windows-amd64.zip -DestinationPath "C:\Program Files\DelGuard"
   
   # 添加到系统PATH
   $env:PATH += ";C:\Program Files\DelGuard"
   ```

3. **验证安装**
   ```powershell
   delguard --version
   ```

#### Linux

1. **下载安装包**
   ```bash
   # Ubuntu/Debian
   wget https://github.com/your-username/delguard/releases/latest/download/delguard-linux-amd64.tar.gz
   
   # CentOS/RHEL
   curl -L -O https://github.com/your-username/delguard/releases/latest/download/delguard-linux-amd64.tar.gz
   ```

2. **解压安装**
   ```bash
   # 解压
   tar -xzf delguard-linux-amd64.tar.gz
   
   # 移动到系统目录
   sudo mv delguard /usr/local/bin/
   
   # 设置执行权限
   sudo chmod +x /usr/local/bin/delguard
   ```

3. **验证安装**
   ```bash
   delguard --version
   ```

#### macOS

1. **使用 Homebrew（推荐）**
   ```bash
   # 添加tap
   brew tap your-username/delguard
   
   # 安装
   brew install delguard
   ```

2. **手动安装**
   ```bash
   # 下载
   curl -L -O https://github.com/your-username/delguard/releases/latest/download/delguard-darwin-amd64.tar.gz
   
   # 解压
   tar -xzf delguard-darwin-amd64.tar.gz
   
   # 移动到系统目录
   sudo mv delguard /usr/local/bin/
   
   # 设置执行权限
   sudo chmod +x /usr/local/bin/delguard
   ```

3. **验证安装**
   ```bash
   delguard --version
   ```

### 方法 2: 从源码构建

#### 前置要求
- Go 1.21 或更高版本
- Git

#### 构建步骤

1. **克隆仓库**
   ```bash
   git clone https://github.com/your-username/delguard.git
   cd delguard
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **构建**
   ```bash
   # 开发构建
   go build -o delguard ./cmd/delguard
   
   # 发布构建（优化）
   go build -ldflags "-s -w" -o delguard ./cmd/delguard
   ```

4. **安装到系统**
   ```bash
   # Linux/macOS
   sudo mv delguard /usr/local/bin/
   
   # Windows (以管理员身份运行)
   move delguard.exe "C:\Program Files\DelGuard\"
   ```

### 方法 3: 包管理器安装

#### Windows (Chocolatey)
```powershell
# 安装 Chocolatey（如果未安装）
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# 安装 DelGuard
choco install delguard
```

#### Windows (Scoop)
```powershell
# 安装 Scoop（如果未安装）
iwr -useb get.scoop.sh | iex

# 添加bucket
scoop bucket add delguard https://github.com/your-username/scoop-delguard.git

# 安装
scoop install delguard
```

#### Linux (APT - Ubuntu/Debian)
```bash
# 添加GPG密钥
curl -fsSL https://delguard.example.com/gpg | sudo apt-key add -

# 添加仓库
echo "deb https://delguard.example.com/apt stable main" | sudo tee /etc/apt/sources.list.d/delguard.list

# 更新并安装
sudo apt update
sudo apt install delguard
```

#### Linux (YUM - CentOS/RHEL)
```bash
# 添加仓库
sudo tee /etc/yum.repos.d/delguard.repo <<EOF
[delguard]
name=DelGuard Repository
baseurl=https://delguard.example.com/yum
enabled=1
gpgcheck=1
gpgkey=https://delguard.example.com/gpg
EOF

# 安装
sudo yum install delguard
```

## 配置

### 初始配置

1. **创建配置目录**
   ```bash
   # Linux/macOS
   mkdir -p ~/.delguard
   
   # Windows
   mkdir "%USERPROFILE%\.delguard"
   ```

2. **生成默认配置**
   ```bash
   delguard config init
   ```

3. **编辑配置文件**
   ```bash
   # Linux/macOS
   nano ~/.delguard/config.json
   
   # Windows
   notepad "%USERPROFILE%\.delguard\config.json"
   ```

### 配置选项

#### 基本配置
```json
{
  "language": "zh-cn",
  "max_file_size": 1073741824,
  "enable_recycle_bin": true,
  "enable_logging": true,
  "log_level": "info"
}
```

#### 高级配置
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
    "enable_system_protection": true,
    "protected_paths": [
      "C:\\Windows",
      "C:\\Program Files",
      "/bin",
      "/sbin",
      "/usr/bin",
      "/usr/sbin"
    ]
  },
  "performance": {
    "enable_performance_monitoring": true,
    "enable_memory_optimization": true,
    "gc_percent": 100,
    "memory_limit_mb": 1024,
    "max_concurrent_operations": 10
  },
  "ui": {
    "enable_colors": true,
    "enable_progress_bar": true,
    "confirmation_timeout": 30
  }
}
```

## 环境变量

DelGuard 支持以下环境变量：

```bash
# 配置文件路径
export DELGUARD_CONFIG="/path/to/config.json"

# 日志级别
export DELGUARD_LOG_LEVEL="debug"

# 禁用颜色输出
export DELGUARD_NO_COLOR="1"

# 启用详细输出
export DELGUARD_VERBOSE="1"
```

## 权限配置

### Linux/macOS

1. **设置sudo权限（可选）**
   ```bash
   # 编辑sudoers文件
   sudo visudo
   
   # 添加以下行（替换username为实际用户名）
   username ALL=(ALL) NOPASSWD: /usr/local/bin/delguard
   ```

2. **设置文件权限**
   ```bash
   # 确保用户有权限访问要删除的文件
   sudo chown -R $USER:$USER /path/to/files
   ```

### Windows

1. **以管理员身份运行**
   - 右键点击命令提示符或PowerShell
   - 选择"以管理员身份运行"

2. **设置执行策略**
   ```powershell
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

## 验证安装

### 基本功能测试

1. **版本检查**
   ```bash
   delguard --version
   ```

2. **帮助信息**
   ```bash
   delguard --help
   ```

3. **配置检查**
   ```bash
   delguard config show
   ```

4. **创建测试文件并删除**
   ```bash
   # 创建测试文件
   echo "test content" > test_file.txt
   
   # 安全删除
   delguard delete --safe test_file.txt
   
   # 验证文件已删除
   ls test_file.txt  # 应该显示文件不存在
   ```

### 性能测试

```bash
# 运行内置基准测试
delguard benchmark --files 1000 --size 1MB
```

## 故障排除

### 常见问题

#### 1. 命令未找到
```bash
# 错误: delguard: command not found

# 解决方案: 检查PATH环境变量
echo $PATH  # Linux/macOS
echo $env:PATH  # Windows PowerShell

# 手动添加到PATH
export PATH=$PATH:/usr/local/bin  # Linux/macOS
$env:PATH += ";C:\Program Files\DelGuard"  # Windows
```

#### 2. 权限被拒绝
```bash
# 错误: Permission denied

# 解决方案: 检查文件权限
ls -la /usr/local/bin/delguard  # Linux/macOS

# 设置执行权限
sudo chmod +x /usr/local/bin/delguard
```

#### 3. 配置文件错误
```bash
# 错误: Invalid configuration file

# 解决方案: 验证JSON格式
delguard config validate

# 重置配置
delguard config reset
```

#### 4. 回收站功能不工作
```bash
# Windows: 检查回收站服务
sc query "Recycle Bin"

# Linux: 检查trash目录
ls -la ~/.local/share/Trash/
```

### 日志分析

1. **启用详细日志**
   ```bash
   delguard --verbose delete file.txt
   ```

2. **查看日志文件**
   ```bash
   # Linux/macOS
   tail -f ~/.delguard/logs/delguard.log
   
   # Windows
   Get-Content -Tail 10 -Wait "$env:USERPROFILE\.delguard\logs\delguard.log"
   ```

### 获取帮助

如果遇到问题，可以通过以下方式获取帮助：

1. **查看文档**: [https://delguard.example.com/docs](https://delguard.example.com/docs)
2. **提交Issue**: [https://github.com/your-username/delguard/issues](https://github.com/your-username/delguard/issues)
3. **社区讨论**: [https://github.com/your-username/delguard/discussions](https://github.com/your-username/delguard/discussions)

## 卸载

### 完全卸载

#### Windows
```powershell
# 删除程序文件
Remove-Item -Recurse -Force "C:\Program Files\DelGuard"

# 删除配置文件
Remove-Item -Recurse -Force "$env:USERPROFILE\.delguard"

# 从PATH中移除（如果手动添加过）
# 需要手动编辑系统环境变量
```

#### Linux/macOS
```bash
# 删除程序文件
sudo rm -f /usr/local/bin/delguard

# 删除配置文件
rm -rf ~/.delguard

# 如果使用包管理器安装，使用对应的卸载命令
# Ubuntu/Debian: sudo apt remove delguard
# CentOS/RHEL: sudo yum remove delguard
# macOS Homebrew: brew uninstall delguard
```

---

安装完成后，您就可以开始使用 DelGuard 了！请查看 [README.md](../README.md) 了解基本使用方法。