#!/bin/sh
# deploy/edge/test-staging-auth.sh
# 本機驗證兩個 Staging 站台的身分驗證行為：以 render.sh 渲染真實樣板，載入 nginx:1.27-alpine，
# 搭配兩個假 upstream 容器（名稱同正式環境的 container_name），以 curl --resolve 逐項檢查。
# 用法：sh deploy/edge/test-staging-auth.sh（需要本機 Docker 與 curl；Nginx 容器內會以 apk 安裝 openssl 產生自簽憑證，需要網路）
set -u

here=$(cd "$(dirname "$0")" && pwd)
image=nginx:1.27-alpine
web_host=staging.example
api_host=api.staging.example
user=tester
password=testpass
token=$(printf '%064d' 0 | tr 0 a)
network=ld-staging-auth-test
edge_name=ld-staging-auth-edge
web_name=littleduck-web-staging
api_name=littleduck-api-staging

# 假 upstream 用正式環境的容器名稱，本機已有同名容器時中止，避免誤刪
for name in "$edge_name" "$web_name" "$api_name"; do
    if [ -n "$(docker ps -aq --filter "name=^${name}\$")" ]; then
        echo "本機已有名為 $name 的容器，為避免誤操作而中止" >&2
        exit 1
    fi
done
if docker network inspect "$network" >/dev/null 2>&1; then
    echo "本機已有名為 $network 的 docker network，為避免誤操作而中止" >&2
    exit 1
fi

work=$(mktemp -d)
created=""
cleanup() {
    for name in $created; do
        docker rm -f "$name" >/dev/null 2>&1
    done
    docker network rm "$network" >/dev/null 2>&1
    rm -rf "$work"
}
trap cleanup EXIT

docker network create "$network" >/dev/null || { echo "無法建立 docker network" >&2; exit 1; }

# 假 upstream：回應固定字串，用來辨識請求送達哪一個容器
start_upstream() {
    name=$1; port=$2; label=$3
    docker run -d --name "$name" --network "$network" "$image" sh -c \
        "printf 'server { listen $port; location / { return 200 \"$label \$request_method \$uri\\\\n\"; } }\\n' > /etc/nginx/conf.d/default.conf && exec nginx -g 'daemon off;'" \
        >/dev/null || { echo "無法啟動 $name" >&2; exit 1; }
    created="$created $name"
}
start_upstream "$web_name" 3000 web-upstream
start_upstream "$api_name" 8080 api-upstream

# 渲染真實樣板
mkdir -p "$work/sites"
FRONTEND_STAGING_HOST=$web_host API_STAGING_HOST=$api_host STAGING_GATE_TOKEN=$token \
    sh "$here/render.sh" "$work/sites" web-staging api-staging >/dev/null || { echo "渲染樣板失敗" >&2; exit 1; }

# Edge Nginx：站台目錄掛在與主機相同的 /etc/nginx/edge-sites，容器內產生自簽憑證與 $6$ 帳密檔後啟動
docker run -d --name "$edge_name" --network "$network" -p 127.0.0.1::443 \
    -v "$work/sites:/etc/nginx/edge-sites" \
    -e WEB_HOST="$web_host" -e API_HOST="$api_host" -e AUTH_USER="$user" -e AUTH_PASSWORD="$password" \
    "$image" sh -c '
        set -e
        apk add -q --no-cache openssl >/dev/null
        for h in "$WEB_HOST" "$API_HOST"; do
            mkdir -p "/etc/letsencrypt/live/$h"
            openssl req -x509 -newkey rsa:2048 -nodes -days 1 -subj "/CN=$h" \
                -keyout "/etc/letsencrypt/live/$h/privkey.pem" -out "/etc/letsencrypt/live/$h/fullchain.pem" 2>/dev/null
        done
        printf "%s:%s\n" "$AUTH_USER" "$(mkpasswd -m sha512 "$AUTH_PASSWORD")" > /etc/nginx/edge-sites/littleduck-staging.htpasswd
        echo "include /etc/nginx/edge-sites/*.conf;" > /etc/nginx/conf.d/default.conf
        nginx -t
        touch /tmp/ready
        exec nginx -g "daemon off;"
    ' >/dev/null || { echo "無法啟動 Edge Nginx" >&2; exit 1; }
created="$created $edge_name"

ready=0
i=0
while [ $i -lt 60 ]; do
    if docker exec "$edge_name" test -f /tmp/ready 2>/dev/null; then
        ready=1
        break
    fi
    if [ -z "$(docker ps -q --filter "name=^${edge_name}\$")" ]; then
        break
    fi
    sleep 1
    i=$((i + 1))
done
if [ $ready -ne 1 ]; then
    echo "Edge Nginx 未就緒（nginx -t 失敗或逾時）：" >&2
    docker logs "$edge_name" >&2
    exit 1
fi
sleep 1
port=$(docker port "$edge_name" 443/tcp | head -n 1 | sed 's/.*://')
echo "nginx -t 通過，Edge Nginx 在 127.0.0.1:$port"

fail=0
pass() { echo "PASS: $1"; }
failc() { echo "FAIL: $1：$2"; fail=1; }

# 用法：req <host> <其他 curl 參數…>；印出「狀態碼」，回應標頭存 $work/headers、本文存 $work/body
req() {
    host=$1; shift
    curl -sk -o "$work/body" -D "$work/headers" -w '%{http_code}' --max-time 10 \
        --resolve "$host:$port:127.0.0.1" "$@"
}
web="https://$web_host:$port"
api="https://$api_host:$port"

# --- 前端 Staging ---
code=$(req "$web_host" "$web/")
[ "$code" = 401 ] && pass "前端：不帶帳密 / → 401" || failc "前端：不帶帳密 / → 401" "實際 $code"

code=$(req "$web_host" -u "$user:$password" "$web/")
if [ "$code" = 200 ] && grep -qi "^set-cookie: ld_staging_gate=$token" "$work/headers" && grep -q '^web-upstream' "$work/body"; then
    pass "前端：帶帳密 / → 200 並發通行 cookie"
else
    failc "前端：帶帳密 / → 200 並發通行 cookie" "實際 $code"
fi

code=$(req "$web_host" -H "Cookie: ld_staging_gate=$token" -H 'Authorization: Bearer x' "$web/api/v1/articles")
if [ "$code" = 200 ] && grep -q '^api-upstream GET /api/v1/articles' "$work/body"; then
    pass "前端：正確 cookie＋Bearer 經 /api/ 送達後端"
else
    failc "前端：正確 cookie＋Bearer 經 /api/ 送達後端" "實際 $code，本文 $(cat "$work/body")"
fi

code=$(req "$web_host" -H 'Cookie: ld_staging_gate=wrong' "$web/")
[ "$code" = 401 ] && pass "前端：錯誤 cookie → 401" || failc "前端：錯誤 cookie → 401" "實際 $code"

code=$(req "$web_host" -H 'Authorization: Bearer x' "$web/api/v1/articles")
[ "$code" = 401 ] && pass "前端：只帶 Bearer → 401" || failc "前端：只帶 Bearer → 401" "實際 $code"

# --- 後端 Staging ---
code=$(req "$api_host" "$api/api/v1/articles")
[ "$code" = 401 ] && pass "後端：不帶帳密 /api/v1/articles → 401" || failc "後端：不帶帳密 /api/v1/articles → 401" "實際 $code"

code=$(req "$api_host" "$api/swagger/index.html")
[ "$code" = 401 ] && pass "後端：不帶帳密 /swagger/index.html → 401" || failc "後端：不帶帳密 /swagger/index.html → 401" "實際 $code"

code=$(req "$api_host" -u "$user:$password" "$api/health")
if [ "$code" = 200 ] && grep -q '^api-upstream GET /health' "$work/body"; then
    pass "後端：帶帳密 /health → 200"
else
    failc "後端：帶帳密 /health → 200" "實際 $code"
fi

code=$(req "$api_host" -X POST -d 'x=1' "$api/api/v1/payments/ecpay/callback")
if [ "$code" != 401 ] && grep -q '^api-upstream POST /api/v1/payments/ecpay/callback' "$work/body"; then
    pass "後端：不帶帳密 POST 綠界通知送達後端"
else
    failc "後端：不帶帳密 POST 綠界通知送達後端" "實際 $code"
fi

if [ $fail -eq 0 ]; then
    echo "ALL PASS"
else
    exit 1
fi
