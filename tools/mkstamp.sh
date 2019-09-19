#!/bin/bash
#
# mean to be passed as workspace_status_command to bazel so that we can use them
# for docker image stamping: see https://github.com/bazelbuild/rules_docker#stamping
#
set -e

# According to GitHub documentation GITHUB_ are reserved to internal use so we can rely on them
# to set a proper tag when bazel runs inside GH actions.
#
# In this case, you can access to the build logs of an image by using the URL below.
# https://github.com/bitnami-labs/jenkins-plugins-resolver/commit/COMMIT_SHA/checks
#
# Example:
#
# gcr.io/bitnami-labs/jenkins-plugins-resolver:master-dca6b4bfc97325ff175d575d30b4546e1bc99e92
# https://github.com/bitnami-labs/jenkins-plugins-resolver/commit/dca6b4bfc97325ff175d575d30b4546e1bc99e92/checks
#
if [[ -n "$GITHUB_SHA" ]]; then
    echo IMAGE_TAG "master-${GITHUB_SHA}"
else
    # We default something reasonable in case we run docker_push rules outside GitHub.
    echo IMAGE_TAG "local-$(git rev-parse HEAD)"
fi
