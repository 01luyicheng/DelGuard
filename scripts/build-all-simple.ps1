# DelGuard 跨平台构建脚本
Write-Host "DelGuard 跨平台构建" -ForegroundColor Cyan

$env:CGO_ENABLED = "0"
New-Item -ItemType Directory -Path "build" -Force | Out-Null

$platforms = @(
    @{OS="windows"; Arch="amd64"; Ext=".exe"},
    @{OS="linux"; Arch="amd64"; Ext=""},
    @{OS="darwin"; Arch="amd64"; Ext=""},
    @{OS="darwin"; Arch="arm64"; Ext=""}
)

$success = 0
foreach ($platform in $platforms) {
    $os = $platform.OS
    $arch = $platform.Arch
    $ext = $platform.Ext
    $output = "build/delguard-$os-$arch$ext"
    
    Write-Host "构建 $os/$arch..." -NoNewline
    
    try {
        $env:GOOS = $os
        $env:GOARCH = $arch
        & go build -ldflags="-s -w" -o $output .
        
        if (Test-Path $output) {
            Write-Host " 成功" -ForegroundColor Green
            $success++
        } else {
            Write-Host " 失败" -ForegroundColor Red
        }
    } catch {
        Write-Host " 失败" -ForegroundColor Red
    }
}

Write-Host "构建结果: $success/$($platforms.Count) 成功" -ForegroundColor $(if ($success -eq $platforms.Count) { "Green" } else { "Yellow" })

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue