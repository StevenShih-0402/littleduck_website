#!/bin/sh
# deploy/ci/smoke-frontend.sh
# 前端冒煙測試：/、/blog 必須回 2xx，且 /runtime-config.js 內容含預期的 API 位址；任一不符即失敗
# 預期 API 位址等於前端 base（Staging 同源）時，另檢查 /api/v1/articles 經 Nginx 轉送回 2xx
# 設了 SMOKE_BASIC_AUTH（帳號:密碼，Staging 用）時每個請求帶帳密，另不帶帳密請求 /，非 401 即失敗（驗證未生效）

# 用法：[SMOKE_BASIC_AUTH=帳號:密碼] smoke-frontend.sh <前端 base URL> <預期 API base URL>
# 例：smoke-frontend.sh https://staging.stevenshih-littleduck.com https://staging.stevenshih-littleduck.com
set -eu

base=${1:?用法：smoke-frontend.sh <前端 base URL> <預期 API base URL>}
expected_api=${2:?缺預期 API base URL}
base=${base%/}
expected_api=${expected_api%/}
auth=${SMOKE_BASIC_AUTH:-}
failed=0

# 用法：fetch <URL> [帶帳密：1] [其他 curl 參數…]；連線失敗時 curl 的 -w 仍會印出 000，結束碼非 0 只需吞掉
fetch() {
    url=$1
    with_auth=${2:-0}
    [ "$#" -ge 2 ] && shift 2 || shift 1
    if [ "$with_auth" = 1 ] && [ -n "$auth" ]; then
        curl -sS --max-time 10 --retry 3 --retry-delay 2 -u "$auth" "$@" "$url" || true
    else
        curl -sS --max-time 10 --retry 3 --retry-delay 2 "$@" "$url" || true
    fi
}

paths="/ /blog"
[ "$expected_api" = "$base" ] && paths="$paths /api/v1/articles"
for path in $paths; do
    code=$(fetch "$base$path" 1 -o /dev/null -w '%{http_code}')
    case "$code" in
        2??) echo "通過 $path → $code" ;;
        *) echo "失敗 $path → $code" >&2; failed=1 ;;
    esac
done

config=$(fetch "$base/runtime-config.js" 1)
case "$config" in
    *"\"apiBaseUrl\":\"$expected_api\""*) echo "通過 /runtime-config.js → 含 $expected_api" ;;
    *) echo "失敗 /runtime-config.js → 未含預期的 API 位址 $expected_api（實際：$config）" >&2; failed=1 ;;
esac

if [ -n "$auth" ]; then
    code=$(fetch "$base/" 0 -o /dev/null -w '%{http_code}')
    case "$code" in
        401) echo "通過 不帶帳密 / → 401" ;;
        *) echo "失敗 不帶帳密 / → $code，預期 401（驗證未生效）" >&2; failed=1 ;;
    esac
fi
exit "$failed"
