package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"strings"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	v1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

var coreDNSPatchTemplate = `
.:53 {
    errors
    health
    rewrite name regex {{ .DomainName}} istio-ingressgateway.istio-system.svc.cluster.local
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

// CoreDNSPatch contains values to fill template with
type CoreDNSPatch struct {
	DomainName string
}

// patchCoreDNS takes kubeclient and cluster domain as a parameter to patch coredns config
// e.g. domainName: `(.*)\.local\.kyma\.dev`
func patchCoreDNS(kubeClient kubernetes.Interface, domainName string, log logger.Interface) (cm v1.ConfigMap, err error) {
	// TODO: Refactor
	err = retry.Do(func() error {
		_, err := kubeClient.AppsV1().Deployments("kube-system").Get(context.TODO(), "coredns", metav1.GetOptions{})
		if err != nil {
			if apierr.IsNotFound(err) {
				log.Info("CoreDNS not found, skipping CoreDNS config patch")
				return nil
			}
			return err
		}

		coreFile, err := generateCorefile(domainName)
		if err != nil {
			return err
		}
		configMaps := kubeClient.CoreV1().ConfigMaps("kube-system")
		coreDNSConfigMap, err := configMaps.Get(context.TODO(), "coredns", metav1.GetOptions{})
		if err != nil {
			if apierr.IsNotFound(err) {
				log.Info("Corefile not found, creating new CoreDNS config")
				newCM, err := configMaps.Create(context.TODO(), getNewCoreDNSConfigMap(coreFile), metav1.CreateOptions{})
				cm = *newCM
				if err != nil {
					log.Warn("Could not create new CoreDNS Corefile config")
				}
				return nil
			}
			return err
		}

		if strings.Contains(coreDNSConfigMap.Data["Corefile"], domainName) {
			log.Info("CoreDNS config already contains proper domain rule")
			return nil
		}

		coreDNSConfigMap.Data["Corefile"] = coreFile
		jsontext, err := json.Marshal(coreDNSConfigMap)
		if err != nil {
			return err
		}

		log.Info("Patching CoreDNS config")
		_, err = configMaps.Patch(context.TODO(), "coredns", types.StrategicMergePatchType, jsontext, metav1.PatchOptions{DryRun: []string{""}})
		if err != nil {
			return err
		}

		return nil
	}, retryOptions...)

	if err != nil {
		return cm, err
	}

	return cm, nil
}

func getNewCoreDNSConfigMap(data string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "coredns"},
		Data:       map[string]string{"Corefile": data},
	}
}

func generateCorefile(domainName string) (coreFile string, err error) {
	coreDNSPatch := CoreDNSPatch{DomainName: domainName}
	patchTemplate := template.Must(template.New("").Parse(coreDNSPatchTemplate))
	patchBuffer := new(bytes.Buffer)
	if err = patchTemplate.Execute(patchBuffer, coreDNSPatch); err != nil {
		return
	}

	coreFile = patchBuffer.String()
	return
}
