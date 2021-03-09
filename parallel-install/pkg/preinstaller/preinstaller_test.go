package preinstaller

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller/mocks"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	"os"
	"path"
	"regexp"
	"testing"
)

func TestPreInstaller_InstallCRDs(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions()

	t.Run("should install CRDs", func(t *testing.T) {
		// given
		resourceParserMock := mocks.AllowResourceParserMock{}
		resourceApplierMock := mocks.AllowResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		input := resourceInfoInput{
			resourceType:             "CustomResourceDefinition",
			dirSuffix:                "crds",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, len(output.Installed), 2)
		assert.Zero(t, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, expectedSecondPath))
	})

	t.Run("should not install CRDs due to incorrect input resource type", func(t *testing.T) {
		// given
		resourceParserMock := mocks.AllowResourceParserMock{}
		resourceApplierMock := mocks.AllowResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		input := resourceInfoInput{
			resourceType:             "typeDifferentThanCrd",
			dirSuffix:                "crds",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Zero(t, len(output.Installed))
		assert.Equal(t, len(output.NotInstalled), 2)

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedSecondComponent, expectedSecondPath))
	})
}

func TestPreInstaller_CreateNamespaces(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions()

	t.Run("should create namespaces", func(t *testing.T) {
		// given
		resourceParserMock := mocks.AllowResourceParserMock{}
		resourceApplierMock := mocks.AllowResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		input := resourceInfoInput{
			resourceType:             "typeDifferentThanNamespace",
			dirSuffix:                "namespaces",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Zero(t, len(output.Installed))
		assert.Equal(t, len(output.NotInstalled), 2)

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp1/ns.yaml")
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp2/ns.yaml")
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedSecondComponent, expectedSecondPath))
	})

	t.Run("should not create namespaces due to incorrect input resource type", func(t *testing.T) {
		// given
		resourceParserMock := mocks.AllowResourceParserMock{}
		resourceApplierMock := mocks.AllowResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		input := resourceInfoInput{
			resourceType:             "Namespace",
			dirSuffix:                "namespaces",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, len(output.Installed), 2)
		assert.Zero(t, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp1/ns.yaml")
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp2/ns.yaml")
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, expectedSecondPath))
	})
}

func TestPreInstaller_install(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions()

	t.Run("should install resources", func(t *testing.T) {
		// given
		resourceParserMock := mocks.AllowResourceParserMock{}
		resourceApplierMock := mocks.AllowResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		input := resourceInfoInput{
			resourceType:             "CustomResourceDefinition",
			dirSuffix:                "crds",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, len(output.Installed), 2)
		assert.Zero(t, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, expectedSecondPath))
	})

	t.Run("should partially install resources due to incorrect resource format and resource type different than input info", func(t *testing.T) {
		// given
		resourceParserMock := mocks.MixedResourceParserMock{}
		resourceApplierMock := mocks.MixedResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/partiallycorrect")
		input := resourceInfoInput{
			resourceType:             "CustomResourceDefinition",
			dirSuffix:                "crds",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, len(output.Installed), 2)
		assert.Equal(t, len(output.NotInstalled), 3)

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, expectedSecondPath))

		expectedThirdComponent := "comp3"
		expectedThirdPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp3/crd-incorrect.yaml")
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedThirdComponent, expectedThirdPath))

		expectedFourthComponent := "comp4"
		expectedFourthPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp4/ns.yaml")
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFourthComponent, expectedFourthPath))

		expectedFifthComponent := "comp5"
		expectedFifthPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp5/ns-incorrect.yaml")
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFifthComponent, expectedFifthPath))
	})

	t.Run("should fail to install resources", func(t *testing.T) {
		t.Run("due to error about not existing installation resources path", func(t *testing.T) {
			// given
			resourceParserMock := mocks.MixedResourceParserMock{}
			resourceApplierMock := mocks.MixedResourceApplierMock{}
			i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
			input := resourceInfoInput{
				resourceType:             "CustomResourceDefinition",
				dirSuffix:                "crds",
				installationResourcePath: "notExistingPath",
			}

			// when
			output, err := i.install(input)

			// then
			assert.Error(t, err)
			expectedError := "no such file or directory"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
			assert.Zero(t, len(output.Installed))
		})

		t.Run("due to no components detected in installation resources path", func(t *testing.T) {
			// given
			resourceParserMock := mocks.MixedResourceParserMock{}
			resourceApplierMock := mocks.MixedResourceApplierMock{}
			i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/nocomponents")
			input := resourceInfoInput{
				resourceType:             "CustomResourceDefinition",
				dirSuffix:                "crds",
				installationResourcePath: resourcePath,
			}

			// when
			output, err := i.install(input)

			// then
			assert.NoError(t, err)
			assert.Zero(t, len(output.Installed))
			assert.Zero(t, len(output.NotInstalled))
		})

		t.Run("due to parser error", func(t *testing.T) {
			// given
			resourceParserMock := mocks.DenyResourceParserMock{}
			resourceApplierMock := mocks.DenyResourceApplierMock{}
			i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/incorrect")
			input := resourceInfoInput{
				resourceType:             "CustomResourceDefinition",
				dirSuffix:                "crds",
				installationResourcePath: resourcePath,
			}

			// when
			output, err := i.install(input)

			// then
			assert.NoError(t, err)
			assert.Zero(t, len(output.Installed))
			assert.Equal(t, len(output.NotInstalled), 2)

			expectedFirstComponent := "comp1"
			expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
			assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFirstComponent, expectedFirstPath))

			expectedSecondComponent := "comp2"
			expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
			assert.True(t, containsFileWithDetails(output.NotInstalled, expectedSecondComponent, expectedSecondPath))
		})

		t.Run("due to applier error", func(t *testing.T) {
			// given
			resourceParserMock := mocks.AllowResourceParserMock{}
			resourceApplierMock := mocks.DenyResourceApplierMock{}
			i := NewPreInstaller(&resourceApplierMock, &resourceParserMock, cfg, dynamicClient, retryOptions)
			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/incorrect")
			input := resourceInfoInput{
				resourceType:             "CustomResourceDefinition",
				dirSuffix:                "crds",
				installationResourcePath: resourcePath,
			}

			// when
			output, err := i.install(input)

			// then
			assert.NoError(t, err)
			assert.Zero(t, len(output.Installed))
			assert.Equal(t, len(output.NotInstalled), 2)

			expectedFirstComponent := "comp1"
			expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
			assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFirstComponent, expectedFirstPath))

			expectedSecondComponent := "comp2"
			expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
			assert.True(t, containsFileWithDetails(output.NotInstalled, expectedSecondComponent, expectedSecondPath))
		})
	})
}

func getTestingConfig() Config {
	return Config{
		Log:                      logger.NewLogger(true),
		InstallationResourcePath: "installationResourcePath",
	}
}

func getTestingRetryOptions() []retry.Option {
	return []retry.Option{
		retry.Delay(0),
		retry.Attempts(1),
		retry.DelayType(retry.FixedDelay),
	}
}

func getTestingResourcesDirectory() string {
	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	return path.Join(currentDir, "/../test/data/resources")
}

func containsFileWithDetails(files []File, component string, path string) bool {
	for _, file := range files {
		if file.component == component && file.path == path {
			return true
		}
	}

	return false
}
