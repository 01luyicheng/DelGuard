#!/usr/bin/env pwsh
# DelGuard Enhanced Universal Installer for Windows (Fixed Version)
# Compatible with PowerShell 5.1+ and PowerShell 7+
# Version: 2.1.1 (Fixed)

[CmdletBinding()]
param(
    [string]$InstallPath = "",
    [switch]$Force,
    [switch]$Quiet,
    [switch]$Uninstall,
    [switch]$Help,
    [switch]$SystemWide,
    [switch]$SkipAliases,
    [switch]$CheckOnly,
    [string]$Version = "latest"
)

# Set strict error handling
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Global variables
$Script:InstallationLog = @()

# Enhanced color output functions with fallback
function Write-ColorOutput {
    param(
        [string]$Message, 
        [string]$Color = "White",
        [switch]$NoNewline
    )
    
    if (-not $Quiet) {
        try {
            $params = @{
                Object = $Message
                ForegroundColor = $Color
            }
            if ($NoNewline) { $params.NoNewline = $true }
            Write-Host @params
        } catch {
            # Fallback for systems without color support
            Write-Host $Message
        }
    }
    
    # Add to log
    $Script:InstallationLog += "$(Get-Date -Format 'HH:mm:ss') [$Color] $Message"
}

function Write-Success { 
    param([string]$Message) 
    Write-ColorOutput "✓ $Message" "Green" 
}

function Write-Warning { 
    param([string]$Message) 
    Write-ColorOutput "⚠ $Message" "Yellow" 
}

function Write-Error { 
    param([string]$Message) 
    Write-ColorOutput "✗ $Message" "Red" 
}

function Write-Info { 
    param([string]$Message) 
    Write-ColorOutput "ℹ $Message" "Cyan" 
}

function Write-Header { 
    param([string]$Message) 
    Write-ColorOutput $Message "Magenta" 
}

function Show-Help {
    Write-Host @"
DelGuard Enhanced Universal Installer for Windows (Fixed)

USAGE:
    .\install.ps1 [OPTIONS]

OPTIONS:
    -InstallPath <path>    Install to specific directory (default: `$env:USERPROFILE\bin)
    -Force                 Force overwrite existing installation
    -Quiet                 Suppress output messages
    -Uninstall            Remove DelGuard installation
    -SystemWide           Install system-wide (requires admin rights)
    -SkipAliases          Skip PowerShell alias configuration
    -CheckOnly            Only check installation status
    -Version <version>    Install specific version (default: latest)
    -Help                 Show this help message

EXAMPLES:
    .\install.ps1                                    # Install to default location
    .\install.ps1 -InstallPath "C:\tools"           # Install to custom location
    .\install.ps1 -Force                            # Force reinstall
    .\install.ps1 -SystemWide                       # Install system-wide
    .\install.ps1 -SkipAliases                      # Install without aliases
    .\install.ps1 -CheckOnly                        # Check installation status
    .\install.ps1 -Uninstall                        # Remove DelGuard

AFTER INSTALLATION:
    Restart PowerShell and use these commands:
    - del <file>     # Safe delete (replaces Windows del)
    - rm <file>      # Safe delete (Unix style)
    - cp <src> <dst> # Safe copy
    - delguard --help # Full help

"@
}

function Test-Administrator {
    try {
        $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
        $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
        return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    } catch {
        Write-Warning "Could not determine administrator status: $($_.Exception.Message)"
        return $false
    }
}

function Test-ExecutionPolicy {
    try {
        $policy = Get-ExecutionPolicy -Scope CurrentUser
        if ($policy -eq "Restricted") {
            Write-Warning "PowerShell execution policy is Restricted"
            Write-Info "You may need to run: Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser"
            return $false
        }
        return $true
    } catch {
        Write-Warning "Could not check execution policy: $($_.Exception.Message)"
        return $true  # Assume it's OK if we can't check
    }
}

function Get-PowerShellProfiles {
    $profiles = @()
    
    # PowerShell 7+ profiles
    if ($PSVersionTable.PSVersion.Major -ge 7) {
        $profiles += @(
            "$env:USERPROFILE\Documents\PowerShell\Microsoft.PowerShell_profile.ps1",
            "$env:USERPROFILE\Documents\PowerShell\Profile.ps1"
        )
    }
    
    # Windows PowerShell 5.1 profiles
    if ($PSVersionTable.PSVersion.Major -eq 5) {
        $profiles += @(
            "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1",
            "$env:USERPROFILE\Documents\WindowsPowerShell\Profile.ps1"
        )
    }
    
    # Built-in profile variables (more reliable)
    try {
        if ($PROFILE.CurrentUserCurrentHost) { $profiles += $PROFILE.CurrentUserCurrentHost }
        if ($PROFILE.CurrentUserAllHosts) { $profiles += $PROFILE.CurrentUserAllHosts }
    } catch {
        Write-Warning "Could not access `$PROFILE variables"
    }
    
    return ($profiles | Sort-Object -Unique | Where-Object { $_ })
}

function Test-InstallationStatus {
    Write-Info "Checking DelGuard installation status..."
    
    # Check common installation paths
    $CommonPaths = @(
        "$env:USERPROFILE\bin\delguard.exe",
        "$env:ProgramFiles\DelGuard\delguard.exe",
        "$env:LOCALAPPDATA\DelGuard\delguard.exe"
    )
    
    $FoundInstallations = @()
    foreach ($Path in $CommonPaths) {
        if (Test-Path $Path -ErrorAction SilentlyContinue) {
            $FoundInstallations += $Path
        }
    }
    
    if ($FoundInstallations.Count -eq 0) {
        Write-Warning "DelGuard is not installed"
        return $false
    }
    
    Write-Success "Found DelGuard installations:"
    foreach ($Installation in $FoundInstallations) {
        try {
            $VersionOutput = & $Installation --version 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Info "  $Installation - $VersionOutput"
            } else {
                Write-Info "  $Installation - Version check failed"
            }
        } catch {
            Write-Info "  $Installation - Version unknown"
        }
    }
    
    # Check PATH
    try {
        $DelGuardCommand = Get-Command delguard -ErrorAction SilentlyContinue
        if ($DelGuardCommand) {
            Write-Success "DelGuard found in PATH: $($DelGuardCommand.Source)"
        } else {
            Write-Warning "DelGuard not found in PATH"
        }
    } catch {
        Write-Warning "Could not check PATH for DelGuard"
    }
    
    # Check PowerShell aliases
    $ProfilePaths = Get-PowerShellProfiles
    $AliasConfigured = $false
    foreach ($ProfilePath in $ProfilePaths) {
        if (Test-Path $ProfilePath -ErrorAction SilentlyContinue) {
            try {
                $Content = Get-Content $ProfilePath -Raw -ErrorAction SilentlyContinue
                if ($Content -and $Content.Contains("DelGuard")) {
                    Write-Success "DelGuard aliases configured in: $ProfilePath"
                    $AliasConfigured = $true
                    break
                }
            } catch {
                Write-Warning "Could not read profile: $ProfilePath"
            }
        }
    }
    
    if (-not $AliasConfigured) {
        Write-Warning "DelGuard aliases not configured"
    }
    
    return $true
}

function Build-DelGuard {
    Write-Info "Building DelGuard executable..."
    
    if (-not (Test-Path "go.mod" -ErrorAction SilentlyContinue)) {
        Write-Error "go.mod not found. Please run this script from DelGuard project root."
        return $false
    }
    
    # Check if executable already exists
    if (Test-Path "delguard.exe" -ErrorAction SilentlyContinue) {
        Write-Success "DelGuard executable already exists"
        return $true
    }
    
    # Check if Go is installed
    try {
        $null = Get-Command go -ErrorAction Stop
    } catch {
        Write-Error "Go is not installed. Please install Go first."
        Write-Info "Visit: https://golang.org/dl/"
        return $false
    }
    
    try {
        # Set build environment
        $env:CGO_ENABLED = "0"
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        
        Write-Info "Building DelGuard executable..."
        
        # Try to build with main.go first
        if (Test-Path "main.go" -ErrorAction SilentlyContinue) {
            Write-Info "Building with main.go..."
            $buildOutput = & go build -ldflags "-s -w" -o "delguard.exe" "main.go" 2>&1
        } else {
            Write-Info "Building entire project..."
            $buildOutput = & go build -ldflags "-s -w" -o "delguard.exe" 2>&1
        }
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Build failed: $buildOutput"
            return $false
        }
        
        if (-not (Test-Path "delguard.exe" -ErrorAction SilentlyContinue)) {
            Write-Error "Build completed but executable not found."
            return $false
        }
        
        Write-Success "DelGuard executable built successfully"
        return $true
    } catch {
        Write-Error "Build error: $($_.Exception.Message)"
        return $false
    }
}

function Install-DelGuard {
    # Determine install path
    if ($SystemWide) {
        if (-not (Test-Administrator)) {
            Write-Error "System-wide installation requires administrator privileges."
            Write-Info "Please run PowerShell as Administrator or use user installation."
            return $false
        }
        $FinalInstallPath = if ($InstallPath) { $InstallPath } else { "$env:ProgramFiles\DelGuard" }
    } else {
        $FinalInstallPath = if ($InstallPath) { $InstallPath } else { "$env:USERPROFILE\bin" }
    }
    
    $ExecutablePath = Join-Path $FinalInstallPath "delguard.exe"
    
    Write-Info "Installing DelGuard to: $FinalInstallPath"
    
    # Create install directory
    if (-not (Test-Path $FinalInstallPath -ErrorAction SilentlyContinue)) {
        try {
            $null = New-Item -ItemType Directory -Path $FinalInstallPath -Force -ErrorAction Stop
            Write-Success "Created install directory: $FinalInstallPath"
        } catch {
            Write-Error "Failed to create install directory: $($_.Exception.Message)"
            return $false
        }
    }
    
    # Check if already installed
    if ((Test-Path $ExecutablePath -ErrorAction SilentlyContinue) -and -not $Force) {
        Write-Warning "DelGuard is already installed at: $ExecutablePath"
        Write-Warning "Use -Force to overwrite existing installation"
        return $false
    }
    
    # Copy executable
    try {
        Copy-Item "delguard.exe" $ExecutablePath -Force -ErrorAction Stop
        Write-Success "Installed executable to: $ExecutablePath"
    } catch {
        Write-Error "Failed to copy executable: $($_.Exception.Message)"
        return $false
    }
    
    # Add to PATH
    try {
        if ($SystemWide) {
            $SystemPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
            if (-not $SystemPath.Contains($FinalInstallPath)) {
                $NewPath = if ($SystemPath) { "$SystemPath;$FinalInstallPath" } else { $FinalInstallPath }
                [Environment]::SetEnvironmentVariable("PATH", $NewPath, "Machine")
                Write-Success "Added to system PATH: $FinalInstallPath"
            } else {
                Write-Info "Already in system PATH: $FinalInstallPath"
            }
        } else {
            $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
            if (-not $UserPath -or -not $UserPath.Contains($FinalInstallPath)) {
                $NewPath = if ($UserPath) { "$UserPath;$FinalInstallPath" } else { $FinalInstallPath }
                [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
                Write-Success "Added to user PATH: $FinalInstallPath"
            } else {
                Write-Info "Already in user PATH: $FinalInstallPath"
            }
        }
        
        # Update current session PATH
        $env:PATH = "$env:PATH;$FinalInstallPath"
    } catch {
        Write-Warning "Failed to add to PATH: $($_.Exception.Message)"
        Write-Info "Please add manually: $FinalInstallPath"
    }
    
    return $true
}

function Install-PowerShellAliases {
    if ($SkipAliases) {
        Write-Info "Skipping PowerShell alias configuration"
        return $true
    }
    
    Write-Info "Configuring PowerShell aliases..."
    
    # Get profile paths
    $ProfilePaths = Get-PowerShellProfiles
    
    if ($ProfilePaths.Count -eq 0) {
        Write-Warning "No PowerShell profile paths found"
        return $false
    }
    
    $ConfigBlock = @"

# DelGuard Safe Delete Tool Configuration
# Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
# Version: DelGuard 2.1.1 Enhanced (Fixed)

if (-not `$global:DelGuardConfigured) {
    try {
        `$delguardPath = (Get-Command delguard -ErrorAction SilentlyContinue).Source
        
        if (`$delguardPath) {
            # Remove existing conflicting aliases safely
            @('del', 'rm', 'cp', 'copy') | ForEach-Object {
                try {
                    if (Get-Alias `$_ -ErrorAction SilentlyContinue) {
                        Remove-Item "Alias:`$_" -Force -ErrorAction SilentlyContinue
                    }
                } catch { }
            }
            
            # Create safe wrapper functions with enhanced parameter handling
            function global:del {
                param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
                if (`$Arguments -contains '--install' -or `$Arguments -contains '--uninstall') {
                    Write-Warning "Use 'delguard `$(`$Arguments -join ' ')' for installation commands"
                    return
                }
                try {
                    & delguard @Arguments
                } catch {
                    Write-Error "DelGuard error: `$(`$_.Exception.Message)"
                }
            }
            
            function global:rm {
                param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
                try {
                    & delguard @Arguments
                } catch {
                    Write-Error "DelGuard error: `$(`$_.Exception.Message)"
                }
            }
            
            function global:cp {
                param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
                try {
                    & delguard --copy @Arguments
                } catch {
                    Write-Error "DelGuard error: `$(`$_.Exception.Message)"
                }
            }
            
            function global:copy {
                param([Parameter(ValueFromRemainingArguments=`$true)][string[]]`$Arguments)
                try {
                    & delguard --copy @Arguments
                } catch {
                    Write-Error "DelGuard error: `$(`$_.Exception.Message)"
                }
            }
            
            # Set global flag to prevent duplicate loading
            `$global:DelGuardConfigured = `$true
            
            Write-Host "DelGuard Safe Delete Tool Loaded (Enhanced Fixed)" -ForegroundColor Green
            Write-Host "Commands: del, rm, cp, copy, delguard" -ForegroundColor Cyan
            Write-Host "Use 'delguard --help' for detailed help" -ForegroundColor Gray
        } else {
            Write-Warning "DelGuard executable not found in PATH"
        }
    } catch {
        Write-Warning "Failed to configure DelGuard aliases: `$(`$_.Exception.Message)"
    }
}
# End DelGuard Configuration

"@
    
    $Success = $false
    foreach ($ProfilePath in $ProfilePaths) {
        try {
            $ProfileDir = Split-Path $ProfilePath -Parent
            if (-not (Test-Path $ProfileDir -ErrorAction SilentlyContinue)) {
                $null = New-Item -ItemType Directory -Path $ProfileDir -Force -ErrorAction Stop
                Write-Success "Created profile directory: $ProfileDir"
            }
            
            # Check existing content
            $ExistingContent = ""
            if (Test-Path $ProfilePath -ErrorAction SilentlyContinue) {
                try {
                    $ExistingContent = Get-Content $ProfilePath -Raw -ErrorAction Stop
                } catch {
                    Write-Warning "Could not read existing profile: $ProfilePath"
                    continue
                }
            }
            
            # Remove old DelGuard configuration
            if ($ExistingContent -and $ExistingContent.Contains("# DelGuard")) {
                if (-not $Force) {
                    Write-Warning "DelGuard configuration already exists in: $ProfilePath"
                    Write-Warning "Use -Force to overwrite"
                    continue
                }
                # Remove existing DelGuard configuration
                $ExistingContent = $ExistingContent -replace '(?s)# DelGuard.*?# End DelGuard Configuration\r?\n?', ''
            }
            
            # Add new configuration
            $NewContent = $ExistingContent + $ConfigBlock
            Set-Content $ProfilePath $NewContent -Encoding UTF8 -ErrorAction Stop
            Write-Success "Updated PowerShell profile: $ProfilePath"
            $Success = $true
            break
        } catch {
            Write-Warning "Failed to update profile: $ProfilePath - $($_.Exception.Message)"
        }
    }
    
    if (-not $Success) {
        Write-Warning "Failed to update any PowerShell profile"
        Write-Info "You can manually add DelGuard to your PATH and create aliases"
    }
    
    return $Success
}

function Uninstall-DelGuard {
    Write-Info "Uninstalling DelGuard..."
    
    # Find and remove executable
    $ExecutablePaths = @(
        "$env:USERPROFILE\bin\delguard.exe",
        "$env:ProgramFiles\DelGuard\delguard.exe",
        "$env:LOCALAPPDATA\DelGuard\delguard.exe"
    )
    
    if ($InstallPath) {
        $ExecutablePaths += "$InstallPath\delguard.exe"
    }
    
    foreach ($ExePath in $ExecutablePaths) {
        if (Test-Path $ExePath -ErrorAction SilentlyContinue) {
            try {
                Remove-Item $ExePath -Force -ErrorAction Stop
                Write-Success "Removed executable: $ExePath"
                
                # Remove from PATH
                $InstallDir = Split-Path $ExePath -Parent
                
                # Remove from user PATH
                try {
                    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
                    if ($UserPath -and $UserPath.Contains($InstallDir)) {
                        $NewPath = ($UserPath -split ';' | Where-Object { $_ -ne $InstallDir }) -join ';'
                        [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
                        Write-Success "Removed from user PATH: $InstallDir"
                    }
                } catch {
                    Write-Warning "Failed to remove from user PATH: $($_.Exception.Message)"
                }
                
                # Remove from system PATH if admin
                if (Test-Administrator) {
                    try {
                        $SystemPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
                        if ($SystemPath -and $SystemPath.Contains($InstallDir)) {
                            $NewPath = ($SystemPath -split ';' | Where-Object { $_ -ne $InstallDir }) -join ';'
                            [Environment]::SetEnvironmentVariable("PATH", $NewPath, "Machine")
                            Write-Success "Removed from system PATH: $InstallDir"
                        }
                    } catch {
                        Write-Warning "Failed to remove from system PATH: $($_.Exception.Message)"
                    }
                }
                
                # Remove empty directory
                try {
                    if ((Get-ChildItem $InstallDir -ErrorAction SilentlyContinue).Count -eq 0) {
                        Remove-Item $InstallDir -Force -ErrorAction SilentlyContinue
                        Write-Success "Removed empty directory: $InstallDir"
                    }
                } catch {
                    # Ignore errors when removing directory
                }
            } catch {
                Write-Warning "Failed to remove: $ExePath - $($_.Exception.Message)"
            }
        }
    }
    
    # Remove from PowerShell profiles
    $ProfilePaths = Get-PowerShellProfiles
    foreach ($ProfilePath in $ProfilePaths) {
        if (Test-Path $ProfilePath -ErrorAction SilentlyContinue) {
            try {
                $Content = Get-Content $ProfilePath -Raw -ErrorAction SilentlyContinue
                if ($Content -and $Content.Contains("# DelGuard")) {
                    $NewContent = $Content -replace '(?s)# DelGuard.*?# End DelGuard Configuration\r?\n?', ''
                    Set-Content $ProfilePath $NewContent -Encoding UTF8 -ErrorAction Stop
                    Write-Success "Removed DelGuard configuration from: $ProfilePath"
                }
            } catch {
                Write-Warning "Failed to clean profile: $ProfilePath - $($_.Exception.Message)"
            }
        }
    }
    
    Write-Success "DelGuard uninstalled successfully!"
    Write-Info "Please restart PowerShell to complete the removal."
}

function Save-InstallationLog {
    if ($Script:InstallationLog.Count -gt 0) {
        try {
            $LogPath = Join-Path $env:TEMP "delguard-install-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"
            $Script:InstallationLog | Out-File -FilePath $LogPath -Encoding UTF8
            Write-Info "Installation log saved to: $LogPath"
        } catch {
            Write-Warning "Could not save installation log: $($_.Exception.Message)"
        }
    }
}

# Main execution
try {
    if ($Help) {
        Show-Help
        exit 0
    }
    
    # Check execution policy
    if (-not (Test-ExecutionPolicy)) {
        Write-Warning "PowerShell execution policy may prevent script execution"
    }
    
    Write-Header "=== DelGuard Enhanced Universal Installer for Windows (Fixed) ==="
    Write-Info "PowerShell Version: $($PSVersionTable.PSVersion)"
    Write-Info "Operating System: $([System.Environment]::OSVersion.VersionString)"
    Write-Info "Administrator: $(if (Test-Administrator) { 'Yes' } else { 'No' })"
    
    if ($CheckOnly) {
        $result = Test-InstallationStatus
        Save-InstallationLog
        exit $(if ($result) { 0 } else { 1 })
    }
    
    if ($Uninstall) {
        Uninstall-DelGuard
        Save-InstallationLog
        exit 0
    }
    
    # Installation process
    Write-Info "Starting DelGuard installation..."
    
    if (-not (Build-DelGuard)) {
        Write-Error "Build failed. Installation aborted."
        Save-InstallationLog
        exit 1
    }
    
    if (-not (Install-DelGuard)) {
        Write-Error "Installation failed."
        Save-InstallationLog
        exit 1
    }
    
    $aliasResult = Install-PowerShellAliases
    if (-not $aliasResult) {
        Write-Warning "PowerShell alias configuration failed, but DelGuard was installed successfully."
        Write-Info "You can still use 'delguard' command directly."
    }
    
    Write-Success "=== Installation Complete ==="
    Write-Info "DelGuard has been installed successfully!"
    Write-Info ""
    Write-Info "NEXT STEPS:"
    Write-Info "1. Restart PowerShell or run: . `$PROFILE"
    Write-Info "2. Test with: delguard --version"
    Write-Info "3. Check status with: .\install.ps1 -CheckOnly"
    Write-Info "4. Use these safe commands:"
    Write-Info "   • del <file>     - Safe delete (replaces Windows del)"
    Write-Info "   • rm <file>      - Safe delete (Unix style)"
    Write-Info "   • cp <src> <dst> - Safe copy"
    Write-Info "   • delguard --help - Full help and options"
    Write-Info ""
    Write-Success "Happy safe deleting!"
    
    # Test installation
    Write-Info "Testing installation..."
    try {
        $TestResult = & "delguard.exe" --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "✓ DelGuard is working correctly"
        } else {
            Write-Warning "⚠ DelGuard may not be working properly"
        }
    } catch {
        Write-Warning "⚠ Could not test DelGuard installation: $($_.Exception.Message)"
    }
    
    Save-InstallationLog
    
} catch {
    Write-Error "Unexpected error: $($_.Exception.Message)"
    Write-Info "Stack trace: $($_.ScriptStackTrace)"
    Save-InstallationLog
    exit 1
}