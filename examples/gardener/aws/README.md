# Provision an AWS cluster with Gardener

## Overview

This example shows you how you can use Gardener to provision a cluster on Amazon Web Services provider. For the example to work, you need to configure Gardener and AWS to enable mutual access. 


## Installation

### Configure Gardener and AWS


1. In Gardener, click **+Create project** to set up a new project on Gardener. 

    ![Create Project](../assets/create-project.png)

2. Go to **Secrets** > **Amazon Web Services**. Click **?** and copy the AWS IAM policy. You will need this information for AWS to grant the access to Gardener.

    ![Copy policy](../assets/copy-policy.png)

3. Go to AWS IAM Console. Use the instructions to create [a new policy](https://gardener.cloud/050-tutorials/content/howto/gardener_aws/#create-new-policy) and [Gardener technical user](https://gardener.cloud/050-tutorials/content/howto/gardener_aws/#create-a-new-technical-user). Make sure you store the **Access Key Id** and **Secret Access Key** for this user.

4. In Gardener, go to **Secrets** > **Amazon Web Services** and click **+** to add a new secret. Use the **Access Key Id** and **Secret Access Key** provided in the previous step. Add the Secret. 

    ![Add Secret](../assets/add-secret-aws.png)


6. Go to **Members** > **Service Accounts**. Click **+** to add a new service account. 

    ![Add Service Account](../assets/add-service-account.png)

7. Download the `kubeconfig` file for this service account. 

    ![Download kubeconfig](../assets/download-kubeconfig.png)

### Run the example

1. To provision a new cluster on AWS, go to the `hydroform` directory and run:

```
go run ./examples/gardener/main.go -p {project_name} -c {/path/to/gardener/kubeconfig} -s {AWS-secret-name}

```
This command will create a new cluster on AWS a specified project 

2. In Gardener, go to **Clusters**. You should see your cluster listed there.