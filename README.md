# Hydroform

## Overview

Hydroform is an infrastructure SDK you can use to create and manage Kubernetes clusters. Hydroform allows you to manage your clusters on the desired target provider and location. 

The currently supported providers include:

- Google Cloud Platform
- Gardener

## Usage

Hydroform is a Go package you can use with any program to: 

- Create and provision the cluster on a selected cloud provider.
- Check the status of the cluster.
- Fetch the kubeconfig file to communicate with the cluster.
- Delete the cluster along with the configuration. 

### Actions 

The `actions` is a Hydroform subpackage brings even more extensibility to the standard Hydroform functionality. You can run actions before and after each Hydroform operation. You can also combine the actions in a sequence to run them in a specific order.

### Examples

Follow the links to view Hydroform usage examples: 
* [GCP](https://github.com/kyma-incubator/hydroform/tree/master/examples/gcp)
* [Gardener](https://github.com/kyma-incubator/hydroform/blob/master/examples/gardener/main.go)

For details, see [this](https://godoc.org/github.com/kyma-incubator/hydroform) documentation.