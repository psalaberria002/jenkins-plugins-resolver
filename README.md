![](https://github.com/bitnami-labs/jenkins-plugins-resolver/workflows/Continuous%20Deployment/badge.svg)

# jenkins-plugins-resolver

Go tools to manage Jenkins plugins resolution, such as transitive dependencies graph computation and download.

These tools are thought to be integrated with _configuration-as-code_ Jenkins deployments.
In this kind of scenario, you may want to have a _"list"_ describing the plugins that your Jenkins project requires. However, if you leave Jenkins to install these plugins, it will install other dependencies that may break your deployment as there are incompatibilities between them or requirements that you will only notice at runtime.

In order to avoid this kind of issues, it makes sense to have an offline tool that, given a list of plugins, it renders a fully-qualified list of plugins so Jenkins will start with no additional operations.

Therefore, we need to tasks here:

1. Compute a fully-qualified list of plugins from a small subset of _"required"_ plugins.
2. Donwload a list of plugins in the `JENKINS_HOME/plugins` folder.

Ideally, the `JENKINS_HOME/plugins` should be removed any time we modify the list of plugins and restart/re-deploy Jenkins.


# Quickstart

According to the tasks we have mentioned we need to perform two tasks.

## Compute a fully-qualified list of plugins from a small subset of them.

The `jpresolver` tool assumes that you have a [**project file**](#project-file) file describing the plugins that your Jenkins project depends on:

```
# JSON
{
  "dependencies": {
    "google-login": "1.4"
  }
}


# YAML
dependencies:
  google-login: 1.4


# Jsonnet
local auth_deps = {
    'google-login': '1.4',
};

{
    dependencies: auth_deps,
}
```

This tool works in the same that other package manager tools do: It will inspect the [**project file**](#project-file) and will create a [**lock file**](#lock-file) with all the required plugins set. This means that it will resolve transitive dependencies and will warn you if any of the project dependencies is incompatible (older) than one of those transitive dependencies.

We can run this tool using the docker image:

> **NOTE**: Take a look at the ["Running the tools"](#running-the-tools) section if you prefer to run the binaries directly.

```
$ docker run --rm -v $PWD:/ws -w /ws gcr.io/bitnami-labs/jenkins-plugins-resolver:latest
2019/09/19 12:37:31 > fetching google-login:1.4 metadata...
2019/09/19 12:37:34 > fetching mailer:1.6 metadata...
```

After this run, the lock file has been generated:

```
{
  "plugins": [
    {
      "name": "google-login",
      "version": "1.4"
    },
    {
      "name": "mailer",
      "version": "1.6"
    }
  ]
}
```

As you can see, the lock file contains an additional dependency.

## Download a list of plugins

This tool is thought to download the plugins described in the lock file to the `JENKINS_HOME/plugins` folder. This process may vary depending on your deployment workflows.

### Kubernetes

If you are using kubernetes, you may use this tool within an init container to download the list of plugins (from a `configmap`, for example) directly to the plugins folder (persisted `volume` mounted with write permissions).

### Virtual Machines

If you are using a virtual machine, you may use this tool as part of the Jenkins start command/service so it prepares the plugins before Jenkins actually starts.

Independtly to your DevOps, let's assume you are running the docker image in the target environment and the lock file is in your current directory:

```
$ docker run --rm -e JENKINS_HOME -v $JENKINS_HOME:$JENKINS_HOME -v $PWD:/ws -w /ws gcr.io/bitnami-labs/jenkins-plugins-downloader:latest
2019/09/19 12:57:32 > downloading mailer:1.6...
2019/09/19 12:57:32 > downloading google-login:1.4...
2019/09/19 12:57:35 done!
```

> **NOTE**: The `-v $JENKINS_HOME:$JENKINS_HOME` flag will mount our Jenkins home path in the same location of the container. The `-e JENKINS_HOME` flag will allow the tool to auto-discover this location and `-v $PWD:/ws -w /ws` will allow to find the lock file in the current directory.

## How to find incompatibilities

This feature is intrinsic to the `jpresolver` tool. Example:

```yml
dependencies:
  google-login: 1.4
  # mailer will force an incompatibility as google-login requires mailer:1.6
  mailer: 1.1
```

```
$ docker run --rm -v $PWD:/ws -w /ws gcr.io/bitnami-labs/jenkins-plugins-resolver:latest
2019/09/24 17:03:51 > fetching google-login:1.4 metadata...
2019/09/24 17:03:51  There were found some incompatibilities:
2019/09/24 17:03:51   ├── mailer:1.1 (a newer version was locked: mailer:1.6):
2019/09/24 17:03:51   │   └── google-login:1.4 > mailer:1.6
```

## Use cache

If you want to speed up the local resolution process, you can use the `-working-dir` flag to cache the plugins information:

```
docker run --rm -v $PWD:/ws -w /ws gcr.io/bitnami-labs/jenkins-plugins-resolver:latest -working-dir /ws/.jenkins
```

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

This CLI allows to resolve transitive plugins dependencies from a project file.

> **NOTE**: If there are incompatibilities between input and the computed output the tool will warn about it to fix it.

## Inputs

### Working directory

The working directory can be configured via `-working-dir` flag. This directory will be used for different purposes (see #working-directories).

### Lock file

The computed list of plugins will be written in the file specified via `-output` flag (defaults to the relative `<input>.lock`). It will follow the same schema that it is expected from the input.

### Project file

The project settings and plugins dependencies can be provided via `-input` flag (defaults to the relative `plugins.json` file). It must follow the following schema:

```
{
  "dependencies": {
    "google-login": "1.4"
  }
}
```

> **NOTE**: You can use YAML and Jsonnet formats too.

# jpdownloader

This CLI allows to download a list of plugins.

## Inputs

### Working directory

The working directory can be configured via `-working-dir` flag. This directory will be used for different purposes (see #working-directories).

### Output directory

The downloaded list of plugins will be copied to the output directory specified via `-output-dir` flag (defaults to the `JENKINS_HOME/plugins` folder).

### Lock file

The list of plugins can be provided via `-input` flag (defaults to the relative `plugins.json.lock` file). It must follow the following schema:

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


# Working directories

The working directories will mainly work as a filesystem cache to avoid unnecessary computation after consecutive runs.

- `workdir/jpi` will be used to store jpi archives (jenkins plugins).
- `workdir/meta` will be used to store the plugins metadata.
- `workdir/graph` will be used to store the graphs from different runs.
