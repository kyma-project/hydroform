package deployment

import (
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_patchCoreDNS(t *testing.T) {
	log := logger.NewLogger(true)
	coreDNSCM := fakeCoreDNSCM()

	t.Run("test skipping coreDNS patch", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset(coreDNSCM)

		// when
		cm, err := patchCoreDNS(kubeClient, `(.*)\.loacal\.kyma\.dev`, log)

		// then
		require.NoError(t, err)
		require.Empty(t, cm.Data)
	})
}

func fakeCoreDNSCM() *v1.ConfigMap {
	domainData := make(map[string]string)
	domainData["Corefile"] = `
	.:53 {
		errors
		health
		rewrite name regex (.*)\.local\.kyma\.dev istio-ingressgateway.istio-system.svc.cluster.local
		ready
		kubernetes cluster.local in-addr.arpa ip6.arpa {
		  pods insecure
		  fallthrough in-addr.arpa ip6.arpa
		}
		hosts /etc/coredns/NodeHosts {
		  reload 1s
		  fallthrough
		}
		prometheus :9153
		forward . /etc/resolv.conf
		cache 30
		loop
		reload
		loadbalance
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
