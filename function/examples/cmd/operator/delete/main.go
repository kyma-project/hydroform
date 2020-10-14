package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/hydroform/function-examples/internal/client"
	xunstruct "github.com/kyma-incubator/hydroform/function-examples/internal/unstructured"
	"github.com/kyma-incubator/hydroform/function/pkg/operator"
)

func main() {
	ctx := context.Background()
	ul, err := xunstruct.FromString(ctx, sampleData)
	if err != nil {
		log.Fatal(err)
	}

	c := client.MapClient{
		ApiVersion: "test.me.plz/v1alpha1",
		Kind:       "Sample",
		Group:      "test.me.pl",
		Resource:   "samples",
		Data:       ul,
	}

	log.WithField("dataLen", len(c.Data.Items)).Info("starting")

	u := xunstruct.NewSample("test", "test-ns")
	o := operator.NewGenericOperator(&c, u)

	if err := o.Delete(ctx, operator.DeleteOptions{}); err != nil {
		log.Fatal(err)
	}

	log.WithField("dataLen", len(c.Data.Items)).Info("done")
}

const sampleData = `apiVersion: test.me.plz/v1alpha1
kind: Sample
metadata:
  name: test
  namespace: test-ns`
