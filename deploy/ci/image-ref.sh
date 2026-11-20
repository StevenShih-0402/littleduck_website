#!/bin/sh
# deploy/ci/image-ref.sh
# 印出本次 commit 對應的映像不可變 tag：<GHCR repo>:<version>-<git 短 sha>，Staging 與 Production 部署都以它為準

# 用法：image-ref.sh [backend|frontend]（預設 backend）
# commit sha 依序讀 Woodpecker 的 CI_COMMIT_SHA、Drone 的 DRONE_COMMIT_SHA（切換觀察期的 fallback，移除 Drone 時一併拿掉），
# 都沒有時用本機 git 的 HEAD。
set -eu

component=${1:-backend}
case "$component" in
    backend)
        image_repo=${LITTLEDUCK_IMAGE_REPO:-ghcr.io/stevenshih-0402/littleduck-api}
        version=$(sh "$(dirname "$0")/backend-version.sh")
        ;;
    frontend)
        image_repo=${LITTLEDUCK_WEB_IMAGE_REPO:-ghcr.io/stevenshih-0402/littleduck-web}
        version=$(sh "$(dirname "$0")/frontend-version.sh")
        ;;
    *) echo "未知元件：$component（可用 backend、frontend）" >&2; exit 1 ;;
esac
sha=${CI_COMMIT_SHA:-${DRONE_COMMIT_SHA:-$(git rev-parse HEAD)}}
echo "$image_repo:$version-$(printf '%s' "$sha" | cut -c1-7)"
