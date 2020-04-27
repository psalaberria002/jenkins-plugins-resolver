#!/bin/bash
#
# mean to be passed as workspace_status_command to bazel so that we can use them
# for docker image stamping: see https://github.com/bazelbuild/rules_docker#stamping
#
set -e

: "${BUILD_TAG:=local}"
: "${GIT_COMMIT:=$(git rev-parse HEAD)}"

# According to GitHub documentation CI is always set inside GH actions.
# https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables
if [[ -n "$CI" ]]; then
    BUILD_TAG="master"
fi

echo IMAGE_TAG "${BUILD_TAG}-${GIT_COMMIT}"
echo STABLE_GIT_COMMIT "${GIT_COMMIT}"
