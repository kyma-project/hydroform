package installation

import (
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/stretchr/testify/assert"
)

type logger struct{}

func (l logger) Infof(format string, a ...interface{}) {}

func Test_Options(t *testing.T) {

	// given
	crModFunc := func(installation *v1alpha1.Installation) {}
	logger := logger{}

	installationOptions := &installationOptions{
		installationCRModificationFunc: func(_ *v1alpha1.Installation) { t.Fatal("invalid function called") },
	}

	// when
	optionsFuncs := []InstallationOption{
		WithInstallationCRModification(crModFunc),
		WithLogger(logger),
	}

	for _, f := range optionsFuncs {
		f.apply(installationOptions)
	}

	// then
	assert.Equal(t, logger, installationOptions.logger)
	installationOptions.installationCRModificationFunc(nil)
}
