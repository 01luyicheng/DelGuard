# DelGuard v1.1.0 发布说明

## 🎉 新版本发布

我们很高兴地宣布 DelGuard v1.1.0 正式发布！这个版本带来了全新的智能提示系统和增强的错误处理功能。

## ✨ 新功能亮点

### 🔔 智能提示系统
- **删除提示**: 每次删除文件后都会显示友好的提示信息
  ```
  DelGuard: [文件名]已被移动到回收站
  ```

- **覆盖保护**: 在覆盖文件前自动备份并显示提示
  ```
  DelGuard: [文件名] 原文件已备份到回收站，准备覆盖
  ```

- **目录删除**: 删除目录时显示专门的提示信息
  ```
  DelGuard: [目录名] 目录已被移动到回收站
  ```

### 🛡️ 智能错误处理
我们新增了详细的错误处理机制，为每种错误类型提供具体的解决建议。

## 📦 下载

### Windows
- `DelGuard-v1.1.0.exe` - Windows 64位可执行文件

### 从源码编译
```bash
git clone https://github.com/01luyicheng/DelGuard.git
cd DelGuard
go build -o DelGuard.exe .
```

## 🚀 快速开始

1. **下载** 适合您系统的可执行文件
2. **安装** 将文件放置在系统PATH中
3. **使用** 运行 `DelGuard 文件名` 开始安全删除

## 📖 使用示例

### 基本删除
```bash
DelGuard document.txt
# 输出: DelGuard: document.txt已被移动到回收站
```

### 批量删除
```bash
DelGuard *.log
# 输出: DelGuard: 成功删除3个文件到回收站
```

### 覆盖保护
```bash
DelGuard newfile.txt oldfile.txt
# 输出: DelGuard: oldfile.txt 原文件已备份到回收站，准备覆盖
```

## 📄 许可证

本项目采用 MIT 许可证 - 详见 LICENSE 文件。