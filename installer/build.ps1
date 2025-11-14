<#
.SYNOPSIS
    DelGuard安装程序构建脚本（PowerShell版本）

.DESCRIPTION
    自动化构建DelGuard MSI安装包的PowerShell脚本

.EXAMPLE
    .\build.ps1
#>

# 设置变量
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$InstallerDir = $PSScriptRoot
$OutputDir = Join-Path $InstallerDir "output"
$Configuration = "Release"
$WixPath = "C:\Program Files (x86)\WiX Toolset v3.11\bin"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "DelGuard 安装程序构建脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 检查WiX工具集
if (-not (Test-Path "$WixPath\candle.exe")) {
    Write-Error "错误：未找到WiX工具集，请先安装WiX Toolset v3.11"
    Write-Host "下载地址：https://wixtoolset.org/releases/" -ForegroundColor Yellow
    exit 1
}

# 创建输出目录
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# 设置WiX环境
$env:PATH = "$WixPath;$env:PATH"

Write-Host "`n1. 构建主程序..." -ForegroundColor Green
Set-Location $ProjectRoot

try {
    dotnet restore
    if ($LASTEXITCODE -ne 0) { throw "主程序还原失败" }
    
    dotnet build -c $Configuration
    if ($LASTEXITCODE -ne 0) { throw "主程序构建失败" }
    
    dotnet publish -c $Configuration -r win-x64 --self-contained false
    if ($LASTEXITCODE -ne 0) { throw "主程序发布失败" }
    
    Write-Host "主程序构建完成" -ForegroundColor Green
}
catch {
    Write-Error "错误：$($_.Exception.Message)"
    exit 1
}

Write-Host "`n2. 构建自定义操作..." -ForegroundColor Green
Set-Location $InstallerDir

try {
    dotnet build DelGuard.CustomActions.csproj -c $Configuration
    if ($LASTEXITCODE -ne 0) { throw "自定义操作构建失败" }
    
    Write-Host "自定义操作构建完成" -ForegroundColor Green
}
catch {
    Write-Error "错误：$($_.Exception.Message)"
    exit 1
}

Write-Host "`n3. 编译WiX源文件..." -ForegroundColor Green
try {
    & candle.exe -nologo -out "$OutputDir\" -arch x64 Product.wxs
    if ($LASTEXITCODE -ne 0) { throw "WiX编译失败" }
    
    Write-Host "WiX源文件编译完成" -ForegroundColor Green
}
catch {
    Write-Error "错误：$($_.Exception.Message)"
    exit 1
}

Write-Host "`n4. 链接生成MSI..." -ForegroundColor Green
try {
    & light.exe -nologo -out "$OutputDir\DelGuard.msi" "$OutputDir\Product.wixobj" -ext WixUIExtension -ext WixUtilExtension
    if ($LASTEXITCODE -ne 0) { throw "WiX链接失败" }
    
    Write-Host "MSI安装包生成完成" -ForegroundColor Green
}
catch {
    Write-Error "错误：$($_.Exception.Message)"
    exit 1
}

Write-Host "`n5. 验证安装包..." -ForegroundColor Green
try {
    # 检查MSI文件是否存在
    $msiPath = "$OutputDir\DelGuard.msi"
    if (-not (Test-Path $msiPath)) {
        throw "MSI文件未生成"
    }
    
    # 显示文件信息
    $msiInfo = Get-Item $msiPath
    Write-Host "安装包验证通过" -ForegroundColor Green
    Write-Host "文件大小: $($msiInfo.Length) 字节" -ForegroundColor Green
}
catch {
    Write-Error "错误：$($_.Exception.Message)"
    exit 1
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "构建完成！" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "安装包位置: $OutputDir\DelGuard.msi" -ForegroundColor Cyan

$msiInfo = Get-Item "$OutputDir\DelGuard.msi"
Write-Host "文件大小: $($msiInfo.Length) 字节" -ForegroundColor Cyan

Write-Host "`n安装说明:" -ForegroundColor Yellow
Write-Host "  1. 双击 DelGuard.msi 运行安装程序" -ForegroundColor White
Write-Host "  2. 按照安装向导完成安装" -ForegroundColor White
Write-Host "  3. 安装完成后可在命令行使用 delguard 命令" -ForegroundColor White

Write-Host "`n卸载说明:" -ForegroundColor Yellow
Write-Host "  - 控制面板 -> 程序和功能 -> 找到 DelGuard -> 卸载" -ForegroundColor White
Write-Host "  - 或使用命令：msiexec /x {产品代码}" -ForegroundColor White

Write-Host "`n构建脚本执行完毕！" -ForegroundColor Green

# 返回项目根目录
Set-Location $ProjectRoot

exit 0