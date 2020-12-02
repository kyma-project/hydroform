# Parallel Install

## Overview

The `parallel-install` library can install and uninstall Kyma on the already existing clusters.
The library allows you to configure several parallel workers. This feature significantly reduces the time of the operation.
> **NOTE:** Parallel processing works only if the components (Helm releases) are independent of each other.

## Usage

The top-level interface for library users is defined in the `installation` package, in the `Installer` interface.
Before starting the installation or uninstallation process, you need to provide a complete configuration
by creating an instance of the `Installation` struct.
To do so, provide the `installation.NewInstallation` function with necessary parameters:

| Parameter | Type | Example value | Description |
| --- | --- | --- | --- |
| prerequisites | `[][]string` | `{"cluster-essentials", "kyma-system"}, {"istio", "istio-system"},` | Array of the component's name and Namespace pairs. These components will be installed first, linearly, in a declared order. |
| componentsYaml | `string` | - | Content of the [Installation CR](https://kyma-project.io/docs/#custom-resource-installation). Components will be extracted and installed in parallel. |
| overridesYaml | `[]string` | `{ "foo: bar", "val: example" }` | List of Helm overrides. The latter the override, the higher is its priority. |
| resourcesPath | `string` | `/go/src/github.com/kyma-project/kyma/resources` | Path to the Kyma resources. |
| cfg | `config.Config` | - | Specifies fine-grained configuration for the installation process. See below for explanation. |

`config.Config` options:

| Parameter | Type | Example value | Description |
| --- | --- | --- | --- |
| WorkersCount | `int` | `4` | Number of parallel workers used for the `install` or `uninstall` operation. |
| CancelTimeoutSeconds | `int` | `900` | Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client. |
| QuitTimeoutSeconds | `int` | `1200` | After this time install/delete operation is aborted and returns an error to the user. Worker goroutines may still be working in the background. Must be greater than CancelTimeoutSeconds. |
| HelmTimeoutSeconds | `int` | `360` | Timeout for the underlying Helm client. |
| BackoffInitialIntervalSeconds | `int` | `1` | Initial interval used for exponent backoff retry policy. |
| BackoffMaxElapsedTimeSeconds | `int` | `30` | Maximum time used for exponential backoff retry policy. |
| Log | `func(format string, v ...interface{})` | `fmt.Printf` | Function used for logging. |

>**NOTE:** This library also fetches overrides from ConfigMaps present in the cluster. However, overrides provided through `NewInstallation` have a higher priority.

Once you have a configured `Installation` instance, use the following functions accordingly. You need to provide a kubeconfig pointing to a cluster for each function.

- `StartKymaInstallation` - Starts the installation process. First, prerequisites are installed linearly. Then, the components' installation continues in parallel.
- `StartKymaUninstallation` - Starts the uninstallation process. The library uninstalls the components first, then it proceeds with the prerequisites' uninstallation in reverse order.

### Example

To learn how to use the library to install Kyma on a Gardener cluster, see this [example](../parallel-install/example/example.go).
