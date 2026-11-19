#!/bin/sh
# deploy/ci/smoke-backend.sh
# 後端冒煙測試：對關鍵讀取端點各發一次請求，任一非 2xx 即失敗
# 設了 SMOKE_BASIC_AUTH（帳號:密碼，Staging 用）時每個請求帶帳密，另不帶帳密請求文章列表，非 401 即失敗（驗證未生效）

# 用法：[SMOKE_BASIC_AUTH=帳號:密碼] smoke-backend.sh <base URL>，例：smoke-backend.sh https://api.staging.stevenshih-littleduck.com
set -eu

base=${1:?用法：smoke-backend.sh <base URL>}
auth=${SMOKE_BASIC_AUTH:-}
failed=0

# 用法：http_code <URL> [帶帳密：1]；連線失敗時 curl 的 -w 仍會印出 000，結束碼非 0 只需吞掉
http_code() {
    if [ "${2:-0}" = 1 ] && [ -n "$auth" ]; then
        curl -sS -o /dev/null -w '%{http_code}' --max-time 10 --retry 3 --retry-delay 2 -u "$auth" "$1" || true
    else
        curl -sS -o /dev/null -w '%{http_code}' --max-time 10 --retry 3 --retry-delay 2 "$1" || true
    fi
}

for path in /health /api/v1/articles; do
    code=$(http_code "$base$path" 1)
    case "$code" in
        2??) echo "通過 $path → $code" ;;
        *) echo "失敗 $path → $code" >&2; failed=1 ;;
    esac
done

if [ -n "$auth" ]; then
    code=$(http_code "$base/api/v1/articles")
    case "$code" in
        401) echo "通過 不帶帳密 /api/v1/articles → 401" ;;
        *) echo "失敗 不帶帳密 /api/v1/articles → $code，預期 401（驗證未生效）" >&2; failed=1 ;;
    esac
fi
exit "$failed"
