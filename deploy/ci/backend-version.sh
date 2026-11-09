#!/bin/sh
# deploy/ci/backend-version.sh
# 印出後端版本號：取自 backend/cmd/server/main.go 的 swag 註解 `@version`（與 Swagger 文件顯示的版本一致）

set -eu

main_go=${1:-backend/cmd/server/main.go}
version=$(sed -n 's|^//[[:space:]]*@version[[:space:]]*\([0-9][0-9A-Za-z.+-]*\)[[:space:]]*$|\1|p' "$main_go" | head -n 1)
[ -n "$version" ] || { echo "在 $main_go 找不到 @version 註解" >&2; exit 1; }
echo "$version"
