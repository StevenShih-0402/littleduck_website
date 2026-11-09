#!/bin/sh
# deploy/ci/vercel-deploy.sh
# 以 Vercel CLI 建置並部署前端，完成後對部署網址冒煙；preview 視為 Staging，production 只在 promote 時執行

# 用法：vercel-deploy.sh <preview|production>
# 需要 VERCEL_TOKEN、VERCEL_ORG_ID、VERCEL_PROJECT_ID；production 另需 FRONTEND_HOST。
# Preview 預設受 Vercel Authentication 保護，冒煙時以 VERCEL_BYPASS_SECRET（Protection Bypass for Automation）通過。
# 在 repo 根目錄執行：Vercel 專案的 Root Directory 設為 frontend，CLI 會自己切進去。
set -eu

target=${1:?用法：vercel-deploy.sh <preview|production>}
vercel_version=${VERCEL_CLI_VERSION:-65.12.1}
: "${VERCEL_TOKEN:?未設定 VERCEL_TOKEN}"

npm install -g --silent "vercel@$vercel_version"

case "$target" in
    preview)
        vercel pull --yes --environment=preview --token="$VERCEL_TOKEN"
        vercel build --token="$VERCEL_TOKEN"
        url=$(vercel deploy --prebuilt --token="$VERCEL_TOKEN")
        echo "Preview 網址：$url"
        ;;
    production)
        : "${FRONTEND_HOST:?未設定 FRONTEND_HOST}"
        vercel pull --yes --environment=production --token="$VERCEL_TOKEN"
        vercel build --prod --token="$VERCEL_TOKEN"
        vercel deploy --prebuilt --prod --token="$VERCEL_TOKEN"
        url="https://$FRONTEND_HOST"
        ;;
    *) echo "未知目標：$target" >&2; exit 1 ;;
esac

# 冒煙：首頁回 2xx 才算成功
header=""
[ -n "${VERCEL_BYPASS_SECRET:-}" ] && [ "$target" = preview ] && header="x-vercel-protection-bypass: $VERCEL_BYPASS_SECRET"
for attempt in 1 2 3; do
    if [ -n "$header" ]; then
        wget -q -O /dev/null --header "$header" "$url/" && { echo "冒煙通過：$url/"; exit 0; }
    else
        wget -q -O /dev/null "$url/" && { echo "冒煙通過：$url/"; exit 0; }
    fi
    sleep 5
done
echo "冒煙失敗：$url/ 未回 2xx" >&2
exit 1
