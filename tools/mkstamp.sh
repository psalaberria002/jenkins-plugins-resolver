#!/bin/bash
#
# mean to be passed as workspace_status_command to bazel so that we can use them
# for docker image stamping: see https://github.com/bazelbuild/rules_docker#stamping
#
set -e

# According to GitHub documentation GITHUB_ACTOR is:
#
# https://help.github.com/en/articles/virtual-environments-for-github-actions#default-environment-variables
# The name of the person or app that initiated the workflow. For example, octocat.
#
# We default something reasonable in case we run docker_push rules outside GitHub.
: "${GITHUB_ACTOR:=local-${USER}}"

echo IMAGE_TAG "${GITHUB_ACTOR}-$(git rev-parse HEAD)"
