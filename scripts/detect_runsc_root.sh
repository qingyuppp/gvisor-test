#!/usr/bin/env bash
# detect_runsc_root.sh  自动探测当前主机上活跃的 runsc rootDir。
# 使用: 直接执行输出一个 rootDir;  加 --all 列出所有候选及统计。
# 退出码: 0 成功找到; 非0 未找到。

set -euo pipefail

show_all=false
if [[ ${1-} == "--all" ]]; then
  show_all=true
fi

uid=$(id -u)
CANDIDATES=(
  "${RUNSC_ROOT-}"
  "/run/runsc"
  "/var/run/runsc"
  "/run/user/${uid}/runsc"
  "/tmp/runsc"
)

# 去重 & 过滤空
uniq_candidates=()
for d in "${CANDIDATES[@]}"; do
  [[ -n "$d" ]] || continue
  skip=false
  for o in "${uniq_candidates[@]}"; do
    [[ $o == $d ]] && skip=true && break
  done
  $skip || uniq_candidates+=("$d")
done

pattern='*_sand:*\.state'

score_root() {
  local root=$1
  local count=0
  local newest=0
  shopt -s nullglob
  for f in "${root}"/*_sand:*\.state; do
    [[ -f $f ]] || continue
    ((count++)) || true
    mtime=$(stat -c %Y "$f" 2>/dev/null || echo 0)
    if (( mtime > newest )); then
      newest=$mtime
    fi
  done
  echo "${count};${newest};${root}"
}

rows=()
for d in "${uniq_candidates[@]}"; do
  [[ -d $d ]] || continue
  rows+=("$(score_root "$d")")
done

# 额外扫描 /run 与 /var/run 二级目录，发现模式文件后加入
while IFS= read -r f; do
  r=$(dirname "$f")
  # 确保没有已在列表
  found=false
  for d in "${uniq_candidates[@]}"; do [[ $d == $r ]] && found=true && break; done
  $found || rows+=("$(score_root "$r")")
  uniq_candidates+=("$r")
done < <(find /run /var/run -maxdepth 2 -type f -name '*_sand:*\.state' 2>/dev/null | head -200)

# 过滤 count=0
filtered=()
for line in "${rows[@]}"; do
  IFS=';' read -r c t r <<<"$line"
  (( c > 0 )) && filtered+=("$line")
done

if (( ${#filtered[@]} == 0 )); then
  echo "未找到任何 runsc state 文件; 确认容器已使用 --runtime=runsc 启动" >&2
  exit 1
fi

# 按 count desc, newest desc 排序
sorted=$(printf '%s\n' "${filtered[@]}" | sort -t';' -k1,1nr -k2,2nr)
if $show_all; then
  printf 'count newestEpoch rootDir\n'
  echo "$sorted" | sed 's/;/ /g'
fi
best=$(echo "$sorted" | head -1 | awk -F';' '{print $3}')

if ! $show_all; then
  echo "$best"
else
  echo "\n推荐 rootDir: $best"
fi
