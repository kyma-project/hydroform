//Package namespace implements the logic for deploying/deleting Kyma installer namespace.
//
//The code in the package uses the user-provided function for logging.
package namespace

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Namespace supports deploying/deleting Kyma installer namespace
type Namespace struct {
	KubeClient kubernetes.Interface
	Log        logger.Interface
}

func (ns *Namespace) DeployInstallerNamespace() error {
	ns.Log.Info("Deploying kyma-installer namespace")

	_, err := ns.KubeClient.CoreV1().Namespaces().Get(context.Background(), "kyma-installer", metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			nsErr := ns.createNamespace()
			if nsErr != nil {
				return fmt.Errorf("Unable to create kyma-installer namespace. Error: %v", nsErr)
			}
		} else {
			return fmt.Errorf("Unable to get kyma-installer namespace. Error: %v", err)
		}
	} else {
		nsErr := ns.updateNamespace()
		if nsErr != nil {
			return fmt.Errorf("Unable to update kyma-installer namespace. Error: %v", nsErr)
		}
	}

	return nil
}

func (ns *Namespace) createNamespace() error {
	_, err := ns.KubeClient.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
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

func (ns *Namespace) updateNamespace() error {
	_, err := ns.KubeClient.CoreV1().Namespaces().Update(context.Background(), &v1.Namespace{
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
