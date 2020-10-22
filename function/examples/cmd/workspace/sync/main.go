/*
* CODE GENERATED AUTOMATICALLY WITH devops/internal/config
 */

package main

import (
	"context"
	"fmt"

	xunstruct "github.com/kyma-incubator/hydroform/function-examples/internal/unstructured"

	"time"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	ul, err := xunstruct.FromString(ctx, sampleData)
	if err != nil {
		log.Fatal(err)
	}

	configuration := workspace.Cfg{
		Name:      "test1",
		Namespace: "default",
	}

	buildClient := func(ns string, resource schema.GroupVersionResource) client.Client {
		return &xclint.MapClient{
			ApiVersion: fmt.Sprintf("%s/%s", resource.Group, resource.Version),
			Kind:       "Function",
			Group:      resource.Group,
			Resource:   resource.Resource,
			Data:       ul,
			Namespace:  ns,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = workspace.Synchronise(ctx, configuration, "/tmp", buildClient)
	if err != nil {
		panic(err.Error())
	}

	log.Println("Syncing completed.")
}

const sampleData = `apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: test1
  namespace: default
spec:
  deps: |-
    {
      "name": "test1",
      "version": "0.0.1",
      "dependencies": {}
    }
  maxReplicas: 1
  minReplicas: 1
  resources:
    limits:
      cpu: 100m
      memory: 128Mi
    requests:
      cpu: 50m
      memory: 64Mi
  runtime: nodejs12
  source: |-
    module.exports = {
        main: function (event, context) {
            return 'Hello Serverless'
        }
    }`
