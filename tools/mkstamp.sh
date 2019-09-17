#!/bin/bash
#
# mean to be passed as workspace_status_command to bazel so that we can use them
# for docker image stamping: see https://github.com/bazelbuild/rules_docker#stamping
#
set -e

# According to jenkins documentation BUILD_TAG is:
#
# String of jenkins-${JOB_NAME}-${BUILD_NUMBER}.
#
# We default something reasonable in case we run docker_push rules outside jenkins.
: "${BUILD_TAG:=local-${USER}}"

echo IMAGE_TAG "${BUILD_TAG}-$(git rev-parse HEAD)"
