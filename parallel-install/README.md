# Parallel Install

## Overview

The `parallel-install` library can deploy and uninstall Kyma on the already existing clusters.
The library allows you to configure several parallel workers. This feature significantly reduces the time of the operation.
> **NOTE:** Parallel processing works only if the components (Helm releases) are independent of one another.

## Usage

The top-level interface for library users is defined in the `deployment` package, in the `Installer` interface.
Before starting the deployment or uninstallation process, you need to provide a complete configuration
by creating an instance of the `Deployment` struct.
To do so, provide the `deployment.NewDeployment` function with necessary parameters:

| Parameter | Type | Example value | Description |
| --- | --- | --- | --- |
| prerequisites | `[][]string` | `{"cluster-essentials", "kyma-system"}, {"istio", "istio-system"},` | Array of the component's name and Namespace pairs. These components will be deployed first, linearly, in a declared order. |
| componentsYaml | `string` | - | Content of the [Installation CR](https://kyma-project.io/docs/#custom-resource-installation). Components will be extracted and deployed in parallel. |
| overridesYaml | `[]string` | `{ "foo: bar", "val: example" }` | List of Helm overrides. The latter the override, the higher is its priority. |
| resourcesPath | `string` | `/go/src/github.com/kyma-project/kyma/resources` | Path to the Kyma resources directory. It contains subdirectories with all Kyma components' charts |
| cfg | `config.Config` | - | Specifies fine-grained configuration for the deployment process. See the table with `config.Config` configuration options for details. |
| processUpdates  | `chan<- deployment.ProcessUpdate` | - | The library caller can pass a channel to retrieve updates of the running installation or uninstallation process. |

See all available configuration options for the `config.Config` type:

| Parameter | Type | Example value | Description |
| --- | --- | --- | --- |
| WorkersCount | `int` | `4` | Number of parallel workers used for the `deploy` or `uninstall` operation. |
| CancelTimeout | `time.Duration` | `900 * time.Second` | Time after which the workers' context is canceled. Pending worker goroutines (if any) may continue if blocked by a Helm client. |
| QuitTimeout | `time.Duration` | `1200 * time.Second` | Time after which the `deploy` or `uninstall` operation is aborted and returns an error to the user. Worker goroutines may still be working in the background. This value must be greater than the value for CancelTimeout. |
| HelmTimeoutSeconds | `int` | `360` | Timeout for the underlying Helm client. |
| BackoffInitialIntervalSeconds | `int` | `1` | Initial interval used for exponential backoff retry policy. |
| BackoffMaxElapsedTimeSeconds | `int` | `30` | Maximum time used for exponential backoff retry policy. |
| Log | `func(format string, v ...interface{})` | `fmt.Printf` | Function used for logging. To modify the logging behavior, set a custom logging function. E.g. for disabling any log-output, just provide an empty logging function implementation (`func(f string, v ...interface{}){}`) |
| Profile | `string` | `evaluation` | Deployment profile. The possible values are: "evaluation", "production", "".  |

>**NOTE:** This library also fetches overrides from ConfigMaps present in the cluster. However, overrides provided through `NewDeployment` have a higher priority.

Once you have a configured `Deployment` instance, use the following functions accordingly. You need to provide a kubeconfig pointing to a cluster for each function.

- `StartKymaDeployment` - Starts the deployment process. First, prerequisites are deployed linearly. Then, the components' deployment continues in parallel.
- `StartKymaUninstallation` - Starts the uninstallation process. The library uninstalls the components first, then it proceeds with the prerequisites' uninstallation in reverse order.

### Example

To learn how to use the library to deploy Kyma on a Gardener cluster, see this [example](../parallel-install/example/example.go).
