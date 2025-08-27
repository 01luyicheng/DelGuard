#!/bin/bash
#
# DelGuard 一键更新脚本 - Linux版本
#
# 功能：
# - 自动检查并更新 DelGuard 安全删除工具到最新版本
# - 备份现有版本
# - 验证更新结果
#
# 使用方法：
# ./update.sh         # 检查并更新DelGuard
# ./update.sh --force # 强制更新到最新版本
# ./update.sh --check # 仅检查是否有更新可用

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # 无颜色

# 常量定义
REPO_URL="https://github.com/01luyicheng/DelGuard"
RELEASE_API="https://api.github.com/repos/01luyicheng/DelGuard/releases/latest"
APP_NAME="DelGuard"
EXECUTABLE_NAME="delguard"

# 默认参数
FORCE=false
CHECK_ONLY=false

# 解析命令行参数
for arg in "$@"; do
  case $arg in
    --force)
      FORCE=true
      shift
      ;;
    --check)
      CHECK_ONLY=true
      shift
      ;;
    *)
      # 未知参数
      ;;
  esac
done

# 显示横幅
show_banner() {
  echo -e "${MAGENTA}"
  echo "╔══════════════════════════════════════════════════════════════╗"
  echo "║                                                              ║"
  echo "║                🔄 DelGuard 一键更新工具                      ║"
  echo "║                                                              ║"
  echo "╚══════════════════════════════════════════════════════════════╝"
  echo -e "${NC}"
  echo ""
}

# 查找已安装的DelGuard
find_installed_delguard() {
  # 检查常见安装位置
  local possible_locations=(
    "/usr/local/bin/$EXECUTABLE_NAME"
    "/usr/bin/$EXECUTABLE_NAME"
    "$HOME/.local/bin/$EXECUTABLE_NAME"
    "$HOME/bin/$EXECUTABLE_NAME"
  )
  
  for location in "${possible_locations[@]}"; do
    if [ -f "$location" ]; then
      echo "$location"
      return 0
    fi
  done
  
  # 尝试从PATH中查找
  local from_path=$(which $EXECUTABLE_NAME 2>/dev/null)
  if [ -n "$from_path" ]; then
    echo "$from_path"
    return 0
  fi
  
  return 1
}

# 获取已安装版本
get_installed_version() {
  local executable_path="$1"
  
  if [ -f "$executable_path" ] && [ -x "$executable_path" ]; then
    local output=$("$executable_path" --version 2>/dev/null)
    if [ -n "$output" ]; then
      # 提取版本号（假设格式为 "DelGuard v1.2.3" 或类似）
      local version=$(echo "$output" | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -n 1)
      if [ -n "$version" ]; then
        echo "$version"
        return 0
      fi
    fi
  fi
  
  echo "未知"
  return 1
}

# 获取最新版本信息
get_latest_release() {
  echo -e "${CYAN}获取最新版本信息...${NC}"
  
  if ! curl -s "$RELEASE_API" > /tmp/delguard_release.json; then
    echo -e "${RED}获取版本信息失败${NC}"
    return 1
  fi
  
  return 0
}

# 获取系统架构
get_system_architecture() {
  local arch=$(uname -m)
  case $arch in
    x86_64)
      echo "amd64"
      ;;
    aarch64|arm64)
      echo "arm64"
      ;;
    i386|i686)
      echo "386"
      ;;
    *)
      echo "amd64" # 默认
      ;;
  esac
}

# 下载文件
download_file() {
  local url="$1"
  local output_path="$2"
  
  echo -e "${CYAN}下载文件: $url${NC}"
  if ! curl -L -o "$output_path" "$url"; then
    echo -e "${RED}下载失败: $url${NC}"
    return 1
  fi
  echo -e "${GREEN}下载完成: $output_path${NC}"
  return 0
}

# 主程序
show_banner

# 查找已安装的DelGuard
installed_path=$(find_installed_delguard)
if [ -z "$installed_path" ]; then
  echo -e "${RED}未找到已安装的DelGuard。请先安装DelGuard。${NC}"
  exit 1
fi

install_dir=$(dirname "$installed_path")
echo -e "${GREEN}已找到DelGuard: $installed_path${NC}"

# 获取已安装版本
installed_version=$(get_installed_version "$installed_path")
echo -e "${CYAN}当前版本: $installed_version${NC}"

# 获取最新版本
if ! get_latest_release; then
  exit 1
fi

latest_version=$(grep -o '"tag_name": *"[^"]*"' /tmp/delguard_release.json | cut -d'"' -f4 | sed 's/^v//')
echo -e "${CYAN}最新版本: $latest_version${NC}"

# 比较版本
update_available=false
if [ "$FORCE" = true ] || [ "$installed_version" != "$latest_version" ] && [ "$installed_version" != "未知" ]; then
  update_available=true
fi

if [ "$update_available" = false ]; then
  echo -e "${GREEN}DelGuard已经是最新版本。${NC}"
  exit 0
fi

echo -e "${YELLOW}发现新版本！${NC}"

# 如果只是检查更新，则退出
if [ "$CHECK_ONLY" = true ]; then
  echo -e "${CYAN}有可用更新。使用不带 --check 参数的命令来执行更新。${NC}"
  exit 0
fi

# 确认更新
read -p "是否更新到最新版本？(Y/N) " confirmation
if [ "$confirmation" != "Y" ] && [ "$confirmation" != "y" ]; then
  echo -e "${YELLOW}更新已取消。${NC}"
  exit 0
fi

# 确定下载URL
arch=$(get_system_architecture)
asset_name="${APP_NAME}-linux-${arch}.tar.gz"
download_url=$(grep -o '"browser_download_url": *"[^"]*'${asset_name}'"' /tmp/delguard_release.json | cut -d'"' -f4)

if [ -z "$download_url" ]; then
  echo -e "${RED}未找到适合的安装包: $asset_name${NC}"
  exit 1
fi

# 创建临时目录
temp_dir="/tmp/delguard-update"
rm -rf "$temp_dir" 2>/dev/null
mkdir -p "$temp_dir"

# 下载文件
archive_path="$temp_dir/$asset_name"
if ! download_file "$download_url" "$archive_path"; then
  exit 1
fi

# 解压文件
echo -e "${CYAN}解压安装包...${NC}"
tar -xzf "$archive_path" -C "$temp_dir"

# 备份当前可执行文件
backup_path="${installed_path}.backup"
cp "$installed_path" "$backup_path"
echo -e "${CYAN}已备份当前版本到: $backup_path${NC}"

# 停止可能正在运行的DelGuard进程
pkill -f "$installed_path" 2>/dev/null

# 复制新文件
extracted_exe=$(find "$temp_dir" -name "$EXECUTABLE_NAME" -type f | head -n 1)
if [ -n "$extracted_exe" ]; then
  cp "$extracted_exe" "$installed_path"
  chmod +x "$installed_path"
  echo -e "${GREEN}已更新到: $installed_path${NC}"
else
  echo -e "${RED}在安装包中未找到可执行文件，恢复备份...${NC}"
  cp "$backup_path" "$installed_path"
  chmod +x "$installed_path"
  echo -e "${RED}更新失败：在安装包中未找到可执行文件${NC}"
  exit 1
fi

# 清理临时文件
rm -rf "$temp_dir"

# 验证更新
new_version=$(get_installed_version "$installed_path")
echo -e "${GREEN}DelGuard已成功更新到版本: $new_version${NC}"