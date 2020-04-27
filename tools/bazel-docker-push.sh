#!/bin/bash

# Run non "latest" container_push targets in paralle.
#
# We cannot just call `bazel run` because the bazel CLI holds
# lock. The bazel-run.sh helper uses the `--script_path` to let
# bazel just build a wrapper shell script that it later runs
# without holding the bazel CLI process alive for the whole duration.
bazel query 'kind("container_push", //...) except attr(tag, "latest", //...)' | xargs -P20 -I% "$(dirname $0)/bazel-run.sh" %
