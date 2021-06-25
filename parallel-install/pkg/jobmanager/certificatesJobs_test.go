package jobmanager

import (
	"context"
	"fmt"
	"testing"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/require"
	v1alpha1 "istio.io/api/meta/v1alpha1"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	v1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	versioned "istio.io/client-go/pkg/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCertificatesJobs(t *testing.T) {
	t.Run("should annotate Gateway", func(t *testing.T) {
		resetFinishedJobsMap()
		SetLogger(logger.NewLogger(false))

		namespace := "kyma-system"
		gateway := "kyma-gateway"
		newAnnotations := map[string]string{
			"meta.helm.sh/release-name":      "certificates",
			"meta.helm.sh/release-namespace": "istio-system",
		}

		kubeClient := fake.NewSimpleClientset()
		ic := versioned.NewSimpleClientset()

		sampleGw := &v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gateway,
				Namespace: namespace,
			},
			Status:   v1alpha1.IstioStatus{},
			TypeMeta: metav1.TypeMeta{},
			Spec:     networkingv1beta1.Gateway{},
		}
		_, err := ic.NetworkingV1beta1().Gateways(namespace).Create(context.TODO(), sampleGw, metav1.CreateOptions{})
		if err != nil {
			t.Logf("Error creating gatewat: %s", err)
		}

		config := &installConfig.Config{
			WorkersCount: 1,
		}
		err = annotateCertificatesGateway{}.execute(config, kubeClient, ic, context.TODO())
		if err != nil {
			fmt.Println(err)
		}
		gw, err := ic.NetworkingV1beta1().Gateways(namespace).Get(context.TODO(), gateway, metav1.GetOptions{})
		fmt.Println("AFTER GW GET")
		if err != nil {
			fmt.Println(err)
		}
		annomap := gw.GetAnnotations()
		require.Equal(t, newAnnotations, annomap)
	})

}
