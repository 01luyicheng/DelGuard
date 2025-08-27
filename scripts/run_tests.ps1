# DelGuard 测试运行脚本

param(
    [string]$TestType = "all",  # all, unit, integration, benchmark
    [switch]$Coverage,
    [switch]$Verbose,
    [string]$Output = "test_results"
)

$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

Write-ColorOutput "=== DelGuard 测试套件 ===" "Cyan"
Write-ColorOutput "测试类型: $TestType" "Yellow"

# 创建输出目录
if (!(Test-Path $Output)) {
    New-Item -ItemType Directory -Path $Output -Force | Out-Null
}

# 测试参数
$TestArgs = @()
if ($Verbose) {
    $TestArgs += "-v"
}

if ($Coverage) {
    $TestArgs += "-coverprofile=$Output/coverage.out"
    $TestArgs += "-covermode=atomic"
}

try {
    switch ($TestType.ToLower()) {
        "unit" {
            Write-ColorOutput "🧪 运行单元测试..." "Yellow"
            & go test $TestArgs ./internal/...
            if ($LASTEXITCODE -ne 0) { throw "单元测试失败" }
        }
        
        "integration" {
            Write-ColorOutput "🔗 运行集成测试..." "Yellow"
            & go test $TestArgs ./tests/integration/...
            if ($LASTEXITCODE -ne 0) { throw "集成测试失败" }
        }
        
        "benchmark" {
            Write-ColorOutput "⚡ 运行性能测试..." "Yellow"
            & go test -bench=. -benchmem ./tests/benchmarks/... | Tee-Object "$Output/benchmark_results.txt"
            if ($LASTEXITCODE -ne 0) { throw "性能测试失败" }
        }
        
        "all" {
            Write-ColorOutput "🧪 运行单元测试..." "Yellow"
            & go test $TestArgs ./internal/...
            if ($LASTEXITCODE -ne 0) { throw "单元测试失败" }
            
            Write-ColorOutput "🔗 运行集成测试..." "Yellow"
            & go test $TestArgs ./tests/integration/...
            if ($LASTEXITCODE -ne 0) { throw "集成测试失败" }
            
            Write-ColorOutput "⚡ 运行性能测试..." "Yellow"
            & go test -bench=. -benchmem ./tests/benchmarks/... | Tee-Object "$Output/benchmark_results.txt"
            if ($LASTEXITCODE -ne 0) { throw "性能测试失败" }
        }
        
        default {
            throw "未知的测试类型: $TestType"
        }
    }
    
    # 生成覆盖率报告
    if ($Coverage -and (Test-Path "$Output/coverage.out")) {
        Write-ColorOutput "📊 生成覆盖率报告..." "Yellow"
        & go tool cover -html="$Output/coverage.out" -o "$Output/coverage.html"
        & go tool cover -func="$Output/coverage.out" | Tee-Object "$Output/coverage_summary.txt"
        
        Write-ColorOutput "覆盖率报告已生成: $Output/coverage.html" "Green"
    }
    
    Write-ColorOutput "✅ 所有测试通过!" "Green"
    
} catch {
    Write-ColorOutput "❌ 测试失败: $($_.Exception.Message)" "Red"
    exit 1
}

Write-ColorOutput "=== 测试完成 ===" "Cyan"