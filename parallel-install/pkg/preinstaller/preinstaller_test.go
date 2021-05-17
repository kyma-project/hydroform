package preinstaller

import (
	"fmt"
	"path"
	"regexp"
	"testing"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller/mocks"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

func TestPreInstaller_InstallCRDs(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions()
	resourceName := "name"
	crdResource := fixCrdResourceWith(resourceName)

	t.Run("should install CRDs", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		cfg.InstallationResourcePath = resourcePath
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

		cfg.InstallationResourcePath = resourcePath
		pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		resourceParser.On("ParseFile", pathToFirstResource).Return(crdResource, nil)
		pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		resourceParser.On("ParseFile", pathToSecondResource).Return(crdResource, nil)

		resourceApplier.On("Apply", crdResource).Return(nil)

		// when
		output, err := i.InstallCRDs()

		// then
		assert.NoError(t, err)
		assert.Equal(t, 2, len(output.Installed))
		assert.Zero(t, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, pathToFirstResource))

		expectedSecondComponent := "comp2"
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, pathToSecondResource))
	})

}

func TestPreInstaller_CreateNamespaces(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions()
	resourceName := "name"
	namespaceResource := fixNamespaceResourceWith(resourceName)

	t.Run("should create namespaces", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		cfg.InstallationResourcePath = resourcePath
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

		pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp1/ns.yaml")
		resourceParser.On("ParseFile", pathToFirstResource).Return(namespaceResource, nil)
		pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp2/ns.yaml")
		resourceParser.On("ParseFile", pathToSecondResource).Return(namespaceResource, nil)

		resourceApplier.On("Apply", namespaceResource).Return(nil)

		// when
		output, err := i.CreateNamespaces()

		// then
		assert.NoError(t, err)
		assert.Equal(t, 2, len(output.Installed))
		assert.Zero(t, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, pathToFirstResource))

		expectedSecondComponent := "comp2"
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, pathToSecondResource))
	})

}

func TestPreInstaller_install(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	retryOptions := getTestingRetryOptions()
	resourceName := "name"
	crdResource := fixCrdResourceWith(resourceName)
	namespaceResource := fixNamespaceResourceWith(resourceName)

	t.Run("should install CRDs", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		resourceParser.On("ParseFile", pathToFirstResource).Return(crdResource, nil)
		pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		resourceParser.On("ParseFile", pathToSecondResource).Return(crdResource, nil)

		resourceApplier.On("Apply", crdResource).Return(nil)

		input := resourceInfoInput{
			resourceType:             "CustomResourceDefinition",
			dirSuffix:                "crds",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, 2, len(output.Installed))
		assert.Zero(t, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, pathToFirstResource))

		expectedSecondComponent := "comp2"
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, pathToSecondResource))
	})

	t.Run("should not install CRDs due to incorrect input resource type", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		resourceParser.On("ParseFile", pathToFirstResource).Return(crdResource, nil)
		pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		resourceParser.On("ParseFile", pathToSecondResource).Return(crdResource, nil)

		resourceApplier.On("Apply", crdResource).Return(nil)

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
		assert.Equal(t, 2, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFirstComponent, pathToFirstResource))

		expectedSecondComponent := "comp2"
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedSecondComponent, pathToSecondResource))
	})

	t.Run("should create namespaces", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp1/ns.yaml")
		resourceParser.On("ParseFile", pathToFirstResource).Return(namespaceResource, nil)
		pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp2/ns.yaml")
		resourceParser.On("ParseFile", pathToSecondResource).Return(namespaceResource, nil)

		resourceApplier.On("Apply", namespaceResource).Return(nil)

		input := resourceInfoInput{
			resourceType:             "Namespace",
			dirSuffix:                "namespaces",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, 2, len(output.Installed))
		assert.Zero(t, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, pathToFirstResource))

		expectedSecondComponent := "comp2"
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, pathToSecondResource))
	})

	t.Run("should not create namespaces due to incorrect input resource type", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp1/ns.yaml")
		resourceParser.On("ParseFile", pathToFirstResource).Return(namespaceResource, nil)
		pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/namespaces/comp2/ns.yaml")
		resourceParser.On("ParseFile", pathToSecondResource).Return(namespaceResource, nil)

		resourceApplier.On("Apply", namespaceResource).Return(nil)

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
		assert.Equal(t, 2, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFirstComponent, pathToFirstResource))

		expectedSecondComponent := "comp2"
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedSecondComponent, pathToSecondResource))
	})

	t.Run("should partially install resources due to incorrect resource format and resource type different than input info", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/partiallycorrect")
		pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		resourceParser.On("ParseFile", pathToFirstResource).Return(crdResource, nil)
		pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		resourceParser.On("ParseFile", pathToSecondResource).Return(crdResource, nil)
		pathToThirdResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp3/crd.yaml")
		resourceParser.On("ParseFile", pathToThirdResource).Return(nil, errors.New("Parser error"))
		pathToFourthResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp4/ns.yaml")
		resourceParser.On("ParseFile", pathToFourthResource).Return(namespaceResource, nil)
		pathToFifthResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp5/ns.yaml")
		resourceParser.On("ParseFile", pathToFifthResource).Return(nil, errors.New("Parser error"))

		resourceApplier.On("Apply", crdResource).Return(nil)
		resourceApplier.On("Apply", namespaceResource).Return(nil)
		resourceApplier.On("Apply", nil).Return(errors.New("Applier error"))

		input := resourceInfoInput{
			resourceType:             "CustomResourceDefinition",
			dirSuffix:                "crds",
			installationResourcePath: resourcePath,
		}

		// when
		output, err := i.install(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, 2, len(output.Installed))
		assert.Equal(t, 3, len(output.NotInstalled))

		expectedFirstComponent := "comp1"
		assert.True(t, containsFileWithDetails(output.Installed, expectedFirstComponent, pathToFirstResource))

		expectedSecondComponent := "comp2"
		assert.True(t, containsFileWithDetails(output.Installed, expectedSecondComponent, pathToSecondResource))

		expectedThirdComponent := "comp3"
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedThirdComponent, pathToThirdResource))

		expectedFourthComponent := "comp4"
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFourthComponent, pathToFourthResource))

		expectedFifthComponent := "comp5"
		assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFifthComponent, pathToFifthResource))
	})

	t.Run("should fail to install resources", func(t *testing.T) {
		t.Run("due to error about not existing installation resources path", func(t *testing.T) {
			// given
			resourceParser := &mocks.ResourceParser{}
			resourceApplier := &mocks.ResourceApplier{}
			i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
			pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
			resourceParser.On("ParseFile", pathToFirstResource).Return(crdResource, nil)
			pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
			resourceParser.On("ParseFile", pathToSecondResource).Return(crdResource, nil)

			resourceApplier.On("Apply", crdResource).Return(nil)

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
			resourceParser := &mocks.ResourceParser{}
			resourceApplier := &mocks.ResourceApplier{}
			i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)
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
			resourceParser := &mocks.ResourceParser{}
			resourceApplier := &mocks.ResourceApplier{}
			i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/incorrect")
			pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
			resourceParser.On("ParseFile", pathToFirstResource).Return(nil, errors.New("Parser error"))
			pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
			resourceParser.On("ParseFile", pathToSecondResource).Return(nil, errors.New("Parser error"))

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
			assert.Equal(t, 2, len(output.NotInstalled))

			expectedFirstComponent := "comp1"
			assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFirstComponent, pathToFirstResource))

			expectedSecondComponent := "comp2"
			assert.True(t, containsFileWithDetails(output.NotInstalled, expectedSecondComponent, pathToSecondResource))
		})

		t.Run("due to applier error", func(t *testing.T) {
			// given
			resourceParser := &mocks.ResourceParser{}
			resourceApplier := &mocks.ResourceApplier{}
			i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient, retryOptions)

			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/incorrect")
			pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
			resourceParser.On("ParseFile", pathToFirstResource).Return(crdResource, nil)
			pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
			resourceParser.On("ParseFile", pathToSecondResource).Return(crdResource, nil)

			resourceApplier.On("Apply", crdResource).Return(errors.New("Applier error"))

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
			assert.Equal(t, 2, len(output.NotInstalled))

			expectedFirstComponent := "comp1"
			expectedFirstPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
			assert.True(t, containsFileWithDetails(output.NotInstalled, expectedFirstComponent, expectedFirstPath))

			expectedSecondComponent := "comp2"
			expectedSecondPath := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
			assert.True(t, containsFileWithDetails(output.NotInstalled, expectedSecondComponent, expectedSecondPath))
		})
	})
}

func Test_addLabel(t *testing.T) {
	t.Run("should not add label when label is empty", func(t *testing.T) {
		// given
		obj := &unstructured.Unstructured{}
		label := ""
		value := "value"

		// when
		addLabel(obj, label, value)

		// then
		labels := obj.GetLabels()
		assert.Nil(t, labels)
	})

	t.Run("should add label when label is not empty and object had no labels", func(t *testing.T) {
		// given
		obj := &unstructured.Unstructured{}
		label := "label"
		value := "value"

		// when
		addLabel(obj, label, value)

		// then
		labels := obj.GetLabels()
		assert.NotNil(t, labels)
		assert.Equal(t, 1, len(labels))

		// and then
		value, ok := labels[label]
		assert.True(t, ok)
		assert.Equal(t, value, "value")
	})

	t.Run("should add label when label is not empty and object had any label", func(t *testing.T) {
		// given
		obj := fixResourceWithLabel("label1")
		label := "label"
		value := "value"

		// when
		addLabel(obj, label, value)

		// then
		labels := obj.GetLabels()
		assert.NotNil(t, labels)
		assert.Equal(t, 2, len(labels))

		// and then
		value, ok := labels[label]
		assert.True(t, ok)
		assert.Equal(t, value, "value")
	})

	t.Run("should override label when label is not empty and object had given label", func(t *testing.T) {
		// given
		obj := fixResourceWithLabel("label")
		label := "label"
		value := "newValue"

		// when
		addLabel(obj, label, value)

		// then
		labels := obj.GetLabels()
		assert.NotNil(t, labels)
		assert.Equal(t, 1, len(labels))

		// and then
		value, ok := labels[label]
		assert.True(t, ok)
		assert.Equal(t, value, "newValue")
	})
}

func getTestingConfig() Config {
	return Config{
		Log:                      logger.NewLogger(true),
		InstallationResourcePath: "installationResourcePath",
		KubeconfigSource: config.KubeconfigSource{
			Path:    "path",
			Content: "",
		},
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
	return path.Join(test.GetTestDataDirectory(), "resources")
}

func containsFileWithDetails(files []File, component string, path string) bool {
	for _, file := range files {
		if file.component == component && file.path == path {
			return true
		}
	}

	return false
}

func fixCrdResourceWith(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"group": "group",
			},
		},
	}
}

func fixNamespaceResourceWith(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
}

func fixResourceWithLabel(label string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					label: "value",
				},
			},
		},
	}
}

func getPreInstaller(applier ResourceApplier, parser ResourceParser, cfg Config, dynamicClient dynamic.Interface, retryOptions []retry.Option) *PreInstaller {
	return &PreInstaller{
		applier:       applier,
		parser:        parser,
		cfg:           cfg,
		dynamicClient: dynamicClient,
		retryOptions:  retryOptions,
	}
}
