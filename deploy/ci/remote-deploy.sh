#!/bin/sh
# deploy/ci/remote-deploy.sh
# 在 Hetzner 主機上執行：以 image.txt 的映像啟動指定服務（預設 api）、120 秒內等到 healthy，成功後才更新 Nginx 站台

# 用法：remote-deploy.sh <部署目錄> <compose 檔名> <container_name> [service]
# service 預設 api；映像環境變數名為 LITTLEDUCK_<SERVICE 大寫>_IMAGE（api → LITTLEDUCK_API_IMAGE、web → LITTLEDUCK_WEB_IMAGE）。
# 部署目錄內需有 .env（主機上手動放置）與 CD 上傳的 compose 檔、image.txt、littleduck-*.conf、install-sites.sh。
set -eu

deploy_dir=${1:?用法：remote-deploy.sh <部署目錄> <compose 檔> <容器名> [service]}
compose_file=${2:?缺 compose 檔名}
container=${3:?缺容器名}
service=${4:-api}
image_var="LITTLEDUCK_$(printf '%s' "$service" | tr 'a-z-' 'A-Z_')_IMAGE"
timeout_sec=${HEALTH_TIMEOUT:-120}

cd "$deploy_dir"
[ -f .env ] || { echo "$deploy_dir/.env 不存在，請依 runbook 放置" >&2; exit 1; }
image=$(cat image.txt)
export "$image_var=$image"
echo "部署映像（$image_var）：$image"

docker compose -f "$compose_file" pull "$service"
docker compose -f "$compose_file" up -d

elapsed=0
while [ "$elapsed" -lt "$timeout_sec" ]; do
    status=$(docker inspect -f '{{.State.Health.Status}}' "$container" 2>/dev/null || echo missing)
    if [ "$status" = healthy ]; then
        echo "$container 已 healthy（${elapsed} 秒）"
        sh ./install-sites.sh .
        exit 0
    fi
    sleep 5
    elapsed=$((elapsed + 5))
done

echo "$container 在 ${timeout_sec} 秒內未 healthy（最後狀態：$status）" >&2
docker compose -f "$compose_file" ps
docker compose -f "$compose_file" logs --tail 50 "$service"
exit 1
