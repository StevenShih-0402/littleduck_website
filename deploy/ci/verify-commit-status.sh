#!/bin/sh
# deploy/ci/verify-commit-status.sh
# Production 部署前置檢查：該 commit 的 Woodpecker push 建置（CD Staging）必須全部成功，否則不准部署 Production。
# Woodpecker 的 Deploy API 不檢查原 pipeline 是否成功，所以由本腳本補上這一關。

# 用法：verify-commit-status.sh <owner/repo> <commit sha>
# 讀 GitHub 的 combined status（public repo 不需 token），只看 context 以 ci/woodpecker/push/ 開頭的項目，
# 每個 context 取最新一筆（GitHub 依建立時間由新到舊排序）。
# 環境變數：
#   REQUIRED_CONTEXTS  必須存在的 context，空白分隔；預設為兩條 CD Staging workflow
#   GITHUB_API_URL     GitHub API 位址，預設 https://api.github.com（測試時指向本機 fixture）
# 新增 push 觸發的 workflow 時，記得同步 REQUIRED_CONTEXTS 的預設值（見 docs/cicd.md）。
set -eu

repo=${1:?用法：verify-commit-status.sh <owner/repo> <sha>}
sha=${2:?缺 commit sha}
api=${GITHUB_API_URL:-https://api.github.com}
required=${REQUIRED_CONTEXTS:-ci/woodpecker/push/cd-backend-staging ci/woodpecker/push/cd-frontend-staging}
prefix=ci/woodpecker/push/

# 每行一筆「context<TAB>state」，同一個 context 只留最新一筆
latest=$(curl -fsS "$api/repos/$repo/commits/$sha/status?per_page=100" \
    | jq -r --arg p "$prefix" '
        [.statuses[] | select(.context | startswith($p))]
        | reduce .[] as $s ({}; if has($s.context) then . else . + {($s.context): $s.state} end)
        | to_entries[] | "\(.key)\t\(.value)"')

problems=""
for ctx in $required; do
    if ! printf '%s\n' "$latest" | cut -f1 | grep -qxF "$ctx"; then
        problems="$problems
  $ctx：missing"
    fi
done

bad=$(printf '%s\n' "$latest" | awk -F '\t' 'NF == 2 && $2 != "success" { printf "  %s：%s\n", $1, $2 }')
if [ -n "$bad" ]; then
    problems="$problems
$bad"
fi

if [ -z "$problems" ]; then
    echo "$sha 的 Woodpecker push 建置全部為 success，可以部署 Production"
else
    echo "$sha 的 Woodpecker push 建置未全部通過，只能 Deploy push 建置全部成功的 commit：$problems" >&2
    exit 1
fi
