# DelGuard Installation Script
param(
    [switch]$Force,
    [switch]$Uninstall,
    [switch]$WhatIf
)

$ErrorActionPreference = "Stop"

function Write-Success { 
    param([string]$Message) 
    Write-Host "Success: $Message" -ForegroundColor Green 
}

function Write-Warning { 
    param([string]$Message) 
    Write-Host "Warning: $Message" -ForegroundColor Yellow 
}

function Write-Error { 
    param([string]$Message) 
    Write-Host "Error: $Message" -ForegroundColor Red 
}

function Write-Info { 
    param([string]$Message) 
    Write-Host "Info: $Message" -ForegroundColor Cyan 
}

function Test-PowerShellVersion {
    $version = $PSVersionTable.PSVersion
    Write-Info "Detected PowerShell version: $($version.ToString())"
    
    if ($version.Major -ge 7) {
        Write-Success "PowerShell 7.x compatible"
        return "pwsh"
    } elseif ($version.Major -eq 5 -and $version.Minor -ge 1) {
        Write-Success "PowerShell 5.1 compatible"
        return "powershell"
    } else {
        Write-Error "Unsupported PowerShell version: $version"
        exit 1
    }
}

function Build-DelGuard {
    Write-Info "Building DelGuard..."
    
    if (-not (Test-Path "go.mod")) {
        Write-Error "go.mod not found. Please run this script in DelGuard project root directory"
        exit 1
    }
    
    if ($WhatIf) {
        Write-Info "[Preview] Will execute: go build -o delguard.exe"
        return $true
    }
    
    try {
        $output = & go build -o delguard.exe 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Build failed: $output"
            return $false
        }
        Write-Success "Build completed"
        return $true
    } catch {
        Write-Error "Build error: $($_.Exception.Message)"
        return $false
    }
}

function Install-DelGuard {
    # Try Program Files first, fallback to user directory
    $installPath = Join-Path $env:ProgramFiles "DelGuard"
    $userInstallPath = Join-Path $env:LOCALAPPDATA "DelGuard"
    
    Write-Info "Attempting to install DelGuard to: $installPath"
    
    if ($WhatIf) {
        Write-Info "[Preview] Will try to create directory: $installPath"
        Write-Info "[Preview] If failed, will use: $userInstallPath"
        Write-Info "[Preview] Will copy file: delguard.exe"
        Write-Info "[Preview] Will add to PATH environment variable"
        return $true
    }
    
    # Try system-wide installation first
    try {
        if (-not (Test-Path $installPath)) {
            New-Item -ItemType Directory -Path $installPath -Force | Out-Null
            Write-Success "Created system install directory: $installPath"
        }
        $finalInstallPath = $installPath
    } catch {
        Write-Warning "Cannot install to system directory (admin rights required)"
        Write-Info "Falling back to user directory: $userInstallPath"
        
        try {
            if (-not (Test-Path $userInstallPath)) {
                New-Item -ItemType Directory -Path $userInstallPath -Force | Out-Null
                Write-Success "Created user install directory: $userInstallPath"
            }
            $finalInstallPath = $userInstallPath
        } catch {
            Write-Error "Cannot create user install directory: $($_.Exception.Message)"
            return $false
        }
    }
    
    $targetPath = Join-Path $finalInstallPath "delguard.exe"
    
    try {
        if (Test-Path "delguard.exe") {
            Copy-Item "delguard.exe" $targetPath -Force
            Write-Success "Copied executable to: $targetPath"
        } else {
            Write-Error "Executable not found: delguard.exe"
            return $false
        }
        
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($currentPath -notlike "*$finalInstallPath*") {
            $newPath = "$currentPath;$finalInstallPath"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-Success "Added to user PATH environment variable: $finalInstallPath"
        }
        
        return $true
    } catch {
        Write-Error "Installation error: $($_.Exception.Message)"
        return $false
    }
}

function Install-PowerShellAliases {
    $profilePath = $PROFILE.CurrentUserAllHosts
    Write-Info "Configuring PowerShell aliases: $profilePath"
    
    $aliasContent = @'

# DelGuard Safe Delete Tool Aliases
if (-not $env:DELGUARD_LOADED) {
    $env:DELGUARD_LOADED = "1"
    Write-Host "DelGuard Safe Delete Tool Enabled" -ForegroundColor Green
    Write-Host "del: Safe delete (Windows style)" -ForegroundColor Gray
    Write-Host "rm:  Safe delete (Unix style)" -ForegroundColor Gray  
    Write-Host "cp:  Safe copy" -ForegroundColor Gray
    Write-Host "Use --help for detailed help" -ForegroundColor Gray
}

function DelGuard-Delete { delguard @args }
function DelGuard-Copy { delguard cp @args }

Set-Alias -Name del -Value DelGuard-Delete -Force
Set-Alias -Name rm -Value DelGuard-Delete -Force  
Set-Alias -Name cp -Value DelGuard-Copy -Force
'@

    if ($WhatIf) {
        Write-Info "[Preview] Will add aliases to: $profilePath"
        return $true
    }
    
    try {
        $profileDir = Split-Path $profilePath -Parent
        if (-not (Test-Path $profileDir)) {
            New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
        }
        
        if (Test-Path $profilePath) {
            $existingContent = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
            if ($existingContent -like "*DELGUARD_LOADED*") {
                if (-not $Force) {
                    Write-Warning "DelGuard aliases already exist, use -Force to override"
                    return $true
                }
            }
        }
        
        Add-Content $profilePath $aliasContent
        Write-Success "PowerShell aliases configured"
        return $true
    } catch {
        Write-Error "Alias configuration error: $($_.Exception.Message)"
        return $false
    }
}

function Uninstall-DelGuard {
    Write-Info "Uninstalling DelGuard..."
    
    $installPath = Join-Path $env:ProgramFiles "DelGuard"
    $targetPath = Join-Path $installPath "delguard.exe"
    
    if ($WhatIf) {
        Write-Info "[Preview] Will delete file: $targetPath"
        Write-Info "[Preview] Will remove from PATH: $installPath"
        return
    }
    
    try {
        if (Test-Path $targetPath) {
            Remove-Item $targetPath -Force
            Write-Success "Deleted: $targetPath"
        }
        
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($currentPath -like "*$installPath*") {
            $newPath = $currentPath -replace [regex]::Escape(";$installPath"), ""
            $newPath = $newPath -replace [regex]::Escape("$installPath;"), ""
            $newPath = $newPath -replace [regex]::Escape($installPath), ""
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-Success "Removed from PATH environment variable"
        }
        
        Write-Success "DelGuard uninstallation completed"
    } catch {
        Write-Error "Uninstallation error: $($_.Exception.Message)"
    }
}

# Main execution
Write-Host "=== DelGuard Installation Script ===" -ForegroundColor Magenta

if ($Uninstall) {
    Uninstall-DelGuard
    exit 0
}

$shell = Test-PowerShellVersion

if (-not (Build-DelGuard)) {
    Write-Error "Build failed, installation aborted"
    exit 1
}

if (-not (Install-DelGuard)) {
    Write-Error "Installation failed"
    exit 1
}

if (-not (Install-PowerShellAliases)) {
    Write-Warning "PowerShell alias configuration failed, but DelGuard installed successfully"
}

if (-not $WhatIf) {
    Write-Success "DelGuard installation completed!"
    Write-Info "Please restart PowerShell or run the following command to load aliases:"
    Write-Host ". `$PROFILE" -ForegroundColor Yellow
    Write-Info "Then you can use these commands:"
    Write-Host "del filename   # Safe delete file" -ForegroundColor Green
    Write-Host "rm filename    # Safe delete file (Unix style)" -ForegroundColor Green
    Write-Host "cp source dest # Safe copy file" -ForegroundColor Green
} else {
    Write-Info "Preview mode completed. Run without -WhatIf to perform actual installation"
}