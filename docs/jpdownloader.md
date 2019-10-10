
< [Prev](lock-file.md) (*Lock File*) | [Home](../README.md) >

___

# jpdownloader

This CLI allows to download a list of plugins.

## Inputs

### Lock file

The [lock file](lock-file.md) can be provided via `-input` flag (defaults to the relative `plugins.json.lock` file).

### Output directory

The downloaded list of plugins will be copied to the output directory specified via `-output-dir` flag (defaults to the `JENKINS_HOME/plugins` folder).

### Working directory

The working directory can be configured via `-working-dir` flag. It defaults to `HOME/.jenkins`.

The working directories will mainly work as a [filesystem cache](#cache) to avoid unnecessary computation after consecutive runs.

- `workdir/jpi` will be used to store jpi archives (jenkins plugins).

## Cache

If you want to speed up the local resolution process, you must use the same `-working-dir` between different runs. This directory will keep a local copy of the plugins so consecutive runs will avoid downloading plugins.

## Environments

This tool is thought to download the plugins described in the [lock file](lock-file.md) to the `JENKINS_HOME/plugins` folder. This process may vary depending on your environment and deployment workflows.

### Kubernetes

If you are using kubernetes, you may use this tool within an init container to download the list of plugins (from a `configmap`, for example) directly to the plugins folder (persisted `volume` mounted with write permissions).

### Local

If you are using a virtual machine, hosted server, local machine, etc, you may use this tool as part of the Jenkins start command/service so it prepares the plugins before Jenkins actually starts.

Independently to your DevOps, let's assume you are running the tool in the target environment and the lock file is in your current directory:

#### Binaries

```console
$ jpdownloader
2019/09/19 12:57:32 > downloading mailer:1.6...
2019/09/19 12:57:32 > downloading google-login:1.4...
2019/09/19 12:57:35 done!
```

#### Docker images

```console
$ docker run --rm -e JENKINS_HOME -v $JENKINS_HOME:$JENKINS_HOME -v $PWD:/ws -w /ws gcr.io/bitnami-labs/jenkins-plugins-downloader:latest
2019/09/19 12:57:32 > downloading mailer:1.6...
2019/09/19 12:57:32 > downloading google-login:1.4...
2019/09/19 12:57:35 done!
```

> **NOTE**: The `-v $JENKINS_HOME:$JENKINS_HOME` flag will mount our Jenkins home path in the same location of the container. The `-e JENKINS_HOME` flag will allow the tool to auto-discover this location and `-v $PWD:/ws -w /ws` will allow to find the lock file in the current directory.

___

< [Prev](lock-file.md) (*Lock File*) | [Home](../README.md) >
****
