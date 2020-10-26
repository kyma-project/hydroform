# Function Examples

## Overview

This module contains examples showing how to use Function API.

## Operators

Operators allow you to perform these two operations on the given resources: `apply` and `delete`. 
Both operations can accept callbacks that will be executed before and/or after each operation.

The library contains two implementations of an operator interface: `GenericOperator` and `TriggersOperator`.

Examples:

* [Apply operation](./cmd/operator/apply/main.go)
* [Delete operation](./cmd/operator/delete/main.go)

## Callbacks

Callbacks are optional Functions that you can provide to execute them before and/or after each operation. You can chain callbacks by providing multiple Functions.

Examples:

* [Pre-operation](./cmd/callbacks/pre/main.go)
* [Post-operation](./cmd/callbacks/pre/main.go)

## Manager

The Manager allows you to control the hierarchy of the operators in the parent-children relation. It handles the life cycle of objects created by the operators and tracks the references between the objects.

Example:
* [Sibling-children owner references](./cmd/manager/main.go) 

## Function Workspace

The function workspace is a configuration __yaml__ file and was designed to work with Kyma CLI `apply` and `sync` commands to quickly create or update kubernetes resources.

It aggregates properties of a function object and other function-related objects e.g. trigger.

The example of configuration file may look like this: 

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

Examples:

* [initialize workspace](./cmd/workspace/init/main.go)
* [synchronize workspace](./cmd/workspace/sync/main.go)
