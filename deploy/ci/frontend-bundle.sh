#!/bin/sh
# deploy/ci/frontend-bundle.sh
# 組出要上傳到 Hetzner 的前端部署包 bundle/：compose 檔、渲染好的 Nginx 站台、主機端腳本與映像 tag

# 用法：frontend-bundle.sh <staging|production>
# 前置：工作目錄有 .image-ref（image-ref.sh frontend 的輸出）；staging 需 FRONTEND_STAGING_HOST 與 STAGING_GATE_TOKEN、production 需 FRONTEND_HOST。
set -eu

env_name=${1:?用法：frontend-bundle.sh <staging|production>}
case "$env_name" in
    staging) compose=docker-compose.staging.yml; site=web-staging ;;
    production) compose=docker-compose.prod.yml; site=web ;;
    *) echo "未知環境：$env_name" >&2; exit 1 ;;
esac
[ -s .image-ref ] || { echo "缺少 .image-ref" >&2; exit 1; }

# 通行值會原樣寫進 Nginx 設定（map 正規式與 Set-Cookie），限定十六進位以免破壞語法
# 以整串比對（grep 是逐行比對，多行值會被放過）
if [ "$env_name" = staging ]; then
    token=${STAGING_GATE_TOKEN:-}
    case "$token" in
        *[!0-9a-f]*) token_ok=0 ;;
        *) [ "${#token}" -ge 32 ] && token_ok=1 || token_ok=0 ;;
    esac
    if [ "$token_ok" -ne 1 ]; then
        echo "STAGING_GATE_TOKEN 未設定或格式錯誤（需 32 字元以上十六進位）" >&2
        exit 1
    fi
fi

rm -rf bundle
mkdir -p bundle
cp "deploy/web/$compose" bundle/
cp deploy/edge/install-sites.sh deploy/ci/remote-deploy.sh bundle/
cp .image-ref bundle/image.txt
sh deploy/edge/render.sh bundle "$site"
ls -l bundle
