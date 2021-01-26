//Package prerequisites implements the logic preparing the cluster for Kyma installation.
//It also contains the code to clean up the prerequisites.
//
//The code in the package uses the user-provided function for logging.
package prerequisites

import (
	"context"
	"fmt"
	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/components"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Prerequisites supports installation / uninstallation of Kyma prerequisites
type Prerequisites struct {
	Context       context.Context
	KubeClient    kubernetes.Interface
	Prerequisites []components.KymaComponent
	Log           *zap.SugaredLogger
}

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
func (p *Prerequisites) InstallPrerequisites() <-chan error {

	statusChan := make(chan error)

	go func() {
		defer close(statusChan)

		p.Log.Info("Deploying kyma-installer namespace")

		_, err := p.KubeClient.CoreV1().Namespaces().Get(context.Background(), "kyma-installer", metav1.GetOptions{})

		if err != nil {
			if errors.IsNotFound(err) {
				nsErr := p.createNamespace()
				if nsErr != nil {
					statusChan <- fmt.Errorf("Unable to create kyma-installer namespace. Error: %v", nsErr)
					return
				}
			} else {
				statusChan <- fmt.Errorf("Unable to get kyma-installer namespace. Error: %v", err)
				return
			}
		} else {
			nsErr := p.updateNamespace()
			if nsErr != nil {
				statusChan <- fmt.Errorf("Unable to update kyma-installer namespace. Error: %v", nsErr)
				return
			}
		}

		for _, prerequisite := range p.Prerequisites {
			//TODO: Is there a better way to find out if Context is canceled?
			if p.Context.Err() != nil {
				//Context is canceled or timed-out. Skip processing
				p.Log.Infof("%s Finishing work: %v", logPrefix, p.Context.Err())
				return //TODO: Consider returning information about "processing skipped because of timeout" via statusChan
			}

			p.Log.Infof("%s Installing component %s ", logPrefix, prerequisite.Name)
			//installation step
			err := prerequisite.Deploy(p.Context)
			if err != nil {
				p.Log.Infof("%s Error installing prerequisite %s: %v (The installation will not continue)", logPrefix, prerequisite.Name, err)
				statusChan <- err
				return
			}
			statusChan <- nil //TODO: Is this necessary?
		}
	}()

	return statusChan
}

func (p *Prerequisites) createNamespace() error {
	_, err := p.KubeClient.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kyma-installer",
			Labels: map[string]string{"istio-injection": "disabled", "kyma-project.io/installation": ""},
		},
	}, metav1.CreateOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (p *Prerequisites) updateNamespace() error {
	_, err := p.KubeClient.CoreV1().Namespaces().Update(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kyma-installer",
			Labels: map[string]string{"istio-injection": "disabled", "kyma-project.io/installation": ""},
		},
	}, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	return nil
}

//UninstallPrerequisites tries to uninstall all provided prerequisites.
//The function does not quit on errors - it tries to uninstall everything.
//
//The function supports the Context cancellation.
//The cancellation is not immediate.
//If the "cancel" signal appears during the uninstallation step (it's a blocking operation),
//such a "cancel" condition is detected only after that step is over, and UninstallPrerequisites returns without an error.
func (p *Prerequisites) UninstallPrerequisites() <-chan error {

	statusChan := make(chan error)

	go func() {
		defer close(statusChan)

		for i := len(p.Prerequisites) - 1; i >= 0; i-- {
			prereq := p.Prerequisites[i]
			//TODO: Is there a better way to find out if Context is canceled?
			if p.Context.Err() != nil {
				//Context is canceled or timed-out. Skip processing
				prereq.Log.Errorf("%s Finishing work: %v", logPrefix, p.Context.Err())
				return //TODO: Consider returning information about "processing skipped because of timeout" via statusChan
			}

			prereq.Log.Infof("%s Uninstalling component %s ", logPrefix, prereq.Name)
			//uninstallation step
			err := prereq.Uninstall(p.Context)
			if err != nil {
				prereq.Log.Errorf("%s Error uninstalling prerequisite %s: %v (The uninstallation continues anyway)", logPrefix, prereq.Name, err)
				statusChan <- err
			}
			statusChan <- nil //TODO: Is this necessary?
		}

		// TODO: Delete namespace deletion once xip-patch is gone.
		p.Log.Info("Deleting kyma-installer namespace")
		err := p.KubeClient.CoreV1().Namespaces().Delete(context.Background(), "kyma-installer", metav1.DeleteOptions{})
		if err != nil {
			statusChan <- fmt.Errorf("Unable to delete kyma-installer namespace. Error: %v", err)
			return
		}
	}()

	return statusChan
}
