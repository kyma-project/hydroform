//Package prerequisites implements logic preparing the cluster for Kyma installation.
//It also contains the code to clean up the prerequisites.
//
//The code in the package uses the user-provided function for logging.
package prerequisites

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const logPrefix = "[prerequisites/prerequisites.go]"

//InstallPrerequisites tries to install all provided prerequisites.
//The function quits on the first encountered error because all prerequisites must be installed in order to start Kyma installation.
//
//The function supports the Context cancellation.
//The cancellation is not immediate.
//If the "cancel" signal appears during the installation step (it's a blocking operation),
//such a "cancel" condition is detected only after the step is over, and InstallPrerequisites returns without an error.
//
//prerequisites provide information about all Components that are considered prerequisites for Kyma installation.
//Such components are installed sequentially, in the same order as in the provided slice.
func InstallPrerequisites(ctx context.Context, kubeClient kubernetes.Interface, prerequisites []components.Component) <-chan error {

	statusChan := make(chan error)

	go func() {
		defer close(statusChan)

		config.Log("Creating kyma-installer namespace")
		_, err := kubeClient.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "kyma-installer",
				Labels: map[string]string{"istio-injection": "disabled", "kyma-project.io/installation": ""},
			},
		}, metav1.CreateOptions{})

		if err != nil {
			statusChan <- fmt.Errorf("Unable to create kyma-installer namespace. Error: %v", err)
			return
		}

		for _, prerequisite := range prerequisites {
			//TODO: Is there a better way to find out if Context is canceled?
			if ctx.Err() != nil {
				//Context is canceled or timed-out. Skip processing
				config.Log("%s Finishing work: %v", logPrefix, ctx.Err())
				return //TODO: Consider returning information about "processing skipped because of timeout" via statusChan
			}

			config.Log("%s Installing component %s ", logPrefix, prerequisite.Name)
			//installation step
			err := prerequisite.InstallComponent(ctx)
			if err != nil {
				config.Log("%s Error installing prerequisite %s: %v (The installation will not continue)", logPrefix, prerequisite.Name, err)
				statusChan <- err
				return
			}
			statusChan <- nil //TODO: Is this necessary?
		}
	}()

	return statusChan
}

//UninstallPrerequisites tries to uninstall all provided prerequisites.
//The function does not quit on errors - it tries to uninstall everything.
//
//The function supports the Context cancellation.
//The cancellation is not immediate.
//If the "cancel" signal appears during the uninstallation step (it's a blocking operation),
//such a "cancel" condition is detected only after that step is over, and UninstallPrerequisites returns without an error.
func UninstallPrerequisites(ctx context.Context, kubeClient kubernetes.Interface, prerequisites []components.Component) <-chan error {

	statusChan := make(chan error)

	go func() {
		defer close(statusChan)

		for i := len(prerequisites) - 1; i >= 0; i-- {
			prereq := prerequisites[i]
			//TODO: Is there a better way to find out if Context is canceled?
			if ctx.Err() != nil {
				//Context is canceled or timed-out. Skip processing
				config.Log("%s Finishing work: %v", logPrefix, ctx.Err())
				return //TODO: Consider returning information about "processing skipped because of timeout" via statusChan
			}

			config.Log("%s Uninstalling component %s ", logPrefix, prereq.Name)
			//uninstallation step
			err := prereq.UninstallComponent(ctx)
			if err != nil {
				config.Log("%s Error uninstalling prerequisite %s: %v (The uninstallation continues anyway)", logPrefix, prereq.Name, err)
				statusChan <- err
			}
			statusChan <- nil //TODO: Is this necessary?
		}

		// TODO: Delete namespace deletion once xip-patch is gone.
		config.Log("Deleting kyma-installer namespace")
		err := kubeClient.CoreV1().Namespaces().Delete(context.Background(), "kyma-installer", metav1.DeleteOptions{})
		if err != nil {
			statusChan <- fmt.Errorf("Unable to delete kyma-installer namespace. Error: %v", err)
			return
		}
	}()

	return statusChan
}
