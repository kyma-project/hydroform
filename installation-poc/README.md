# Parallel installation library

## Overview

The `parallel-install` package allows you to install and uninstall Kyma on already existing clusters.

## Usage

Before starting an installation/uninstallation you need to provide a configuration by creating an instance of the `Installation` struct. To do so, use the `installation.NewInstallation` function. 

| Parameter | Type | Example value | Description |
| --- | --- | --- | --- |
| prerequisites | `[][]string` | `{"cluster-essentials", "kyma-system"}, {"istio", "istio-system"},` | Array of component's name and namespace pairs. This components will be installed first in a declared order linearly. |
| componentsYaml | `string` | - | Content of the [Installation CR](https://kyma-project.io/docs/#custom-resource-installation). Components will be extracted and installed in parallel. |
| overridesYaml | `[]string` | `{ "foo: bar", "val: example" }` | List of Helm overrides. The latter the override, the higher is its priority. |
| resourcesPath | `string` | `/go/src/github.com/kyma-project/kyma/resources` | Path to the Kyma resources. |
| concurrency | `int` | `3` | Specifies how many components install simultaneously. |

>**NOTE:** Library also fetches overrides from Config Maps present in the cluster. However, overrides provided through the `NewInstallation` have higher priority.

Use the following functions accordingly. You need to provide a kubeconfig pointing to a cluster to each function.

- `StartKymaInstallation` - Starts installation. Prerequisites are installed first linearly, then components' installation continues in parallel. 
- `StartKymaUninstallation` - Starts uninstallation. Library uninstalls components first, then proceeds with prerequisites' uninstallation in reverse order.

## Example 

//TODO: once main.go is moved to a folder and works provide instructions how to use it
