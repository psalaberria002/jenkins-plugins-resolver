#!/bin/bash
#
# mean to be passed as workspace_status_command to bazel so that we can use them
# for docker image stamping: see https://github.com/bazelbuild/rules_docker#stamping
#
set -e

# According to GitHub documentation GITHUB_ are reserved to internal use so we can rely on them
# to set a proper tag when bazel runs inside GH actions.
#
# We default something reasonable in case we run docker_push rules outside GitHub.
if [[ -n "$GITHUB_SHA" ]]; then
    echo IMAGE_TAG "gh-actions-${GITHUB_SHA}"
else
    echo IMAGE_TAG "local-$(git rev-parse HEAD)"
fi
