# Install

## Overview

Package `install` contains library used to install Kyma on already existing clusters.


## Usage

The installation process consist of two phases:
- `PrepareInstallation` - in this phase Tiller is installed and Kyma Installer is deployed to the cluster along with the default configuration.
- `StartInstallation` - in this phase the Installation is triggered by labeling Installation Custom Resource.

### Example

The example of using the library to install Kyma on properly configured Minikube cluster can be found [here](https://github.com/kyma-incubator/hydroform/tree/master/install/examples/example.go). 
