package preinstaller

import (
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type resourceType struct {
	name      string
	decoder runtime.Decoder
	validator func(data []byte, decoder runtime.Decoder) bool
	applier func(data []byte, kubeClient kubernetes.Interface) error
}

func newCrdPreInstallerResource() *resourceType {
	return &resourceType{
		name: "crds",
		decoder: initializeDecoder(),
		validator: func(data []byte, decoder runtime.Decoder) bool {
			_, _, err := decoder.Decode(data, nil, nil)
			if err != nil {
				return false
			}

			return true
		},
		applier: func(data []byte, kubeClient kubernetes.Interface) error {
			// TODO

			return nil
		},
	}
}

func newNamespacePreInstallerResource() *resourceType {
	return &resourceType{
		name: "namespaces",
		decoder: initializeDecoder(),
		validator: func(data []byte, decoder runtime.Decoder) bool {
			_, _, err := decoder.Decode(data, nil, nil)
			if err != nil {
				return false
			}

			return true
		},
		applier: func(data []byte, kubeClient kubernetes.Interface) error {
			// TODO

			return nil
		},
	}
}

func initializeDecoder() runtime.Decoder {
	sch := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(sch)
	_ = apiextv1.AddToScheme(sch)

	return serializer.NewCodecFactory(sch).UniversalDeserializer()
}
