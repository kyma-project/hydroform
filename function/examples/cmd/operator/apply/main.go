package main

import (
	"context"
	"os"

	"github.com/kyma-incubator/hydroform/function-examples/internal/client"
	xunstruct "github.com/kyma-incubator/hydroform/function-examples/internal/unstructured"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	ctx := context.Background()

	c := client.MapClient{
		Data: unstructured.UnstructuredList{},
	}
	u := xunstruct.NewSample("test", "test-ns")
	o := operator.NewGenericOperator(&c, u)

	if err := o.Apply(ctx, operator.ApplyOptions{}); err != nil {
		log.Fatal(err)
	}

	for _, item := range c.Data.Items {
		if err := yaml.NewEncoder(os.Stdout).Encode(&item.Object); err != nil {
			log.Fatal(err)
		}
	}
}
