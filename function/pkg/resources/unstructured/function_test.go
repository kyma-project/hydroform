package unstructured

//import (
//	"fmt"
//	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
//	"path"
//	"strings"
//	"testing"
//
//	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
//	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
//	"github.com/onsi/gomega"
//)
//
//type testPropertyData struct {
//	field    []string
//	expected interface{}
//}
//
//func testPropertyDataSlice(cfg workspace.Cfg) []testPropertyData {
//	return []testPropertyData{
//		{
//			field:    []string{"metadata", "name"},
//			expected: cfg.Name,
//		},
//		{
//			field:    []string{"metadata", "labels"},
//			expected: cfg.Labels,
//		},
//		{
//			field:    []string{"apiVersion"},
//			expected: "serverless.kyma-project.io/v1alpha1",
//		},
//		{
//			field:    []string{"spec", "source"},
//			expected: "test-js-handler",
//		},
//		{
//			field:    []string{"spec", "runtime"},
//			expected: cfg.Runtime,
//		},
//	}
//}
//
//func readFileTestNode(filename string) ([]byte, error) {
//	_, realFilename := path.Split(filename)
//	switch workspace.FileName(realFilename) {
//	case workspace.FileNameHandlerPy:
//		return []byte("test-python-handler"), nil
//	case workspace.FileNameRequirementsTxt:
//		return []byte("test-python-requirements"), nil
//	case workspace.FileNameHandlerJs:
//		return []byte("test-js-handler"), nil
//	case workspace.FileNamePackageJSON:
//		return []byte("test-js-deps"), nil
//	default:
//		return []byte{}, nil
//	}
//}
//
//var (
//	cfgTestFull = workspace.Cfg{
//		Name:      "test",
//		Namespace: "test",
//		Labels: map[string]string{
//			"test": "me",
//		},
//		Source: workspace.SourceInline{},
//		Runtime: types.Nodejs10,
//		Resources: struct {
//			Limits   workspace.ResourceList `yaml:"limits"`
//			Requests workspace.ResourceList `yaml:"requests"`
//		}{
//			Limits: map[workspace.ResourceName]interface{}{
//				workspace.ResourceNameCPU:    "1",
//				workspace.ResourceNameMemory: "10m",
//			},
//			Requests: map[workspace.ResourceName]interface{}{
//				workspace.ResourceNameCPU:    "1",
//				workspace.ResourceNameMemory: "10m",
//			},
//		},
//	}
//	cfgTestJustLimits = workspace.Cfg{
//		Name:      "test",
//		Namespace: "test",
//		Labels: map[string]string{
//			"test": "me",
//		},
//		Source: workspace.SourceInline{},
//		Runtime: types.Nodejs10,
//		Resources: struct {
//			Limits   workspace.ResourceList `yaml:"limits"`
//			Requests workspace.ResourceList `yaml:"requests"`
//		}{
//			Limits: map[workspace.ResourceName]interface{}{
//				workspace.ResourceNameCPU:    "1",
//				workspace.ResourceNameMemory: "10m",
//			},
//		},
//	}
//	cfgTestNoResources = workspace.Cfg{
//		Name:      "test",
//		Namespace: "test",
//		Labels: map[string]string{
//			"test": "me",
//		},
//		Runtime: types.Nodejs10,
//		Source: workspace.SourceInline{},
//	}
//	cfgTestNoResourcesAndLabels = workspace.Cfg{
//		Name:      "test",
//		Namespace: "test",
//		Runtime:   types.Nodejs10,
//		Source: workspace.SourceInline{},
//	}
//)
//
//func Test_NewFunctionError(t *testing.T) {
//	_, err := NewFunction(workspace.Cfg{
//		Runtime: types.Nodejs12,
//	})
//	gomega.NewWithT(t).Expect(err).Should(gomega.HaveOccurred())
//}
//
//func Test_NewFunction(t *testing.T) {
//	for _, cfg := range []workspace.Cfg{
//		cfgTestFull,
//		cfgTestJustLimits,
//		cfgTestNoResources,
//		cfgTestNoResourcesAndLabels,
//	} {
//		result, err := newFunction(cfg, readFileTestNode)
//		gomega.NewWithT(t).Expect(err).ShouldNot(gomega.HaveOccurred())
//
//		testDataSlice := testPropertyDataSlice(cfg)
//		for _, prop := range testDataSlice {
//			name := strings.Join(prop.field, ".")
//			t.Run(fmt.Sprintf("%s should be correct", name), func(t *testing.T) {
//				g := gomega.NewWithT(t)
//				value, found, err := unstructured.NestedFieldNoCopy(result.Object, prop.field...)
//				g.Expect(err).ShouldNot(gomega.HaveOccurred())
//				g.Expect(found).To(gomega.BeTrue())
//				g.Expect(value).To(gomega.Equal(prop.expected))
//			})
//		}
//	}
//}
