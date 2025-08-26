# DelGuard Common Installation Functions
# Shared functions for all PowerShell installation scripts

# Load configuration
function Get-InstallConfig {
    param(
        [string]$ConfigPath = "config/install-config.json"
    )
    
    if (Test-Path $ConfigPath) {
        try {
            $config = Get-Content $ConfigPath -Raw | ConvertFrom-Json
            return $config
        }
        catch {
            Write-Warning "Failed to load config from $ConfigPath, using defaults"
        }
    }
    
    # Return default configuration if file not found
    return @{
        default_settings = @{
            install_path = @{
                windows = "$env:USERPROFILE\bin"
            }
            aliases = @{
                enabled = $true
                commands = @("del", "rm", "cp", "copy")
            }
        }
        security = @{
            require_admin_for_system_install = $true
            backup_existing_aliases = $true
        }
    }
}

# Enhanced color output functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$ForegroundColor = "White",
        [string]$BackgroundColor = $null,
        [switch]$NoNewline
    )
    
    $params = @{
        Object = $Message
        ForegroundColor = $ForegroundColor
    }
    
    if ($BackgroundColor) {
        $params.BackgroundColor = $BackgroundColor
    }
    
    if ($NoNewline) {
        $params.NoNewline = $true
    }
    
    Write-Host @params
}

function Write-Header {
    param([string]$Message)
    Write-ColorOutput "`n=== $Message ===" -ForegroundColor Magenta
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ️  $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "✅ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "❌ $Message" -ForegroundColor Red
}

function Write-Progress {
    param(
        [string]$Activity,
        [string]$Status,
        [int]$PercentComplete = -1
    )
    
    if ($PercentComplete -ge 0) {
        Write-Progress -Activity $Activity -Status $Status -PercentComplete $PercentComplete
    } else {
        Write-Progress -Activity $Activity -Status $Status
    }
}

# System detection functions
function Get-SystemInfo {
    return @{
        OS = [System.Environment]::OSVersion.Platform
        OSVersion = [System.Environment]::OSVersion.Version
        Architecture = [System.Environment]::Is64BitOperatingSystem
        PowerShellVersion = $PSVersionTable.PSVersion
        IsAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
        UserName = [System.Environment]::UserName
        MachineName = [System.Environment]::MachineName
    }
}

function Test-AdminRights {
    return ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
}

function Test-GoInstallation {
    try {
        $goVersion = & go version 2>$null
        if ($LASTEXITCODE -eq 0) {
            return @{
                Installed = $true
                Version = $goVersion
            }
        }
    }
    catch {
        # Go not found
    }
    
    return @{
        Installed = $false
        Version = $null
    }
}

# Path management functions
function Add-ToPath {
    param(
        [string]$Path,
        [switch]$SystemWide,
        [switch]$Force
    )
    
    # Check if path already exists
    $currentPath = if ($SystemWide) {
        [Environment]::GetEnvironmentVariable("PATH", "Machine")
    } else {
        [Environment]::GetEnvironmentVariable("PATH", "User")
    }
    
    if ($currentPath -split ';' -contains $Path) {
        Write-Info "Path already exists in environment: $Path"
        return $true
    }
    
    try {
        $newPath = if ($currentPath) { "$currentPath;$Path" } else { $Path }
        
        if ($SystemWide) {
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
        } else {
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        }
        
        # Update current session
        $env:PATH = "$env:PATH;$Path"
        
        Write-Success "Added to PATH: $Path"
        return $true
    }
    catch {
        Write-Error "Failed to add to PATH: $($_.Exception.Message)"
        return $false
    }
}

function Remove-FromPath {
    param(
        [string]$Path,
        [switch]$SystemWide
    )
    
    try {
        $currentPath = if ($SystemWide) {
            [Environment]::GetEnvironmentVariable("PATH", "Machine")
        } else {
            [Environment]::GetEnvironmentVariable("PATH", "User")
        }
        
        $pathArray = $currentPath -split ';' | Where-Object { $_ -ne $Path }
        $newPath = $pathArray -join ';'
        
        if ($SystemWide) {
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
        } else {
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        }
        
        Write-Success "Removed from PATH: $Path"
        return $true
    }
    catch {
        Write-Error "Failed to remove from PATH: $($_.Exception.Message)"
        return $false
    }
}

# File operations
function Backup-File {
    param(
        [string]$FilePath,
        [string]$BackupSuffix = ".delguard-backup"
    )
    
    if (Test-Path $FilePath) {
        $backupPath = "$FilePath$BackupSuffix"
        try {
            Copy-Item $FilePath $backupPath -Force
            Write-Info "Backed up: $FilePath -> $backupPath"
            return $backupPath
        }
        catch {
            Write-Warning "Failed to backup $FilePath: $($_.Exception.Message)"
            return $null
        }
    }
    
    return $null
}

function Restore-File {
    param(
        [string]$BackupPath,
        [string]$OriginalPath
    )
    
    if (Test-Path $BackupPath) {
        try {
            Copy-Item $BackupPath $OriginalPath -Force
            Remove-Item $BackupPath -Force
            Write-Success "Restored: $OriginalPath"
            return $true
        }
        catch {
            Write-Error "Failed to restore $OriginalPath: $($_.Exception.Message)"
            return $false
        }
    }
    
    return $false
}

# Installation validation
function Test-Installation {
    param(
        [string]$ExecutablePath,
        [string[]]$ExpectedCommands = @("--version", "--help")
    )
    
    if (-not (Test-Path $ExecutablePath)) {
        return @{
            Valid = $false
            Error = "Executable not found: $ExecutablePath"
        }
    }
    
    foreach ($command in $ExpectedCommands) {
        try {
            $result = & $ExecutablePath $command 2>$null
            if ($LASTEXITCODE -ne 0) {
                return @{
                    Valid = $false
                    Error = "Command failed: $ExecutablePath $command"
                }
            }
        }
        catch {
            return @{
                Valid = $false
                Error = "Exception running: $ExecutablePath $command - $($_.Exception.Message)"
            }
        }
    }
    
    return @{
        Valid = $true
        Error = $null
    }
}

# Cleanup functions
function Remove-EmptyDirectories {
    param([string[]]$Paths)
    
    foreach ($path in $Paths) {
        if (Test-Path $path) {
            try {
                $items = Get-ChildItem $path -Force
                if ($items.Count -eq 0) {
                    Remove-Item $path -Force
                    Write-Info "Removed empty directory: $path"
                }
            }
            catch {
                Write-Warning "Could not remove directory $path: $($_.Exception.Message)"
            }
        }
    }
}

# Export functions for use in other scripts
Export-ModuleMember -Function *