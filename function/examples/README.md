# Function Examples

## Overview

This module contains examples showing how to use Function API.

## Operators

Operators allow you to perform these two operations on the given resources: `apply` and `delete`. 
Both operations can accept callbacks that will be executed before and/or after each operation.

The library contains two implementations of an operator interface: `GenericOperator` and `TriggersOperator`.

See how the examples use specific libraries to:

* [Apply resources](./cmd/operator/apply/main.go).
* [Delete resources](./cmd/operator/delete/main.go).

## Callbacks

Callbacks are optional Functions that you can provide to execute them before and/or after each operation. You can chain callbacks by providing multiple Functions.

See how the examples use specific libraries to handle:

* [Pre-operation callbacks](./cmd/callbacks/pre/main.go).
* [Post-operation callbacks](./cmd/callbacks/pre/main.go).

## Manager

The Manager allows you to control the hierarchy of the operators in the parent-children relation. It handles the life cycle of objects created by the operators and tracks the references between the objects.

See how the examples use specific libraries to [handle sibling-children owner references](./cmd/manager/main.go).

## Function Workspace

The Function workspace is a configuration YAML file that was designed to work with the Kyma CLI `init`, `apply`, and `sync` commands to quickly create or update Kubernetes resources.

It aggregates properties of a Function object and other Function-related objects, such as triggers.

A sample configuration file looks as follows:

```yaml
name: example
namespace: example-ns
runtime: python38
resource:
    limits:
        cpu: 100m
        memory: 128Mi
    requests:
        cpu: 50m
        memory: 64Mi
source:
    sourceType: inline
    sourcePath: /tmp/test-fn-git
triggers:
    version: v1
    type: t1
    source: src1
```

See how the examples use specific libraries to:

* [Initialize the workspace locally (`init`)](./cmd/workspace/init/main.go).
* [Apply your workspace on a cluster (`apply`)](./cmd/workspace/apply/main.go).
* [Fetch cluster resources to synchronize your local workspace (`sync`)](./cmd/workspace/sync/main.go).

## Docker

Docker allow you to perform operation on the given resources: `run`.

See how the examples use specific libraries to:
