package installation

import (
	"github.com/kyma-incubator/hydroform/install/k8s"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type InstallationOption interface {
	apply(*installationOptions)
}

type optionFunc func(*installationOptions)

func (f optionFunc) apply(o *installationOptions) {
	f(o)
}

type installationOptions struct {
	logger                         Logger
	installationCRModificationFunc func(installation *v1alpha1.Installation)
}

func WithInstallationCRModification(modFunc func(installation *v1alpha1.Installation)) InstallationOption {
	return optionFunc(func(o *installationOptions) {
		o.installationCRModificationFunc = modFunc
	})
}

func WithLogger(logger Logger) InstallationOption {
	return optionFunc(func(o *installationOptions) {
		o.logger = logger
	})
}

func DefaultDecoder() (runtime.Decoder, error) {
	resourceScheme, err := k8s.DefaultScheme()
	if err != nil {
		return nil, err
	}
	codecs := serializer.NewCodecFactory(resourceScheme)
	decoder := codecs.UniversalDeserializer()

	return decoder, nil
}
