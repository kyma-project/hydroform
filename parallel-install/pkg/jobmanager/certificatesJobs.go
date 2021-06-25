package jobmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"k8s.io/client-go/kubernetes"

	istio "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Register job using implemented interface

type annotateCertificatesGateway struct{}

// No deprecation planned; 06/2021
var _ = register(annotateCertificatesGateway{})

func (j annotateCertificatesGateway) when() (component, executionTime) {
	return component("certificates"), Pre
}

func (j annotateCertificatesGateway) identify() jobName {
	return jobName("annotateCertificatesGateway")
}

// This job increases the PVC-size of the logging component to 30GB.
// This will be triggered before the deployment of its corresponding component.
func (j annotateCertificatesGateway) execute(cfg *config.Config, kubeClient kubernetes.Interface, ic istio.Interface, ctx context.Context) error {

	namespace := "kyma-system"
	gateway := "kyma-gateway"
	newAnnotations := map[string]string{
		"meta.helm.sh/release-name":      "certificates",
		"meta.helm.sh/release-namespace": "istio-system",
	}

	gw, err := ic.NetworkingV1beta1().Gateways(namespace).Get(context.TODO(), gateway, metav1.GetOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not fetch Gateway `%s` with error: %s", gateway, err))
	}
	gwAnnotations := gw.GetAnnotations()

	if gwAnnotations == nil {
		gwAnnotations = make(map[string]string)
	}

	for k, v := range newAnnotations {
		gwAnnotations[k] = v
	}

	gw.SetAnnotations(gwAnnotations)

	gw, err = ic.NetworkingV1beta1().Gateways(namespace).Update(context.TODO(), gw, metav1.UpdateOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not fetch Gateway: %s", gateway))
	}

	gwAnnotations = gw.GetAnnotations()
	for k, v := range newAnnotations {
		if gwAnnotations[k] != v {
			return errors.New(fmt.Sprintf("Annotation was not applied: %s: %s", k, v))
		}
	}
	return nil
}
