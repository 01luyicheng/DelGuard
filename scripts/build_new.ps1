# DelGuard 新架构构建脚本

param(
    [string]$Target = "windows",
    [string]$Arch = "amd64",
    [switch]$Release,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

# 项目信息
$ProjectName = "delguard"
$Version = "2.0.0"
$BuildDir = "build"
$CmdDir = "cmd/delguard"

# 颜色输出函数
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

Write-ColorOutput "=== DelGuard 构建脚本 v2.0 ===" "Cyan"
Write-ColorOutput "目标平台: $Target/$Arch" "Yellow"

# 创建构建目录
if (!(Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
}

# 设置构建参数
$env:GOOS = $Target
$env:GOARCH = $Arch

# 构建标志
$BuildFlags = @()
if ($Release) {
    $BuildFlags += "-ldflags", "-s -w -X main.Version=$Version"
    Write-ColorOutput "发布模式构建" "Green"
} else {
    Write-ColorOutput "调试模式构建" "Yellow"
}

# 输出文件名
$OutputName = "$ProjectName"
if ($Target -eq "windows") {
    $OutputName += ".exe"
}

$OutputPath = Join-Path $BuildDir $OutputName

try {
    Write-ColorOutput "正在构建..." "Yellow"
    
    # 执行构建
    if ($Verbose) {
        & go build $BuildFlags -v -o $OutputPath "./$CmdDir"
    } else {
        & go build $BuildFlags -o $OutputPath "./$CmdDir"
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✅ 构建成功!" "Green"
        Write-ColorOutput "输出文件: $OutputPath" "Cyan"
        
        # 显示文件信息
        $FileInfo = Get-Item $OutputPath
        Write-ColorOutput "文件大小: $([math]::Round($FileInfo.Length / 1MB, 2)) MB" "White"
        
        # 如果是发布模式，创建压缩包
        if ($Release) {
            $ZipName = "$ProjectName-$Version-$Target-$Arch.zip"
            $ZipPath = Join-Path $BuildDir $ZipName
            
            Write-ColorOutput "创建发布包..." "Yellow"
            Compress-Archive -Path $OutputPath -DestinationPath $ZipPath -Force
            Write-ColorOutput "发布包: $ZipPath" "Green"
        }
        
    } else {
        Write-ColorOutput "❌ 构建失败!" "Red"
        exit 1
    }
    
} catch {
    Write-ColorOutput "❌ 构建过程中发生错误: $($_.Exception.Message)" "Red"
    exit 1
}

Write-ColorOutput "=== 构建完成 ===" "Cyan"