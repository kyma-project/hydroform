package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/hydroform/function-examples/internal/client"
	xunstruct "github.com/kyma-project/hydroform/function-examples/internal/unstructured"
	"github.com/kyma-project/hydroform/function/pkg/operator"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	ctx := context.Background()

	c := client.MapClient{
		Data: unstructured.UnstructuredList{},
	}
	u := xunstruct.NewSample("test", "test-ns")
	o := operator.NewGenericOperator(&c, u)

	if err := o.Apply(ctx, operator.ApplyOptions{
		Options: operator.Options{
			Callbacks: operator.Callbacks{
				Pre: []func(interface{}, error) error{
					func(i interface{}, e error) error {
						u, ok := i.(*unstructured.Unstructured)
						if !ok {
							return fmt.Errorf("unexpected type")
						}
						log.WithFields(u.Object).Info("applying object")
						return nil
					},
				},
			},
		},
	}); err != nil {
		log.Fatal(err)
	}
}
