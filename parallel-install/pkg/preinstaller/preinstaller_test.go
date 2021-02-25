package preinstaller

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller/mocks"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	"os"
	"path"
	"regexp"
	"testing"
	"time"
)

func TestPreInstaller_InstallCRDs(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions(cfg)

	t.Run("should install CRDs", func(t *testing.T) {
		// given
		resourceApplierMock := mocks.AllowResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, cfg, dynamicClient, retryOptions)
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		input := resourceInfoInput{
			resourceType:             "CRD",
			dirSuffix:                "crds",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.True(t, len(output.installed) == 2)
		assert.True(t, len(output.notInstalled) == 0)

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		assert.True(t, containsFileWithDetails(output.installed, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		assert.True(t, containsFileWithDetails(output.installed, expectedSecondComponent, expectedSecondPath))
	})
}

func TestPreInstaller_CreateNamespaces(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions(cfg)

	t.Run("should create namespaces", func(t *testing.T) {
		// given
		resourceApplierMock := mocks.AllowResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, cfg, dynamicClient, retryOptions)
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
		assert.True(t, len(output.installed) == 2)
		assert.True(t, len(output.notInstalled) == 0)

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp1/ns.yaml")
		assert.True(t, containsFileWithDetails(output.installed, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp2/ns.yaml")
		assert.True(t, containsFileWithDetails(output.installed, expectedSecondComponent, expectedSecondPath))
	})
}

func TestPreInstaller_install(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions(cfg)
	resourceManager := NewDefaultResourceManager(dynamicClient, retryOptions)
	resourceApplier := NewGenericResourceApplier(cfg.Log, resourceManager)

	t.Run("should install resources", func(t *testing.T) {
		// given
		resourceApplierMock := mocks.AllowResourceApplierMock{}
		i := NewPreInstaller(&resourceApplierMock, cfg, dynamicClient, retryOptions)
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		input := resourceInfoInput{
			resourceType:             "CRD",
			dirSuffix:                "crds",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.True(t, len(output.installed) == 2)
		assert.True(t, len(output.notInstalled) == 0)

		expectedFirstComponent := "comp1"
		expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		assert.True(t, containsFileWithDetails(output.installed, expectedFirstComponent, expectedFirstPath))

		expectedSecondComponent := "comp2"
		expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		assert.True(t, containsFileWithDetails(output.installed, expectedSecondComponent, expectedSecondPath))
	})

	t.Run("should fail to install resources", func(t *testing.T) {
		t.Run("due to error about not existing installation resources path", func(t *testing.T) {
			// given
			i := NewPreInstaller(resourceApplier, cfg, dynamicClient, retryOptions)
			input := resourceInfoInput{
				resourceType:             "CRD",
				dirSuffix:                "crds",
				installationResourcePath: "notExistingPath",
			}

			// when
			output, err := i.install(input)

			// then
			expectedError := "no such file or directory"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
			assert.True(t, len(output.installed) == 0)
		})

		t.Run("due to no components detected in installation resources path", func(t *testing.T) {
			// given
			i := NewPreInstaller(resourceApplier, cfg, dynamicClient, retryOptions)
			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/nocomponents")
			input := resourceInfoInput{
				resourceType:             "CRD",
				dirSuffix:                "crds",
				installationResourcePath: resourcePath,
			}

			// when
			output, err := i.install(input)

			// then
			assert.NoError(t, err)
			assert.True(t, len(output.installed) == 0)
			assert.True(t, len(output.notInstalled) == 0)
		})

		t.Run("due to applier error", func(t *testing.T) {
			// given
			resourceApplierMock := mocks.DenyResourceApplierMock{}
			i := NewPreInstaller(&resourceApplierMock, cfg, dynamicClient, retryOptions)
			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/incorrect")
			input := resourceInfoInput{
				resourceType:             "CRD",
				dirSuffix:                "crds",
				installationResourcePath: resourcePath,
			}

			// when
			output, err := i.install(input)

			// then
			assert.NoError(t, err)
			assert.True(t, len(output.installed) == 0)
			assert.True(t, len(output.notInstalled) == 2)

			expectedFirstComponent := "comp1"
			expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
			assert.True(t, containsFileWithDetails(output.notInstalled, expectedFirstComponent, expectedFirstPath))

			expectedSecondComponent := "comp2"
			expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
			assert.True(t, containsFileWithDetails(output.notInstalled, expectedSecondComponent, expectedSecondPath))

		})
	})
}

func getTestingConfig() config.Config {
	return config.Config{
		BackoffInitialIntervalSeconds: 3,
		BackoffMaxElapsedTimeSeconds:  60 * 5,
		Log:                           logger.NewLogger(true),
		InstallationResourcePath:      "123",
	}
}

func getTestingRetryOptions(cfg config.Config) []retry.Option {
	return []retry.Option{
		retry.Delay(time.Duration(cfg.BackoffInitialIntervalSeconds) * time.Second),
		retry.Attempts(uint(cfg.BackoffMaxElapsedTimeSeconds / cfg.BackoffInitialIntervalSeconds)),
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
