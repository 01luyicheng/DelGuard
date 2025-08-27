#!/usr/bin/env pwsh
# DelGuard Code Quality Check Script

param(
    [switch]$Fix = $false,
    [switch]$Verbose = $false
)

Write-Host "🔍 DelGuard Code Quality Check Started..." -ForegroundColor Green

# Check Go environment
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Go is not installed or not in PATH" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Go Version: $(go version)" -ForegroundColor Green

# 1. Format code
Write-Host "`n📝 Formatting code..." -ForegroundColor Yellow
go fmt ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Code formatting failed" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Code formatting completed" -ForegroundColor Green

# 2. Tidy dependencies
Write-Host "`n📦 Tidying dependencies..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Dependency tidying failed" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Dependencies tidied" -ForegroundColor Green

# 3. Static analysis
Write-Host "`n🔍 Running static analysis..." -ForegroundColor Yellow
go vet ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Static analysis found issues" -ForegroundColor Red
    if (-not $Fix) {
        exit 1
    }
}
Write-Host "✅ Static analysis completed" -ForegroundColor Green

# 4. Check golangci-lint
if (Get-Command golangci-lint -ErrorAction SilentlyContinue) {
    Write-Host "`n🔧 Running golangci-lint..." -ForegroundColor Yellow
    if ($Fix) {
        golangci-lint run --fix
    } else {
        golangci-lint run
    }
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ golangci-lint found issues" -ForegroundColor Red
        if (-not $Fix) {
            exit 1
        }
    }
    Write-Host "✅ golangci-lint check completed" -ForegroundColor Green
} else {
    Write-Host "⚠️  golangci-lint not installed, skipping advanced checks" -ForegroundColor Yellow
    Write-Host "   Install command: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" -ForegroundColor Cyan
}

# 5. Build test
Write-Host "`n🏗️  Building project..." -ForegroundColor Yellow
go build -o delguard.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Build successful" -ForegroundColor Green

# 6. Run tests
Write-Host "`n🧪 Running tests..." -ForegroundColor Yellow
if (Test-Path "tests") {
    go test ./tests/... -v
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Tests failed" -ForegroundColor Red
        if (-not $Fix) {
            exit 1
        }
    }
    Write-Host "✅ Tests passed" -ForegroundColor Green
} else {
    Write-Host "⚠️  Test directory not found" -ForegroundColor Yellow
}

# 7. Security check
Write-Host "`n🔒 Running security checks..." -ForegroundColor Yellow
if (Get-Command gosec -ErrorAction SilentlyContinue) {
    gosec ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Security check found issues" -ForegroundColor Red
        if (-not $Fix) {
            exit 1
        }
    }
    Write-Host "✅ Security check completed" -ForegroundColor Green
} else {
    Write-Host "⚠️  gosec not installed, skipping security checks" -ForegroundColor Yellow
    Write-Host "   Install command: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest" -ForegroundColor Cyan
}

Write-Host "`n🎉 Code quality check completed!" -ForegroundColor Green

if ($Verbose) {
    Write-Host "`n📊 Project Statistics:" -ForegroundColor Cyan
    $goFiles = (Get-ChildItem -Recurse -Filter "*.go" | Where-Object { $_.FullName -notmatch "vendor|\.git" }).Count
    $totalLines = (Get-ChildItem -Recurse -Filter "*.go" | Where-Object { $_.FullName -notmatch "vendor|\.git" } | Get-Content | Measure-Object -Line).Lines
    Write-Host "   Go files: $goFiles" -ForegroundColor White
    Write-Host "   Total lines: $totalLines" -ForegroundColor White
}