# DelGuard 构建脚本 (PowerShell)
# 适用于 Windows 系统

param(
    [string]$Version = "1.0.0",
    [string]$Target = "all",
    [switch]$Debug = $false,
    [switch]$Test = $false,
    [switch]$Clean = $false,
    [switch]$Install = $false,
    [switch]$Uninstall = $false,
    [string]$OutputDir = "build"
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色输出
$colors = @{
    Red = "`e[31m"
    Green = "`e[32m"
    Yellow = "`e[33m"
    Blue = "`e[34m"
    Magenta = "`e[35m"
    Cyan = "`e[36m"
    Reset = "`e[0m"
}

function Write-Color {
    param([string]$Text, [string]$Color = "White")
    Write-Host "$($colors[$Color])$Text$($colors.Reset)"
}

# 版本信息
$BuildDate = Get-Date -Format "yyyy-MM-dd"
$GitCommit = if (Get-Command git -ErrorAction SilentlyContinue) {
    git rev-parse --short HEAD
} else { "unknown" }

# 构建参数
$LDFLAGS = @(
    "-s", "-w"
    "-X main.Version=$Version"
    "-X main.BuildDate=$BuildDate"
    "-X main.GitCommit=$GitCommit"
)

if ($Debug) {
    $LDFLAGS = @(
        "-X main.Version=$Version"
        "-X main.BuildDate=$BuildDate"
        "-X main.GitCommit=$GitCommit"
    )
}

# 目标平台
$targets = @(
    @{OS = "windows"; Arch = "amd64"; Suffix = ".exe"}
    @{OS = "windows"; Arch = "386"; Suffix = ".exe"}
    @{OS = "linux"; Arch = "amd64"; Suffix = ""}
    @{OS = "linux"; Arch = "386"; Suffix = ""}
    @{OS = "darwin"; Arch = "amd64"; Suffix = ""}
    @{OS = "darwin"; Arch = "arm64"; Suffix = ""}
)

function Test-GoEnvironment {
    Write-Color "检查Go环境..." "Cyan"
    
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Color "错误: 未找到Go编译器" "Red"
        Write-Color "请访问 https://golang.org/dl/ 安装Go" "Yellow"
        exit 1
    }
    
    $goVersion = go version
    Write-Color "Go版本: $goVersion" "Green"
    
    # 检查依赖
    Write-Color "检查依赖..." "Cyan"
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Color "错误: 依赖检查失败" "Red"
        exit 1
    }
    
    Write-Color "依赖检查完成" "Green"
}

function Invoke-SecurityTests {
    Write-Color "运行安全测试..." "Cyan"
    
    go test -v -run "TestSecuritySuite" .
    if ($LASTEXITCODE -ne 0) {
        Write-Color "警告: 安全测试失败" "Yellow"
    } else {
        Write-Color "安全测试通过" "Green"
    }
}

function Invoke-UnitTests {
    Write-Color "运行单元测试..." "Cyan"
    
    go test -v -race -coverprofile=coverage.out .
    if ($LASTEXITCODE -ne 0) {
        Write-Color "错误: 单元测试失败" "Red"
        exit 1
    }
    
    # 生成覆盖率报告
    go tool cover -html=coverage.out -o coverage.html
    Write-Color "测试覆盖率报告已生成: coverage.html" "Green"
    
    Write-Color "单元测试通过" "Green"
}

function Build-Target {
    param($OS, $Arch, $Suffix)
    
    $outputName = "delguard-$OS-$Arch$Suffix"
    $outputPath = Join-Path $OutputDir $outputName
    
    Write-Color "构建 $OS/$Arch..." "Cyan"
    
    $env:GOOS = $OS
    $env:GOARCH = $Arch
    $env:CGO_ENABLED = "0"
    
    $buildCmd = @(
        "go", "build"
        "-ldflags", ($LDFLAGS -join " ")
        "-o", $outputPath
        "."
    )
    
    & $buildCmd[0] $buildCmd[1..($buildCmd.Length-1)]
    
    if ($LASTEXITCODE -ne 0) {
        Write-Color "错误: 构建 $OS/$Arch 失败" "Red"
        return $false
    }
    
    # 计算文件哈希
    $hash = Get-FileHash -Path $outputPath -Algorithm SHA256
    $hash.Hash | Out-File "$outputPath.sha256" -Encoding ASCII
    
    Write-Color "构建成功: $outputPath (SHA256: $($hash.Hash))" "Green"
    return $true
}

function Build-All {
    Write-Color "开始构建所有目标..." "Cyan"
    
    # 创建输出目录
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir | Out-Null
    }
    
    $successCount = 0
    $totalCount = $targets.Count
    
    foreach ($target in $targets) {
        if (Build-Target -OS $target.OS -Arch $target.Arch -Suffix $target.Suffix) {
            $successCount++
        }
    }
    
    Write-Color "构建完成: $successCount/$totalCount 成功" "Green"
    
    if ($successCount -eq $totalCount) {
        # 创建压缩包
        Compress-Builds
    }
}

function Compress-Builds {
    Write-Color "创建压缩包..." "Cyan"
    
    try {
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        
        $zipPath = Join-Path $OutputDir "delguard-v$Version.zip"
        if (Test-Path $zipPath) {
            Remove-Item $zipPath -Force
        }
        
        [System.IO.Compression.ZipFile]::CreateFromDirectory($OutputDir, $zipPath)
        
        Write-Color "压缩包已创建: $zipPath" "Green"
    } catch {
        Write-Color "警告: 创建压缩包失败: $_" "Yellow"
    }
}

function Install-Local {
    Write-Color "安装到本地系统..." "Cyan"
    
    # 检查管理员权限
    $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
    
    if (-not $isAdmin) {
        Write-Color "警告: 需要管理员权限进行系统安装" "Yellow"
        Write-Color "请使用管理员权限重新运行脚本" "Yellow"
        return
    }
    
    # 构建Windows版本
    if (Build-Target -OS "windows" -Arch "amd64" -Suffix ".exe") {
        $exePath = Join-Path $OutputDir "delguard-windows-amd64.exe"
        $installDir = "$env:ProgramFiles\DelGuard"
        
        # 创建安装目录
        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir | Out-Null
        }
        
        # 复制可执行文件
        Copy-Item $exePath "$installDir\delguard.exe" -Force
        
        # 添加到PATH
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if (-not $currentPath.Contains($installDir)) {
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$installDir", "User")
        }
        
        Write-Color "安装完成: $installDir\delguard.exe" "Green"
        Write-Color "请重新打开终端以使用 delguard 命令" "Green"
    }
}

function Uninstall-Local {
    Write-Color "卸载本地安装..." "Cyan"
    
    $installDir = "$env:ProgramFiles\DelGuard"
    
    if (Test-Path $installDir) {
        Remove-Item $installDir -Recurse -Force
        Write-Color "已删除安装目录: $installDir" "Green"
    }
    
    # 从PATH中移除
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    $newPath = $currentPath -replace [regex]::Escape(";$installDir"), ""
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    
    Write-Color "卸载完成" "Green"
}

function Clean-Build {
    Write-Color "清理构建文件..." "Cyan"
    
    $itemsToClean = @(
        "build",
        "coverage.out",
        "coverage.html",
        "*.exe",
        "*.test"
    )
    
    foreach ($item in $itemsToClean) {
        if (Test-Path $item) {
            if (Test-Path $item -PathType Container) {
                Remove-Item $item -Recurse -Force
            } else {
                Remove-Item $item -Force
            }
            Write-Color "已删除: $item" "Green"
        }
    }
    
    Write-Color "清理完成" "Green"
}

# 主程序
function Main {
    Write-Color "=== DelGuard 构建脚本 v$Version ===" "Magenta"
    Write-Color "构建日期: $BuildDate" "Cyan"
    Write-Color "Git提交: $GitCommit" "Cyan"
    Write-Color ""
    
    if ($Clean) {
        Clean-Build
        return
    }
    
    if ($Uninstall) {
        Uninstall-Local
        return
    }
    
    # 检查环境
    Test-GoEnvironment
    
    if ($Test) {
        Invoke-UnitTests
        Invoke-SecurityTests
    } else {
        if ($Target -eq "all") {
            Build-All
        } else {
            $parts = $Target.Split("-")
            if ($parts.Count -eq 2) {
                Build-Target -OS $parts[0] -Arch $parts[1] -Suffix $(if ($parts[0] -eq "windows") { ".exe" } else { "" })
            } else {
                Write-Color "错误: 无效的目标格式" "Red"
                Write-Color "示例: windows-amd64, linux-amd64, darwin-arm64" "Yellow"
                exit 1
            }
        }
    }
    
    if ($Install) {
        Install-Local
    }
    
    Write-Color "=== 构建完成 ===" "Green"
}

# 执行主程序
Main