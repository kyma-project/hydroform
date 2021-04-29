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
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
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
func patchCoreDNS(kubeClient kubernetes.Interface, domainName string, log logger.Interface) (cm *v1.ConfigMap, err error) {
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

		configMaps := kubeClient.CoreV1().ConfigMaps("kube-system")
		coreDNSConfigMap, exists, err := findCoreDNSConfigMap(configMaps, log)
		if err != nil {
			return err
		}

		if strings.Contains(coreDNSConfigMap.Data["Corefile"], domainName) {
			log.Info("CoreDNS config already contains proper domain rule")
			return nil
		}

		coreFile, err := generateCorefile(domainName)
		if err != nil {
			return err
		}
		if exists {
			log.Info("Patching CoreDNS config")
			cm, err = patchCoreDNSConfigMap(configMaps, coreDNSConfigMap, coreFile, log)
			if err != nil {
				return err
			}
		} else {
			log.Info("Corefile not found, creating new CoreDNS config")
			cm, err = createCoreDNSConfigMap(configMaps, coreFile, log)
			if err != nil {
				return err
			}
		}

		return nil
	}, retryOptions...)

	if err != nil {
		return cm, err
	}

	return cm, nil
}

func findCoreDNSConfigMap(configMaps corev1.ConfigMapInterface, log logger.Interface) (cm *v1.ConfigMap, exists bool, err error) {
	cm, err = configMaps.Get(context.TODO(), "coredns", metav1.GetOptions{})
	if err != nil {
		if apierr.IsNotFound(err) {
			return cm, false, nil
		}
		return cm, exists, err
	}
	return cm, true, nil
}

func patchCoreDNSConfigMap(configMaps corev1.ConfigMapInterface, coreDNSConfigMap *v1.ConfigMap, coreFile string, log logger.Interface) (cm *v1.ConfigMap, err error) {
	coreDNSConfigMap.Data["Corefile"] = coreFile
	jsontext, err := json.Marshal(coreDNSConfigMap)
	if err != nil {
		return cm, err
	}

	cm, err = configMaps.Patch(context.TODO(), "coredns", types.StrategicMergePatchType, jsontext, metav1.PatchOptions{})
	if err != nil {
		return cm, err
	}
	return
}

func createCoreDNSConfigMap(configMaps corev1.ConfigMapInterface, coreFile string, log logger.Interface) (cm *v1.ConfigMap, err error) {
	cm, err = configMaps.Create(context.TODO(), getNewCoreDNSConfigMap(coreFile), metav1.CreateOptions{})
	if err != nil {
		log.Error("Could not create new CoreDNS Corefile config")
		return cm, err
	}
	return
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
