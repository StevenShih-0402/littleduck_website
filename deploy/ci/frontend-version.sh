#!/bin/sh
# deploy/ci/frontend-version.sh
# 印出前端版本號：取自 frontend/package.json 頂層的 "version"（以 sed 讀取，不依賴 jq／node）

set -eu

package_json=${1:-frontend/package.json}
version=$(sed -n 's|^[[:space:]]*"version"[[:space:]]*:[[:space:]]*"\([0-9][0-9A-Za-z.+-]*\)".*|\1|p' "$package_json" | head -n 1)
[ -n "$version" ] || { echo "在 $package_json 找不到 version" >&2; exit 1; }
echo "$version"
