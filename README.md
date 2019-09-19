![](https://github.com/bitnami-labs/jenkins-plugins-resolver/workflows/Continuous%20Integration/badge.svg)

# jenkins-plugins-resolver

Go tools to manage Jenkins plugins resolution, such as transitive dependencies graph computation and download

# Development

This project uses the standard [go mod](https://blog.golang.org/using-go-modules) tool to manage and vendor Go dependencies.

Aditionally, it uses [bazel](https://bazel.build) to manage the project build.

## Install bazel

Feel free to follow [official installation instructions](https://docs.bazel.build/versions/master/install.html).

For macOS users:

```
brew install bazelbuild/tap/bazel
# or, to upgrade it
brew upgrade bazelbuild/tap/bazel
```

## Building with bazel

Go code built with bazel requires all the `.go` source files and all the imports to be declared in the `BUILD.bazel` files that live in each Go package directory (and the same applies to the vendored sources).

Due to the predictability of Go this operation can be fully automated with `gazelle`:

```
bazel run //:gazelle
```

This will create missing `BUILD.bazel` files and patch existing ones (preserving comments and other rules).

> **NOTE**: Remember to run it if you notice your bazel builds fail mentioing that the Go compiler cannot find a given package.

However, `go mod` will remove these `BUILD.bazel` files in vendored sources. In order to fix these, we have included a small snippet to maintain them easily:

```
scripts/go-vendor
```

Finally, you can build the project:

```
bazel build //cmd/jpresolver:jpresolver
bazel build //cmd/jpdownloader:jpdownloader
```

## Running the tools

Once you build them with bazel, you can run them from the bazel workspace:

```
bazel-bin/cmd/jpresolver/...
bazel-bin/cmd/jpdownloader/..
```

You can also run them with bazel directly:

```
bazel run //cmd/jpresolver:jpresolver -- -h
bazel run //cmd/jpdownloader:jpdownloader -- -h
```

# jpresolver

This CLI allows to resolve transitive plugins dependencies from a list of plugins.

> **NOTE**: If there are incompatibilities between input and the computed output the tool will warn about it to fix it.

## Inputs

### Working directory

The working directory can be configured via `-working-dir` flag. This directory will be used for different purposes (see #working-directories).

### Output

The computed list of plugins will be written in the file specified via `-output` flag. It will follow the same schema that it is expected from the input.

### Input

The list of plugins must be provided via `-input` flag. It must follow the following schema:

```
{
    "plugins": [
        {
            "name": "kubernetes",
            "version": "1.18.2",
        }
    ]
}
```

## How to use it

Let's suppose we want to install the `kubernetes` plugin. We need to create a file (ie, _input.json_) with the following content):

```
{
    "plugins": [
        {
            "name": "kubernetes",
            "version": "1.18.2",
        }
    ]
}
```

We can run this tool either locally or using the container image:

```
$ bazel run //cmd/jpresolver:jpresolver -- -working-dir $PWD/.jenkins -input $PWD/input.json -output $PWD/output.json
INFO: Analyzed target //cmd/jpresolver:jpresolver (0 packages loaded, 0 targets configured).
INFO: Found 1 target...
Target //cmd/jpresolver:jpresolver up-to-date:
  bazel-bin/cmd/jpresolver/darwin_amd64_stripped/jpresolver
INFO: Elapsed time: 0.421s, Critical Path: 0.03s
INFO: 0 processes.
INFO: Build completed successfully, 1 total action
INFO: Running command line: bazel-bin/cmd/jpresolver/darwin_amd64_stripped/jpresolver -working-dir REDACTEDPWD/.jenkins -input REDACTEDPWD/input.json -output REDACTEDPWD/output.json
2019/09/16 12:37:58 #10> fetching kubernetes:1.18.2 metadata...
...
INFO: Build completed successfully, 1 total action
```

```
$ docker run -v $PWD:/workspace gcr.io/bitnami-labs/jenkins-plugin-resolver -t -- -input /workspace/input.json -output /workspace/output.json
2019/09/16 12:37:58 #10> fetching kubernetes:1.18.2 metadata...
...
```

We will get the following differences:

```
$ diff -ur input.json output.json
--- input.json	2019-09-16 12:37:43.000000000 +0200
+++ output.json	2019-09-16 12:37:58.000000000 +0200
@@ -1,8 +1,68 @@
 {
   "plugins": [
     {
+      "name": "apache-httpcomponents-client-4-api",
+      "version": "4.5.5-3.0"
+    },
+    {
+      "name": "authentication-tokens",
+      "version": "1.3"
+    },
+    {
+      "name": "cloudbees-folder",
+      "version": "6.9"
+    },
+    {
+      "name": "credentials",
+      "version": "2.2.0"
+    },
+    {
+      "name": "credentials-binding",
+      "version": "1.12"
+    },
+    {
+      "name": "docker-commons",
+      "version": "1.14"
+    },
+    {
+      "name": "durable-task",
+      "version": "1.30"
+    },
+    {
+      "name": "jackson2-api",
+      "version": "2.9.9"
+    },
+    {
       "name": "kubernetes",
       "version": "1.18.2"
+    },
+    {
+      "name": "kubernetes-credentials",
+      "version": "0.4.0"
+    },
+    {
+      "name": "plain-credentials",
+      "version": "1.5"
+    },
+    {
+      "name": "scm-api",
+      "version": "2.2.6"
+    },
+    {
+      "name": "structs",
+      "version": "1.19"
+    },
+    {
+      "name": "variant",
+      "version": "1.3"
+    },
+    {
+      "name": "workflow-api",
+      "version": "2.35"
+    },
+    {
+      "name": "workflow-step-api",
+      "version": "2.20"
     }
   ]
-}
+}
```

# jpdownloader

This CLI allows to download a list of plugins.

## Inputs

### Working directory

The working directory can be configured via `-working-dir` flag. This directory will be used for different purposes (see #working-directories).

### Input

The list of plugins must be provided via `-input` flag. It must follow the following schema:

```
{
    "plugins": [
        {
            "name": "kubernetes",
            "version": "1.18.2",
        }
    ]
}
```

## How to use it

Let's suppose we want to install the `kubernetes` plugin. We need to create a file (ie, _input.json_) with the following content):

```
{
    "plugins": [
        {
            "name": "kubernetes",
            "version": "1.18.2",
        }
    ]
}
```

We can resolve the list of dependencies first (`jpresolver`) and then run this tool either locally or using the container image:

```
$ bazel run //cmd/jpdownloader:jpdownloader -- -working-dir $PWD/.jenkins -input $PWD/input.json
INFO: Analyzed target //cmd/jpdownloader:jpdownloader (0 packages loaded, 0 targets configured).
INFO: Found 1 target...
Target //cmd/jpdownloader:jpdownloader up-to-date:
  bazel-bin/cmd/jpdownloader/darwin_amd64_stripped/jpdownloader
INFO: Elapsed time: 0.421s, Critical Path: 0.03s
INFO: 0 processes.
INFO: Build completed successfully, 1 total action
INFO: Running command line: bazel-bin/cmd/jpdownloader/darwin_amd64_stripped/jpdownloader -working-dir REDACTEDPWD/.jenkins -input REDACTEDPWD/input.json
2019/09/16 13:30:59 # 3> downloading credentials:2.2.0...
2019/09/16 13:30:59 # 7> downloading docker-commons:1.14...
2019/09/16 13:30:59 # 5> downloading credentials-binding:1.12...
2019/09/16 13:30:59 # 6> downloading durable-task:1.30...
2019/09/16 13:30:59 # 4> downloading cloudbees-folder:6.9...
2019/09/16 13:30:59 # 8> downloading jackson2-api:2.9.9...
2019/09/16 13:30:59 # 9> downloading kubernetes:1.18.2...
2019/09/16 13:30:59 #10> downloading kubernetes-credentials:0.4.0...
2019/09/16 13:30:59 # 2> downloading authentication-tokens:1.3...
2019/09/16 13:30:59 # 1> downloading apache-httpcomponents-client-4-api:4.5.5-3.0...
2019/09/16 13:31:01 # 6> downloaded durable-task:1.30.
2019/09/16 13:31:01 # 6> downloading plain-credentials:1.5...
2019/09/16 13:31:01 #10> downloaded kubernetes-credentials:0.4.0.
2019/09/16 13:31:01 #10> downloading scm-api:2.2.6...
2019/09/16 13:31:01 # 7> downloaded docker-commons:1.14.
2019/09/16 13:31:01 # 7> downloading structs:1.19...
2019/09/16 13:31:01 # 2> downloaded authentication-tokens:1.3.
2019/09/16 13:31:01 # 2> downloading variant:1.3...
2019/09/16 13:31:01 # 4> downloaded cloudbees-folder:6.9.
2019/09/16 13:31:01 # 4> downloading workflow-api:2.35...
2019/09/16 13:31:02 # 6> downloaded plain-credentials:1.5.
2019/09/16 13:31:02 # 6> downloading workflow-step-api:2.20...
2019/09/16 13:31:02 # 5> downloaded credentials-binding:1.12.
2019/09/16 13:31:02 # 2> downloaded variant:1.3.
2019/09/16 13:31:02 #10> downloaded scm-api:2.2.6.
2019/09/16 13:31:02 # 7> downloaded structs:1.19.
2019/09/16 13:31:02 # 4> downloaded workflow-api:2.35.
2019/09/16 13:31:02 # 6> downloaded workflow-step-api:2.20.
2019/09/16 13:31:03 # 3> downloaded credentials:2.2.0.
2019/09/16 13:31:04 # 1> downloaded apache-httpcomponents-client-4-api:4.5.5-3.0.
2019/09/16 13:31:04 # 8> downloaded jackson2-api:2.9.9.
2019/09/16 13:31:12 # 9> downloaded kubernetes:1.18.2.
2019/09/16 13:31:12 done!
INFO: Build completed successfully, 1 total action
```

```
$ docker run -v $PWD:/workspace gcr.io/bitnami-labs/jenkins-plugin-resolver -t -- -input /workspace/input.json -output /workspace/output.json
2019/09/16 13:30:59 # 3> downloading credentials:2.2.0...
2019/09/16 13:30:59 # 7> downloading docker-commons:1.14...
2019/09/16 13:30:59 # 5> downloading credentials-binding:1.12...
2019/09/16 13:30:59 # 6> downloading durable-task:1.30...
2019/09/16 13:30:59 # 4> downloading cloudbees-folder:6.9...
2019/09/16 13:30:59 # 8> downloading jackson2-api:2.9.9...
2019/09/16 13:30:59 # 9> downloading kubernetes:1.18.2...
2019/09/16 13:30:59 #10> downloading kubernetes-credentials:0.4.0...
2019/09/16 13:30:59 # 2> downloading authentication-tokens:1.3...
2019/09/16 13:30:59 # 1> downloading apache-httpcomponents-client-4-api:4.5.5-3.0...
2019/09/16 13:31:01 # 6> downloaded durable-task:1.30.
2019/09/16 13:31:01 # 6> downloading plain-credentials:1.5...
2019/09/16 13:31:01 #10> downloaded kubernetes-credentials:0.4.0.
2019/09/16 13:31:01 #10> downloading scm-api:2.2.6...
2019/09/16 13:31:01 # 7> downloaded docker-commons:1.14.
2019/09/16 13:31:01 # 7> downloading structs:1.19...
2019/09/16 13:31:01 # 2> downloaded authentication-tokens:1.3.
2019/09/16 13:31:01 # 2> downloading variant:1.3...
2019/09/16 13:31:01 # 4> downloaded cloudbees-folder:6.9.
2019/09/16 13:31:01 # 4> downloading workflow-api:2.35...
2019/09/16 13:31:02 # 6> downloaded plain-credentials:1.5.
2019/09/16 13:31:02 # 6> downloading workflow-step-api:2.20...
2019/09/16 13:31:02 # 5> downloaded credentials-binding:1.12.
2019/09/16 13:31:02 # 2> downloaded variant:1.3.
2019/09/16 13:31:02 #10> downloaded scm-api:2.2.6.
2019/09/16 13:31:02 # 7> downloaded structs:1.19.
2019/09/16 13:31:02 # 4> downloaded workflow-api:2.35.
2019/09/16 13:31:02 # 6> downloaded workflow-step-api:2.20.
2019/09/16 13:31:03 # 3> downloaded credentials:2.2.0.
2019/09/16 13:31:04 # 1> downloaded apache-httpcomponents-client-4-api:4.5.5-3.0.
2019/09/16 13:31:04 # 8> downloaded jackson2-api:2.9.9.
2019/09/16 13:31:12 # 9> downloaded kubernetes:1.18.2.
2019/09/16 13:31:12 done!
```

## How to find incompatibilities

This feature is intrinsic to the `jpresolver` tool. Example:

```
$ cat testdata/inputs.json
{
  "plugins": [
    {
      "name": "google-login",
      "version": "1.4"
    },
    {
      "name": "mailer",
      "version": "1.1"
    }
  ]
}

$ jpresolver -input testdata/inputs.json
2019/09/16 14:29:53  There were found some incompatibilities:
2019/09/16 14:29:53   ├── mailer:1.1:
2019/09/16 14:29:53   │   └── google-login:1.4 > mailer:1.6
2019/09/16 14:29:53
2019/09/16 14:29:53  You should bump the version in the input and evaluate the required changes.
```

# Working directories

The working directories will mainly work as a filesystem cache to avoid unnecessary computation after consecutive runs.

- `workdir/jpi` will be used to store jpi archives (jenkins plugins).
- `workdir/meta` will be used to store the plugins metadata.
- `workdir/graph` will be used to store the graphs from different runs.
