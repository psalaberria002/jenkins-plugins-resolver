
< [Prev](project-file.md) (*Project File*) | [Next](lock-file.md) (*Lock file*) >

___

# jpresolver

This CLI allows to resolve transitive plugins dependencies from a project file.

If there are incompatibilities between the project dependencies and the transitive dependencies the tool will warn about it and it won't generate the [lock file](#lock-file).

## Usage

```console
jpresolver -input plugins.json -war jenkins.war -optional
```

## Inputs

### Project file

The [project file](project-file.md) can be provided via `-input` flag (defaults to the relative `plugins.json` file).

### Jenkins .war file

The Jenkins war file can be provided via `-war` flag and it will be used to improve the dependencies resolution from the provided Jenkins version.

### Optional

If you are interested on downloading optional dependencies too, you can provide the `-optional` flag.

### Lock file

The [lock file](lock-file.md) will be written in the `<input-file-basename>-lock.json` file.

Examples:

| **Input**             | **Output**
| --------------------- | --------------------------
| `plugins.json`        | `plugins-lock.json`
| `myproject.prod.json` | `myproject.prod-lock.json`

### Working directory

The working directory can be configured via `-working-dir` flag. It defaults to `HOME/.jenkins`.

The working directories will mainly work as a [filesystem cache](#cache) to avoid unnecessary computation after consecutive runs.

- `workdir/jpi` will be used to store jpi archives (jenkins plugins).
- `workdir/meta` will be used to store the plugins metadata.
- `workdir/graph` will be used to store the plugins dependencies graph from different runs.
- `workdir/war` will be used to store the Jenkins war detached plugins from different runs.

## Cache

If you want to speed up the local resolution process, you must use the same `-working-dir` between different runs. This directory will keep a local copy of the plugins, metadata and graphs so consecutive runs will avoid downloading plugins, computing their metadata, etc.

## How to find incompatibilities

This feature is intrinsic to the `jpresolver` tool. Example:

```yml
dependencies:
  google-login: 1.4
  # mailer will force an incompatibility as google-login requires mailer:1.6
  mailer: 1.1
```

```console
$ jpresolver -input plugins.yml
2019/10/09 23:37:46 Computing graph...
2019/10/09 23:37:46 Recorded graph to disk: /Users/jotadrilo/.jpr/graph/4db6fb76c293c35ac1bdb25838028d6db8cec47be70dfe0a71ed3fd155b93bc7.graph
2019/10/09 23:37:46  There were found some incompatibilities:
2019/10/09 23:37:46   └── Requester: mailer:1.1 (project file)
2019/10/09 23:37:46       Cause: Some plugins require a newer version.
2019/10/09 23:37:46         └── google-login:1.4 (project file) > mailer:1.6
2019/10/09 23:37:46
```
___

< [Prev](project-file.md) (*Project File*) | [Next](lock-file.md) (*Lock file*) >
