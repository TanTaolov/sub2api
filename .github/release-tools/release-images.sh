#!/usr/bin/env bash
# Simplified release images: one Linux amd64 image, pushed to GHCR only.
set -euo pipefail
: "${RELEASE_VERSION:?}" "${RELEASE_SHA:?}" "${GITHUB_REPOSITORY:?}"
owner=${GITHUB_REPOSITORY%%/*}
registry="ghcr.io/${owner,,}/sub2api"
args=(--platform linux/amd64 --file .release-context/amd64/Dockerfile
  --label "org.opencontainers.image.version=$RELEASE_VERSION"
  --label "org.opencontainers.image.revision=$RELEASE_SHA"
  --label "org.opencontainers.image.source=https://github.com/$GITHUB_REPOSITORY"
  --tag "$registry:$RELEASE_VERSION-amd64"
  --tag "$registry:$RELEASE_VERSION")
# Prereleases must not move the moving tags.
if [[ $RELEASE_VERSION != *-* ]]; then args+=(--tag "$registry:latest"); fi
args+=(--push)
docker buildx build "${args[@]}" .release-context/amd64
