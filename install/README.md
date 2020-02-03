# Install

## Overview

The [`install`](https://godoc.org/github.com/kyma-incubator/hydroform/install/installation) package allows you to install Kyma on already existing clusters.


## Usage

The installation process consist of two phases triggered by the following functions: 
* `PrepareInstallation` which creates all necessary Kyma resources, installs  Tiller, and deploys the Kyma Installer to the cluster along with the default configuration.
* `StartInstallation` which triggers the installation by labeling the Installation Custom Resource. 

### Example

To learn how to use the library to install Kyma on a properly configured Minikube cluster, see this [example](../install/example/example.go).
