# Parallel Install

## Overview

The `parallel-install` library can deploy and uninstall Kyma on the already existing clusters.
The library allows you to configure several parallel workers. This feature significantly reduces the time of the operation.
> **NOTE:** Parallel processing works only if the components (Helm releases) are independent of one another.

## Usage

The top-level interface for library users is defined in the `deployment` package, in the `Installer` interface.
Before starting the deployment or uninstallation process, you need to provide a complete configuration
by creating an instance of the `Deployment` struct.
To do so, provide the `deployment.NewDeployment` function with the necessary parameters:

| Parameter      | Type                              | Example value | Description                                                                                                                            |
| -------------- | --------------------------------- | ------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| overrides      | `deployment.Overrides{}`          | -             | An instance of `deployment.Overrides` including all overrides that Helm must consider.                                                 |
| cfg            | `config.Config`                   | -             | Specifies fine-grained configuration for the deployment process. See the table with `config.Config` configuration options for details. |
| processUpdates | `chan<- deployment.ProcessUpdate` | -             | The library caller can pass a channel to retrieve updates of the running installation or uninstallation process.                       |

See all available configuration options for the `config.Config` type:

| Parameter                     | Type                                    | Example value                                                     | Description                                                                                                                                                                                                                |
| ----------------------------- | --------------------------------------- | ----------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| WorkersCount                  | `int`                                   | `4`                                                               | Number of parallel workers used for the `deploy` or `uninstall` operation.                                                                                                                                                 |
| CancelTimeout                 | `time.Duration`                         | `900 * time.Second`                                               | Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client.                                                                                            |
| QuitTimeout                   | `time.Duration`                         | `1200 * time.Second`                                              | Time after which the `deploy` or `uninstall` operation is aborted and returns an error to the user. Worker goroutines may still be working in the background. This value must be greater than the value for CancelTimeout. |
| HelmTimeoutSeconds            | `int`                                   | `360`                                                             | Timeout for the underlying Helm client.                                                                                                                                                                                    |
| BackoffInitialIntervalSeconds | `int`                                   | `1`                                                               | Initial interval used for exponential backoff retry policy.                                                                                                                                                                |
| BackoffMaxElapsedTimeSeconds  | `int`                                   | `30`                                                              | Maximum time used for exponential backoff retry policy.                                                                                                                                                                    |
| Log                           | `func(format string, v ...interface{})` | `fmt.Printf`                                                      | Function used for logging. To modify the logging behavior, set a custom logging function. For example, to disable any log output, provide an empty logging function implementation (`func(f string, v ...interface{}){}`). |
| Profile                       | `string`                                | `evaluation`                                                      | Deployment profile. The possible values are: "evaluation", "production", "".                                                                                                                                               |
| ComponentsListFile            | `string`                                | `/kyma/components.yaml`                                           | List of prerequisites and components used by the installer library.                                                                                                                                                        |
| ResourcePath                  | `string`                                | `$GOPATH/src/github.com/kyma-project/kyma/resources`              | Path to Kyma resources.                                                                                                                                                                                                    |
| InstallationResourcePath      | `string`                                | `$GOPATH/src/github.com/kyma-project/kyma/installation/resources` | Path to Kyma installation resources.                                                                                                                                                                                       |
| Version                       | `string`                                | `1.18.1`                                                          | The Kyma version.                                                                                                                                                                                                          |

>**NOTE:** This library also fetches overrides from ConfigMaps present in the cluster. However, overrides provided through `NewDeployment` have a higher priority.

Once you have a configured `Deployment` instance, use the following functions accordingly. You need to provide a kubeconfig pointing to a cluster for each function.

- `StartKymaDeployment` - Starts the deployment process. First, prerequisites are deployed linearly. Then, the components' deployment continues in parallel.
- `StartKymaUninstallation` - Starts the uninstallation process. The library uninstalls the components first, then it proceeds with the prerequisites' uninstallation in reverse order.
- `ReadKymaMetadata` - Retrieves Kyma metadata, such as Kyma version.

### Example

To learn how to use the library to deploy Kyma on a Gardener cluster, see this [example](../parallel-install/example/example.go).

## Utility Packages
The `parallel-install` library provides you with utility packages that helps you with Kyma installation.

### Download Package
The `download` package allows you to download remote files.

`download.GetFile` downloads a single file and returns the path to the downloaded file. If the destination directory does not exist, it is created automatically. The required parameters are:
| Parameter | Type     | Example value                                                                         | Description                                                                                                                            |
| --------- | -------- | ------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| file      | `string` | `https://storage.googleapis.com/kyma-mps-dev-artifacts/prometheus-config-updater.zip` | A URL or a local path. If it is a URL, the file is downloaded. If it is a local path, there's a check whether the file exists locally. |
| dstDir    | `string` | `myWorkspace/files`                                                                   | Path to which the file is downloaded.                                                                                                  |
  
`download.GetFiles` downloads a list of files and returns the paths to the downloaded files. If the destination directory does not exist, it is created automatically. The required parameters are:
| Parameter | Type       | Example value                                                                                                                                                                            | Description                                                                                                                                |
| --------- | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| files     | `[]string` | `[]string{"https://storage.googleapis.com/kyma-mps-dev-artifacts/prometheus-config-updater.zip", "https://storage.googleapis.com/kyma-mps-dev-artifacts/avs-bridge-noparent-1.3.5.tgz"}` | A list of URLs or local paths. For each URL, the file is downloaded. For each local path, there's a check whether the file exists locally. |
| dstDir    | `string`   | `myWorkspace/files`                                                                                                                                                                      | Path to which the files are downloaded.                                                                                                    |

### Archive Package
The `archive` package allows you to decompress `zip` and `tar.gz`/`tgz` files.

`archive.Unzip` extracts a `zip` file, moving all the files and folders within the zip file to an output directory. The required parameters are:
| Parameter | Type     | Example value                                     | Description                                                                              |
| --------- | -------- | ------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| src       | `string` | `myWorkspace/files/prometheus-config-updater.zip` | Path to the `zip` file.                                                                  |
| dst       | `string` | `myWorkspace/files/prometheusConfigUpdater`       | Path to the output directory in which files and folders within the `zip` file are moved. |

`archive.Untar` extracts a `tar.gz`/`tgz` file, moving all the files and folders within the `tar.gz`/`tgz` file to an output directory. The required parameters are:
| Parameter | Type     | Example value                                     | Description                                                                                       |
| --------- | -------- | ------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| src       | `string` | `myWorkspace/files/avs-bridge-noparent-1.3.5.tgz` | Path to the `tar.gz`/`tgz` file.                                                                  |
| dst       | `string` | `myWorkspace/files/avsBridgeNoparent`             | Path to the output directory in which files and folders within the `tar.gz`/`tgz` file are moved. |

### Git Package
The `git` package allows you to clone Git repositories.  

`git.CloneRepo` clones a Git repository. The required parameters are:
| Parameter | Type     | Example value                          | Description                                                                                                                                                              |
| --------- | -------- | -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| url       | `string` | `https://github.com/kyma-project/kyma` | URL to the Git repository.                                                                                                                                               |
| dstPath   | `string` | `myWorkspace/repos/kyma`               | Path to which the repository is cloned.                                                                                                                                  |
| rev       | `string` | `master`                               | Revision which is used for checking out the repository. It can be `master`, a release version (e.g. `1.4.1`), a commit hash (e.g. `34edf09a`), or a PR (e.g. `PR-9486`). |
