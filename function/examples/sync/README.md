# Synchronize the Function's workspace

## Overview

This example shows how to download the current Function's configuration from the cluster in the form of the `config.yaml` file. It synchronizes your configuration in the local workspace with the cluster configuration using the Function API.

## Usage

```bash
Usage:
	sync {NAME} [ --kubeconfig={FILE} ] [ --output={DIR} ] [options]

Options:
	-n --namespace={NAMESPACE}  Choose the Namespace for your Function (`default` is set as the default one).
	--debug                     Enable verbose output.
	-h --help                   Show available options.
	--version                   Show the example's version.
```