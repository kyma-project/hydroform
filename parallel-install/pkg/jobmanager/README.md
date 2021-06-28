# `jobManager` Package

The `jobManager` implements a `job` interface, which enables a clean automated Kyma deployment in regards of configuring the cluster and the components, resources and other features of Kyma. The terms "Deployment" and "Deploy" are used in the context of installing Kyma on an empty cluster, or to upgrade Kyma from an older to a newer version.

## Workflow / General mechanism

### Features of this package

- Supports only single linear upgrade: A &#8594; B && B &#8594; C; NOT A &#8594; C. This is due to the fact that Kyma only supports single linear upgrades.
- Jobs themselves cover the checks to determine if they should be triggered. &#8594; They should be triggered if the cluster is not in the wanted state.
- Announce deprecation of jobs, for example, as in-line comment.
- The jobManager only supports Kyma `deploy` and not `delete`, to prevent that developers misuse jobs to clean up dirty left-overs from `kyma alpha delete`.
- When the deploy of Kyma fails, the global post-jobs do not run.
- When the deploy of a component fails, the component-based post-jobs do not run.
- Jobs do run async to each other.
- CancelContext is propagated to give developers the opportunity to cancel the deploy.


- This mechanism supports jobs for two different use cases: The __component-based__ jobs and the __global/component-independent__ jobs
  - __Component-based__:
    - Check whether the component is installed on the cluster or must be newly installed; and only trigger if it must be installed.
    - It is possible to trigger jobs before and after a deployment of a component.
  - __Global / Component-independent__:
    - Always trigger the job when installing or upgrading Kyma.
    - It is to trigger jobs before and after the deployment of Kyma.
    - Call component-independent jobs `global` jobs to stick to the naming convention of our helm charts.

### How does it work?

This package registers, manages, and triggers certain jobs to have a fully-automated installation or migration. This package has two (hash)maps to manage the workload: One for `pre`-jobs and one for `post`-jobs. In the (hash)maps, the key is the name of the component the jobs belong to, and the value is a slice of the jobs.

Furthermore, the `jobManager` package has a `duration` variable for benchmarking, which is fetched at the end of a Kyma deployment.

Jobs are implemented within the `jobManager` package in `go`-files, one for each component, using the specific `job` interface. Then, the implemented interface is registered using `register(job)` in the same file.
To implement the `job` interface, the newly created jobs must implement the `execute(*config.Config, kubernetes.Interface)` function, which takes the installation config and a kubernetes interface as input, so that the jobs can interact with the cluster. The return value must be an error. Additionally, the `when()` function must be implemented, which returns the component the job is bound to and whether it should run pre or post the deployment. The `identify()` function also needs to be implemented to have a unique identifier for each job. If the active solution for tagging jobs as deprecated is chosen, then the `deprecate` function also must be implemented - more in the next section.

The jobManager is used by the `deployment` package and in the `engine` package . At the hooks, during the deployment phase, each hook only has to check if the key for the wanted component is present in the pre- or post-map. If it's present, the jobs in the map are trigged, if not, nothing must be done.

Retries for the jobs are not handled by the jobManager. Retries should be implemented by the jobs themselves, because it's more flexible and the interface is easy to manage. Also, the check if the logic of the job should be executed stays inside of the job, and is not implemented by the jobManager.

### Deprecation of Jobs

At this point, we stick to the passiv solution; we may switch to the active solution in the future. The deprecation version must be documented; alternatively, the fact that there is no specific deprecation version or that deprecation is unnecessary.

Passive:
- Tag (as a comment above the register call) at which defined Kyma version the jobs should be deprecated.

Active:
- Go-Build-Tags cannot be used for this use case.
- Add `deprecation` function to job-interface, which returns at which Kyma version the job should be deprecated. Before the job is executed, the deprecation function is called to check whether it is already deprecated. If deprecated, an Error should be thrown to block the CI.

### Workflow of the jobManager
<img src="./pictures/migration-logic-diagram.png?raw=true">


## Trade-Offs

The jobs of this package cannot handle every situation that may come up in the cluster, because we do not know what the setup/usage of the customer's cluster looks like - for example, which provisioner is used, and especially regarding the access rights of the user who is deploying Kyma. Thus, an additional migration guide will be needed in the future, as before. Let us demonstrate this on the `increaseLoggingPvcSize job`:
- The option `allowVolumeExpansion` of the respective `PVC` must be set to `true`. If it's not, it must be changed. To do this, the provided kubeconfig must have admin rights, and the hypervisor must allow it:
   - __k3d:__ Using a local cluster to deploy Kyma on the sample job fails, since k3d is missing a plugin to expand existing volumes. 

      ```
      Ignoring the PVC: didn't find a plugin capable of expanding the volume; waiting for an external controller to process this PVC.
      ```

   - __Azure:__  When Kyma is to be deployed on an Azure cluster, the disk expand is only allowed on an unattached disk. Due to some GitHub Issues, this feature will be added in the future.

      ```console
      error expanding volume "kyma-system/storage-logging-loki-0" of plugin "kubernetes.io/azure-disk": azureDisk - disk resize is only supported on Unattached disk, current disk state: Attached, already attached to /subscriptions/68266e60-bb03-40e0-935d-531fac39f8c1/resourceGroups/shoot--berlin--jh-02/providers/Microsoft.Compute/virtualMachines/shoot--berlin--jh-02-worker-jz1n6-z1-6d9c5-cvn2b
      ```

As before, during the deploy upgrading the cluster, the user gets the necessary information to make sure the deploy of Kyma works for them. In other cases, the jobManager works as a Go-based solution instead of using certain Helm features or shell scripts.

## Additions
[Original PoC](https://github.com/kyma-project/community/blob/main/internal/proposals/migration-hooks/migration-hooks-proposal.md)
