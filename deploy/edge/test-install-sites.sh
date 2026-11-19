#!/bin/sh
# deploy/edge/test-install-sites.sh
# 本機驗證 install-sites.sh 的憑證與帳密檔檢查：以帶 mygo-v2 label 的假 Nginx 容器模擬主機，不 mock 被測腳本。
# 用法：sh deploy/edge/test-install-sites.sh（需要本機 Docker）
set -u

here=$(cd "$(dirname "$0")" && pwd)
script="$here/install-sites.sh"
image=nginx:1.27-alpine
label_filter="label=com.docker.compose.project=mygo-v2"

if [ -n "$(docker ps -q --filter "$label_filter")" ]; then
    echo "本機已有執行中的 mygo-v2 容器，為避免誤操作而中止" >&2
    exit 1
fi

work=$(mktemp -d)
container=""
cleanup() {
    [ -z "$container" ] || docker rm -f "$container" >/dev/null 2>&1
    rm -rf "$work"
}
trap cleanup EXIT

container=$(docker run -d \
    --label com.docker.compose.project=mygo-v2 \
    --label com.docker.compose.service=nginx \
    "$image") || { echo "無法啟動測試容器" >&2; exit 1; }

# 憑證只建立在容器內，主機上沒有同路徑檔案
docker exec "$container" sh -c 'mkdir -p /etc/letsencrypt/live/test.example && touch /etc/letsencrypt/live/test.example/fullchain.pem /etc/letsencrypt/live/test.example/privkey.pem'
# 帳密檔同樣只建立在容器內
docker exec "$container" sh -c 'mkdir -p /etc/nginx/edge-sites && touch /etc/nginx/edge-sites/littleduck-staging.htpasswd'

mkdir -p "$work/src-ok" "$work/src-missing" "$work/src-auth-ok" "$work/src-auth-missing" "$work/bin"
cat > "$work/src-ok/littleduck-test.conf" <<'CONF'
# test
# ssl_certificate /etc/letsencrypt/live/test.example/fullchain.pem;
# ssl_certificate_key /etc/letsencrypt/live/test.example/privkey.pem;
CONF
cat > "$work/src-missing/littleduck-test.conf" <<'CONF'
# ssl_certificate /etc/letsencrypt/live/missing.example/fullchain.pem;
CONF
cat > "$work/src-auth-ok/littleduck-test.conf" <<'CONF'
# ssl_certificate /etc/letsencrypt/live/test.example/fullchain.pem;
# auth_basic_user_file /etc/nginx/edge-sites/littleduck-staging.htpasswd;
CONF
cat > "$work/src-auth-missing/littleduck-test.conf" <<'CONF'
# ssl_certificate /etc/letsencrypt/live/test.example/fullchain.pem;
# auth_basic_user_file /etc/nginx/edge-sites/missing.htpasswd;
CONF

real_docker=$(command -v docker)
cat > "$work/bin/docker" <<WRAP
#!/bin/sh
# ps 回傳容器 ID 後立刻把容器停掉，之後的 docker exec 會得到真實的失敗（容器未執行，回 1）
if [ "\$1" = ps ]; then
    id=\$("$real_docker" "\$@") || exit \$?
    "$real_docker" stop "\$id" >/dev/null
    echo "\$id"
    exit 0
fi
exec "$real_docker" "\$@"
WRAP
chmod +x "$work/bin/docker"

fail=0
pass() { echo "PASS: $1"; }
failc() { echo "FAIL: $1：$2"; fail=1; }

new_sites() { d=$(mktemp -d -p "$work"); echo "$d"; }

# 案例 1：憑證存在
sites=$(new_sites)
out=$(EDGE_SITES_DIR="$sites" sh "$script" "$work/src-ok" 2>&1); rc=$?
echo "$out"
if [ $rc -eq 0 ] && echo "$out" | grep -q '站台設定已更新並 reload' && [ -f "$sites/littleduck-test.conf" ]; then
    pass "憑證存在"
else
    failc "憑證存在" "rc=$rc"
fi

# 案例 2：憑證不存在
sites=$(new_sites)
out=$(EDGE_SITES_DIR="$sites" sh "$script" "$work/src-missing" 2>&1); rc=$?
echo "$out"
if [ $rc -ne 0 ] && echo "$out" | grep -q '請先依 runbook 核發' && [ -z "$(ls -A "$sites")" ]; then
    pass "憑證不存在"
else
    failc "憑證不存在" "rc=$rc"
fi

# 案例 3：帳密檔存在
sites=$(new_sites)
out=$(EDGE_SITES_DIR="$sites" sh "$script" "$work/src-auth-ok" 2>&1); rc=$?
echo "$out"
if [ $rc -eq 0 ] && echo "$out" | grep -q '站台設定已更新並 reload' && [ -f "$sites/littleduck-test.conf" ]; then
    pass "帳密檔存在"
else
    failc "帳密檔存在" "rc=$rc"
fi

# 案例 4：帳密檔不存在
sites=$(new_sites)
out=$(EDGE_SITES_DIR="$sites" sh "$script" "$work/src-auth-missing" 2>&1); rc=$?
echo "$out"
if [ $rc -ne 0 ] && echo "$out" | grep -q '請先依 runbook 建立' && [ -z "$(ls -A "$sites")" ]; then
    pass "帳密檔不存在"
else
    failc "帳密檔不存在" "rc=$rc"
fi

# 案例 5：無法檢查（容器已停止；須放最後，會停掉假容器）
sites=$(new_sites)
out=$(PATH="$work/bin:$PATH" EDGE_SITES_DIR="$sites" sh "$script" "$work/src-ok" 2>&1); rc=$?
echo "$out"
if [ $rc -ne 0 ] && echo "$out" | grep -q '無法在 Nginx 容器內檢查' && ! echo "$out" | grep -q '請先依 runbook 核發' && [ -z "$(ls -A "$sites")" ]; then
    pass "無法檢查"
else
    failc "無法檢查" "rc=$rc"
fi

if [ $fail -eq 0 ]; then
    echo "ALL PASS"
else
    exit 1
fi
