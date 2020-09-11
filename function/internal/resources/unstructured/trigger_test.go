package unstructured

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/function/internal/workspace"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
	"testing"
)

var (
	cfgTestTriggersFull = workspace.Cfg{
		Name:      "trigger-test",
		Namespace: "trigger-test",
		Labels: map[string]interface{}{
			"test": "me",
		},
	}
)

func Test_NewTrigger(t *testing.T) {
	for _, cfg := range []workspace.Cfg{cfgTestTriggersFull} {
		ref := NewFunctionOwnerReference("test", "test")
		result, err := newFunction(cfg, readFileTestNode, ref.Object)
		gomega.NewWithT(t).Expect(err).ShouldNot(gomega.HaveOccurred())

		testDataSlice := testPropertyDataSlice(cfg)
		for _, prop := range testDataSlice {
			name := strings.Join(prop.field, ".")
			t.Run(fmt.Sprintf("%s should be correct", name), func(t *testing.T) {
				g := gomega.NewWithT(t)
				value, found, err := unstructured.NestedFieldNoCopy(result.Object, prop.field...)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
				g.Expect(found).To(gomega.BeTrue())
				g.Expect(value).To(gomega.Equal(prop.expected))
			})
		}
	}
}
