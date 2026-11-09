#!/bin/sh
# deploy/ci/image-ref.sh
# 印出本次 commit 對應的後端映像不可變 tag：<GHCR repo>:<version>-<git 短 sha>，Staging 與 Production 部署都以它為準

set -eu

image_repo=${LITTLEDUCK_IMAGE_REPO:-ghcr.io/stevenshih-0402/littleduck-api}
sha=${DRONE_COMMIT_SHA:-$(git rev-parse HEAD)}
version=$(sh "$(dirname "$0")/backend-version.sh")
echo "$image_repo:$version-$(printf '%s' "$sha" | cut -c1-7)"
