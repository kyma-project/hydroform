package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/kyma-project/hydroform/function-examples/internal/client"
	xunstruct "github.com/kyma-project/hydroform/function-examples/internal/unstructured"
	"github.com/kyma-project/hydroform/function/pkg/manager"
	"github.com/kyma-project/hydroform/function/pkg/operator"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	ctx := context.Background()

	c := client.MapClient{
		ApiVersion: "test.me.plz/v1alpha1",
		Kind:       "Sample",
		Group:      "test.me.pl",
		Resource:   "samples",
		Data:       unstructured.UnstructuredList{},
	}

	uParent := xunstruct.NewSample("parent", "test-ns")
	uChild1 := xunstruct.NewSample("child1", "test-ns")
	uChild2 := xunstruct.NewSample("child2", "test-ns")
	uSibling := xunstruct.NewSample("sibling", "test-ns")

	m := manager.NewManager()
	m.AddParent(operator.NewGenericOperator(&c, uParent), []operator.Operator{
		operator.NewGenericOperator(&c, uChild1, uChild2),
		operator.NewGenericOperator(&c, uSibling),
	})

	if err := m.Do(ctx, manager.Options{SetOwnerReferences: true}); err != nil {
		log.Fatal(err)
	}

	for _, item := range c.Data.Items {
		func() {
			defer fmt.Println("---")
			if err := yaml.NewEncoder(os.Stdout).Encode(&item.Object); err != nil {
				log.Fatal(err)
			}
		}()
	}
}
