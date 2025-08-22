#!/usr/bin/env pwsh
# DelGuard Windows 一键安装（无管理员权限）
# 用法（本地仓库内执行）:
#   pwsh -File scripts/install.ps1
#   pwsh -File scripts/install.ps1 -DefaultInteractive
#
# 用法（远程下载，需事先导出 GitHub 仓库 owner/repo）:
#   $env:DELGUARD_GITHUB_REPO = "YourOrg/DelGuard"
#   iwr -useb https://raw.githubusercontent.com/$env:DELGUARD_GITHUB_REPO/main/scripts/install.ps1 | iex
#   # 可添加 -DefaultInteractive: iwr ... | iex; Install-DelGuard -DefaultInteractive
#
# 若发布资产包含 .sha256，将自动校验

[CmdletBinding(DefaultParameterSetName="Run")]
param(
  [Parameter(ParameterSetName="Run")]
  [switch]$DefaultInteractive
)

$ErrorActionPreference = "Stop"

function Ensure-Path {
  param([string]$Dir)
  if (-not (Test-Path -Path $Dir)) { New-Item -ItemType Directory -Path $Dir -Force | Out-Null }
}

function Add-UserPath {
  param([string]$Dir)
  $current = [Environment]::GetEnvironmentVariable("PATH", "User")
  if ($null -eq $current -or $current -notlike "*$Dir*") {
    $newPath = if ([string]::IsNullOrEmpty($current)) { $Dir } else { "$current;$Dir" }
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Host "已将 $Dir 添加到用户 PATH（新开终端生效）"
  }
}

function Get-OSArch {
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
  switch ($arch) {
    "X64"   { return "amd64" }
    "Arm64" { return "arm64" }
    default { throw "不支持的架构：$arch" }
  }
}

function Download-File {
  param([string]$Url, [string]$Dest)
  Invoke-WebRequest -UseBasicParsing -Uri $Url -OutFile $Dest
}

function Install-FromRemote {
  $repo = $env:DELGUARD_GITHUB_REPO
  if ([string]::IsNullOrEmpty($repo)) {
    throw "未找到本地二进制，也未设置 DELGUARD_GITHUB_REPO（形如 YourOrg/DelGuard）以执行远程安装。"
  }
  $arch = Get-OSArch
  $asset = "delguard-windows-$arch.exe"
  $base  = "https://github.com/$repo/releases/latest/download"
  $urlBin = "$base/$asset"
  $urlSha = "$urlBin.sha256"

  $tmp = New-TemporaryFile
  $tmpSha = [System.IO.Path]::ChangeExtension($tmp.FullName, ".sha256")

  Write-Host "下载 $urlBin ..."
  Download-File -Url $urlBin -Dest $tmp.FullName

  $shaOk = $true
  try {
    Download-File -Url $urlSha -Dest $tmpSha.FullName
  } catch {
    $shaOk = $false
    Write-Host "未找到校验文件，跳过校验。"
  }
  if ($shaOk) {
    Write-Host "正在校验 SHA256..."
    $expected = (Get-Content $tmpSha.FullName | Select-Object -First 1).Split()[0]
    $actual = (Get-FileHash $tmp.FullName -Algorithm SHA256).Hash
    if ($expected.ToUpperInvariant() -ne $actual.ToUpperInvariant()) {
      throw "校验失败：期望 $expected 实际 $actual"
    }
  }

  $targetDir = Join-Path $env:LOCALAPPDATA "Programs\DelGuard"
  Ensure-Path $targetDir
  Move-Item -Force $tmp.FullName (Join-Path $targetDir "delguard.exe")
  Write-Host "已安装到 $targetDir\delguard.exe"
}

function Install-DelGuard {
  param([switch]$DefaultInteractive)

  # 修复路径解析问题，处理$MyInvocation.MyCommand.Path为null的情况
  $scriptPath = $MyInvocation.MyCommand.Path
  if ([string]::IsNullOrEmpty($scriptPath)) {
    # 如果脚本路径为空，尝试使用当前工作目录
    $scriptPath = $PWD.Path
    $root = $scriptPath
  } else {
    $root = Split-Path -Parent -Path $scriptPath
  }
  
  $projDir = $null
  if ($null -ne $root) {
    try { $projDir = (Resolve-Path (Join-Path $root "..")).Path } catch {}
  }
  $targetDir = Join-Path $env:LOCALAPPDATA "Programs\DelGuard"
  Ensure-Path $targetDir

  # 修复：优先检查构建目录中的实际文件
  $foundBinary = $false
  
  # 1. 优先检查build目录中的DelGuard.exe
  if ($projDir) {
    $buildDir = Join-Path $projDir "build"
    $delguardBuild = Join-Path $buildDir "DelGuard.exe"
    if (Test-Path $delguardBuild) {
      Write-Host "发现构建好的DelGuard.exe，正在安装..."
      Copy-Item -Force $delguardBuild (Join-Path $targetDir "delguard.exe")
      $foundBinary = $true
    }
  }
  
  # 2. 检查项目根目录的其他可能位置
  if (-not $foundBinary -and $projDir) {
    $possibleLocations = @(
      (Join-Path $projDir "delguard.exe"),
      (Join-Path $projDir "DelGuard.exe"),
      (Join-Path $projDir "build\delguard-windows-amd64.exe")
    )
    
    foreach ($location in $possibleLocations) {
      if (Test-Path $location) {
        Write-Host "发现二进制文件: $location"
        Copy-Item -Force $location (Join-Path $targetDir "delguard.exe")
        $foundBinary = $true
        break
      }
    }
  }

  # 3. 如果本地没有找到，尝试远程下载
  if (-not $foundBinary) {
    try {
      Install-FromRemote
    } catch {
      # 4. 最后尝试本地构建
      if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "无法安装：未找到可用二进制，也无法下载，且未检测到 Go 构建环境。"
      }
      if (-not $projDir) { throw "无法定位源码目录进行本地构建。" }
      
      Write-Host "正在本地构建DelGuard..."
      Push-Location $projDir
      try {
        & go build -o (Join-Path $targetDir "delguard.exe") .
        Write-Host "本地构建完成"
      } catch {
        throw "本地构建失败: $_"
      } finally {
        Pop-Location
      }
    }
  }

  # 确保二进制文件已安装
  $finalExe = Join-Path $targetDir "delguard.exe"
  if (-not (Test-Path $finalExe)) {
    throw "安装失败：未能找到或创建 delguard.exe"
  }

  Add-UserPath $targetDir

  # 运行内置安装别名（写入当前用户配置，无需管理员）
  try {
    $args = @("--install")
    if ($DefaultInteractive) { $args += "--default-interactive" }
    & $finalExe @args
  } catch {
    Write-Warning "安装别名时出现问题: $_"
    Write-Host "可以手动运行: $finalExe --install"
  }

  Write-Host "`n✅ 安装完成！"
  Write-Host "请新开一个 PowerShell 或 CMD 窗口使用："
  Write-Host "  del -i file.txt   # 交互删除"
  Write-Host "  del -rf folder    # 递归强制删除"
  Write-Host "  delguard --help   # 查看帮助"
}

# 允许以 "iwr ... | iex" 方式执行：默认调用 Install-DelGuard
if ($PSBoundParameters.Count -eq 0 -and $args.Count -eq 0) {
  Install-DelGuard
} elseif ($PSCmdlet.ParameterSetName -eq "Run") {
  Install-DelGuard -DefaultInteractive:$DefaultInteractive
} else {
  Install-DelGuard
}