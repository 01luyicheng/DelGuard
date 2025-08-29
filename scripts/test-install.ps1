# DelGuard Windows 安装测试脚本
# 自动化测试安装过程的各个环节

param(
    [switch]$Detailed
)

# 导入错误处理库
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ErrorHandlerPath = Join-Path $ScriptDir "lib\error-handler.ps1"

if (Test-Path $ErrorHandlerPath) {
    . $ErrorHandlerPath
    Initialize-ErrorHandler
} else {
    # 基本错误处理
    function Write-Info { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Blue }
    function Write-Success { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
    function Write-Warning { param([string]$Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
    function Write-Error { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }
    function Write-Header { param([string]$Message) Write-Host $Message -ForegroundColor Cyan; Write-Host ("=" * 50) }
}

# 测试配置
$TestDir = "$env:TEMP\delguard-install-test"
$InstallScript = Join-Path $ScriptDir "install.ps1"
$VerifyScript = Join-Path $ScriptDir "verify-install.ps1"
$RepairScript = Join-Path $ScriptDir "repair-install.ps1"

# 测试结果
$TestsPassed = 0
$TestsFailed = 0
$TestResults = @()

# 测试函数
function Invoke-Test {
    param(
        [string]$TestName,
        [scriptblock]$TestFunction
    )
    
    Write-Header "🧪 测试: $TestName"
    
    try {
        $result = & $TestFunction
        if ($result) {
            Write-Success "测试通过: $TestName"
            $script:TestsPassed++
            $script:TestResults += "✅ $TestName"
        } else {
            Write-Error "测试失败: $TestName"
            $script:TestsFailed++
            $script:TestResults += "❌ $TestName"
        }
    } catch {
        Write-Error "测试异常: $TestName - $($_.Exception.Message)"
        $script:TestsFailed++
        $script:TestResults += "❌ $TestName (异常)"
    }
    
    Write-Host ""
}

# 准备测试环境
function Initialize-TestEnvironment {
    Write-Info "准备测试环境..."
    
    # 创建测试目录
    if (-not (Test-Path $TestDir)) {
        New-Item -Path $TestDir -ItemType Directory -Force | Out-Null
    }
    
    # 检查现有安装
    $existingDelguard = Get-Command delguard -ErrorAction SilentlyContinue
    if ($existingDelguard) {
        Write-Info "发现现有DelGuard: $($existingDelguard.Source)"
    }
    
    return $true
}

# 清理测试环境
function Clear-TestEnvironment {
    Write-Info "清理测试环境..."
    
    # 删除测试目录
    if (Test-Path $TestDir) {
        Remove-Item $TestDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Success "已删除测试目录"
    }
    
    return $true
}

# 测试1: 脚本语法检查
function Test-ScriptSyntax {
    Write-Info "检查PowerShell脚本语法..."
    
    $scripts = @($InstallScript, $VerifyScript, $RepairScript)
    
    foreach ($script in $scripts) {
        if (Test-Path $script) {
            try {
                $null = [System.Management.Automation.PSParser]::Tokenize((Get-Content $script -Raw), [ref]$null)
                Write-Success "语法检查通过: $(Split-Path $script -Leaf)"
            } catch {
                Write-Error "语法错误: $(Split-Path $script -Leaf) - $($_.Exception.Message)"
                return $false
            }
        } else {
            Write-Warning "脚本不存在: $(Split-Path $script -Leaf)"
        }
    }
    
    return $true
}

# 测试2: 系统要求检查
function Test-SystemRequirements {
    Write-Info "检查系统要求..."
    
    # 检查PowerShell版本
    if ($PSVersionTable.PSVersion.Major -ge 5) {
        Write-Success "PowerShell版本支持: $($PSVersionTable.PSVersion)"
    } else {
        Write-Error "PowerShell版本过低: $($PSVersionTable.PSVersion)"
        return $false
    }
    
    # 检查操作系统
    $osInfo = Get-CimInstance Win32_OperatingSystem
    if ($osInfo.Version -ge "10.0") {
        Write-Success "操作系统支持: $($osInfo.Caption)"
    } else {
        Write-Error "操作系统版本过低: $($osInfo.Caption)"
        return $false
    }
    
    # 检查架构
    $supportedArchs = @("AMD64", "ARM64", "x86")
    if ($env:PROCESSOR_ARCHITECTURE -in $supportedArchs) {
        Write-Success "系统架构支持: $env:PROCESSOR_ARCHITECTURE"
    } else {
        Write-Warning "未测试的架构: $env:PROCESSOR_ARCHITECTURE"
    }
    
    return $true
}

# 测试3: 网络连接
function Test-NetworkConnectivity {
    Write-Info "测试网络连接..."
    
    try {
        $null = Invoke-WebRequest -Uri "https://api.github.com" -UseBasicParsing -TimeoutSec 10
        Write-Success "GitHub API连接正常"
        return $true
    } catch {
        Write-Error "无法连接到GitHub API: $($_.Exception.Message)"
        return $false
    }
}

# 测试4: 权限检查
function Test-Permissions {
    Write-Info "测试权限..."
    
    # 测试临时目录写入权限
    $testFile = Join-Path $env:TEMP "delguard-permission-test.tmp"
    try {
        "test" | Out-File $testFile -Force
        Remove-Item $testFile -Force
        Write-Success "临时目录可写"
    } catch {
        Write-Error "临时目录不可写: $($_.Exception.Message)"
        return $false
    }
    
    # 检查执行策略
    $executionPolicy = Get-ExecutionPolicy
    if ($executionPolicy -in @("RemoteSigned", "Unrestricted", "Bypass")) {
        Write-Success "执行策略允许: $executionPolicy"
    } else {
        Write-Warning "执行策略可能阻止脚本运行: $executionPolicy"
    }
    
    return $true
}

# 测试5: 错误处理库
function Test-ErrorHandler {
    Write-Info "测试错误处理库..."
    
    if (Test-Path $ErrorHandlerPath) {
        try {
            $null = [System.Management.Automation.PSParser]::Tokenize((Get-Content $ErrorHandlerPath -Raw), [ref]$null)
            Write-Success "错误处理库语法正确"
            return $true
        } catch {
            Write-Error "错误处理库语法错误: $($_.Exception.Message)"
            return $false
        }
    } else {
        Write-Error "错误处理库不存在"
        return $false
    }
}

# 测试6: 模块导入
function Test-ModuleImport {
    Write-Info "测试模块导入..."
    
    try {
        # 测试导入错误处理库
        if (Test-Path $ErrorHandlerPath) {
            . $ErrorHandlerPath
            Write-Success "错误处理库导入成功"
        }
        return $true
    } catch {
        Write-Error "模块导入失败: $($_.Exception.Message)"
        return $false
    }
}

# 生成测试报告
function New-TestReport {
    Write-Header "📊 测试报告"
    
    Write-Host "测试完成时间: $(Get-Date)"
    Write-Host "通过测试: $TestsPassed"
    Write-Host "失败测试: $TestsFailed"
    Write-Host "总计测试: $($TestsPassed + $TestsFailed)"
    Write-Host ""
    
    Write-Host "详细结果:"
    foreach ($result in $TestResults) {
        Write-Host "  $result"
    }
    Write-Host ""
    
    if ($TestsFailed -eq 0) {
        Write-Success "🎉 所有测试通过！安装脚本准备就绪。"
        return $true
    } else {
        Write-Error "❌ 有 $TestsFailed 个测试失败，请修复后重试。"
        return $false
    }
}

# 主函数
function Main {
    Write-Header "🛡️  DelGuard Windows 安装测试套件"
    
    # 准备测试环境
    Initialize-TestEnvironment
    
    # 运行测试
    Invoke-Test "脚本语法检查" { Test-ScriptSyntax }
    Invoke-Test "系统要求检查" { Test-SystemRequirements }
    Invoke-Test "网络连接测试" { Test-NetworkConnectivity }
    Invoke-Test "权限检查" { Test-Permissions }
    Invoke-Test "错误处理库测试" { Test-ErrorHandler }
    Invoke-Test "模块导入测试" { Test-ModuleImport }
    
    # 清理测试环境
    Clear-TestEnvironment
    
    # 生成报告
    $success = New-TestReport
    
    if (-not $success) {
        exit 1
    }
}

# 运行主函数
Main