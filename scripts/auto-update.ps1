# DelGuard Auto-Update Script
# Handles automatic updates and version checking

param(
    [switch]$CheckOnly,
    [switch]$Force,
    [switch]$Quiet,
    [string]$UpdateChannel = "stable"
)

# Import common functions
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$commonScript = Join-Path (Split-Path -Parent $scriptDir) "scripts\install-common.ps1"
if (Test-Path $commonScript) {
    . $commonScript
}

# Configuration
$UpdateConfig = @{
    RepositoryUrl = "https://api.github.com/repos/delguard/delguard"
    DownloadBaseUrl = "https://github.com/delguard/delguard/releases/download"
    CurrentVersion = "2.1.0"
    Channels = @{
        stable = "latest"
        beta = "prerelease"
        dev = "development"
    }
}

function Get-LatestVersion {
    param(
        [string]$Channel = "stable"
    )
    
    try {
        Write-Info "Checking for updates from $($UpdateConfig.RepositoryUrl)..."
        
        # Use GitHub API to get latest release
        $apiUrl = if ($Channel -eq "stable") {
            "$($UpdateConfig.RepositoryUrl)/releases/latest"
        } else {
            "$($UpdateConfig.RepositoryUrl)/releases"
        }
        
        $response = Invoke-RestMethod -Uri $apiUrl -Method Get -TimeoutSec 30
        
        if ($Channel -eq "stable") {
            return @{
                Version = $response.tag_name -replace '^v', ''
                DownloadUrl = $response.assets | Where-Object { $_.name -like "*windows*" -or $_.name -like "*win*" } | Select-Object -First 1 -ExpandProperty browser_download_url
                ReleaseNotes = $response.body
                PublishedAt = $response.published_at
            }
        } else {
            # Handle prerelease/beta versions
            $prerelease = $response | Where-Object { $_.prerelease -eq $true } | Select-Object -First 1
            if ($prerelease) {
                return @{
                    Version = $prerelease.tag_name -replace '^v', ''
                    DownloadUrl = $prerelease.assets | Where-Object { $_.name -like "*windows*" -or $_.name -like "*win*" } | Select-Object -First 1 -ExpandProperty browser_download_url
                    ReleaseNotes = $prerelease.body
                    PublishedAt = $prerelease.published_at
                }
            }
        }
        
        return $null
    }
    catch {
        Write-Error "Failed to check for updates: $($_.Exception.Message)"
        return $null
    }
}

function Compare-Versions {
    param(
        [string]$Current,
        [string]$Latest
    )
    
    try {
        $currentVersion = [System.Version]::Parse($Current)
        $latestVersion = [System.Version]::Parse($Latest)
        
        return $latestVersion.CompareTo($currentVersion)
    }
    catch {
        # Fallback to string comparison
        return [string]::Compare($Latest, $Current, $true)
    }
}

function Get-CurrentVersion {
    # Try to get version from installed DelGuard
    $delguardPaths = @(
        "$env:USERPROFILE\bin\delguard.exe",
        "$env:ProgramFiles\DelGuard\delguard.exe",
        "$env:LOCALAPPDATA\DelGuard\delguard.exe"
    )
    
    foreach ($path in $delguardPaths) {
        if (Test-Path $path) {
            try {
                $versionOutput = & $path --version 2>$null
                if ($LASTEXITCODE -eq 0 -and $versionOutput) {
                    # Extract version from output like "DelGuard v2.1.0 - Safe Delete Tool"
                    if ($versionOutput -match 'v?(\d+\.\d+\.\d+)') {
                        return $matches[1]
                    }
                }
            }
            catch {
                continue
            }
        }
    }
    
    return $UpdateConfig.CurrentVersion
}

function Download-Update {
    param(
        [string]$DownloadUrl,
        [string]$OutputPath
    )
    
    try {
        Write-Info "Downloading update from: $DownloadUrl"
        Write-Progress -Activity "Downloading DelGuard Update" -Status "Starting download..." -PercentComplete 0
        
        # Create temporary directory
        $tempDir = Join-Path $env:TEMP "delguard-update-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
        New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
        
        $tempFile = Join-Path $tempDir "delguard-update.zip"
        
        # Download with progress
        $webClient = New-Object System.Net.WebClient
        $webClient.DownloadProgressChanged += {
            param($sender, $e)
            Write-Progress -Activity "Downloading DelGuard Update" -Status "Downloaded $($e.BytesReceived) of $($e.TotalBytesToReceive) bytes" -PercentComplete $e.ProgressPercentage
        }
        
        $webClient.DownloadFileCompleted += {
            param($sender, $e)
            Write-Progress -Activity "Downloading DelGuard Update" -Completed
        }
        
        $webClient.DownloadFileAsync((New-Object System.Uri($DownloadUrl)), $tempFile)
        
        # Wait for download to complete
        while ($webClient.IsBusy) {
            Start-Sleep -Milliseconds 100
        }
        
        $webClient.Dispose()
        
        if (Test-Path $tempFile) {
            Write-Success "Download completed: $tempFile"
            return $tempFile
        } else {
            Write-Error "Download failed: File not found"
            return $null
        }
    }
    catch {
        Write-Error "Download failed: $($_.Exception.Message)"
        return $null
    }
}

function Install-Update {
    param(
        [string]$UpdateFile,
        [string]$InstallPath
    )
    
    try {
        Write-Info "Installing update from: $UpdateFile"
        
        # Extract update (assuming it's a zip file)
        $extractPath = Join-Path (Split-Path $UpdateFile) "extracted"
        Expand-Archive -Path $UpdateFile -DestinationPath $extractPath -Force
        
        # Find the executable
        $newExecutable = Get-ChildItem -Path $extractPath -Name "delguard.exe" -Recurse | Select-Object -First 1
        if (-not $newExecutable) {
            Write-Error "DelGuard executable not found in update package"
            return $false
        }
        
        $newExecutablePath = $newExecutable.FullName
        
        # Backup current installation
        if (Test-Path $InstallPath) {
            $backupPath = "$InstallPath.backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
            Copy-Item $InstallPath $backupPath -Force
            Write-Info "Backed up current version to: $backupPath"
        }
        
        # Install new version
        Copy-Item $newExecutablePath $InstallPath -Force
        Write-Success "Updated DelGuard executable: $InstallPath"
        
        # Verify installation
        $verification = Test-Installation -ExecutablePath $InstallPath
        if ($verification.Valid) {
            Write-Success "Update installation verified successfully"
            
            # Clean up backup if verification passed
            if (Test-Path "$InstallPath.backup-*") {
                Remove-Item "$InstallPath.backup-*" -Force
                Write-Info "Cleaned up backup files"
            }
            
            return $true
        } else {
            Write-Error "Update verification failed: $($verification.Error)"
            
            # Restore backup
            if (Test-Path $backupPath) {
                Copy-Item $backupPath $InstallPath -Force
                Write-Info "Restored previous version from backup"
            }
            
            return $false
        }
    }
    catch {
        Write-Error "Update installation failed: $($_.Exception.Message)"
        return $false
    }
}

function Show-UpdateInfo {
    param($UpdateInfo)
    
    Write-Header "DelGuard Update Available"
    Write-Info "Current Version: $(Get-CurrentVersion)"
    Write-Info "Latest Version:  $($UpdateInfo.Version)"
    Write-Info "Published:       $($UpdateInfo.PublishedAt)"
    
    if ($UpdateInfo.ReleaseNotes) {
        Write-Info "Release Notes:"
        Write-Host $UpdateInfo.ReleaseNotes -ForegroundColor Gray
    }
}

# Main execution
function Main {
    Write-Header "DelGuard Auto-Update"
    
    $currentVersion = Get-CurrentVersion
    Write-Info "Current version: $currentVersion"
    
    $latestInfo = Get-LatestVersion -Channel $UpdateChannel
    if (-not $latestInfo) {
        Write-Error "Could not retrieve update information"
        exit 1
    }
    
    $comparison = Compare-Versions -Current $currentVersion -Latest $latestInfo.Version
    
    if ($comparison -gt 0) {
        # Update available
        Show-UpdateInfo -UpdateInfo $latestInfo
        
        if ($CheckOnly) {
            Write-Info "Update available. Use without -CheckOnly to install."
            exit 0
        }
        
        if (-not $Force -and -not $Quiet) {
            $response = Read-Host "Do you want to install this update? (y/N)"
            if ($response -notmatch '^[Yy]') {
                Write-Info "Update cancelled by user"
                exit 0
            }
        }
        
        # Find current installation
        $installPath = $null
        $searchPaths = @(
            "$env:USERPROFILE\bin\delguard.exe",
            "$env:ProgramFiles\DelGuard\delguard.exe",
            "$env:LOCALAPPDATA\DelGuard\delguard.exe"
        )
        
        foreach ($path in $searchPaths) {
            if (Test-Path $path) {
                $installPath = $path
                break
            }
        }
        
        if (-not $installPath) {
            Write-Error "Could not find current DelGuard installation"
            exit 1
        }
        
        # Download and install update
        $updateFile = Download-Update -DownloadUrl $latestInfo.DownloadUrl -OutputPath $installPath
        if ($updateFile) {
            $success = Install-Update -UpdateFile $updateFile -InstallPath $installPath
            if ($success) {
                Write-Success "DelGuard has been updated to version $($latestInfo.Version)"
            } else {
                Write-Error "Update installation failed"
                exit 1
            }
        } else {
            Write-Error "Update download failed"
            exit 1
        }
    }
    elseif ($comparison -eq 0) {
        Write-Success "DelGuard is up to date (version $currentVersion)"
    }
    else {
        Write-Info "You are running a newer version ($currentVersion) than the latest release ($($latestInfo.Version))"
    }
}

# Run main function
Main