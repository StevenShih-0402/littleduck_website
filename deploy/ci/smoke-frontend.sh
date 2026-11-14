#!/bin/sh
# deploy/ci/smoke-frontend.sh
# 前端冒煙測試：/、/blog 必須回 2xx，且 /runtime-config.js 內容含預期的 API 位址；任一不符即失敗

# 用法：smoke-frontend.sh <前端 base URL> <預期 API base URL>
# 例：smoke-frontend.sh https://staging.stevenshih-littleduck.com https://api.staging.stevenshih-littleduck.com
set -eu

base=${1:?用法：smoke-frontend.sh <前端 base URL> <預期 API base URL>}
expected_api=${2:?缺預期 API base URL}
base=${base%/}
expected_api=${expected_api%/}
failed=0

for path in / /blog; do
    # 連線失敗時 curl 的 -w 仍會印出 000，結束碼非 0 只需吞掉
    code=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 10 --retry 3 --retry-delay 2 "$base$path") || true
    case "$code" in
        2??) echo "通過 $path → $code" ;;
        *) echo "失敗 $path → $code" >&2; failed=1 ;;
    esac
done

config=$(curl -sS --max-time 10 --retry 3 --retry-delay 2 "$base/runtime-config.js") || config=""
case "$config" in
    *"\"apiBaseUrl\":\"$expected_api\""*) echo "通過 /runtime-config.js → 含 $expected_api" ;;
    *) echo "失敗 /runtime-config.js → 未含預期的 API 位址 $expected_api（實際：$config）" >&2; failed=1 ;;
esac
exit "$failed"
