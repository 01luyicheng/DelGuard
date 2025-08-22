#!/usr/bin/env bash
# DelGuard macOS/Linux 一键安装（用户级，无 sudo）
# 用法（本地仓库内执行，优先使用本地产物/本机构建）:
#   bash scripts/install.sh
#   bash scripts/install.sh --default-interactive
#
# 用法（远程下载，需事先导出 GitHub 仓库 owner/repo）:
#   export DELGUARD_GITHUB_REPO="YourOrg/DelGuard"
#   bash -c "$(curl -fsSL https://raw.githubusercontent.com/${DELGUARD_GITHUB_REPO}/main/scripts/install.sh)" -- --default-interactive
#
# 下载检查：若发布资产包含 .sha256，将自动校验

set -euo pipefail

DEFAULT_INTERACTIVE=0
if [[ "${1:-}" == "--default-interactive" ]]; then
  DEFAULT_INTERACTIVE=1
fi

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" || script_dir=""
proj_dir=""
if [[ -n "${script_dir}" && -d "${script_dir}/.." ]]; then
  proj_dir="$( cd "${script_dir}/.." && pwd )"
fi

target_dir="${HOME}/.local/bin"
mkdir -p "${target_dir}"

# 体系结构与平台判定
detect_os_arch() {
  local os arch
  case "$(uname -s)" in
    Linux)  os="linux" ;;
    Darwin) os="darwin" ;;
    *) echo "不支持的系统：$(uname -s)" >&2; return 1 ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64)   arch="amd64" ;;
    arm64|aarch64)  arch="arm64" ;;
    *) echo "不支持的架构：$(uname -m)" >&2; return 1 ;;
  esac
  echo "${os} ${arch}"
}

# 首选本地二进制
detect_src() {
  [[ -n "${proj_dir}" ]] || return 1
  local os arch
  read -r os arch < <(detect_os_arch) || return 1
  local candidates=(
    "${proj_dir}/build/delguard-${os}-${arch}"
    "${proj_dir}/delguard"
  )
  for c in "${candidates[@]}"; do
    if [[ -f "$c" && -x "$c" ]]; then
      echo "$c"; return 0
    fi
  done
  return 1
}

have() { command -v "$1" >/dev/null 2>&1; }

download_file() {
  # $1=url $2=dest
  if have curl; then
    curl -fsSL "$1" -o "$2"
  elif have wget; then
    wget -qO "$2" "$1"
  else
    echo "缺少下载工具：请安装 curl 或 wget" >&2
    return 1
  fi
}

install_from_remote() {
  local repo="${DELGUARD_GITHUB_REPO:-}"
  if [[ -z "${repo}" ]]; then
    echo "未找到本地二进制，也未设置 DELGUARD_GITHUB_REPO（形如 YourOrg/DelGuard）以执行远程安装。" >&2
    return 1
  fi
  local os arch
  read -r os arch < <(detect_os_arch)

  local asset="delguard-${os}-${arch}"
  local base="https://github.com/${repo}/releases/latest/download"
  local url_bin="${base}/${asset}"
  local url_sha="${url_bin}.sha256"

  local tmp="$(mktemp -t delguard.XXXXXX)"
  trap 'rm -f "$tmp" "$tmp.sha256" 2>/dev/null || true' EXIT

  echo "下载 ${url_bin} ..."
  download_file "${url_bin}" "${tmp}"

  # 可选校验
  if download_file "${url_sha}" "${tmp}.sha256"; then
    echo "正在校验 SHA256..."
    if have sha256sum; then
      # 兼容 "HASH  FILENAME" 或 "HASH FILENAME"
      # 将文件名替换为临时名后校验
      awk '{print $1"  '"${tmp//\//\\/}"'"}' "${tmp}.sha256" | sha256sum -c -
    elif have shasum; then
      local expected
      expected="$(awk '{print $1}' "${tmp}.sha256")"
      local actual
      actual="$(shasum -a 256 "${tmp}" | awk '{print $1}')"
      if [[ "${expected^^}" != "${actual^^}" ]]; then
        echo "校验失败：期望 ${expected} 实际 ${actual}" >&2
        return 1
      fi
    else
      echo "未找到校验工具（sha256sum/shasum），跳过校验。" >&2
    fi
  else
    echo "未找到校验文件，跳过校验。"
  fi

  install -m 0755 "${tmp}" "${target_dir}/delguard"
  echo "已安装到 ${target_dir}/delguard"
}

# 优先本地，其次远程，最后本机构建
if src="$(detect_src)"; then
  cp -f "${src}" "${target_dir}/delguard"
  chmod +x "${target_dir}/delguard"
else
  if ! install_from_remote; then
    # 远程失败则尝试本机构建
    if ! have go; then
      echo "无法安装：未找到可用二进制，也无法下载，且未检测到 Go 构建环境。" >&2
      exit 1
    fi
    if [[ -n "${proj_dir}" ]]; then
      ( cd "${proj_dir}" && go build -o "${target_dir}/delguard" . )
    else
      echo "无法定位源码目录进行本地构建。" >&2
      exit 1
    fi
  fi
fi

# PATH 提示
case ":${PATH}:" in
  *":${HOME}/.local/bin:"*) ;;
  *) echo '提示：请将 ~/.local/bin 加入 PATH，例如在 ~/.bashrc 或 ~/.zshrc 添加： export PATH="$HOME/.local/bin:$PATH"';;
esac

# 安装别名
if [[ "${DEFAULT_INTERACTIVE}" == "1" ]]; then
  "${target_dir}/delguard" --install --default-interactive
else
  "${target_dir}/delguard" --install
fi

echo
echo "安装完成。新开一个终端后可使用："
echo "  rm -i file.txt     # 交互删除"
echo "  rm -rf folder      # 递归强制删除（将进入回收站）"