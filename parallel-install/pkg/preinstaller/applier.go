package preinstaller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/ghodss/yaml"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type resourceApplier interface {
	Apply(manifest string) error
}

type genericResourceApplier struct {
	log logger.Interface
	kubeClient     kubernetes.Interface
	dynamicClient dynamic.Interface
	decoder runtime.Decoder
	retryOptions []retry.Option
}

func newGenericResourceApplier(log logger.Interface, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface, decoder runtime.Decoder, retryOptions []retry.Option) *genericResourceApplier {
	return &genericResourceApplier{
		log: log,
		kubeClient: kubeClient,
		dynamicClient: dynamicClient,
		decoder: decoder,
		retryOptions: retryOptions,
	}
}

func (c *genericResourceApplier) Apply(manifest string) error {
	var _, grpVerKind, err = c.decoder.Decode([]byte(manifest), nil, nil)
	if err != nil {
		return err
	}

	converted, err := convertYamlToJson(manifest)
	if err != nil {
		return err
	}

	resource, err := parseManifest([]byte(converted))
	if err != nil {
		return err
	}

	resourceSchema := schema.GroupVersionResource{
		Group: grpVerKind.Group,
		Version: grpVerKind.Version,
		Resource: pluralForm(grpVerKind.Kind),
	}

	resourceName := resource.GetName()
	obj, err := c.getResource(resourceName, resourceSchema)
	if err != nil {
		return err
	}

	if obj != nil {
		c.log.Infof("Resource: %s already exists. Skipping.", resourceName)
		return nil
	}

	c.log.Infof("Creating resource: %s .", resourceName)

	return c.createResource(resource, resourceSchema)
}

func (c *genericResourceApplier) createResource(resource *unstructured.Unstructured, resourceSchema schema.GroupVersionResource) error {
	var err error
	err = retry.Do(func() error {
		fmt.Println("Creating resource")

		if _, err = c.dynamicClient.Resource(resourceSchema).Create(context.TODO(), resource, metav1.CreateOptions{}); err != nil {
			return err
		}

		fmt.Println("Created resource")

		return nil
	}, c.retryOptions...)

	if err != nil {
		return err
	}

	return nil
}

func (c *genericResourceApplier) getResource(resourceName string, resourceSchema schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	var obj *unstructured.Unstructured
	err := retry.Do(func() error {
		var err error

		if obj, err = c.dynamicClient.Resource(resourceSchema).Get(context.TODO(), resourceName, metav1.GetOptions{}); err != nil {
			return err
		}

		return nil
	}, c.retryOptions...)

	if err != nil {
		return nil, err
	}

	return obj, err
}

func convertYamlToJson(input string) (string, error){
	convertedInput, err := yaml.YAMLToJSON([]byte(input))
	if err != nil {
		return "", err
	}
	if string(convertedInput) != "null" {
		return string(convertedInput), nil
	}

	return "", err
}

func parseManifest(input []byte) (*unstructured.Unstructured, error) {
	var middleware map[string]interface{}
	err := json.Unmarshal(input, &middleware)
	if err != nil {
		return nil, err
	}

	resource := &unstructured.Unstructured{
		Object: middleware,
	}
	return resource, nil
}

func pluralForm(name string) string {
	return fmt.Sprintf("%ss", strings.ToLower(name))
}
