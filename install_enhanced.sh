#!/bin/bash
#
# DelGuard 增强安装脚本 - Linux版本
#
# 功能：
# - 自动下载并安装 DelGuard 安全删除工具
# - 智能语言检测，自动配置界面语言
# - 环境兼容性检查
# - 一键卸载功能
#
# 使用方法：
# ./install_enhanced.sh         # 标准安装
# ./install_enhanced.sh --force # 强制重新安装
# ./install_enhanced.sh --system # 系统级安装（需要sudo权限）
# ./install_enhanced.sh --uninstall # 卸载
# ./install_enhanced.sh --status # 检查安装状态

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
VERSION="2.1.0"

# 默认参数
FORCE=false
SYSTEM_WIDE=false
UNINSTALL=false
STATUS=false

# 解析命令行参数
for arg in "$@"; do
  case $arg in
    --force)
      FORCE=true
      shift
      ;;
    --system)
      SYSTEM_WIDE=true
      shift
      ;;
    --uninstall)
      UNINSTALL=true
      shift
      ;;
    --status)
      STATUS=true
      shift
      ;;
    *)
      # 未知参数
      ;;
  esac
done

# 路径配置
if [ "$SYSTEM_WIDE" = true ]; then
  INSTALL_DIR="/usr/local/bin"
  CONFIG_DIR="/etc/$APP_NAME"
else
  INSTALL_DIR="$HOME/.local/bin"
  CONFIG_DIR="$HOME/.config/$APP_NAME"
fi

EXECUTABLE_PATH="$INSTALL_DIR/$EXECUTABLE_NAME"
LOG_FILE="$CONFIG_DIR/install.log"

# 确保日志目录存在
mkdir -p "$(dirname "$LOG_FILE")" 2>/dev/null

# 日志函数
log() {
  local level="$1"
  local message="$2"
  local timestamp=$(date "+%Y-%m-%d %H:%M:%S")
  local log_message="[$timestamp] [$level] $message"
  
  case $level in
    "INFO")
      echo -e "${CYAN}$log_message${NC}"
      ;;
    "ERROR")
      echo -e "${RED}$log_message${NC}"
      ;;
    "WARNING")
      echo -e "${YELLOW}$log_message${NC}"
      ;;
    "SUCCESS")
      echo -e "${GREEN}$log_message${NC}"
      ;;
    *)
      echo "$log_message"
      ;;
  esac
  
  echo "$log_message" >> "$LOG_FILE"
}

# 显示横幅
show_banner() {
  echo -e "${MAGENTA}"
  echo "╔══════════════════════════════════════════════════════════════╗"
  echo "║                                                              ║"
  echo "║                    🛡️  DelGuard $VERSION                    ║"
  echo "║                   安全文件删除工具                           ║"
  echo "║                                                              ║"
  echo "╚══════════════════════════════════════════════════════════════╝"
  echo -e "${NC}"
  echo ""
}

# 检查是否为root用户
is_root() {
  [ "$(id -u)" -eq 0 ]
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

# 检查网络连接
check_network() {
  if ! curl -s --head https://api.github.com > /dev/null; then
    return 1
  fi
  return 0
}

# 获取最新版本信息
get_latest_release() {
  log "INFO" "获取最新版本信息..."
  if ! curl -s "$RELEASE_API" > /tmp/delguard_release.json; then
    log "ERROR" "获取版本信息失败"
    return 1
  fi
  return 0
}

# 下载文件
download_file() {
  local url="$1"
  local output_path="$2"
  
  log "INFO" "下载文件: $url"
  if ! curl -L -o "$output_path" "$url"; then
    log "ERROR" "下载失败: $url"
    return 1
  fi
  log "SUCCESS" "下载完成: $output_path"
  return 0
}

# 检查系统环境
check_system_environment() {
  log "INFO" "检查系统环境..."
  
  # 检查操作系统
  local os_name=$(uname -s)
  local os_version=$(uname -r)
  log "INFO" "操作系统: $os_name $os_version"
  
  # 检查发行版（如果是Linux）
  if [ "$os_name" = "Linux" ]; then
    if [ -f /etc/os-release ]; then
      source /etc/os-release
      log "INFO" "Linux发行版: $NAME $VERSION_ID"
    fi
  fi
  
  # 检查shell
  local shell=$(basename "$SHELL")
  log "INFO" "当前Shell: $shell"
  
  # 检查磁盘空间
  local free_space=$(df -h . | awk 'NR==2 {print $4}')
  log "INFO" "可用磁盘空间: $free_space"
  
  # 检查必要工具
  for tool in curl wget unzip tar; do
    if command -v $tool > /dev/null; then
      log "INFO" "已安装: $tool"
    else
      log "WARNING" "未安装: $tool（可能影响安装过程）"
    fi
  done
  
  # 检查系统语言
  local lang="$LANG"
  log "INFO" "系统语言: $lang"
  
  log "SUCCESS" "系统环境检查完成"
}

# 检测系统语言并设置DelGuard语言
set_delguard_language() {
  log "INFO" "检测系统语言..."
  
  # 获取系统语言
  local lang="$LANG"
  log "INFO" "检测到系统语言: $lang"
  
  # 确定DelGuard使用的语言
  local delguard_lang="en-US" # 默认英语
  
  if [[ "$lang" == zh_* ]]; then
    delguard_lang="zh-CN"
    log "INFO" "将使用中文(简体)作为DelGuard界面语言"
  elif [[ "$lang" == ja_* ]]; then
    delguard_lang="ja"
    log "INFO" "将使用日语作为DelGuard界面语言"
  else
    log "INFO" "将使用英语作为DelGuard界面语言"
  fi
  
  # 创建或更新DelGuard语言配置
  local config_file="$CONFIG_DIR/config.json"
  
  # 确保配置目录存在
  mkdir -p "$CONFIG_DIR"
  
  # 读取现有配置（如果存在）
  if [ -f "$config_file" ]; then
    # 尝试更新现有配置
    if command -v jq > /dev/null; then
      # 使用jq更新配置
      jq --arg lang "$delguard_lang" '.language = $lang' "$config_file" > "$config_file.tmp" && mv "$config_file.tmp" "$config_file"
    else
      # 简单替换（不太可靠，但在没有jq的情况下尝试）
      sed -i 's/"language":[[:space:]]*"[^"]*"/"language": "'$delguard_lang'"/g' "$config_file" 2>/dev/null
    fi
  else
    # 创建新配置
    echo '{
  "language": "'$delguard_lang'"
}' > "$config_file"
  fi
  
  log "SUCCESS" "DelGuard语言配置已更新为: $delguard_lang"
}

# 安装 DelGuard
install_delguard() {
  log "INFO" "开始安装 $APP_NAME..."
  
  # 检查权限（系统级安装时）
  if [ "$SYSTEM_WIDE" = true ] && ! is_root; then
    log "ERROR" "系统级安装需要root权限"
    log "ERROR" "请使用sudo运行此脚本"
    return 1
  fi
  
  # 检查系统环境
  check_system_environment
  
  # 检查网络连接
  if ! check_network; then
    log "ERROR" "网络连接检查失败"
    log "ERROR" "无法连接到GitHub，请检查网络连接"
    return 1
  fi
  
  # 检查现有安装
  if [ -f "$EXECUTABLE_PATH" ] && [ "$FORCE" = false ]; then
    log "WARNING" "$APP_NAME 已经安装在 $EXECUTABLE_PATH"
    log "INFO" "使用 --force 参数强制重新安装"
    return 0
  fi
  
  # 获取最新版本
  if ! get_latest_release; then
    return 1
  fi
  
  # 解析版本信息
  local version=$(grep -o '"tag_name": *"[^"]*"' /tmp/delguard_release.json | cut -d'"' -f4)
  log "SUCCESS" "最新版本: $version"
  
  # 确定下载URL
  local arch=$(get_system_architecture)
  local asset_name="${APP_NAME}-linux-${arch}.tar.gz"
  local download_url=$(grep -o '"browser_download_url": *"[^"]*'${asset_name}'"' /tmp/delguard_release.json | cut -d'"' -f4)
  
  if [ -z "$download_url" ]; then
    log "ERROR" "未找到适合的安装包: $asset_name"
    return 1
  fi
  
  log "INFO" "下载URL: $download_url"
  
  # 创建临时目录
  local temp_dir="/tmp/delguard-install"
  rm -rf "$temp_dir" 2>/dev/null
  mkdir -p "$temp_dir"
  
  # 下载文件
  local archive_path="$temp_dir/$asset_name"
  if ! download_file "$download_url" "$archive_path"; then
    return 1
  fi
  
  # 解压文件
  log "INFO" "解压安装包..."
  tar -xzf "$archive_path" -C "$temp_dir"
  
  # 创建安装目录
  mkdir -p "$INSTALL_DIR"
  
  # 复制文件
  local extracted_exe=$(find "$temp_dir" -name "$EXECUTABLE_NAME" -type f | head -n 1)
  if [ -n "$extracted_exe" ]; then
    cp "$extracted_exe" "$EXECUTABLE_PATH"
    chmod +x "$EXECUTABLE_PATH"
    log "SUCCESS" "已安装到: $EXECUTABLE_PATH"
  else
    log "ERROR" "在安装包中未找到可执行文件"
    return 1
  fi
  
  # 添加到 PATH
  add_to_path
  
  # 安装shell别名
  install_shell_aliases
  
  # 创建配置目录
  mkdir -p "$CONFIG_DIR"
  
  # 设置DelGuard语言
  set_delguard_language
  
  # 清理临时文件
  rm -rf "$temp_dir"
  
  log "SUCCESS" "$APP_NAME $version 安装成功！"
  log "INFO" "可执行文件位置: $EXECUTABLE_PATH"
  log "INFO" "配置目录: $CONFIG_DIR"
  log "INFO" ""
  log "INFO" "使用方法:"
  log "INFO" "  delguard file.txt          # 删除文件到回收站"
  log "INFO" "  delguard -p file.txt       # 永久删除文件"
  log "INFO" "  delguard --help            # 查看帮助"
  log "INFO" ""
  log "INFO" "请重新启动终端以使用 delguard 命令"
  
  return 0
}

# 添加到 PATH
add_to_path() {
  # 检查INSTALL_DIR是否已在PATH中
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    local shell_rc=""
    
    # 确定使用的shell配置文件
    if [ -n "$BASH_VERSION" ]; then
      if [ -f "$HOME/.bashrc" ]; then
        shell_rc="$HOME/.bashrc"
      elif [ -f "$HOME/.bash_profile" ]; then
        shell_rc="$HOME/.bash_profile"
      fi
    elif [ -n "$ZSH_VERSION" ]; then
      shell_rc="$HOME/.zshrc"
    fi
    
    if [ -n "$shell_rc" ]; then
      # 添加到PATH
      echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_rc"
      log "SUCCESS" "已添加到PATH: $INSTALL_DIR (在 $shell_rc)"
    else
      log "WARNING" "无法确定shell配置文件，请手动添加 $INSTALL_DIR 到PATH"
    fi
  else
    log "INFO" "PATH中已存在: $INSTALL_DIR"
  fi
}

# 安装shell别名
install_shell_aliases() {
  local shell_rc=""
  
  # 确定使用的shell配置文件
  if [ -n "$BASH_VERSION" ]; then
    if [ -f "$HOME/.bashrc" ]; then
      shell_rc="$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
      shell_rc="$HOME/.bash_profile"
    fi
  elif [ -n "$ZSH_VERSION" ]; then
    shell_rc="$HOME/.zshrc"
  fi
  
  if [ -n "$shell_rc" ]; then
    # 检查是否已经配置了别名
    if ! grep -q "# DelGuard 别名配置" "$shell_rc"; then
      cat >> "$shell_rc" << EOF

# DelGuard 别名配置
if [ -f "$EXECUTABLE_PATH" ]; then
  alias delguard="$EXECUTABLE_PATH"
  alias dg="$EXECUTABLE_PATH"
  # 兼容Unix命令
  alias rm="$EXECUTABLE_PATH"
fi
EOF
      log "SUCCESS" "已添加shell别名配置"
    else
      log "INFO" "shell别名已存在"
    fi
  else
    log "WARNING" "无法确定shell配置文件，请手动配置别名"
  fi
}

# 卸载 DelGuard
uninstall_delguard() {
  log "INFO" "开始卸载 $APP_NAME..."
  
  # 删除可执行文件
  if [ -f "$EXECUTABLE_PATH" ]; then
    rm -f "$EXECUTABLE_PATH"
    log "SUCCESS" "已删除: $EXECUTABLE_PATH"
  fi
  
  # 从PATH中移除
  remove_from_path
  
  # 移除shell别名
  remove_shell_aliases
  
  log "SUCCESS" "$APP_NAME 卸载完成"
  log "INFO" "配置文件保留在: $CONFIG_DIR"
  log "INFO" "如需完全清理，请手动删除配置目录"
  
  return 0
}

# 从PATH中移除
remove_from_path() {
  local shell_rc=""
  
  # 确定使用的shell配置文件
  if [ -n "$BASH_VERSION" ]; then
    if [ -f "$HOME/.bashrc" ]; then
      shell_rc="$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
      shell_rc="$HOME/.bash_profile"
    fi
  elif [ -n "$ZSH_VERSION" ]; then
    shell_rc="$HOME/.zshrc"
  fi
  
  if [ -n "$shell_rc" ]; then
    # 从PATH中移除
    sed -i '/export PATH="'"$INSTALL_DIR"':\$PATH"/d' "$shell_rc" 2>/dev/null
    log "SUCCESS" "已从PATH中移除: $INSTALL_DIR"
  fi
}

# 移除shell别名
remove_shell_aliases() {
  local shell_rc=""
  
  # 确定使用的shell配置文件
  if [ -n "$BASH_VERSION" ]; then
    if [ -f "$HOME/.bashrc" ]; then
      shell_rc="$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
      shell_rc="$HOME/.bash_profile"
    fi
  elif [ -n "$ZSH_VERSION" ]; then
    shell_rc="$HOME/.zshrc"
  fi
  
  if [ -n "$shell_rc" ]; then
    # 删除别名配置块
    sed -i '/# DelGuard 别名配置/,/^fi$/d' "$shell_rc" 2>/dev/null
    log "SUCCESS" "已移除shell别名配置"
  fi
}

# 检查安装状态
check_install_status() {
  echo -e "${MAGENTA}=== DelGuard 安装状态 ===${NC}"
  
  if [ -f "$EXECUTABLE_PATH" ]; then
    echo -e "${GREEN}✓ 已安装${NC}"
    echo "  位置: $EXECUTABLE_PATH"
    
    if [ -x "$EXECUTABLE_PATH" ]; then
      local version=$("$EXECUTABLE_PATH" --version 2>/dev/null)
      if [ -n "$version" ]; then
        echo "  版本: $version"
      else
        echo -e "${YELLOW}  版本: 无法获取${NC}"
      fi
    fi
  else
    echo -e "${RED}✗ 未安装${NC}"
  fi
  
  # 检查PATH
  if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
    echo -e "${GREEN}✓ 已添加到PATH${NC}"
  else
    echo -e "${YELLOW}✗ 未添加到PATH${NC}"
  fi
  
  # 检查别名
  local has_alias=false
  if command -v delguard > /dev/null; then
    echo -e "${GREEN}✓ delguard命令可用${NC}"
    has_alias=true
  fi
  
  if [ "$has_alias" = false ]; then
    echo -e "${YELLOW}✗ delguard命令不可用${NC}"
  fi
  
  # 检查配置目录
  if [ -d "$CONFIG_DIR" ]; then
    echo -e "${GREEN}✓ 配置目录存在: $CONFIG_DIR${NC}"
  else
    echo -e "${YELLOW}✗ 配置目录不存在${NC}"
  fi
}

# 主程序
show_banner

# 根据参数执行相应操作
if [ "$STATUS" = true ]; then
  check_install_status
elif [ "$UNINSTALL" = true ]; then
  uninstall_delguard
else
  install_delguard
fi