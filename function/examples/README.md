# Function Examples

## Overview

This module contains examples showing how to use Function API.

## Operators

Operators allow you to perform these two operations on the given resources: `apply` and `delete`. 
Both operations can accept callbacks that will be executed before and/or after each operation.

The library contains two implementations of an operator interface: `GenericOperator` and `SubscriptionOperator`.

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

It aggregates properties of a Function object and other Function-related objects, such as subscriptions.

A sample configuration file looks as follows:

```yaml
name: function-crazy-karol9
namespace: testme
runtime: nodejs12
source:
  sourceType: inline
  sourcePath: /tmp/cli-test
subscriptions:
  - name: first-function-subscription
    protocol: NATS
    filter:
      dialect: NATS100
      filters:
        - eventSource:
            property: source
            type: exact
            value: ""
          eventType:
            property: type
            type: exact
            value: sap.kyma.custom.test.order.created.v1
```

See how the examples use specific libraries to:

* [Initialize the workspace locally (`init`)](./cmd/workspace/init/main.go).
* [Apply your workspace on a cluster (`apply`)](./cmd/workspace/apply/main.go).
* [Fetch cluster resources to synchronize your local workspace (`sync`)](./cmd/workspace/sync/main.go).

## Docker

Docker allows you to perform the `run` operation on the given resources.

See how the examples use specific libraries to [run resources](./cmd/docker/run/main.go).
