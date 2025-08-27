# DelGuard Quality Assurance Check Script

param(
    [switch]$Fix,
    [string]$Output = "qa_results"
)

$ErrorActionPreference = "Stop"

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

Write-ColorOutput "=== DelGuard Quality Assurance Check ===" "Cyan"

# Create output directory
if (!(Test-Path $Output)) {
    New-Item -ItemType Directory -Path $Output -Force | Out-Null
}

$QAResults = @{
    "CodeFormat" = $false
    "StaticAnalysis" = $false
    "TestCoverage" = $false
}

try {
    # 1. Code format check
    Write-ColorOutput "Checking code format..." "Yellow"
    
    if ($Fix) {
        & go fmt ./...
        Write-ColorOutput "Code format has been automatically fixed" "Green"
    }
    
    $fmtOutput = & go fmt ./... 2>&1
    if ($fmtOutput) {
        Write-ColorOutput "Code format issues found:" "Red"
        Write-Host $fmtOutput
    } else {
        Write-ColorOutput "Code format check passed" "Green"
        $QAResults.CodeFormat = $true
    }
    
    # 2. Static analysis
    Write-ColorOutput "Running static analysis..." "Yellow"
    
    & go vet ./... 2>&1 | Tee-Object "$Output/vet_results.txt"
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "go vet check passed" "Green"
        $QAResults.StaticAnalysis = $true
    } else {
        Write-ColorOutput "go vet found issues, see $Output/vet_results.txt" "Red"
    }
    
    # 3. Test coverage
    Write-ColorOutput "Checking test coverage..." "Yellow"
    
    & go test -coverprofile="$Output/coverage.out" ./internal/... 2>&1 | Tee-Object "$Output/test_results.txt"
    if ($LASTEXITCODE -eq 0) {
        $coverageOutput = & go tool cover -func="$Output/coverage.out" | Select-String "total:"
        if ($coverageOutput) {
            $coveragePercent = ($coverageOutput -split "\s+")[-1]
            Write-ColorOutput "Test coverage: $coveragePercent" "Green"
            
            # Check coverage threshold
            $coverageValue = [float]($coveragePercent -replace "%", "")
            if ($coverageValue -ge 60) {
                $QAResults.TestCoverage = $true
            } else {
                Write-ColorOutput "Test coverage is below 60%" "Yellow"
            }
        }
    } else {
        Write-ColorOutput "Test execution failed" "Red"
    }
    
    # Calculate overall quality score
    $passedChecks = ($QAResults.Values | Where-Object {$_} | Measure-Object).Count
    $totalChecks = $QAResults.Count
    $qualityScore = [math]::Round(($passedChecks / $totalChecks) * 100, 2)
    
    Write-ColorOutput "=== Quality Assurance Check Complete ===" "Cyan"
    Write-ColorOutput "Overall Quality Score: $qualityScore%" $(if($qualityScore -ge 80){"Green"}else{"Yellow"})
    
} catch {
    Write-ColorOutput "Error occurred during quality check: $($_.Exception.Message)" "Red"
    exit 1
}