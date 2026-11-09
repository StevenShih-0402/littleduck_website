#!/bin/sh
# deploy/ci/remote-deploy.sh
# 在 Hetzner 主機上執行：以 image.txt 的映像啟動 api、120 秒內等到 healthy，成功後才更新 Nginx 站台

# 用法：remote-deploy.sh <部署目錄> <compose 檔名> <api container_name>
# 部署目錄內需有 .env（主機上手動放置）與 CD 上傳的 compose 檔、image.txt、littleduck-*.conf、install-sites.sh。
set -eu

deploy_dir=${1:?用法：remote-deploy.sh <部署目錄> <compose 檔> <容器名>}
compose_file=${2:?缺 compose 檔名}
container=${3:?缺容器名}
timeout_sec=${HEALTH_TIMEOUT:-120}

cd "$deploy_dir"
[ -f .env ] || { echo "$deploy_dir/.env 不存在，請依 runbook 放置" >&2; exit 1; }
LITTLEDUCK_API_IMAGE=$(cat image.txt)
export LITTLEDUCK_API_IMAGE
echo "部署映像：$LITTLEDUCK_API_IMAGE"

docker compose -f "$compose_file" pull api
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
docker compose -f "$compose_file" logs --tail 50 api
exit 1
