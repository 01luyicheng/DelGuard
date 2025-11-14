# 创建GitHub Release的PowerShell脚本
# 需要先安装GitHub CLI (gh) 或使用GitHub API

param(
    [string]$Token = "",
    [string]$Tag = "v1.0.0",
    [string]$Title = "DelGuard 1.0.0",
    [string]$Description = "首次发布DelGuard - 安全的Windows文件删除工具",
    [bool]$Prerelease = $false,
    [bool]$Draft = $false
)

$ReleaseNotes = @"
# DelGuard 1.0.0 发布说明

## 🎉 首次发布

DelGuard 是一个强大的 Windows 文件删除工具，用于安全的删除文件和文件夹，将它们移动到回收站而不是直接彻底删除。

## ✨ 核心功能

- **安全删除**：将文件和文件夹移动到回收站，而不是永久删除
- **强制删除**：绕过系统限制删除被占用的文件
- **预览模式**：先显示将要删除的文件，确认后再执行删除
- **递归删除**：支持删除文件夹及其所有子内容
- **静默模式**：无提示删除，适合批处理脚本
- **详细日志**：显示详细的删除过程和结果
- **通配符支持**：支持使用通配符批量删除文件

## 📦 安装方式

### 方法一：使用安装脚本
1. 下载 DelGuard 发布包
2. 以管理员身份运行 `install.bat`
3. 按照提示完成安装
4. 安装完成后，可在命令行直接使用 `delguard` 命令

### 方法二：手动安装
1. 下载 `delguard.exe` 可执行文件
2. 将文件放置在合适的目录（如 `C:\Program Files\DelGuard`）
3. 手动将目录添加到系统 PATH 环境变量

## 🔧 系统要求

- 操作系统：Windows 10/11 (x64)
- 运行时：.NET 8.0 或更高版本
- 权限：普通用户权限（某些操作需要管理员权限）

## 🔒 安全说明

1. **谨慎使用强制删除**：强制删除可能会影响正在运行的程序
2. **系统目录保护**：DelGuard 会阻止删除关键的系统目录
3. **预览模式建议**：对于不熟悉的删除操作，建议先使用预览模式

---

**注意**：使用DelGuard删除文件前，请确保您真的需要删除这些文件。删除操作不可撤销，重要文件请先备份。
"@

Write-Host "创建GitHub Release..." -ForegroundColor Cyan
Write-Host "标签: $Tag" -ForegroundColor White
Write-Host "标题: $Title" -ForegroundColor White

# 尝试使用GitHub CLI (gh)
try {
    # 检查gh是否安装
    $ghVersion = gh --version 2>$null
    if ($ghVersion) {
        Write-Host "使用GitHub CLI创建Release..." -ForegroundColor Green
        
        # 创建发布
        $releaseArgs = @(
            "release", "create", $Tag,
            "--title", $Title,
            "--notes", $ReleaseNotes
        )
        
        if ($Prerelease) {
            $releaseArgs += "--prerelease"
        }
        
        if ($Draft) {
            $releaseArgs += "--draft"
        }
        
        # 添加文件
        $releaseArgs += "installer/output/delguard.exe"
        $releaseArgs += "installer/output/install.bat"
        $releaseArgs += "installer/output/uninstall.bat"
        $releaseArgs += "installer/output/README.md"
        $releaseArgs += "installer/output/LICENSE"
        $releaseArgs += "installer/output/windows.json"
        
        & gh @releaseArgs
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Release创建成功！" -ForegroundColor Green
            Write-Host "访问: https://github.com/01luyicheng/DelGuard/releases/tag/$Tag" -ForegroundColor Cyan
        } else {
            Write-Error "创建Release失败"
        }
    } else {
        throw "GitHub CLI未安装"
    }
}
catch {
    Write-Warning "无法使用GitHub CLI，请手动创建Release或使用脚本"
    Write-Host "手动创建步骤:" -ForegroundColor Yellow
    Write-Host "1. 访问 https://github.com/01luyicheng/DelGuard/releases/new" -ForegroundColor White
    Write-Host "2. 输入标签: $Tag" -ForegroundColor White
    Write-Host "3. 输入标题: $Title" -ForegroundColor White
    Write-Host "4. 复制以下内容作为描述:" -ForegroundColor White
    Write-Host $ReleaseNotes -ForegroundColor Gray
    Write-Host "5. 上传以下文件:" -ForegroundColor White
    Write-Host "  - installer/output/delguard.exe" -ForegroundColor Gray
    Write-Host "  - installer/output/install.bat" -ForegroundColor Gray
    Write-Host "  - installer/output/uninstall.bat" -ForegroundColor Gray
    Write-Host "  - installer/output/README.md" -ForegroundColor Gray
    Write-Host "  - installer/output/LICENSE" -ForegroundColor Gray
    Write-Host "  - installer/output/windows.json" -ForegroundColor Gray
}