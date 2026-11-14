#!/bin/sh
# deploy/ci/smoke-backend.sh
# 後端冒煙測試：對關鍵讀取端點各發一次請求，任一非 2xx 即失敗

# 用法：smoke-backend.sh <base URL>，例：smoke-backend.sh https://api.staging.stevenshih-littleduck.com
set -eu

base=${1:?用法：smoke-backend.sh <base URL>}
failed=0
for path in /health /api/v1/articles; do
    # 連線失敗時 curl 的 -w 仍會印出 000，結束碼非 0 只需吞掉
    code=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 10 --retry 3 --retry-delay 2 "$base$path") || true
    case "$code" in
        2??) echo "通過 $path → $code" ;;
        *) echo "失敗 $path → $code" >&2; failed=1 ;;
    esac
done
exit "$failed"
