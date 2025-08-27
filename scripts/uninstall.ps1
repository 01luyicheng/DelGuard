#!/usr/bin/env pwsh
# DelGuard Universal PowerShell Uninstaller
# Supports: Windows PowerShell 5.1+ and PowerShell 7+ (Cross-platform)
# Author: DelGuard Team
# Version: 1.0

param(
    [string]$InstallPath = "",
    [switch]$Force,
    [switch]$Quiet
)

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

Write-Info "=== DelGuard Universal Uninstaller ==="
Write-Info "Platform: $(if($IsWindowsOS){'Windows'}elseif($IsMacOS){'macOS'}elseif($IsLinuxOS){'Linux'}else{'Unknown'})"

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

# Use provided path or search common locations
$SearchPaths = @()
if ($InstallPath) {
    $SearchPaths += $InstallPath
}
$SearchPaths += @(
    $DefaultInstallPath,
    "/usr/local/bin",
    "/usr/bin",
    "$env:HOME/.local/bin"
)

# Find DelGuard installations
$FoundInstallations = @()
foreach ($SearchPath in $SearchPaths) {
    $ExecutablePath = Join-Path $SearchPath $ExeName
    if (Test-Path $ExecutablePath) {
        $FoundInstallations += $ExecutablePath
    }
}

if ($FoundInstallations.Count -eq 0) {
    Write-Warning "DelGuard installation not found in common locations."
    Write-Info "Searched paths:"
    foreach ($Path in $SearchPaths) {
        Write-Info "  - $Path"
    }
    
    if (-not $Force) {
        Write-Info "Use -Force to clean configuration files anyway."
        exit 1
    }
} else {
    Write-Info "Found DelGuard installations:"
    foreach ($Installation in $FoundInstallations) {
        Write-Info "  - $Installation"
    }
}

# Remove executables
$RemovedCount = 0
foreach ($ExecutablePath in $FoundInstallations) {
    try {
        Remove-Item $ExecutablePath -Force
        Write-Success "Removed executable: $ExecutablePath"
        $RemovedCount++
    } catch {
        Write-Error "Failed to remove: $ExecutablePath - $($_.Exception.Message)"
    }
}

# Remove from PowerShell profiles
$ProfilesUpdated = 0
foreach ($ProfilePath in $ProfilePaths) {
    if (Test-Path $ProfilePath) {
        try {
            $content = Get-Content $ProfilePath -Raw -ErrorAction SilentlyContinue
            if ($content -and $content.Contains("# DelGuard PowerShell Configuration")) {
                # Remove DelGuard configuration block
                $newContent = $content -replace '(?s)# DelGuard PowerShell Configuration.*?# End DelGuard Configuration\r?\n?', ''
                
                # Remove empty lines at the end
                $newContent = $newContent.TrimEnd()
                
                Set-Content $ProfilePath $newContent -Encoding UTF8
                Write-Success "Removed DelGuard configuration from: $ProfilePath"
                $ProfilesUpdated++
            }
        } catch {
            Write-Error "Failed to update profile: $ProfilePath - $($_.Exception.Message)"
        }
    }
}

# Remove from system PATH (Windows only)
if ($IsWindowsOS -and $FoundInstallations.Count -gt 0) {
    try {
        $InstallDir = Split-Path $FoundInstallations[0] -Parent
        $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        
        if ($UserPath -and $UserPath.Contains($InstallDir)) {
            $PathEntries = $UserPath.Split(';') | Where-Object { $_ -ne $InstallDir -and $_ -ne "" }
            $NewPath = $PathEntries -join ';'
            [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
            Write-Success "Removed from user PATH: $InstallDir"
        }
    } catch {
        Write-Warning "Could not update PATH environment variable: $($_.Exception.Message)"
    }
}

# Remove configuration directory (optional)
$ConfigPaths = @()
if ($IsWindowsOS) {
    $ConfigPaths += "$env:APPDATA\DelGuard"
} else {
    $ConfigPaths += "$env:HOME/.config/delguard"
}

foreach ($ConfigPath in $ConfigPaths) {
    if (Test-Path $ConfigPath) {
        if ($Force) {
            try {
                Remove-Item $ConfigPath -Recurse -Force
                Write-Success "Removed configuration directory: $ConfigPath"
            } catch {
                Write-Warning "Could not remove configuration directory: $ConfigPath - $($_.Exception.Message)"
            }
        } else {
            Write-Info "Configuration directory found: $ConfigPath"
            Write-Info "Use -Force to remove configuration files"
        }
    }
}

# Summary
Write-Info ""
Write-Info "=== Uninstallation Summary ==="
Write-Info "Executables removed: $RemovedCount"
Write-Info "Profiles updated: $ProfilesUpdated"

if ($RemovedCount -gt 0 -or $ProfilesUpdated -gt 0) {
    Write-Success "DelGuard has been successfully uninstalled!"
    Write-Info "Please restart your PowerShell session to complete the removal."
} else {
    Write-Warning "No DelGuard components were found or removed."
}

if (-not $Force -and (Test-Path $ConfigPaths[0] -ErrorAction SilentlyContinue)) {
    Write-Info ""
    Write-Info "Note: Configuration files were preserved."
    Write-Info "Run with -Force to remove all DelGuard data."
}