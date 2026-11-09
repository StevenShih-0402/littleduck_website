#!/bin/sh
# deploy/ci/verify-commit-status.sh
# promote 前置檢查：該 commit 的 Drone push 建置（CI＋Staging／Preview）必須是成功，否則不准部署 Production

# 用法：verify-commit-status.sh <owner/repo> <commit sha>
# 讀 GitHub 的 combined status（public repo 不需 token），context 為 Drone 回報的 continuous-integration/drone/push。
set -eu

repo=${1:?用法：verify-commit-status.sh <owner/repo> <sha>}
sha=${2:?缺 commit sha}
context=${STATUS_CONTEXT:-continuous-integration/drone/push}

state=$(curl -fsS "https://api.github.com/repos/$repo/commits/$sha/status" \
    | jq -r --arg c "$context" '[.statuses[] | select(.context == $c)][0].state // "missing"')

if [ "$state" = success ]; then
    echo "$sha 的 $context 為 success，可以部署 Production"
else
    echo "$sha 的 $context 為 $state：只能 promote push 建置全部通過的 commit" >&2
    exit 1
fi
