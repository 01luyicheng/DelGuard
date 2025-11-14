/*
 * 文件用途：DelGuard安装程序兼容性测试脚本
 * 作者：TREA
 * 功能说明：测试安装程序在不同Windows版本和配置下的兼容性
 * 输入：测试配置参数
 * 输出：兼容性测试报告
 * 依赖：PowerShell 5.1+, Windows评估工具包
 * 交互方式：生成测试报告和日志文件
 */

#Requires -Version 5.1

<#
.SYNOPSIS
    DelGuard安装程序兼容性测试脚本

.DESCRIPTION
    该脚本用于测试DelGuard MSI安装程序在不同Windows版本、
    配置和环境下的兼容性和稳定性

.PARAMETER TestPath
    测试安装包路径

.PARAMETER OutputPath
    测试报告输出路径

.PARAMETER TestModes
    测试模式：All, FreshInstall, Upgrade, Uninstall, Compatibility

.EXAMPLE
    .\test-compatibility.ps1 -TestPath ".\output\DelGuard.msi"
    执行完整的兼容性测试

.EXAMPLE
    .\test-compatibility.ps1 -TestPath ".\DelGuard.msi" -TestModes FreshInstall
    仅执行全新安装测试
#>

param(
    [Parameter(Mandatory=$true)]
    [string]$TestPath,
    
    [string]$OutputPath = ".\TestResults",
    
    [ValidateSet("All", "FreshInstall", "Upgrade", "Uninstall", "Compatibility")]
    [string]$TestModes = "All"
)

$ErrorActionPreference = "Continue"
$ProgressPreference = "SilentlyContinue"

# 测试配置
$TestConfig = @{
    ProductName = "DelGuard"
    TestDuration = 0
    SuccessCount = 0
    FailureCount = 0
    WarningCount = 0
    Results = @()
}

# 系统信息
$SystemInfo = @{
    OSVersion = [System.Environment]::OSVersion.Version
    OSName = (Get-CimInstance -ClassName Win32_OperatingSystem).Caption
    Architecture = $env:PROCESSOR_ARCHITECTURE
    TotalMemory = [math]::Round((Get-CimInstance -ClassName Win32_ComputerSystem).TotalPhysicalMemory / 1GB, 2)
    FreeMemory = [math]::Round((Get-CimInstance -ClassName Win32_OperatingSystem).FreePhysicalMemory / 1MB, 2)
    DotNetVersion = (Get-ItemProperty "HKLM:SOFTWARE\Microsoft\NET Framework Setup\NDP\v4\Full\" -Name Release -ErrorAction SilentlyContinue).Release
}

# 日志函数
function Write-TestLog {
    param([string]$Message, [string]$Level = "INFO", [string]$TestName = "")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logEntry = "[$timestamp] [$Level] [$TestName] $Message"
    Write-Host $logEntry
    
    $logFile = Join-Path $OutputPath "CompatibilityTest.log"
    Add-Content -Path $logFile -Value $logEntry
}

function Initialize-TestEnvironment {
    Write-TestLog "初始化测试环境..." "INFO" "Setup"
    
    # 创建输出目录
    if (!(Test-Path $OutputPath)) {
        New-Item -ItemType Directory -Path $OutputPath -Force | Out-Null
    }
    
    # 检查测试文件
    if (!(Test-Path $TestPath)) {
        throw "测试文件不存在: $TestPath"
    }
    
    # 记录系统信息
    Write-TestLog "操作系统: $($SystemInfo.OSName)" "INFO" "System"
    Write-TestLog "系统版本: $($SystemInfo.OSVersion)" "INFO" "System"
    Write-TestLog "系统架构: $($SystemInfo.Architecture)" "INFO" "System"
    Write-TestLog "总内存: $($SystemInfo.TotalMemory) GB" "INFO" "System"
    Write-TestLog "可用内存: $($SystemInfo.FreeMemory) GB" "INFO" "System"
    
    # 检查Windows Installer服务
    $msiService = Get-Service -Name "msiserver" -ErrorAction SilentlyContinue
    if ($msiService -and $msiService.Status -eq "Running") {
        Write-TestLog "Windows Installer服务运行正常" "INFO" "System"
    }
    else {
        Write-TestLog "Windows Installer服务未运行" "WARNING" "System"
    }
}

function Test-InstallationPackage {
    param([string]$PackagePath)
    
    Write-TestLog "检查安装包完整性..." "INFO" "PackageCheck"
    
    try {
        # 检查文件大小
        $fileInfo = Get-Item $PackagePath
        $fileSizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
        Write-TestLog "安装包大小: $fileSizeMB MB" "INFO" "PackageCheck"
        
        if ($fileInfo.Length -eq 0) {
            throw "安装包文件大小为0"
        }
        
        # 使用Windows Installer API检查包
        $installer = New-Object -ComObject WindowsInstaller.Installer
        $database = $installer.OpenDatabase($PackagePath, 0)
        
        if ($database) {
            Write-TestLog "安装包格式验证通过" "INFO" "PackageCheck"
            
            # 检查产品信息
            $view = $database.OpenView("SELECT Value FROM Property WHERE Property = 'ProductName'")
            $view.Execute()
            $record = $view.Fetch()
            if ($record) {
                $productName = $record.StringData(1)
                Write-TestLog "产品名称: $productName" "INFO" "PackageCheck"
            }
            
            $view = $database.OpenView("SELECT Value FROM Property WHERE Property = 'ProductVersion'")
            $view.Execute()
            $record = $view.Fetch()
            if ($record) {
                $productVersion = $record.StringData(1)
                Write-TestLog "产品版本: $productVersion" "INFO" "PackageCheck"
            }
            
            return $true
        }
        else {
            throw "无法打开安装包数据库"
        }
    }
    catch {
        Write-TestLog "安装包检查失败: $($_.Exception.Message)" "ERROR" "PackageCheck"
        return $false
    }
}

function Test-FreshInstallation {
    Write-TestLog "开始全新安装测试..." "INFO" "FreshInstall"
    
    $testResult = @{
        TestName = "Fresh Installation"
        Status = "Unknown"
        Duration = 0
        Details = @()
    }
    
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    
    try {
        # 清理之前的安装（如果存在）
        $productCode = "{12345678-1234-1234-1234-123456789012}"  # 需要替换为实际的产品代码
        msiexec /x $productCode /qn /norestart 2>$null
        
        Start-Sleep -Seconds 2
        
        # 执行全新安装
        $arguments = @(
            "/i", $TestPath
            "/qn"
            "/norestart"
            "/l*v", (Join-Path $OutputPath "FreshInstall.log")
        )
        
        Write-TestLog "执行安装命令: msiexec $arguments" "INFO" "FreshInstall"
        $process = Start-Process -FilePath "msiexec.exe" -ArgumentList $arguments -Wait -PassThru
        
        $testResult.Duration = $stopwatch.Elapsed.TotalSeconds
        
        if ($process.ExitCode -eq 0) {
            Write-TestLog "全新安装成功，耗时: $($testResult.Duration) 秒" "INFO" "FreshInstall"
            $testResult.Status = "Success"
            
            # 验证安装结果
            $installDir = Join-Path $env:ProgramFiles "DelGuard"
            if (Test-Path $installDir) {
                $testResult.Details += "安装目录创建成功: $installDir"
                
                $exePath = Join-Path $installDir "delguard.exe"
                if (Test-Path $exePath) {
                    $testResult.Details += "可执行文件存在: $exePath"
                    
                    # 测试基本功能
                    $testOutput = & "$exePath" --help 2>&1
                    if ($LASTEXITCODE -eq 0) {
                        $testResult.Details += "基本功能测试通过"
                    }
                    else {
                        $testResult.Details += "基本功能测试失败"
                        $testResult.Status = "Warning"
                    }
                }
                else {
                    $testResult.Details += "可执行文件不存在"
                    $testResult.Status = "Failure"
                }
            }
            else {
                $testResult.Details += "安装目录未创建"
                $testResult.Status = "Failure"
            }
        }
        else {
            Write-TestLog "全新安装失败，退出码: $($process.ExitCode)" "ERROR" "FreshInstall"
            $testResult.Status = "Failure"
            $testResult.Details += "安装过程失败，退出码: $($process.ExitCode)"
        }
    }
    catch {
        Write-TestLog "全新安装测试异常: $($_.Exception.Message)" "ERROR" "FreshInstall"
        $testResult.Status = "Failure"
        $testResult.Details += "异常: $($_.Exception.Message)"
    }
    finally {
        $stopwatch.Stop()
    }
    
    return $testResult
}

function Test-Uninstallation {
    Write-TestLog "开始卸载测试..." "INFO" "Uninstall"
    
    $testResult = @{
        TestName = "Uninstallation"
        Status = "Unknown"
        Duration = 0
        Details = @()
    }
    
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    
    try {
        # 执行卸载
        $productCode = "{12345678-1234-1234-1234-123456789012}"  # 需要替换为实际的产品代码
        $arguments = @(
            "/x", $productCode
            "/qn"
            "/norestart"
            "/l*v", (Join-Path $OutputPath "Uninstall.log")
        )
        
        Write-TestLog "执行卸载命令: msiexec $arguments" "INFO" "Uninstall"
        $process = Start-Process -FilePath "msiexec.exe" -ArgumentList $arguments -Wait -PassThru
        
        $testResult.Duration = $stopwatch.Elapsed.TotalSeconds
        
        if ($process.ExitCode -eq 0) {
            Write-TestLog "卸载成功，耗时: $($testResult.Duration) 秒" "INFO" "Uninstall"
            $testResult.Status = "Success"
            
            # 验证卸载结果
            $installDir = Join-Path $env:ProgramFiles "DelGuard"
            if (!(Test-Path $installDir)) {
                $testResult.Details += "安装目录已清理"
            }
            else {
                $testResult.Details += "警告：安装目录仍存在"
                $testResult.Status = "Warning"
            }
        }
        else {
            Write-TestLog "卸载失败，退出码: $($process.ExitCode)" "ERROR" "Uninstall"
            $testResult.Status = "Failure"
            $testResult.Details += "卸载过程失败，退出码: $($process.ExitCode)"
        }
    }
    catch {
        Write-TestLog "卸载测试异常: $($_.Exception.Message)" "ERROR" "Uninstall"
        $testResult.Status = "Failure"
        $testResult.Details += "异常: $($_.Exception.Message)"
    }
    finally {
        $stopwatch.Stop()
    }
    
    return $testResult
}

function Test-Compatibility {
    Write-TestLog "开始兼容性测试..." "INFO" "Compatibility"
    
    $testResult = @{
        TestName = "Compatibility Check"
        Status = "Success"
        Duration = 0
        Details = @()
    }
    
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    
    try {
        # 检查系统版本兼容性
        $osVersion = $SystemInfo.OSVersion
        if ($osVersion.Major -lt 10) {
            $testResult.Details += "警告：Windows版本较旧，可能不完全兼容"
            $testResult.Status = "Warning"
        }
        else {
            $testResult.Details += "系统版本兼容"
        }
        
        # 检查内存
        if ($SystemInfo.TotalMemory -lt 2) {
            $testResult.Details += "警告：系统内存较少，可能影响性能"
            if ($testResult.Status -eq "Success") {
                $testResult.Status = "Warning"
            }
        }
        else {
            $testResult.Details += "系统内存充足"
        }
        
        # 检查磁盘空间
        $systemDrive = $env:SystemDrive
        $diskSpace = Get-CimInstance -ClassName Win32_LogicalDisk -Filter "DeviceID='$systemDrive'"
        $freeSpaceGB = [math]::Round($diskSpace.FreeSpace / 1GB, 2)
        
        if ($freeSpaceGB -lt 1) {
            $testResult.Details += "警告：磁盘空间不足"
            if ($testResult.Status -eq "Success") {
                $testResult.Status = "Warning"
            }
        }
        else {
            $testResult.Details += "磁盘空间充足: $freeSpaceGB GB 可用"
        }
        
        # 检查依赖组件
        $requiredDlls = @("msi.dll", "msiserver", "advapi32.dll")
        foreach ($dll in $requiredDlls) {
            try {
                if ($dll -eq "msiserver") {
                    $service = Get-Service -Name $dll -ErrorAction Stop
                    if ($service.Status -ne "Running") {
                        throw "服务未运行"
                    }
                }
                else {
                    Add-Type -TypeDefinition "public class Test { [System.Runtime.InteropServices.DllImport(`"$dll`")] public static extern void Test(); }" -ErrorAction Stop
                }
                $testResult.Details += "依赖组件正常: $dll"
            }
            catch {
                $testResult.Details += "依赖组件问题: $dll"
                $testResult.Status = "Warning"
            }
        }
        
        $testResult.Duration = $stopwatch.Elapsed.TotalSeconds
        Write-TestLog "兼容性测试完成，耗时: $($testResult.Duration) 秒" "INFO" "Compatibility"
    }
    catch {
        Write-TestLog "兼容性测试异常: $($_.Exception.Message)" "ERROR" "Compatibility"
        $testResult.Status = "Failure"
        $testResult.Details += "异常: $($_.Exception.Message)"
    }
    finally {
        $stopwatch.Stop()
    }
    
    return $testResult
}

function Generate-TestReport {
    param([array]$TestResults)
    
    Write-TestLog "生成测试报告..." "INFO" "Report"
    
    $reportPath = Join-Path $OutputPath "CompatibilityTestReport.html"
    $reportTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    $html = @"
<!DOCTYPE html>
<html>
<head>
    <title>DelGuard兼容性测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .test-result { margin: 10px 0; padding: 10px; border-radius: 5px; }
        .success { background-color: #d4edda; border: 1px solid #c3e6cb; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; }
        .failure { background-color: #f8d7da; border: 1px solid #f5c6cb; }
        .details { margin-top: 10px; font-size: 0.9em; }
        .system-info { background-color: #e7f3ff; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>DelGuard兼容性测试报告</h1>
        <p>测试时间: $reportTime</p>
        <p>测试文件: $TestPath</p>
    </div>
    
    <div class="system-info">
        <h3>系统信息</h3>
        <p><strong>操作系统:</strong> $($SystemInfo.OSName)</p>
        <p><strong>系统版本:</strong> $($SystemInfo.OSVersion)</p>
        <p><strong>系统架构:</strong> $($SystemInfo.Architecture)</p>
        <p><strong>总内存:</strong> $($SystemInfo.TotalMemory) GB</p>
        <p><strong>可用内存:</strong> $($SystemInfo.FreeMemory) GB</p>
    </div>
    
    <div class="summary">
        <h2>测试结果摘要</h2>
"@
    
    $successCount = ($TestResults | Where-Object { $_.Status -eq "Success" }).Count
    $warningCount = ($TestResults | Where-Object { $_.Status -eq "Warning" }).Count
    $failureCount = ($TestResults | Where-Object { $_.Status -eq "Failure" }).Count
    $totalCount = $TestResults.Count
    
    $html += @"
        <p>总测试数: $totalCount</p>
        <p>成功: <span style="color: green;">$successCount</span></p>
        <p>警告: <span style="color: orange;">$warningCount</span></p>
        <p>失败: <span style="color: red;">$failureCount</span></p>
    </div>
    
    <div class="test-results">
        <h2>详细测试结果</h2>
"@
    
    foreach ($result in $TestResults) {
        $statusClass = $result.Status.ToLower()
        $html += @"
        <div class="test-result $statusClass">
            <h3>$($result.TestName)</h3>
            <p><strong>状态:</strong> $($result.Status)</p>
            <p><strong>耗时:</strong> $([math]::Round($result.Duration, 2)) 秒</p>
"@
        
        if ($result.Details.Count -gt 0) {
            $html += "        <div class='details'>"
            $html += "        <p><strong>详细信息:</strong></p>"
            $html += "        <ul>"
            foreach ($detail in $result.Details) {
                $html += "        <li>$detail</li>"
            }
            $html += "        </ul>"
            $html += "        </div>"
        }
        
        $html += "    </div>"
    }
    
    $html += @"
    </div>
</body>
</html>
"@
    
    $html | Out-File -FilePath $reportPath -Encoding UTF8
    Write-TestLog "测试报告已生成: $reportPath" "INFO" "Report"
    
    return $reportPath
}

# 主测试流程
function Main {
    Write-TestLog "开始DelGuard兼容性测试" "INFO" "Main"
    
    try {
        Initialize-TestEnvironment
        
        $allResults = @()
        
        # 安装包检查
        $packageCheck = Test-InstallationPackage -PackagePath $TestPath
        if ($packageCheck) {
            $allResults += @{
                TestName = "Package Validation"
                Status = "Success"
                Duration = 0
                Details = @("安装包格式验证通过")
            }
        }
        else {
            $allResults += @{
                TestName = "Package Validation"
                Status = "Failure"
                Duration = 0
                Details = @("安装包格式验证失败")
            }
        }
        
        # 根据测试模式执行相应测试
        switch ($TestModes) {
            "All" {
                $allResults += Test-FreshInstallation
                $allResults += Test-Uninstallation
                $allResults += Test-Compatibility
            }
            "FreshInstall" {
                $allResults += Test-FreshInstallation
            }
            "Uninstall" {
                $allResults += Test-Uninstallation
            }
            "Compatibility" {
                $allResults += Test-Compatibility
            }
        }
        
        # 生成报告
        $reportPath = Generate-TestReport -TestResults $allResults
        
        # 显示摘要
        Write-Host "`n================================" -ForegroundColor Cyan
        Write-Host "兼容性测试完成！" -ForegroundColor Cyan
        Write-Host "================================" -ForegroundColor Cyan
        
        $successCount = ($allResults | Where-Object { $_.Status -eq "Success" }).Count
        $failureCount = ($allResults | Where-Object { $_.Status -eq "Failure" }).Count
        
        Write-Host "成功测试: $successCount" -ForegroundColor Green
        Write-Host "失败测试: $failureCount" -ForegroundColor Red
        Write-Host "测试报告: $reportPath" -ForegroundColor Yellow
        
        if ($failureCount -eq 0) {
            Write-Host "`n所有测试均通过！安装程序兼容性良好。" -ForegroundColor Green
            exit 0
        }
        else {
            Write-Host "`n发现兼容性问题，请查看详细报告。" -ForegroundColor Red
            exit 1
        }
    }
    catch {
        Write-TestLog "测试过程异常: $($_.Exception.Message)" "ERROR" "Main"
        Write-Host "`n测试失败: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }
}

# 执行主函数
Main