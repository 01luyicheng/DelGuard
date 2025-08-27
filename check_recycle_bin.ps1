# 检查回收站内容
# Check Recycle Bin contents
$shell = New-Object -ComObject Shell.Application
$recycleBin = $shell.Namespace(10)  # 10 is the special folder ID for Recycle Bin

if ($recycleBin -ne $null) {
    $items = $recycleBin.Items()
    Write-Host "Files in Recycle Bin: $($items.Count)"
    
    foreach ($item in $items) {
        Write-Host "File: $($item.Name) - Original Path: $($item.Path)"
    }
} else {
    Write-Host "Cannot access Recycle Bin"
}
