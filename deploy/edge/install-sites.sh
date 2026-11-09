#!/bin/sh
# deploy/edge/install-sites.sh
# 在 Hetzner 主機上安裝渲染好的站台設定到 /opt/edge-sites/，檢查憑證與 nginx -t 後 reload；任一步失敗就還原

# 用法：install-sites.sh <渲染好的 .conf 所在目錄>
# 前置：mygo_v2_anon 的 Nginx 已把主機 /opt/edge-sites 掛到容器 /etc/nginx/edge-sites 並 include *.conf。
# 只動 littleduck-*.conf，不碰 mygo 自己的設定。
set -eu

src_dir=${1:?用法：install-sites.sh <conf 目錄>}
sites_dir=${EDGE_SITES_DIR:-/opt/edge-sites}
backup_dir=$(mktemp -d)
trap 'rm -rf "$backup_dir"' EXIT

nginx_container=$(docker ps -q \
    --filter label=com.docker.compose.project=mygo-v2 \
    --filter label=com.docker.compose.service=nginx)
[ -n "$nginx_container" ] || { echo "找不到 mygo-v2 的 nginx 容器" >&2; exit 1; }
[ -d "$sites_dir" ] || { echo "$sites_dir 不存在，mygo 的 Nginx 尚未設定 include（見 runbook 前置）" >&2; exit 1; }

# 憑證沒核發就放設定會讓整個 Nginx 起不來，先擋下
for conf in "$src_dir"/littleduck-*.conf; do
    for cert in $(grep -o '/etc/letsencrypt/live/[^;]*' "$conf" | sort -u); do
        [ -f "$cert" ] || { echo "$(basename "$conf")：憑證 $cert 不存在，請先依 runbook 核發" >&2; exit 1; }
    done
done

cp -p "$sites_dir"/littleduck-*.conf "$backup_dir"/ 2>/dev/null || true
cp "$src_dir"/littleduck-*.conf "$sites_dir"/

if docker exec "$nginx_container" nginx -t; then
    docker exec "$nginx_container" nginx -s reload
    echo "站台設定已更新並 reload"
else
    echo "nginx -t 失敗，還原舊設定" >&2
    rm -f "$sites_dir"/littleduck-*.conf
    cp -p "$backup_dir"/littleduck-*.conf "$sites_dir"/ 2>/dev/null || true
    exit 1
fi
