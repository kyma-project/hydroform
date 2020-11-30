# Parallel Install

## Overview

The `parallel-install` library can install and uninstall Kyma on the already existing clusters.
The library allows you to configure several concurrent workers. This feature significantly reduces the time of the operation.
Remember that concurrent processing works only if the components (Helm Releases) are independent of each other.

## Usage

Before starting the installation or uninstallation process, you need to provide a complete configuration by creating an instance of the `Installation` struct. To do so, use the `installation.NewInstallation` function.

| Parameter | Type | Example value | Description |
| --- | --- | --- | --- |
| prerequisites | `[][]string` | `{"cluster-essentials", "kyma-system"}, {"istio", "istio-system"},` | Array of the component's name and Namespace pairs. These components will be installed first, linearly, in a declared order. |
| componentsYaml | `string` | - | Content of the [Installation CR](https://kyma-project.io/docs/#custom-resource-installation). Components will be extracted and installed in parallel. |
| overridesYaml | `[]string` | `{ "foo: bar", "val: example" }` | List of Helm overrides. The latter the override, the higher is its priority. |
| resourcesPath | `string` | `/go/src/github.com/kyma-project/kyma/resources` | Path to the Kyma resources. |
| concurrency | `int` | `3` | Specifies how many components are installed simultaneously. |

>**NOTE:** This library also fetches overrides from ConfigMaps present in the cluster. However, overrides provided through `NewInstallation` have a higher priority.

Use the following functions accordingly. You need to provide a kubeconfig pointing to a cluster for each function.

- `StartKymaInstallation` - Starts the installation process. First, prerequisites are installed linearly. Then, the components' installation continues in parallel.
- `StartKymaUninstallation` - Starts the uninstallation process. The library uninstalls the components first, then it proceeds with the prerequisites' uninstallation in reverse order.

### Example

To learn how to use the library to install Kyma on a Gardener cluster, see this [example](../parallel-install/example/example.go).
