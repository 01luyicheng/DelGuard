# DelGuard Windows 构建脚本
# PowerShell 版本的构建脚本

param(
    [Parameter(Position=0)]
    [ValidateSet("clean", "deps", "test", "build", "package", "all", "info")]
    [string]$Command = "all"
)

# 项目信息
$ProjectName = "delguard"
$BuildDir = "build"
$DistDir = "dist"

# 颜色定义
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

# 获取版本信息
function Get-Version {
    try {
        $version = git describe --tags --always --dirty 2>$null
        if ($version) {
            return $version
        }
    }
    catch {
        # Git 不可用或不在仓库中
    }
    return "v0.1.0"
}

# 获取Git提交信息
function Get-GitCommit {
    try {
        $commit = git rev-parse --short HEAD 2>$null
        if ($commit) {
            return $commit
        }
    }
    catch {
        # Git 不可用或不在仓库中
    }
    return "unknown"
}

# 检查依赖
function Test-Dependencies {
    Write-Info "检查构建依赖..."
    
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Error "Go 未安装或不在 PATH 中"
        exit 1
    }
    
    $goVersion = go version
    Write-Success "Go 版本: $goVersion"
}

# 清理构建目录
function Clear-Build {
    Write-Info "清理构建目录..."
    
    if (Test-Path $BuildDir) {
        Remove-Item -Recurse -Force $BuildDir
    }
    
    if (Test-Path $DistDir) {
        Remove-Item -Recurse -Force $DistDir
    }
    
    go clean
    Write-Success "构建目录已清理"
}

# 安装依赖
function Install-Dependencies {
    Write-Info "安装Go依赖..."
    go mod download
    go mod tidy
    Write-Success "依赖安装完成"
}

# 运行测试
function Invoke-Tests {
    Write-Info "运行测试..."
    $testResult = go test -v ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Success "所有测试通过"
    } else {
        Write-Error "测试失败"
        exit 1
    }
}

# 构建单个平台
function Build-Platform {
    param(
        [string]$GOOS,
        [string]$GOARCH,
        [string]$Version,
        [string]$GitCommit,
        [string]$BuildTime
    )
    
    $outputName = "$ProjectName-$Version-$GOOS-$GOARCH"
    if ($GOOS -eq "windows") {
        $outputName += ".exe"
    }
    
    Write-Info "构建 $GOOS/$GOARCH..."
    
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    
    $ldflags = "-X main.Version=$Version -X main.BuildTime=$BuildTime -X main.GitCommit=$GitCommit -s -w"
    
    go build -ldflags $ldflags -o "$DistDir\$outputName" ".\cmd\$ProjectName"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "构建完成: $outputName"
        return $true
    } else {
        Write-Error "构建失败: $GOOS/$GOARCH"
        return $false
    }
}

# 构建所有平台
function Build-All {
    $version = Get-Version
    $gitCommit = Get-GitCommit
    $buildTime = Get-Date -Format "yyyy-MM-dd_HH:mm:ss" -AsUTC
    
    Write-Info "开始构建所有平台..."
    Write-Info "版本: $version"
    Write-Info "Git提交: $gitCommit"
    Write-Info "构建时间: $buildTime"
    
    if (-not (Test-Path $DistDir)) {
        New-Item -ItemType Directory -Path $DistDir | Out-Null
    }
    
    # 支持的平台列表
    $platforms = @(
        @{GOOS="windows"; GOARCH="amd64"},
        @{GOOS="windows"; GOARCH="386"},
        @{GOOS="linux"; GOARCH="amd64"},
        @{GOOS="linux"; GOARCH="386"},
        @{GOOS="linux"; GOARCH="arm64"},
        @{GOOS="darwin"; GOARCH="amd64"},
        @{GOOS="darwin"; GOARCH="arm64"}
    )
    
    $successCount = 0
    $totalCount = $platforms.Count
    
    foreach ($platform in $platforms) {
        if (Build-Platform $platform.GOOS $platform.GOARCH $version $gitCommit $buildTime) {
            $successCount++
        }
    }
    
    Write-Success "构建完成: $successCount/$totalCount 个平台"
}

# 打包发布文件
function New-Packages {
    Write-Info "打包发布文件..."
    
    if (-not (Test-Path $DistDir)) {
        Write-Error "构建目录不存在，请先运行构建"
        exit 1
    }
    
    Push-Location $DistDir
    
    try {
        $files = Get-ChildItem -Name "$ProjectName-*" -File
        
        foreach ($file in $files) {
            if ($file -like "*windows*") {
                # Windows 平台使用 zip
                $zipName = "$file.zip"
                if (Get-Command Compress-Archive -ErrorAction SilentlyContinue) {
                    $items = @($file)
                    if (Test-Path "..\README.md") { $items += "..\README.md" }
                    if (Test-Path "..\LICENSE") { $items += "..\LICENSE" }
                    Compress-Archive -Path $items -DestinationPath $zipName -Force
                    Write-Success "打包完成: $zipName"
                }
            } else {
                # Unix 平台使用 tar.gz (需要 tar 命令)
                $tarName = "$file.tar.gz"
                if (Get-Command tar -ErrorAction SilentlyContinue) {
                    tar -czf $tarName $file ..\README.md ..\LICENSE 2>$null
                    if ($LASTEXITCODE -eq 0) {
                        Write-Success "打包完成: $tarName"
                    }
                }
            }
        }
    }
    finally {
        Pop-Location
    }
}

# 生成校验和
function New-Checksums {
    Write-Info "生成校验和文件..."
    
    Push-Location $DistDir
    
    try {
        $files = Get-ChildItem -Name "*.zip", "*.tar.gz" -File 2>$null
        
        if ($files) {
            $checksums = @()
            foreach ($file in $files) {
                if (Test-Path $file) {
                    $hash = Get-FileHash -Path $file -Algorithm SHA256
                    $checksums += "$($hash.Hash.ToLower())  $file"
                }
            }
            
            if ($checksums) {
                $checksums | Out-File -FilePath "checksums.sha256" -Encoding UTF8
                Write-Success "校验和文件已生成: checksums.sha256"
            }
        }
    }
    finally {
        Pop-Location
    }
}

# 显示构建信息
function Show-BuildInfo {
    Write-Info "构建信息:"
    Write-Host "  项目名称: $ProjectName"
    Write-Host "  版本: $(Get-Version)"
    Write-Host "  Git提交: $(Get-GitCommit)"
    Write-Host "  构建时间: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss UTC' -AsUTC)"
    Write-Host "  Go版本: $(go version)"
}

# 主函数
function Main {
    switch ($Command) {
        "clean" {
            Clear-Build
        }
        "deps" {
            Install-Dependencies
        }
        "test" {
            Invoke-Tests
        }
        "build" {
            Test-Dependencies
            Install-Dependencies
            Build-All
        }
        "package" {
            New-Packages
            New-Checksums
        }
        "all" {
            Test-Dependencies
            Clear-Build
            Install-Dependencies
            Invoke-Tests
            Build-All
            New-Packages
            New-Checksums
            Write-Success "完整构建流程完成！"
        }
        "info" {
            Show-BuildInfo
        }
        default {
            Write-Host "用法: .\build.ps1 {clean|deps|test|build|package|all|info}"
            Write-Host ""
            Write-Host "命令说明:"
            Write-Host "  clean   - 清理构建目录"
            Write-Host "  deps    - 安装依赖"
            Write-Host "  test    - 运行测试"
            Write-Host "  build   - 构建所有平台"
            Write-Host "  package - 打包发布文件"
            Write-Host "  all     - 完整构建流程"
            Write-Host "  info    - 显示构建信息"
            exit 1
        }
    }
}

# 执行主函数
Main