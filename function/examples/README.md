# Function Examples

## Overview

This module contains examples showing how to use Function API.

## Operators

Operators allow you to perform these two operations on the given resources: `apply` and `delete`. 
Both operations can accept callbacks that will be executed before and/or after each operation.

The library contains two implementations of an operator interface: `GenericOperator` and `TriggersOperator`.

Examples:

* [apply operation](./cmd/operator/apply/main.go)
* [Delete operation](./cmd/operator/delete/main.go)

## Callbacks

Callbacks are optional functions provided by function API user to be executed before and/or after each operation. User may chain callbacks by providing multiple functions.

Examples:

* [pre-operation](./cmd/callbacks/pre/main.go)
* [post-operation](./cmd/callbacks/pre/main.go)

## Manager

Manager allows to controll the hierarchy of the operators in `parent` - `children` relation. It handles the life cycle of objects created by the operators and tracks the references between the objects.

Example:
* [sibling-children owner references](./cmd/manager/main.go) 

## Workspace initialization show case

An [example](./cmd/init/main.go) application that shows how to integrate Function API to to initialize the serverless workspace.
