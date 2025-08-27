#!/bin/bash
#
# DelGuard 一键卸载脚本 - Linux版本
#
# 功能：
# - 自动卸载 DelGuard 安全删除工具
# - 清理配置文件和别名设置
#
# 使用方法：
# ./uninstall.sh         # 标准卸载，会提示确认
# ./uninstall.sh --force # 强制卸载，不提示确认
# ./uninstall.sh --keep-config # 卸载但保留配置文件

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # 无颜色

# 常量定义
APP_NAME="DelGuard"
EXECUTABLE_NAME="delguard"

# 默认参数
FORCE=false
KEEP_CONFIG=false

# 解析命令行参数
for arg in "$@"; do
  case $arg in
    --force)
      FORCE=true
      shift
      ;;
    --keep-config)
      KEEP_CONFIG=true
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
  echo "║                🗑️ DelGuard 一键卸载工具                      ║"
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

# 查找配置目录
find_config_dir() {
  # 检查常见配置位置
  local possible_locations=(
    "/etc/$APP_NAME"
    "$HOME/.config/$APP_NAME"
    "$HOME/.$APP_NAME"
  )
  
  for location in "${possible_locations[@]}"; do
    if [ -d "$location" ]; then
      echo "$location"
      return 0
    fi
  done
  
  return 1
}

# 移除shell别名
remove_shell_aliases() {
  local shell_rc_files=(
    "$HOME/.bashrc"
    "$HOME/.bash_profile"
    "$HOME/.zshrc"
  )
  
  for rc_file in "${shell_rc_files[@]}"; do
    if [ -f "$rc_file" ]; then
      # 检查是否包含DelGuard配置
      if grep -q "# DelGuard" "$rc_file"; then
        # 移除DelGuard相关配置
        sed -i '/# DelGuard/,/^fi$/d' "$rc_file" 2>/dev/null
        echo -e "${GREEN}已从shell配置文件移除DelGuard别名: $rc_file${NC}"
      fi
    fi
  done
}

# 主程序
show_banner

# 查找已安装的DelGuard
installed_path=$(find_installed_delguard)
if [ -z "$installed_path" ]; then
  echo -e "${YELLOW}未找到已安装的DelGuard。${NC}"
  exit 0
fi

install_dir=$(dirname "$installed_path")
echo -e "${CYAN}已找到DelGuard: $installed_path${NC}"

# 查找配置目录
config_dir=$(find_config_dir)
if [ -n "$config_dir" ]; then
  echo -e "${CYAN}已找到配置目录: $config_dir${NC}"
fi

# 确认卸载
if [ "$FORCE" = false ]; then
  read -p "确认卸载DelGuard？(Y/N) " confirmation
  if [ "$confirmation" != "Y" ] && [ "$confirmation" != "y" ]; then
    echo -e "${YELLOW}卸载已取消。${NC}"
    exit 0
  fi
fi

# 停止可能正在运行的DelGuard进程
pkill -f "$installed_path" 2>/dev/null

# 删除可执行文件
rm -f "$installed_path"
echo -e "${GREEN}已删除可执行文件: $installed_path${NC}"

# 移除shell别名
remove_shell_aliases

# 处理配置目录
if [ -n "$config_dir" ] && [ "$KEEP_CONFIG" = false ]; then
  rm -rf "$config_dir"
  echo -e "${GREEN}已删除配置目录: $config_dir${NC}"
elif [ -n "$config_dir" ]; then
  echo -e "${CYAN}已保留配置目录: $config_dir${NC}"
fi

echo -e "${GREEN}DelGuard卸载完成！${NC}"