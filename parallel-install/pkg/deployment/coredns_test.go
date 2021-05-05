package deployment

import (
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/require"
	v1apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_patchCoreDNS(t *testing.T) {
	log := logger.NewLogger(true)
	coreDNSConfigMap := fakeCoreDNSConfigMap()
	coreDNSDeployment := fakeCoreDNSDeployment()
	domain := `(.*)\.local\.kyma\.dev`

	t.Run("test skipping coreDNS patch when coreDNS deployment doesn't exist", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset(coreDNSConfigMap)

		// when
		cm, err := patchCoreDNS(kubeClient, domain, log)

		// then
		require.NoError(t, err)
		require.Empty(t, cm.Data)
	})

	t.Run("test skipping coreDNS patch when coreDNS configMap has proper entry", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset(coreDNSConfigMap, coreDNSDeployment)

		// when
		cm, err := patchCoreDNS(kubeClient, domain, log)

		// then
		require.NoError(t, err)
		require.Empty(t, cm.Data)
	})

	t.Run("test patching coreDNS configMap when coreDNS configMap is empty", func(t *testing.T) {
		// given
		emptyConfigMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "coredns",
				Namespace: "kube-system",
			},
			Data: make(map[string]string),
		}
		kubeClient := fake.NewSimpleClientset(coreDNSDeployment, emptyConfigMap)

		// when
		cm, err := patchCoreDNS(kubeClient, domain, log)

		// then
		require.NoError(t, err)
		require.Contains(t, cm.Data["Corefile"], domain)
	})

	t.Run("test patching coreDNS configMap when coreDNS configMap does not contain proper domain", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset(coreDNSDeployment, fakeWrongCoreDNSConfigMap())

		// when
		cm, err := patchCoreDNS(kubeClient, domain, log)

		// then
		require.NoError(t, err)
		require.Contains(t, cm.Data["Corefile"], domain)
	})
}

func fakeCoreDNSConfigMap() *v1.ConfigMap {
	domainData := make(map[string]string)
	domainData["Corefile"] = `
	.:53 {
		errors
		health
		rewrite name regex (.*)\.local\.kyma\.dev istio-ingressgateway.istio-system.svc.cluster.local
		ready
	}
	`

	coreDNSCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: "kube-system",
		},
		Data: domainData,
	}

	return coreDNSCM
}

func fakeWrongCoreDNSConfigMap() *v1.ConfigMap {
	domainData := make(map[string]string)
	domainData["Corefile"] = `
	.:53 {
		errors
		health
		rewrite name regex (.*)\.local\.kymaaa\.dev istio-ingressgateway.istio-system.svc.cluster.local
	}
	`

	coreDNSCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: "kube-system",
		},
		Data: domainData,
	}

	return coreDNSCM
}

func fakeCoreDNSDeployment() *v1apps.Deployment {
	coreDNSDeployment := &v1apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: "kube-system",
		},
	}

	return coreDNSDeployment
}
