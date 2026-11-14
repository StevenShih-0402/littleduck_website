#!/bin/sh
# deploy/ci/frontend-bundle.sh
# 組出要上傳到 Hetzner 的前端部署包 bundle/：compose 檔、渲染好的 Nginx 站台、主機端腳本與映像 tag

# 用法：frontend-bundle.sh <staging|production>
# 前置：工作目錄有 .image-ref（image-ref.sh frontend 的輸出）；staging 需 FRONTEND_STAGING_HOST、production 需 FRONTEND_HOST。
set -eu

env_name=${1:?用法：frontend-bundle.sh <staging|production>}
case "$env_name" in
    staging) compose=docker-compose.staging.yml; site=web-staging ;;
    production) compose=docker-compose.prod.yml; site=web ;;
    *) echo "未知環境：$env_name" >&2; exit 1 ;;
esac
[ -s .image-ref ] || { echo "缺少 .image-ref" >&2; exit 1; }

rm -rf bundle
mkdir -p bundle
cp "deploy/web/$compose" bundle/
cp deploy/edge/install-sites.sh deploy/ci/remote-deploy.sh bundle/
cp .image-ref bundle/image.txt
sh deploy/edge/render.sh bundle "$site"
ls -l bundle
