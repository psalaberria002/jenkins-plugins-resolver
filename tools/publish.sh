#!/bin/bash
set -eu

# Login GCR
docker login https://gcr.io -u _json_key -p "${GCR_BITNAMI_LABS}"

# Publish inmutable tags
tools/bazel-docker-push.sh

# Publish latest tags
tools/bazel-docker-push-latest.sh
