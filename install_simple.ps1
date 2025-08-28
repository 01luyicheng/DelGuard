param(
    [switch]$Force,
    [string]$InstallPath = "$env:USERPROFILE\bin"
)

$ErrorActionPreference = 'Stop'

function Write-Success { param([string]$Message) Write-Host $Message -ForegroundColor Green }
function Write-Warning { param([string]$Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Error { param([string]$Message) Write-Host $Message -ForegroundColor Red }
function Write-Info { param([string]$Message) Write-Host $Message -ForegroundColor Cyan }

Write-Info "=== DelGuard Installer ==="
Write-Info "Install Path: $InstallPath"

$EXECUTABLE_NAME = "delguard.exe"
$EXECUTABLE_PATH = Join-Path $InstallPath $EXECUTABLE_NAME
$SourceExe = Join-Path $PSScriptRoot $EXECUTABLE_NAME

if (-not (Test-Path $SourceExe)) {
    Write-Error "delguard.exe not found"
    exit 1
}

Write-Success "Found source file: $SourceExe"

if ((Test-Path $EXECUTABLE_PATH) -and -not $Force) {
    Write-Warning "DelGuard already installed at: $EXECUTABLE_PATH"
    Write-Warning "Use -Force to reinstall"
    exit 1
}

try {
    # Create install directory
    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
        Write-Success "Created directory: $InstallPath"
    }

    # Copy executable
    Copy-Item $SourceExe $EXECUTABLE_PATH -Force
    Write-Success "Copied executable to: $EXECUTABLE_PATH"

    # Add to PATH
    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($UserPath -eq $null) { $UserPath = "" }
    
    if (-not $UserPath.Contains($InstallPath)) {
        if ($UserPath -eq "") {
            $NewPath = $InstallPath
        } else {
            $NewPath = "$UserPath;$InstallPath"
        }
        [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
        Write-Success "Added to PATH: $InstallPath"
        $env:PATH = "$env:PATH;$InstallPath"
    } else {
        Write-Info "Already in PATH: $InstallPath"
    }

    # PowerShell profiles
    $ProfilePaths = @(
        "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1",
        "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
    )

    $ConfigBlock = @"
# DelGuard PowerShell Configuration
# Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')

`$delguardPath = '$EXECUTABLE_PATH'

if (Test-Path `$delguardPath) {
    try {
        Remove-Item Alias:del -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:rm -Force -ErrorAction SilentlyContinue  
        Remove-Item Alias:cp -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:copy -Force -ErrorAction SilentlyContinue
    } catch { }
    
    function global:del {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "Usage: del [options] files..." -ForegroundColor Yellow
            Write-Host "Options:" -ForegroundColor Yellow
            Write-Host "  -f, --force     Force delete" -ForegroundColor Gray
            Write-Host "  -r, --recursive Recursive delete" -ForegroundColor Gray
            Write-Host "  -v, --verbose   Verbose output" -ForegroundColor Gray
            Write-Host "  --help          Show help" -ForegroundColor Gray
            return
        }
        & `$delguardPath delete @Arguments
    }
    
    function global:rm {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "Usage: rm [options] files..." -ForegroundColor Yellow
            Write-Host "Options:" -ForegroundColor Yellow
            Write-Host "  -f, --force     Force delete" -ForegroundColor Gray
            Write-Host "  -r, --recursive Recursive delete" -ForegroundColor Gray
            Write-Host "  -v, --verbose   Verbose output" -ForegroundColor Gray
            Write-Host "  --help          Show help" -ForegroundColor Gray
            return
        }
        & `$delguardPath delete @Arguments
    }
    
    function global:cp {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "Usage: cp source target" -ForegroundColor Yellow
            return
        }
        Copy-Item @Arguments
    }
    
    function global:copy {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        if (`$Arguments.Count -eq 0) {
            Write-Host "Usage: copy source target" -ForegroundColor Yellow
            return
        }
        Copy-Item @Arguments
    }
    
    function global:delguard {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        & `$delguardPath @Arguments
    }
    
    if (-not `$global:DelGuardAliasesLoaded) {
        Write-Host 'DelGuard aliases loaded successfully' -ForegroundColor Green
        Write-Host 'Available commands: del, rm, cp, copy, delguard' -ForegroundColor Cyan
        `$global:DelGuardAliasesLoaded = `$true
    }
} else {
    Write-Warning "DelGuard executable not found: `$delguardPath"
}
# DelGuard Configuration End
"@

    # Install to PowerShell profiles
    foreach ($ProfilePath in $ProfilePaths) {
        $ProfileDir = Split-Path $ProfilePath -Parent
        
        if (-not (Test-Path $ProfileDir)) {
            New-Item -ItemType Directory -Path $ProfileDir -Force | Out-Null
            Write-Success "Created profile directory: $ProfileDir"
        }
        
        $ExistingContent = ""
        if (Test-Path $ProfilePath) {
            $ExistingContent = Get-Content $ProfilePath -Raw -ErrorAction SilentlyContinue
            if ($ExistingContent -eq $null) { $ExistingContent = "" }
        }
        
        if ($ExistingContent.Contains("# DelGuard PowerShell Configuration")) {
            if (-not $Force) {
                Write-Warning "DelGuard config exists in: $ProfilePath"
                continue
            }
            $ExistingContent = $ExistingContent -replace '(?s)# DelGuard PowerShell Configuration.*?# DelGuard Configuration End\r?\n?', ''
        }
        
        $NewContent = $ExistingContent + "`n" + $ConfigBlock + "`n"
        Set-Content $ProfilePath $NewContent -Encoding UTF8
        Write-Success "Updated PowerShell profile: $ProfilePath"
    }

    Write-Success ""
    Write-Success "=== Installation Complete ==="
    Write-Info "DelGuard installed successfully!"
    Write-Info "Location: $EXECUTABLE_PATH"
    Write-Info ""
    Write-Info "Available commands:"
    Write-Info "  del <file>      - Safe delete file"
    Write-Info "  rm <file>       - Safe delete file"  
    Write-Info "  cp <src> <dst>  - Copy file"
    Write-Info "  delguard        - DelGuard main program"
    Write-Info ""
    Write-Warning "Please restart PowerShell to load aliases, or run:"
    Write-Warning ". `$PROFILE"

    Write-Info ""
    Write-Info "Testing installation..."
    try {
        $TestResult = & $EXECUTABLE_PATH --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "✓ DelGuard is working correctly"
        } else {
            Write-Warning "⚠ DelGuard may not be working properly"
        }
    } catch {
        Write-Warning "⚠ Could not test DelGuard installation"
    }

} catch {
    Write-Error "Installation failed: $($_.Exception.Message)"
    exit 1
}