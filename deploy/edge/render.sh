#!/bin/sh
# deploy/edge/render.sh
# 把本目錄的 *.conf.template 渲染成 Nginx 站台設定：只帶入站台網域變數，保留 $host 等 Nginx 變數

# 用法：DRONE_HOST=… API_HOST=… API_STAGING_HOST=… deploy/edge/render.sh <輸出目錄> [樣板名…]
# 未指定樣板名時渲染全部；樣板用到的網域變數未設定就失敗，避免產出空的 server_name。
set -eu

out_dir=${1:?用法：render.sh <輸出目錄> [樣板名…]}
shift
src_dir=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$out_dir"

if [ "$#" -eq 0 ]; then
    set -- $(cd "$src_dir" && ls *.conf.template | sed 's/\.conf\.template$//')
fi

for name in "$@"; do
    template="$src_dir/$name.conf.template"
    [ -f "$template" ] || { echo "找不到樣板：$template" >&2; exit 1; }
    for var in $(grep -o '\${[A-Z_]*}' "$template" | sort -u | tr -d '${}'); do
        eval "value=\${$var:-}"
        [ -n "$value" ] || { echo "$name：環境變數 $var 未設定" >&2; exit 1; }
    done
    vars=$(grep -o '\${[A-Z_]*}' "$template" | sort -u | tr '\n' ' ')
    envsubst "$vars" < "$template" > "$out_dir/littleduck-$name.conf"
    echo "已產生 $out_dir/littleduck-$name.conf"
done
