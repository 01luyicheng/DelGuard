#!/usr/bin/env pwsh
# DelGuard Universal Cross-Platform Installer
# Works on Windows, macOS, and Linux with PowerShell 7+
# Author: DelGuard Team
# Version: 1.0

param(
    [string]$InstallPath = "",
    [switch]$Force,
    [switch]$Quiet,
    [switch]$Uninstall,
    [switch]$Help
)

# Cross-platform compatibility
$IsWindowsOS = $PSVersionTable.PSVersion.Major -ge 6 ? $IsWindows : ($env:OS -eq "Windows_NT")
$IsMacOS = $PSVersionTable.PSVersion.Major -ge 6 ? $IsMacOS : $false
$IsLinuxOS = $PSVersionTable.PSVersion.Major -ge 6 ? $IsLinux : $false

# Color output functions
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    if (-not $Quiet) {
        switch ($Color) {
            "Red"     { Write-Host $Message -ForegroundColor Red }
            "Green"   { Write-Host $Message -ForegroundColor Green }
            "Yellow"  { Write-Host $Message -ForegroundColor Yellow }
            "Cyan"    { Write-Host $Message -ForegroundColor Cyan }
            "Magenta" { Write-Host $Message -ForegroundColor Magenta }
            default   { Write-Host $Message }
        }
    }
}

function Write-Success { param([string]$Message) Write-ColorOutput $Message "Green" }
function Write-Warning { param([string]$Message) Write-ColorOutput $Message "Yellow" }
function Write-Error { param([string]$Message) Write-ColorOutput $Message "Red" }
function Write-Info { param([string]$Message) Write-ColorOutput $Message "Cyan" }

function Show-Help {
    Write-Host @"
DelGuard Universal Cross-Platform Installer

Usage: pwsh -File universal_install.ps1 [OPTIONS]

Options:
    -InstallPath PATH     Install to specific path
    -Force               Force overwrite existing installation
    -Quiet               Suppress output messages
    -Uninstall           Remove DelGuard installation
    -Help                Show this help message

Default Install Paths:
    Windows: `$env:USERPROFILE\bin
    macOS:   `$HOME/.local/bin
    Linux:   `$HOME/.local/bin

Examples:
    pwsh -File universal_install.ps1
    pwsh -File universal_install.ps1 -Force
    pwsh -File universal_install.ps1 -InstallPath "/usr/local/bin"
    pwsh -File universal_install.ps1 -Uninstall

"@
}

if ($Help) {
    Show-Help
    exit 0
}

# Determine platform-specific settings
if ($IsWindowsOS) {
    $ExeName = "delguard.exe"
    $DefaultInstallPath = "$env:USERPROFILE\bin"
    $PathSeparator = ";"
    $ProfilePaths = @(
        "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1",
        "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
    )
} else {
    $ExeName = "delguard"
    $DefaultInstallPath = "$env:HOME/.local/bin"
    $PathSeparator = ":"
    $ProfilePaths = @(
        "$env:HOME/.config/powershell/Microsoft.PowerShell_profile.ps1"
    )
}

$FinalInstallPath = if ($InstallPath) { $InstallPath } else { $DefaultInstallPath }
$ExecutablePath = Join-Path $FinalInstallPath $ExeName

Write-Info "=== DelGuard Universal Installer ==="
Write-Info "Platform: $(if($IsWindowsOS){'Windows'}elseif($IsMacOS){'macOS'}elseif($IsLinuxOS){'Linux'}else{'Unknown'})"
Write-Info "PowerShell: $($PSVersionTable.PSVersion)"
Write-Info "Install Path: $FinalInstallPath"

# Uninstall mode
if ($Uninstall) {
    Write-Info "Uninstalling DelGuard..."
    
    # Remove executable
    if (Test-Path $ExecutablePath) {
        Remove-Item $ExecutablePath -Force
        Write-Success "Removed executable: $ExecutablePath"
    }
    
    # Remove from PowerShell profiles
    foreach ($ProfilePath in $ProfilePaths) {
        if (Test-Path $ProfilePath) {
            $content = Get-Content $ProfilePath -Raw -ErrorAction SilentlyContinue
            if ($content -and $content.Contains("# DelGuard Configuration")) {
                $newContent = $content -replace '(?s)# DelGuard Configuration.*?# End DelGuard Configuration\r?\n?', ''
                Set-Content $ProfilePath $newContent -Encoding UTF8
                Write-Success "Removed DelGuard configuration from: $ProfilePath"
            }
        }
    }
    
    # Remove from shell configs on Unix-like systems
    if (-not $IsWindowsOS) {
        $ShellConfigs = @("$env:HOME/.bashrc", "$env:HOME/.zshrc", "$env:HOME/.profile")
        foreach ($config in $ShellConfigs) {
            if (Test-Path $config) {
                $content = Get-Content $config -Raw -ErrorAction SilentlyContinue
                if ($content -and $content.Contains("# DelGuard Configuration")) {
                    $newContent = $content -replace '(?s)# DelGuard Configuration.*?# End DelGuard Configuration\r?\n?', ''
                    Set-Content $config $newContent -Encoding UTF8
                    Write-Success "Removed DelGuard configuration from: $config"
                }
            }
        }
    }
    
    Write-Success "DelGuard uninstalled successfully!"
    exit 0
}

# Installation mode
Write-Info "Installing DelGuard..."

# Create install directory
if (-not (Test-Path $FinalInstallPath)) {
    New-Item -ItemType Directory -Path $FinalInstallPath -Force | Out-Null
    Write-Success "Created directory: $FinalInstallPath"
}

# Find source executable
$SourceExe = $null
$PossibleSources = @(
    (Join-Path $PSScriptRoot ".." $ExeName),
    (Join-Path $PSScriptRoot ".." "build" $ExeName),
    (Join-Path (Get-Location) $ExeName)
)

foreach ($source in $PossibleSources) {
    if (Test-Path $source) {
        $SourceExe = $source
        break
    }
}

if (-not $SourceExe) {
    Write-Error "DelGuard executable not found. Please build the project first."
    Write-Error "Run: go build -o $ExeName"
    exit 1
}

# Check if already installed
if ((Test-Path $ExecutablePath) -and -not $Force) {
    Write-Warning "DelGuard is already installed at: $ExecutablePath"
    Write-Warning "Use -Force to overwrite"
    exit 1
}

# Copy executable
Copy-Item $SourceExe $ExecutablePath -Force
Write-Success "Copied executable to: $ExecutablePath"

# Make executable on Unix-like systems
if (-not $IsWindowsOS) {
    & chmod +x $ExecutablePath 2>$null
}

# PowerShell configuration
$ConfigBlock = @"
# DelGuard Configuration
# Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
# Version: DelGuard 1.0 Cross-Platform

`$delguardPath = '$ExecutablePath'

if (Test-Path `$delguardPath) {
    # Remove existing aliases to prevent conflicts
    try {
        Remove-Item Alias:del -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:rm -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:cp -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:copy -Force -ErrorAction SilentlyContinue
    } catch { }
    
    # Define cross-platform alias functions
    function global:del {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        & `$delguardPath -i @Arguments
    }
    
    function global:rm {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        & `$delguardPath -i @Arguments
    }
    
    function global:cp {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        & `$delguardPath --cp @Arguments
    }
    
    function global:delguard {
        param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
        & `$delguardPath @Arguments
    }
    
    # Show loading message only once per session
    if (-not `$global:DelGuardLoaded) {
        Write-Host 'DelGuard loaded successfully' -ForegroundColor Green
        Write-Host 'Commands: del, rm, cp, delguard' -ForegroundColor Cyan
        Write-Host 'Use --help for detailed help' -ForegroundColor Gray
        `$global:DelGuardLoaded = `$true
    }
} else {
    Write-Warning "DelGuard executable not found: `$delguardPath"
}
# End DelGuard Configuration
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
    }
    
    if ($ExistingContent -and $ExistingContent.Contains("# DelGuard Configuration")) {
        if (-not $Force) {
            Write-Warning "DelGuard configuration already exists in: $ProfilePath"
            Write-Warning "Use -Force to overwrite"
            continue
        }
        $ExistingContent = $ExistingContent -replace '(?s)# DelGuard Configuration.*?# End DelGuard Configuration\r?\n?', ''
    }
    
    $NewContent = $ExistingContent + "`n" + $ConfigBlock + "`n"
    Set-Content $ProfilePath $NewContent -Encoding UTF8
    Write-Success "Updated PowerShell profile: $ProfilePath"
}

# Add to PATH
$CurrentPath = $env:PATH
if (-not $CurrentPath.Contains($FinalInstallPath)) {
    if ($IsWindowsOS) {
        $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if (-not $UserPath.Contains($FinalInstallPath)) {
            [Environment]::SetEnvironmentVariable("PATH", "$UserPath$PathSeparator$FinalInstallPath", "User")
            Write-Success "Added to user PATH: $FinalInstallPath"
        }
    } else {
        Write-Info "Please add $FinalInstallPath to your PATH manually:"
        Write-Info "echo 'export PATH=\"$FinalInstallPath:\$PATH\"' >> ~/.bashrc"
        Write-Info "echo 'export PATH=\"$FinalInstallPath:\$PATH\"' >> ~/.zshrc"
    }
}

Write-Success "=== Installation Complete ==="
Write-Info "DelGuard has been installed successfully!"
Write-Info "Available commands: del, rm, cp, delguard"
Write-Info "Restart your shell session to use the new commands."

# Test installation
Write-Info "Testing installation..."
try {
    $TestResult = & $ExecutablePath --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Success "✓ DelGuard is working correctly"
    } else {
        Write-Warning "⚠ DelGuard may not be working properly"
    }
} catch {
    Write-Warning "⚠ Could not test DelGuard installation"
}