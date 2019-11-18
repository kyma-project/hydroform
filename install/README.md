# Install

## Overview

The `install` package contains the library used to install Kyma on already existing clusters.


## Usage

The installation process consist of two phases:
- `PrepareInstallation` - in this phase Tiller is installed and the Kyma Installer is deployed to the cluster along with the default configuration.
- `StartInstallation` - in this phase the Installation is triggered by labeling Installation Custom Resource.

### Example

To see how to use the library to install Kyma on a properly configured Minikube cluster, see this [example](https://github.com/kyma-incubator/hydroform/tree/master/install/examples/example.go).
