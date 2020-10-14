package main

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/hydroform/function-examples/internal/client"
	xunstruct "github.com/kyma-incubator/hydroform/function-examples/internal/unstructured"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	c := client.MapClient{
		ApiVersion: "test.me.plz/v1alpha1",
		Kind:       "Sample",
		Group:      "test.me.pl",
		Resource:   "samples",
		Data:       unstructured.UnstructuredList{},
	}

	u := xunstruct.NewSample("test", "test-ns")
	o := operator.NewGenericOperator(&c, u)
	ctx := context.Background()

	if err := o.Delete(ctx, operator.DeleteOptions{
		Options: operator.Options{
			Callbacks: operator.Callbacks{
				Post: []func(interface{}, error) error{
					func(v interface{}, err error) error {
						log.WithError(err).Error("while deleting object")
						return fmt.Errorf("operation stopped: %s", err)
					},
				},
			},
		},
	}); err != nil {
		log.Fatal(err)
	}
}
