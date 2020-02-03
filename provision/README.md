# Provision

## Overview

The [`provision`](https://godoc.org/github.com/kyma-incubator/hydroform/provision) package provided in this module allows you to create and manage clusters.

## Usage

The package includes the  `provision`, `status`, `credentials`, and `deprovision` functions. Use them to:

- Create and provision the cluster on a selected cloud provider.
- Check the status of the cluster.
- Fetch the `kubeconfig` file to communicate with the cluster.
- Delete the cluster along with the configuration. 

### Actions 

The `actions` Hydroform subpackage brings even more extensibility to the standard Hydroform functionality. You can run actions before and after each Hydroform operation. You can also combine the actions in a sequence to run them in a specific order.

### Examples

Follow the links to view the [usage examples](./examples/README.md).
