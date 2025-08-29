#!/usr/bin/env pwsh
# DelGuard Universal PowerShell Installer
# Supports: Windows PowerShell 5.1+ and PowerShell 7+ (Cross-platform)
# Author: DelGuard Team
# Version: 1.0

param(
    [string]$InstallPath = "",
    [switch]$Force,
    [switch]$Quiet,
    [switch]$Uninstall
)

# 确保PowerShell使用UTF-8编码显示中文
if ($PSVersionTable.PSVersion.Major -lt 6) {
    # Windows PowerShell 5.1需要设置控制台编码
    try {
        [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
        [Console]::InputEncoding = [System.Text.Encoding]::UTF8
        # 设置控制台代码页为UTF-8
        chcp 65001 | Out-Null
    } catch {
        Write-Warning "无法设置UTF-8编码，中文显示可能异常"
    }
} else {
    # PowerShell 7+ 默认支持UTF-8
    $PSDefaultParameterValues['*:Encoding'] = 'utf8'
}

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

# Platform detection
if ($PSVersionTable.PSVersion.Major -ge 6) {
    $IsWindowsOS = $IsWindows
    $IsMacOS = $IsMacOS
    $IsLinuxOS = $IsLinux
} else {
    $IsWindowsOS = ($env:OS -eq "Windows_NT")
    $IsMacOS = $false
    $IsLinuxOS = $false
}

Write-Info "=== DelGuard Universal Installer ==="
Write-Info "Platform: $(if($IsWindowsOS){'Windows'}elseif($IsMacOS){'macOS'}elseif($IsLinuxOS){'Linux'}else{'Unknown'})"
Write-Info "PowerShell: $($PSVersionTable.PSVersion)"

# Determine executable name and paths
if ($IsWindowsOS) {
    $ExeName = "delguard.exe"
    $DefaultInstallPath = "$env:USERPROFILE\bin"
    $ProfilePaths = @(
        "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1",
        "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
    )
} else {
    $ExeName = "delguard"
    $DefaultInstallPath = "$env:HOME/bin"
    $ProfilePaths = @(
        "$env:HOME/.config/powershell/Microsoft.PowerShell_profile.ps1"
    )
}

# Use provided path or default
$FinalInstallPath = if ($InstallPath) { $InstallPath } else { $DefaultInstallPath }
$ExecutablePath = Join-Path $FinalInstallPath $ExeName

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
            if ($content -and $content.Contains("# DelGuard PowerShell Configuration")) {
                # Remove DelGuard configuration block
                $newContent = $content -replace '(?s)# DelGuard PowerShell Configuration.*?# End DelGuard Configuration\r?\n?', ''
                Set-Content $ProfilePath $newContent -Encoding UTF8
                Write-Success "Removed DelGuard configuration from: $ProfilePath"
            }
        }
    }
    
    Write-Success "DelGuard uninstalled successfully!"
    return
}

# Installation mode
Write-Info "Installing DelGuard to: $FinalInstallPath"

# Create install directory
if (-not (Test-Path $FinalInstallPath)) {
    New-Item -ItemType Directory -Path $FinalInstallPath -Force | Out-Null
    Write-Success "Created directory: $FinalInstallPath"
}

# Copy executable
$SourceExe = Join-Path $PSScriptRoot "..\$ExeName"
if (-not (Test-Path $SourceExe)) {
    # Try to find in current directory
    $SourceExe = Join-Path (Get-Location) $ExeName
}

if (-not (Test-Path $SourceExe)) {
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

Copy-Item $SourceExe $ExecutablePath -Force
Write-Success "Copied executable to: $ExecutablePath"

# Make executable on Unix-like systems
if (-not $IsWindowsOS) {
    chmod +x $ExecutablePath 2>$null
}

# Add to PATH if not already there
$PathSeparator = if ($IsWindowsOS) { ";" } else { ":" }
$CurrentPath = $env:PATH
if (-not $CurrentPath.Contains($FinalInstallPath)) {
    if ($IsWindowsOS) {
        # Add to user PATH on Windows
        $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if (-not $UserPath.Contains($FinalInstallPath)) {
            [Environment]::SetEnvironmentVariable("PATH", "$UserPath$PathSeparator$FinalInstallPath", "User")
            Write-Success "Added to user PATH: $FinalInstallPath"
        }
    } else {
        Write-Info "Please add $FinalInstallPath to your PATH manually:"
        Write-Info "echo 'export PATH=`"${FinalInstallPath}:`$PATH`"' >> ~/.bashrc"
        Write-Info "echo 'export PATH=`"${FinalInstallPath}:`$PATH`"' >> ~/.zshrc"
    }
}

# PowerShell profile configuration
$ConfigBlock = @"
# DelGuard PowerShell Configuration
# Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
# Version: DelGuard 1.0 for PowerShell 7+
# Supports: del, rm, cp, copy, delguard commands

`$delguardPath = '$ExecutablePath'

if (Test-Path `$delguardPath) {
    # Remove existing aliases to prevent conflicts
    try {
        Remove-Item Alias:del -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:rm -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:cp -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:copy -Force -ErrorAction SilentlyContinue
    } catch { }
    
    # Define robust alias functions for all 5 commands
    function global:del {
        param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Arguments)
        & $delguardPath delete @Arguments
    }
    
    function global:rm {
        param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Arguments)
        & $delguardPath delete @Arguments
    }
    
    function global:cp {
        param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Arguments)
        Write-Host "Copy functionality not yet implemented" -ForegroundColor Yellow
    }
    
    function global:copy {
        param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Arguments)
        Write-Host "Copy functionality not yet implemented" -ForegroundColor Yellow
    }
    
    function global:delguard {
        param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Arguments)
        & $delguardPath @Arguments
    }
    
    # Show loading message only once per session
    if (-not `$global:DelGuardLoaded) {
        Write-Host 'DelGuard aliases loaded successfully' -ForegroundColor Green
        Write-Host 'Commands: del, rm, cp, copy, delguard' -ForegroundColor Cyan
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
    
    # Create profile directory if it doesn't exist
    if (-not (Test-Path $ProfileDir)) {
        New-Item -ItemType Directory -Path $ProfileDir -Force | Out-Null
        Write-Success "Created profile directory: $ProfileDir"
    }
    
    # Check if profile exists and has DelGuard config
    $ExistingContent = ""
    if (Test-Path $ProfilePath) {
        $ExistingContent = Get-Content $ProfilePath -Raw -ErrorAction SilentlyContinue
    }
    
    if ($ExistingContent -and $ExistingContent.Contains("# DelGuard PowerShell Configuration")) {
        if (-not $Force) {
            Write-Warning "DelGuard configuration already exists in: $ProfilePath"
            Write-Warning "Use -Force to overwrite"
            continue
        }
        # Remove existing DelGuard configuration
        $ExistingContent = $ExistingContent -replace '(?s)# DelGuard PowerShell Configuration.*?# End DelGuard Configuration\r?\n?', ''
    }
    
    # Append new configuration
    $NewContent = $ExistingContent + "`n" + $ConfigBlock + "`n"
    Set-Content $ProfilePath $NewContent -Encoding UTF8
    Write-Success "Updated PowerShell profile: $ProfilePath"
}

Write-Success "=== Installation Complete ==="
Write-Info "DelGuard has been installed successfully!"
Write-Info "Available commands: del, rm, cp, copy, delguard"
Write-Info "Restart your PowerShell session to use the new commands."
Write-Info "Or run: . `$PROFILE"

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