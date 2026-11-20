#!/bin/sh
# deploy/ci/test-verify-commit-status.sh
# 本機驗證 verify-commit-status.sh 的判斷：以 python3 -m http.server 提供假的 GitHub combined status，不 mock 被測腳本。
# 用法：sh deploy/ci/test-verify-commit-status.sh（需要 python3、curl、jq）
set -u

here=$(cd "$(dirname "$0")" && pwd)
script="$here/verify-commit-status.sh"
repo=owner/repo

work=$(mktemp -d)
server_pid=""
cleanup() {
    [ -z "$server_pid" ] || kill "$server_pid" 2>/dev/null
    rm -rf "$work"
}
trap cleanup EXIT

# fixture：每個案例一個 commit sha，路徑同 GitHub API（http.server 會忽略 ?per_page 查詢字串）
fixture() {
    mkdir -p "$work/repos/$repo/commits/$1"
    cat > "$work/repos/$repo/commits/$1/status"
}

fixture all-green <<'JSON'
{"state":"success","statuses":[
  {"context":"ci/woodpecker/push/cd-frontend-staging","state":"success"},
  {"context":"ci/woodpecker/push/cd-backend-staging","state":"success"},
  {"context":"ci/woodpecker/pr/ci-backend","state":"failure"}
]}
JSON

fixture missing-one <<'JSON'
{"state":"success","statuses":[
  {"context":"ci/woodpecker/push/cd-backend-staging","state":"success"}
]}
JSON

fixture one-failure <<'JSON'
{"state":"failure","statuses":[
  {"context":"ci/woodpecker/push/cd-backend-staging","state":"success"},
  {"context":"ci/woodpecker/push/cd-frontend-staging","state":"failure"}
]}
JSON

fixture drone-only <<'JSON'
{"state":"success","statuses":[
  {"context":"continuous-integration/drone/push","state":"success"}
]}
JSON

# 找一個空閒埠啟動 fixture 伺服器
port=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')
python3 -m http.server "$port" --bind 127.0.0.1 --directory "$work" >/dev/null 2>&1 &
server_pid=$!
for _ in $(seq 1 50); do
    curl -fsS -o /dev/null "http://127.0.0.1:$port/repos/$repo/commits/all-green/status" 2>/dev/null && break
    sleep 0.1
done

fail=0
check() {
    name=$1 sha=$2 want=$3
    out=$(GITHUB_API_URL="http://127.0.0.1:$port" sh "$script" "$repo" "$sha" 2>&1); rc=$?
    echo "$out"
    if [ "$rc" -eq "$want" ]; then
        echo "PASS: $name"
    else
        echo "FAIL: $name：預期 rc=$want，實際 rc=$rc"
        fail=1
    fi
}

check "全綠" all-green 0
check "缺一個必要 context" missing-one 1
check "有一個 failure" one-failure 1
check "只有 drone context" drone-only 1

if [ $fail -eq 0 ]; then
    echo "ALL PASS"
else
    exit 1
fi
