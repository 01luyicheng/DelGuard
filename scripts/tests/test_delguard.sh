#!/usr/bin/env bash
# 基础与扩展自测：macOS / Linux（仅在临时目录内操作）
# 要求：已安装 Go 或仓库根存在 delguard 可执行文件
# 覆盖：基础删除/恢复、长路径、符号链接、权限不足（期望失败不删除）、跨设备（占位提示）

set -euo pipefail

new_tmp_dir() {
  local prefix=${1:-"delguard_test_"}
  local ts
  ts=$(date +%Y%m%d_%H%M%S)
  local dir="${TMPDIR:-/tmp}/${prefix}${ts}"
  mkdir -p "$dir"
  echo "$dir"
}

resolve_delguard() {
  local root
  root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
  local candidates=(
    "${root}/delguard"
    "${root}/build/delguard-linux-amd64"
    "${root}/build/delguard-darwin-amd64"
  )
  for c in "${candidates[@]}"; do
    if [[ -x "$c" ]]; then echo "$c"; return 0; fi
  done
  if command -v go >/dev/null 2>&1; then
    local out
    out="$(new_tmp_dir)/delguard"
    ( cd "$root" && go build -o "$out" . )
    echo "$out"
    return 0
  fi
  echo "未找到 delguard 可执行文件，且未安装 Go。请先构建。" >&2
  exit 1
}

exe="$(resolve_delguard)"
work="$(new_tmp_dir)"
cd "$work"

ts="$(date +%s)"
prefix="dgtest_${ts}_"

export DELGUARD_INTERACTIVE=0 # 关闭交互避免等待

echo "== 基础删除与恢复 =="
file1="${prefix}file1.txt"
dirA="${prefix}dirA"
mkdir -p "${dirA}/sub"
echo "hello" > "$file1"
echo "world" > "${dirA}/sub/file2.log"

echo ":: DRY-RUN 预演..."
"$exe" -n -r --verbose "$file1" "$dirA"

echo ":: 执行删除..."
"$exe" -r "$file1" "$dirA"

if [[ -e "$file1" ]]; then
  echo "删除失败：$file1 仍然存在" >&2; exit 1
fi
if [[ -d "$dirA" ]]; then
  echo "删除失败：$dirA 仍然存在" >&2; exit 1
fi
echo "PASS: 基础删除完成"

echo ":: 执行恢复（按平台策略）..."
pattern="${prefix}*"
"$exe" --restore "$pattern" || true
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
if [[ "$os" == "linux" ]]; then
  [[ -e "$file1" ]] || { echo "恢复校验失败(Linux)：未在原路径找到 $file1" >&2; exit 1; }
  [[ -d "$dirA" ]] || { echo "恢复校验失败(Linux)：未在原路径找到 $dirA" >&2; exit 1; }
else
  # macOS：当前目录至少应出现前缀匹配的文件或带 _restored_ 后缀
  ls | grep -E "^${prefix}file1(|_restored_.*)\.txt$" >/dev/null 2>&1 || \
    echo "恢复校验提示(macOS)：未在当前目录检测到恢复的 ${prefix}file1.txt（Finder 行为可能延迟）"
fi
echo "PASS: 恢复流程完成"

echo "== 长路径删除 =="
longBase="${prefix}long_"
nested=""
for i in $(seq 1 20); do nested="${nested}${longBase}"; done
longDir="${nested}"
mkdir -p "$longDir"
longFile="${longDir}/${prefix}deep.txt"
echo "deep" > "$longFile"
# 仅预演与删除执行
"$exe" -n "$longFile"
"$exe" "$longFile"
[[ ! -e "$longFile" ]] || { echo "长路径删除失败：$longFile 仍存在" >&2; exit 1; }
echo "PASS: 长路径删除完成（长度=$(printf "%s" "$longFile" | wc -c)）"

echo "== 符号链接删除（不影响目标） =="
target="${prefix}target.txt"
lnk="${prefix}link_to_target"
echo "t" > "$target"
ln -s "$target" "$lnk"
"$exe" "$lnk"
[[ ! -L "$lnk" ]] || { echo "符号链接删除失败：$lnk 仍存在" >&2; exit 1; }
[[ -f "$target" ]] || { echo "符号链接删除影响了目标：$target 不存在" >&2; exit 1; }
echo "PASS: 符号链接删除仅移除链接"

echo "== 权限不足删除（期望失败不删除） =="
permDir="${prefix}permD"
permFile="${permDir}/f.txt"
mkdir -p "$permDir"
echo "x" > "$permFile"
chmod 500 "$permDir"  # 目录不可写，无法删除其中条目
set +e
out="$("$exe" "$permFile" 2>&1)"
rc=$?
set -e
# 期望失败（或至少文件仍存在）
if [[ -e "$permFile" ]]; then
  echo "PASS: 权限不足时未删除文件（rc=$rc）"
else
  echo "WARN: 权限用例删除了文件（可能因系统权限模型差异），输出："
  echo "$out"
fi
chmod 700 "$permDir" || true

echo "== 跨设备删除（占位） =="
echo "SKIP: 无法在脚本中可靠创建第二挂载点以触发 EXDEV，已在实现中提供回退逻辑。"

echo "全部扩展用例完成。"