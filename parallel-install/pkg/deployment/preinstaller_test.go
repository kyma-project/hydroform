package deployment

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment/mocks"
	"github.com/stretchr/testify/require"
	"path"
	"regexp"
	"testing"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/test"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

func TestPreInstaller_InstallCRDs(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	resourceName := "name"
	crdResource := fixCrdResourceWithGivenName(resourceName)

	t.Run("should install CRDs", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/correct")
		cfg.InstallationResourcePath = resourcePath
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient)

		cfg.InstallationResourcePath = resourcePath
		pathToFirstResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp1/crd.yaml")
		resourceParser.On("ParseFile", pathToFirstResource).Return(crdResource, nil)
		pathToSecondResource := fmt.Sprintf("%s%s", resourcePath, "/crds/comp2/crd.yaml")
		resourceParser.On("ParseFile", pathToSecondResource).Return(crdResource, nil)

		resourceApplier.On("Apply", crdResource).Return(nil)

		// when
		err := i.InstallCRDs()

		// then
		require.NoError(t, err)
		resourceParser.AssertNumberOfCalls(t, "ParseFile", 2)
		resourceApplier.AssertNumberOfCalls(t, "Apply", 2)
	})

}

func TestPreInstaller_install(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	cfg := getTestingConfig()
	resourceName := "name"
	crdResource := fixCrdResourceWithGivenName(resourceName)
	namespaceResource := fixNamespaceResourceWithGivenName(resourceName)

	t.Run("should install CRDs", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient)

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
		require.NoError(t, err)
		resourceParser.AssertNumberOfCalls(t, "ParseFile", 2)
		resourceApplier.AssertNumberOfCalls(t, "Apply", 2)

		require.Equal(t, 2, len(output.Installed))
		require.Zero(t, len(output.NotInstalled))
		require.True(t, containsFileWithDetails(output.Installed, "comp1", pathToFirstResource))
		require.True(t, containsFileWithDetails(output.Installed, "comp2", pathToSecondResource))
	})

	t.Run("should not install CRDs due to incorrect input resource type", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient)

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
		require.NoError(t, err)
		resourceParser.AssertNumberOfCalls(t, "ParseFile", 2)
		resourceApplier.AssertNumberOfCalls(t, "Apply", 0)

		require.Zero(t, len(output.Installed))
		require.Equal(t, 2, len(output.NotInstalled))
		require.True(t, containsFileWithDetails(output.NotInstalled, "comp1", pathToFirstResource))
		require.True(t, containsFileWithDetails(output.NotInstalled, "comp2", pathToSecondResource))
	})

	t.Run("should partially install resources due to incorrect resource format and resource type different than input info", func(t *testing.T) {
		// given
		resourceParser := &mocks.ResourceParser{}
		resourceApplier := &mocks.ResourceApplier{}
		i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient)

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
		require.NoError(t, err)
		resourceParser.AssertNumberOfCalls(t, "ParseFile", 5)
		resourceApplier.AssertNumberOfCalls(t, "Apply", 2)

		require.Equal(t, 2, len(output.Installed))
		require.Equal(t, 3, len(output.NotInstalled))
		require.True(t, containsFileWithDetails(output.Installed, "comp1", pathToFirstResource))
		require.True(t, containsFileWithDetails(output.Installed, "comp2", pathToSecondResource))
		require.True(t, containsFileWithDetails(output.NotInstalled, "comp3", pathToThirdResource))
		require.True(t, containsFileWithDetails(output.NotInstalled, "comp4", pathToFourthResource))
		require.True(t, containsFileWithDetails(output.NotInstalled, "comp5", pathToFifthResource))
	})

	t.Run("should fail to install resources", func(t *testing.T) {
		t.Run("due to error about not existing installation resources path", func(t *testing.T) {
			// given
			resourceParser := &mocks.ResourceParser{}
			resourceApplier := &mocks.ResourceApplier{}
			i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient)

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
			require.Error(t, err)
			resourceParser.AssertNumberOfCalls(t, "ParseFile", 0)
			resourceApplier.AssertNumberOfCalls(t, "Apply", 0)

			expectedError := "no such file or directory"
			receivedError := err.Error()
			matched, err := regexp.MatchString(expectedError, receivedError)
			require.True(t, matched, fmt.Sprintf("Expected error message: %s but got: %s", expectedError, receivedError))
			require.Zero(t, len(output.Installed))
		})

		t.Run("due to no components detected in installation resources path", func(t *testing.T) {
			// given
			resourceParser := &mocks.ResourceParser{}
			resourceApplier := &mocks.ResourceApplier{}
			i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient)
			resourcePath := fmt.Sprintf("%s%s", getTestingResourcesDirectory(), "/nocomponents")

			input := resourceInfoInput{
				resourceType:             "CustomResourceDefinition",
				dirSuffix:                "crds",
				installationResourcePath: resourcePath,
			}

			// when
			output, err := i.install(input)

			// then
			require.NoError(t, err)
			resourceParser.AssertNumberOfCalls(t, "ParseFile", 0)
			resourceApplier.AssertNumberOfCalls(t, "Apply", 0)

			require.Zero(t, len(output.Installed))
			require.Zero(t, len(output.NotInstalled))
		})

		t.Run("due to parser error", func(t *testing.T) {
			// given
			resourceParser := &mocks.ResourceParser{}
			resourceApplier := &mocks.ResourceApplier{}
			i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient)

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
			require.NoError(t, err)
			resourceParser.AssertNumberOfCalls(t, "ParseFile", 2)
			resourceApplier.AssertNumberOfCalls(t, "Apply", 0)

			require.Zero(t, len(output.Installed))
			require.Equal(t, 2, len(output.NotInstalled))
			require.True(t, containsFileWithDetails(output.NotInstalled, "comp1", pathToFirstResource))
			require.True(t, containsFileWithDetails(output.NotInstalled, "comp2", pathToSecondResource))
		})

		t.Run("due to applier error", func(t *testing.T) {
			// given
			resourceParser := &mocks.ResourceParser{}
			resourceApplier := &mocks.ResourceApplier{}
			i := getPreInstaller(resourceApplier, resourceParser, cfg, dynamicClient)

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
			require.NoError(t, err)
			resourceParser.AssertNumberOfCalls(t, "ParseFile", 2)
			resourceApplier.AssertNumberOfCalls(t, "Apply", 2)

			require.Zero(t, len(output.Installed))
			require.Equal(t, 2, len(output.NotInstalled))
			require.True(t, containsFileWithDetails(output.NotInstalled, "comp1", pathToFirstResource))
			require.True(t, containsFileWithDetails(output.NotInstalled, "comp2", pathToSecondResource))
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
		require.Nil(t, labels)
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
		require.NotNil(t, labels)
		require.Equal(t, 1, len(labels))

		// and then
		value, ok := labels[label]
		require.True(t, ok)
		require.Equal(t, value, "value")
	})

	t.Run("should add label when label is not empty and object had any label", func(t *testing.T) {
		// given
		obj := fixResourceWithGivenLabel("label1")
		label := "label"
		value := "value"

		// when
		addLabel(obj, label, value)

		// then
		labels := obj.GetLabels()
		require.NotNil(t, labels)
		require.Equal(t, 2, len(labels))

		// and then
		value, ok := labels[label]
		require.True(t, ok)
		require.Equal(t, value, "value")
	})

	t.Run("should override label when label is not empty and object had given label", func(t *testing.T) {
		// given
		obj := fixResourceWithGivenLabel("label")
		label := "label"
		value := "newValue"

		// when
		addLabel(obj, label, value)

		// then
		labels := obj.GetLabels()
		require.NotNil(t, labels)
		require.Equal(t, 1, len(labels))

		// and then
		value, ok := labels[label]
		require.True(t, ok)
		require.Equal(t, value, "newValue")
	})
}

func getTestingConfig() inputConfig {
	return inputConfig{
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

func containsFileWithDetails(files []file, component string, path string) bool {
	for _, file := range files {
		if file.component == component && file.path == path {
			return true
		}
	}

	return false
}

func fixCrdResourceWithGivenName(name string) *unstructured.Unstructured {
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

func fixNamespaceResourceWithGivenName(name string) *unstructured.Unstructured {
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

func fixResourceWithGivenLabel(label string) *unstructured.Unstructured {
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

func getPreInstaller(applier ResourceApplier, parser ResourceParser, cfg inputConfig, dynamicClient dynamic.Interface) *preInstaller {
	return &preInstaller{
		applier:       applier,
		parser:        parser,
		cfg:           cfg,
		dynamicClient: dynamicClient,
	}
}
