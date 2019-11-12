# Provision a Google Kubernetes Engine (GKE) cluster.

## Overview

This example shows you how to provision a GKE cluster.


## Installation

### Configure GKE

To provision a GKE cluster you need a service account the details of which you will pass as properties when executing this example.

1. Log in to GCP. Run:

```
gcloud auth application-default login
``` 
 
Log in using Google Cloud credentials.

2. Go to **IAM & Admin** > **Service accounts**.

3. Click **Create service account**.

![Create Account](./assets/create-account-gke.png)

4. Provide the details of your account.

![Account Details](./assets/add-details-gke.png)

5. Assign the roles.
>**NOTE**: The roles you need are Compute Admin, Kubernetes Engine Admin, and Service Account User.

![Assign Roles](./assets/assign-roles.png)

6. Optionally, you can grant user access to this service account.

7. Save your configuration. 

8. In the main **Service Account**  view, create and store the service account key for your account.

![Create key](./assets/create-account-key-gke.png)


### Run the example

1. To provision a new cluster on GCP, go to the `hydroform` directory and run:

```
go run ./examples/gcp/main.go -p {project_name} -c /{path/to/service_account_key.json}
```
2. Go to **Kubernetes Engine** > **Clusters**. Your cluster should appear there.