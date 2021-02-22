package preinstaller

import (
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type resourceType struct {
	name      string
	applier resourceApplier
}

func newCrdPreInstallerResource(log logger.Interface, dynamicClient dynamic.Interface, retryOptions []retry.Option) *resourceType {
	return &resourceType{
		name: "crds",
		applier: newGenericResourceApplier(log, dynamicClient, initializeDecoder(), retryOptions),
	}
}

func newNamespacePreInstallerResource(log logger.Interface, dynamicClient dynamic.Interface, retryOptions []retry.Option) *resourceType {
	return &resourceType{
		name: "namespaces",
		applier: newGenericResourceApplier(log, dynamicClient, initializeDecoder(), retryOptions),
	}
}

func initializeDecoder() runtime.Decoder {
	sch := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(sch)
	_ = apiextv1.AddToScheme(sch)

	return serializer.NewCodecFactory(sch).UniversalDeserializer()
}
