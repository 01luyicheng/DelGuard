@echo off
echo.
echo ================================================
echo    DelGuard - Fix PowerShell and Install cp
echo ================================================
echo.

:: Fix PowerShell profile
echo Fixing PowerShell profile...
if exist "%USERPROFILE%\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1" (
    powershell -Command "Get-Content '%USERPROFILE%\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1' | ForEach-Object { $_.Replace('\"C:\Users\21601\AppData\Local\Programs\DelGuard\delguard.exe\" @args', '\"C:\Users\21601\AppData\Local\Programs\DelGuard\delguard.exe\" @args') } | Set-Content '%USERPROFILE%\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1'.tmp"
    move /y "%USERPROFILE%\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1.tmp" "%USERPROFILE%\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1" >nul 2>&1
)

:: Add current directory to PATH
echo Adding DelGuard to PATH...
set "CURRENT_DIR=%~dp0"
set "CURRENT_DIR=%CURRENT_DIR:~0,-1%"

powershell -Command "
    $currentDir = '%CURRENT_DIR%';
    $userPath = [Environment]::GetEnvironmentVariable('PATH', 'User');
    if ($userPath -notlike ('*' + $currentDir + '*')) {
        $newPath = $currentDir + ';' + $userPath;
        [Environment]::SetEnvironmentVariable('PATH', $newPath, 'User');
        Write-Host 'Added to PATH successfully' -ForegroundColor Green;
    } else {
        Write-Host 'Already in PATH' -ForegroundColor Yellow;
    }
"

echo.
echo ================================================
echo    Installation Complete!
echo ================================================
echo.
echo You can now use: cp source.txt dest.txt
echo Please restart Command Prompt or PowerShell
echo.
pause