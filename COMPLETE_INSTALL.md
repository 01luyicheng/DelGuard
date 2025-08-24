# 完成安装：系统级cp命令

## ✅ 当前状态
- ✅ 已创建用户bin目录：`%USERPROFILE%\bin`
- ✅ 已复制delguard.exe到用户bin目录
- ✅ 已创建cp.bat、del.bat、rm.bat批处理文件

## 🎯 最后一步：添加到PATH

### 方法1：图形界面（推荐）
1. **Win + R** → 输入 `sysdm.cpl` → 回车
2. 点击 **"环境变量"**
3. 在 **"用户变量"** 中找到 **PATH**
4. 点击 **"编辑"**
5. 点击 **"新建"**
6. 输入：`C:\Users\21601\bin`
7. 点击 **"确定"** 保存所有窗口
8. **重启命令提示符或PowerShell**

### 方法2：命令行
在 **新的管理员命令提示符** 中运行：
```cmd
setx PATH "%USERPROFILE%\bin;%PATH%"
```

## 🧪 验证安装

重启终端后，运行以下测试：

```cmd
:: 测试cp命令
where cp
cp --help

:: 创建测试文件
echo Hello World > test.txt
cp test.txt backup.txt

:: 验证文件复制
type backup.txt

:: 测试del/rm命令
del test.txt
rm backup.txt
```

## 📝 已安装的文件位置
```
%USERPROFILE%\bin\
├── delguard.exe    (主程序)
├── cp.bat         (cp命令)
├── del.bat        (del命令)
└── rm.bat         (rm命令)
```

## 🚀 使用方法

现在你可以在任何位置使用：
```cmd
cp source.txt dest.txt        # 安全复制
cp -r source_dir dest_dir     # 递归复制目录
cp -f file.txt new.txt        # 强制覆盖
cp -i file.txt new.txt        # 交互模式
```

## ❗ 如果仍然无效

1. **检查PATH**：
   ```cmd
   echo %PATH%
   ```

2. **手动验证**：
   ```cmd
   "%USERPROFILE%\bin\cp.bat" --help
   ```

3. **重新启动**：完全关闭所有终端窗口，重新打开

## ✅ 安装完成

完成以上步骤后，你的DelGuard cp命令将在系统任何位置可用！