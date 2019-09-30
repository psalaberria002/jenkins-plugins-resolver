![](https://github.com/bitnami-labs/jenkins-plugins-resolver/workflows/Continuous%20Deployment/badge.svg)

# Jenkins Plugins Resolver

Jenkins Plugin Resolver is a Go application to manage Jenkins projects.

This application manages your project dependencies, by resolving their transitive dependencies offline (no need to run Jenkins) and downloading the full list of dependencies in the Jenkins plugins folder.

## Installation

You can either use the Docker images or build the binaries.

### Docker

You need to [install docker](https://runnable.com/docker/getting-started/) to use the docker images.

```shell
docker pull gcr.io/bitnami-labs/jenkins-plugins-resolver:latest
docker pull gcr.io/bitnami-labs/jenkins-plugins-downloader:latest
```

### Binaries

You need to [install bazel](#install-bazel) to build the binaries locally.

```shell
bazel build //cmd/jpresolver:jpresolver
bazel build //cmd/jpdownloader:jpdownloader
```

## Usage

Write your [project file](docs/project-file.md) with your project dependencies.

The `jpresolver` tool will [resolve the transitive dependencies](docs/jpresolver.md) for your project file.

```shell
jpresolver -optional
```

> **NOTE**: The `-optional` flag will enable resolving optional dependencies too.

Once it runs, it will generate a [lock file](docs/lock-file.md) describing all the project dependencies (required plugins and transitive dependencies).

The `jpdownloader` tool is meant to run in a Jenkins environment. It will [download the project dependencies](docs/jpdownloader.md) in your Jenkins plugins folder (`JENKINS_HOME/plugins`).

```shell
jpdownloader
```

## Development

This project uses the standard [go mod](https://blog.golang.org/using-go-modules) tool to manage and vendor Go dependencies.

Aditionally, it uses [bazel](https://bazel.build) to manage the project build.

### Install bazel

Feel free to follow [official installation instructions](https://docs.bazel.build/versions/master/install.html).

### Building with bazel

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

### Running with bazel

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
