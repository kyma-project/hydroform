package preinstaller

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	"os"
	"path"
	"regexp"
	"testing"
	"time"
)

//func TestPreInstaller_InstallCRDs(t *testing.T) {
//	scheme := runtime.NewScheme()
//	dynamicClient := fake.NewSimpleDynamicClient(scheme)
//	cfg := getTestingConfig()
//	retryOptions := getTestingRetryOptions(cfg)
//	resourceManager := NewResourceManager(dynamicClient, retryOptions)
//	resourceApplier := NewGenericResourceApplier(cfg.Log, *resourceManager)
//
//	t.Run("should install CRDs", func(t *testing.T) {
//		// TODO
//	})
//}

//func TestPreInstaller_CreateNamespaces(t *testing.T) {
//	scheme := runtime.NewScheme()
//	dynamicClient := fake.NewSimpleDynamicClient(scheme)
//	cfg := getTestingConfig()
//	retryOptions := getTestingRetryOptions(cfg)
//	resourceManager := NewResourceManager(dynamicClient, retryOptions)
//	resourceApplier := NewGenericResourceApplier(cfg.Log, *resourceManager)
//
//	t.Run("should create namespaces", func(t *testing.T) {
//		// TODO
//	})
//}

func TestPreInstaller_apply(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions(cfg)
	resourceManager := NewResourceManager(dynamicClient, retryOptions)
	resourceApplier := NewGenericResourceApplier(cfg.Log, *resourceManager)

	t.Run("should fail to create a resource", func(t *testing.T) {
		t.Run("due to error about not existing installation resources path", func(t *testing.T) {
			// given
			cfg.InstallationResourcePath = "notExistingPath"
			i := NewPreInstaller(resourceApplier, cfg, dynamicClient, retryOptions)
			resources := resourceType{name: "crds"}

			// when
			output, err := i.apply(resources)

			// then
			expectedError := "no such file or directory"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			assert.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
			assert.True(t, len(output.installed) == 0)
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
