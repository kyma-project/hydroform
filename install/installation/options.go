package installation

import (
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
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
