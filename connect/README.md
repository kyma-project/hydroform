# Connect

## Overview

The [`connect`](https://godoc.org/github.com/kyma-incubator/hydroform/connect) package provided in this module allows third-party applications to easily integrate with Kyma.

## Usage

The package includes functions to do the following:

- `connect` : Create a secure connection to the Kyma Application Connector
- `registerService`, `updateService`, `deleteService` : Register / Update / Delete a third-party service with Kyma
- `sendEvent` : Send an event, which can, in turn, trigger subscribed lambda functions.
- `getSubscribedEvents` : Get a list of all active events for the application. This can be useful to make sure that we're not sending notifications regarding inactive events.
- `renewCertificateSigningRequest`, `revokeCertificate` : Renew / Revoke the client's certificates 

### Examples

Follow the links to view the [usage examples](./examples).
