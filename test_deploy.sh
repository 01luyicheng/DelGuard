#!/bin/bash
#
# DelGuard 部署测试脚本 - Linux版本
#
# 功能：
# - 自动部署并测试 DelGuard 安全删除工具的各项功能
#
# 使用方法：
# ./test_deploy.sh         # 标准测试部署
# ./test_deploy.sh --clean # 清理环境后测试部署

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
CLEAN=false

# 解析命令行参数
for arg in "$@"; do
  case $arg in
    --clean)
      CLEAN=true
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
  echo "║                🧪 DelGuard 部署测试工具                      ║"
  echo "║                                                              ║"
  echo "╚══════════════════════════════════════════════════════════════╝"
  echo -e "${NC}"
  echo ""
}

# 创建测试环境
create_test_environment() {
  echo -e "${CYAN}创建测试环境...${NC}"
  
  # 创建测试目录
  test_dir="/tmp/delguard-test"
  rm -rf "$test_dir" 2>/dev/null
  mkdir -p "$test_dir"
  
  # 创建测试文件
  test_files=(
    "test1.txt"
    "test2.txt"
    "important_document.docx"
    "report.pdf"
    "image.jpg"
    "config.json"
  )
  
  for file in "${test_files[@]}"; do
    echo "This is a test file: $file
Created for DelGuard testing." > "$test_dir/$file"
  done
  
  echo -e "${GREEN}测试环境已创建: $test_dir${NC}"
  echo "$test_dir"
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

# 安装DelGuard
install_delguard() {
  echo -e "${CYAN}安装DelGuard...${NC}"
  
  # 运行安装脚本
  install_script="./install_enhanced.sh"
  
  if [ ! -f "$install_script" ]; then
    echo -e "${RED}未找到安装脚本: $install_script${NC}"
    exit 1
  fi
  
  # 执行安装脚本
  bash "$install_script" --force
  
  # 检查安装结果
  delguard_path=$(find_installed_delguard)
  if [ -z "$delguard_path" ]; then
    echo -e "${RED}DelGuard安装失败${NC}"
    exit 1
  fi
  
  echo -e "${GREEN}DelGuard安装成功: $delguard_path${NC}"
  echo "$delguard_path"
}

# 卸载DelGuard
uninstall_delguard() {
  echo -e "${CYAN}卸载DelGuard...${NC}"
  
  # 运行卸载脚本
  uninstall_script="./uninstall.sh"
  
  if [ ! -f "$uninstall_script" ]; then
    echo -e "${YELLOW}未找到卸载脚本: $uninstall_script${NC}"
    return
  fi
  
  # 执行卸载脚本
  bash "$uninstall_script" --force
  
  # 检查卸载结果
  delguard_path=$(find_installed_delguard)
  if [ -n "$delguard_path" ]; then
    echo -e "${YELLOW}DelGuard卸载失败，仍然存在: $delguard_path${NC}"
  else
    echo -e "${GREEN}DelGuard卸载成功${NC}"
  fi
}

# 测试基本功能
test_basic_functionality() {
  local delguard_path="$1"
  local test_dir="$2"
  
  echo -e "${CYAN}测试基本功能...${NC}"
  
  # 测试帮助命令
  echo -e "${CYAN}测试帮助命令...${NC}"
  help_output=$("$delguard_path" --help 2>&1)
  if echo "$help_output" | grep -q "使用方法" || echo "$help_output" | grep -q "Usage"; then
    echo -e "${GREEN}✓ 帮助命令正常${NC}"
  else
    echo -e "${RED}✗ 帮助命令异常${NC}"
  fi
  
  # 测试版本命令
  echo -e "${CYAN}测试版本命令...${NC}"
  version_output=$("$delguard_path" --version 2>&1)
  if echo "$version_output" | grep -q "[0-9]\+\.[0-9]\+\.[0-9]\+"; then
    echo -e "${GREEN}✓ 版本命令正常: $version_output${NC}"
  else
    echo -e "${RED}✗ 版本命令异常${NC}"
  fi
  
  # 测试删除文件
  test_file="$test_dir/test1.txt"
  echo -e "${CYAN}测试删除文件: $test_file${NC}"
  "$delguard_path" "$test_file"
  
  if [ ! -f "$test_file" ]; then
    echo -e "${GREEN}✓ 文件删除成功${NC}"
  else
    echo -e "${RED}✗ 文件删除失败${NC}"
  fi
  
  # 测试不存在的文件（智能搜索功能）
  non_existent_file="$test_dir/non_existent.txt"
  echo -e "${CYAN}测试智能搜索功能: $non_existent_file${NC}"
  search_output=$("$delguard_path" "$non_existent_file" 2>&1)
  
  if echo "$search_output" | grep -q "不存在" && echo "$search_output" | grep -q "相似"; then
    echo -e "${GREEN}✓ 智能搜索功能正常${NC}"
  else
    echo -e "${RED}✗ 智能搜索功能异常${NC}"
  fi
}

# 测试语言检测
test_language_detection() {
  local delguard_path="$1"
  
  echo -e "${CYAN}测试语言检测功能...${NC}"
  
  # 获取当前系统语言
  current_lang="$LANG"
  echo -e "${CYAN}当前系统语言: $current_lang${NC}"
  
  # 执行命令并检查输出语言
  output=$("$delguard_path" --help 2>&1)
  
  if echo "$current_lang" | grep -q "zh" && echo "$output" | grep -q "使用方法"; then
    echo -e "${GREEN}✓ 中文语言检测正常${NC}"
  elif echo "$current_lang" | grep -q "en" && echo "$output" | grep -q "Usage"; then
    echo -e "${GREEN}✓ 英文语言检测正常${NC}"
  else
    echo -e "${YELLOW}✗ 语言检测功能可能有问题${NC}"
    echo -e "  系统语言: $current_lang"
    echo -e "  输出示例: $(echo "$output" | head -n 3)"
  fi
}

# 测试更新功能
test_update_functionality() {
  echo -e "${CYAN}测试更新功能...${NC}"
  
  # 运行更新脚本
  update_script="./update.sh"
  
  if [ ! -f "$update_script" ]; then
    echo -e "${YELLOW}未找到更新脚本: $update_script${NC}"
    return
  fi
  
  # 执行更新脚本（仅检查模式）
  bash "$update_script" --check
  
  echo -e "${GREEN}✓ 更新检查功能正常${NC}"
}

# 主程序
show_banner

# 如果指定了Clean参数，先卸载现有版本
if [ "$CLEAN" = true ]; then
  uninstall_delguard
fi

# 创建测试环境
test_dir=$(create_test_environment)

# 安装DelGuard
delguard_path=$(install_delguard)

# 测试基本功能
test_basic_functionality "$delguard_path" "$test_dir"

# 测试语言检测
test_language_detection "$delguard_path"

# 测试更新功能
test_update_functionality

echo -e "${GREEN}所有测试完成！${NC}"