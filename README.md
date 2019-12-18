# Hydroform

## Overview

Hydroform is an infrastructure SDK you can use to create and manage Kubernetes clusters. Hydroform allows you to manage your clusters on the desired target provider and location. 

The currently supported platforms include:

- Google Cloud Platform
- Gardener

## Usage

Hydroform is a [Go package](https://godoc.org/github.com/kyma-incubator/hydroform) that you can use with any program. It gives you the following commands: `provision`, `status`, `credentials`, and `deprovision`. Use them to: 

- Create and provision the cluster on a selected cloud provider.
- Check the status of the cluster.
- Fetch the `kubeconfig` file to communicate with the cluster.
- Delete the cluster along with the configuration. 

