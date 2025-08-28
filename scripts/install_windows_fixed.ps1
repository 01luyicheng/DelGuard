# DelGuard Windows Installation Script
# Version: 2.0.0
# Description: Auto install DelGuard file deletion protection tool

param(
    [string]$InstallPath = "$env:ProgramFiles\DelGuard",
    [string]$ServiceName = "DelGuardService",
    [switch]$Silent = $false,
    [switch]$CreateDesktopShortcut = $true,
    [switch]$AddToPath = $true,
    [switch]$StartService = $true
)

# Set error handling
$ErrorActionPreference = "Stop"

# Color output functions
function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput $Message "Green" }
function Write-Warning { param([string]$Message) Write-ColorOutput $Message "Yellow" }
function Write-Error { param([string]$Message) Write-ColorOutput $Message "Red" }
function Write-Info { param([string]$Message) Write-ColorOutput $Message "Cyan" }

# Check administrator privileges
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Get system information
function Get-SystemInfo {
    return @{
        OS = (Get-WmiObject -Class Win32_OperatingSystem).Caption
        Architecture = $env:PROCESSOR_ARCHITECTURE
        PowerShellVersion = $PSVersionTable.PSVersion.ToString()
        DotNetVersion = [System.Runtime.InteropServices.RuntimeInformation]::FrameworkDescription
    }
}

# Build DelGuard executable
function Build-DelGuard {
    param([string]$ProjectRoot)
    
    Write-Info "Building DelGuard executable..."
    
    # Check Go environment
    if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
        throw "Go compiler not found. Please install Go language environment first."
    }
    
    # Set working directory
    Push-Location $ProjectRoot
    
    try {
        # Download dependencies
        Write-Info "Downloading Go module dependencies..."
        & go mod download
        & go mod tidy
        
        # Build Windows executable
        Write-Info "Compiling Windows executable..."
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        & go build -ldflags "-s -w" -o "delguard.exe" "./cmd/delguard"
        
        if ($LASTEXITCODE -ne 0) {
            throw "Compilation failed"
        }
        
        Write-Success "DelGuard compiled successfully"
    }
    finally {
        Pop-Location
    }
}

# Create installation directory
function New-InstallDirectory {
    param([string]$Path)
    
    Write-Info "Creating installation directory: $Path"
    
    if (Test-Path $Path) {
        Write-Warning "Installation directory already exists, will perform overwrite installation"
        # Stop service if running
        Stop-DelGuardService -Silent
    } else {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }
    
    Write-Success "Installation directory created successfully"
}

# Copy files
function Copy-DelGuardFiles {
    param([string]$SourcePath, [string]$DestPath)
    
    Write-Info "Copying DelGuard files to installation directory..."
    
    # Get current script directory
    $ScriptDir = Split-Path -Parent $MyInvocation.ScriptName
    $ProjectRoot = Split-Path -Parent $ScriptDir
    
    # Check and build executable
    $ExePath = Join-Path $ProjectRoot "delguard.exe"
    if (-not (Test-Path $ExePath)) {
        Write-Warning "Pre-compiled executable not found, attempting to build..."
        Build-DelGuard -ProjectRoot $ProjectRoot
    }
    
    # Copy main program
    if (Test-Path $ExePath) {
        Copy-Item $ExePath $DestPath -Force
        Write-Success "Main program copied successfully"
    } else {
        throw "DelGuard main program not found: $ExePath"
    }
    
    # Copy configuration files
    $ConfigDir = Join-Path $ProjectRoot "configs"
    if (Test-Path $ConfigDir) {
        $DestConfigDir = Join-Path $DestPath "configs"
        Copy-Item $ConfigDir $DestConfigDir -Recurse -Force
        Write-Success "Configuration files copied successfully"
    }
    
    # Copy documentation
    $DocsDir = Join-Path $ProjectRoot "docs"
    if (Test-Path $DocsDir) {
        $DestDocsDir = Join-Path $DestPath "docs"
        Copy-Item $DocsDir $DestDocsDir -Recurse -Force
        Write-Success "Documentation files copied successfully"
    }
}

# Create configuration file
function New-ConfigFile {
    param([string]$InstallPath)
    
    Write-Info "Creating configuration file..."
    
    $ConfigPath = Join-Path $InstallPath "config.yaml"
    $ConfigContent = @"
# DelGuard Configuration File
# Auto-generated at: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

app:
  name: "DelGuard"
  version: "2.0.0"
  log_level: "info"
  data_dir: "$InstallPath\data"

monitor:
  enabled: true
  watch_paths:
    - "C:\Users"
    - "D:\"
  exclude_paths:
    - "C:\Windows"
    - "C:\Program Files"
  file_types:
    - ".doc"
    - ".docx"
    - ".xls"
    - ".xlsx"
    - ".ppt"
    - ".pptx"
    - ".pdf"
    - ".txt"
    - ".jpg"
    - ".png"
    - ".mp4"
    - ".mp3"

restore:
  backup_dir: "$InstallPath\backups"
  max_backup_size: "10GB"
  retention_days: 30

search:
  index_enabled: true
  index_update_interval: "1h"
  max_results: 1000

security:
  enable_encryption: true
  require_admin: false
  audit_log: true
"@
    
    $ConfigContent | Out-File -FilePath $ConfigPath -Encoding UTF8
    Write-Success "Configuration file created successfully: $ConfigPath"
}

# Register Windows service
function Register-DelGuardService {
    param([string]$InstallPath, [string]$ServiceName)
    
    Write-Info "Registering Windows service..."
    
    $ExePath = Join-Path $InstallPath "delguard.exe"
    $ServiceDescription = "DelGuard File Deletion Protection Service"
    
    # Check if service already exists
    $ExistingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($ExistingService) {
        Write-Warning "Service already exists, updating..."
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        & sc.exe delete $ServiceName
        Start-Sleep -Seconds 2
    }
    
    # Create service
    $CreateResult = & sc.exe create $ServiceName binPath= "`"$ExePath`" service" start= auto DisplayName= "DelGuard File Protection Service" depend= ""
    
    if ($LASTEXITCODE -eq 0) {
        # Set service description
        & sc.exe description $ServiceName $ServiceDescription
        
        # Set service recovery options
        & sc.exe failure $ServiceName reset= 86400 actions= restart/5000/restart/10000/restart/20000
        
        Write-Success "Windows service registered successfully"
    } else {
        throw "Service registration failed: $CreateResult"
    }
}

# Start service
function Start-DelGuardService {
    param([string]$ServiceName)
    
    Write-Info "Starting DelGuard service..."
    
    try {
        Start-Service -Name $ServiceName
        Write-Success "Service started successfully"
        
        # Verify service status
        $Service = Get-Service -Name $ServiceName
        if ($Service.Status -eq "Running") {
            Write-Success "Service running status is normal"
        } else {
            Write-Warning "Service status abnormal: $($Service.Status)"
        }
    } catch {
        Write-Error "Service startup failed: $($_.Exception.Message)"
        throw
    }
}

# Stop service
function Stop-DelGuardService {
    param([string]$ServiceName = $ServiceName, [switch]$Silent)
    
    $Service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($Service -and $Service.Status -eq "Running") {
        if (-not $Silent) {
            Write-Info "Stopping DelGuard service..."
        }
        Stop-Service -Name $ServiceName -Force
        if (-not $Silent) {
            Write-Success "Service stopped"
        }
    }
}

# Add to system PATH
function Add-ToSystemPath {
    param([string]$InstallPath)
    
    Write-Info "Adding to system PATH environment variable..."
    
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($CurrentPath -notlike "*$InstallPath*") {
        $NewPath = $CurrentPath + ";" + $InstallPath
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "Machine")
        Write-Success "Added to system PATH"
    } else {
        Write-Info "Installation path already exists in PATH"
    }
}

# Create command aliases
function New-CommandAliases {
    param([string]$InstallPath)
    
    Write-Info "Creating command aliases..."
    
    # Create batch files for CMD compatibility
    $BatchFiles = @{
        "del.bat" = "@echo off`n`"$InstallPath\delguard.exe`" delete %*"
        "rm.bat" = "@echo off`n`"$InstallPath\delguard.exe`" delete %*"
        "delguard.bat" = "@echo off`n`"$InstallPath\delguard.exe`" %*"
    }
    
    foreach ($BatchFile in $BatchFiles.Keys) {
        $BatchPath = Join-Path $InstallPath $BatchFile
        $BatchFiles[$BatchFile] | Out-File -FilePath $BatchPath -Encoding ASCII
        Write-Success "Created batch file: $BatchFile"
    }
    
    # Create PowerShell profile aliases
    $ProfilePath = $PROFILE.AllUsersAllHosts
    $ProfileDir = Split-Path $ProfilePath -Parent
    
    if (-not (Test-Path $ProfileDir)) {
        New-Item -ItemType Directory -Path $ProfileDir -Force | Out-Null
    }
    
    $AliasContent = @"

# DelGuard Aliases
Set-Alias -Name delguard -Value "$InstallPath\delguard.exe" -Force
Set-Alias -Name del -Value "$InstallPath\delguard.exe" -Force
Set-Alias -Name rm -Value "$InstallPath\delguard.exe" -Force

function DelGuard-Delete { & "$InstallPath\delguard.exe" delete @args }
function DelGuard-Search { & "$InstallPath\delguard.exe" search @args }
function DelGuard-Restore { & "$InstallPath\delguard.exe" restore @args }

Set-Alias -Name dg-del -Value DelGuard-Delete -Force
Set-Alias -Name dg-search -Value DelGuard-Search -Force
Set-Alias -Name dg-restore -Value DelGuard-Restore -Force
"@
    
    if (Test-Path $ProfilePath) {
        $ExistingContent = Get-Content $ProfilePath -Raw
        if ($ExistingContent -notlike "*DelGuard Aliases*") {
            Add-Content -Path $ProfilePath -Value $AliasContent
            Write-Success "Added PowerShell aliases to profile"
        } else {
            Write-Info "PowerShell aliases already exist"
        }
    } else {
        $AliasContent | Out-File -FilePath $ProfilePath -Encoding UTF8
        Write-Success "Created PowerShell profile with aliases"
    }
    
    Write-Success "Command aliases created successfully"
}

# Create desktop shortcut
function New-DesktopShortcut {
    param([string]$InstallPath)
    
    Write-Info "Creating desktop shortcut..."
    
    $WshShell = New-Object -ComObject WScript.Shell
    $DesktopPath = [Environment]::GetFolderPath("Desktop")
    $ShortcutPath = Join-Path $DesktopPath "DelGuard.lnk"
    $ExePath = Join-Path $InstallPath "delguard.exe"
    
    $Shortcut = $WshShell.CreateShortcut($ShortcutPath)
    $Shortcut.TargetPath = $ExePath
    $Shortcut.WorkingDirectory = $InstallPath
    $Shortcut.Description = "DelGuard File Deletion Protection Tool"
    $Shortcut.Save()
    
    Write-Success "Desktop shortcut created successfully"
}

# Verify installation
function Test-Installation {
    param([string]$InstallPath, [string]$ServiceName)
    
    Write-Info "Verifying installation..."
    
    $Issues = @()
    
    # Check main program
    $ExePath = Join-Path $InstallPath "delguard.exe"
    if (-not (Test-Path $ExePath)) {
        $Issues += "Main program file does not exist"
    }
    
    # Check configuration file
    $ConfigPath = Join-Path $InstallPath "config.yaml"
    if (-not (Test-Path $ConfigPath)) {
        $Issues += "Configuration file does not exist"
    }
    
    # Check service
    $Service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if (-not $Service) {
        $Issues += "Windows service not registered"
    } elseif ($Service.Status -ne "Running") {
        $Issues += "Service not running"
    }
    
    if ($Issues.Count -eq 0) {
        Write-Success "Installation verification passed"
        return $true
    } else {
        Write-Error "Installation verification failed:"
        foreach ($Issue in $Issues) {
            Write-Error "  - $Issue"
        }
        return $false
    }
}

# Main installation process
function Install-DelGuard {
    try {
        Write-ColorOutput "
╔══════════════════════════════════════════════════════════════╗
║                    DelGuard Installer                        ║
║                     Version: 2.0.0                          ║
╚══════════════════════════════════════════════════════════════╝
" "Cyan"

        # Check administrator privileges
        if (-not (Test-Administrator)) {
            throw "Administrator privileges required to install DelGuard. Please run PowerShell as administrator."
        }

        # Display system information
        if (-not $Silent) {
            $SystemInfo = Get-SystemInfo
            Write-Info "System Information:"
            Write-Info "  Operating System: $($SystemInfo.OS)"
            Write-Info "  Architecture: $($SystemInfo.Architecture)"
            Write-Info "  PowerShell Version: $($SystemInfo.PowerShellVersion)"
            Write-Info "  .NET Version: $($SystemInfo.DotNetVersion)"
            Write-Info ""
            Write-Info "Installation Path: $InstallPath"
            Write-Info "Service Name: $ServiceName"
            Write-Info ""
        }

        # Confirm installation
        if (-not $Silent) {
            $Confirm = Read-Host "Continue with installation? (Y/n)"
            if ($Confirm -eq "n" -or $Confirm -eq "N") {
                Write-Info "Installation cancelled"
                return
            }
        }

        Write-Info "Starting DelGuard installation..."

        # 1. Create installation directory
        New-InstallDirectory -Path $InstallPath

        # 2. Copy files
        Copy-DelGuardFiles -SourcePath "." -DestPath $InstallPath

        # 3. Create configuration file
        New-ConfigFile -InstallPath $InstallPath

        # 4. Register Windows service
        Register-DelGuardService -InstallPath $InstallPath -ServiceName $ServiceName

        # 5. Add to PATH
        if ($AddToPath) {
            Add-ToSystemPath -InstallPath $InstallPath
        }

        # 6. Create command aliases
        New-CommandAliases -InstallPath $InstallPath

        # 7. Create shortcuts
        if ($CreateDesktopShortcut) {
            New-DesktopShortcut -InstallPath $InstallPath
        }

        # 8. Start service
        if ($StartService) {
            Start-DelGuardService -ServiceName $ServiceName
        }

        # 8. Verify installation
        $InstallSuccess = Test-Installation -InstallPath $InstallPath -ServiceName $ServiceName

        if ($InstallSuccess) {
            Write-Success "
╔══════════════════════════════════════════════════════════════╗
║                   DelGuard Installation Successful!          ║
╚══════════════════════════════════════════════════════════════╝"
            Write-Success "Installation Path: $InstallPath"
            Write-Success "Service Status: Running"
            Write-Success "Configuration File: $(Join-Path $InstallPath 'config.yaml')"
            Write-Info ""
            Write-Info "Usage:"
            Write-Info "  Command Line: delguard --help"
            Write-Info "  Service Management: services.msc"
        } else {
            throw "Installation verification failed"
        }

    } catch {
        Write-Error "Installation failed: $($_.Exception.Message)"
        Write-Error "Please check the error information and try again"
        exit 1
    }
}

# Execute installation
Install-DelGuard