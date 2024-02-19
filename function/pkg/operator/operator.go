package operator

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/util/retry"

	"github.com/kyma-project/hydroform/function/pkg/client"
)

type Callback = func(interface{}, error) error

//go:generate mockgen -source=operator.go -destination=automock/operator.go

type Operator interface {
	Apply(context.Context, ApplyOptions) error
	Delete(context.Context, DeleteOptions) error
}

func applyObject(ctx context.Context, c client.Client, u unstructured.Unstructured,
	stages []string) (*unstructured.Unstructured, client.PostStatusEntry, error) {
	// Check if object exists
	response, err := c.Get(ctx, u.GetName(), metav1.GetOptions{})
	objFound := response != nil
	isNotFoundErr := errors.IsNotFound(err)
	if err != nil && !isNotFoundErr {
		statusEntryFailed := client.NewPostStatusEntryApplyFailed(u)
		return &u, statusEntryFailed, err
	}

	// If object is up to date return
	var equal bool
	if objFound {
		equal, err = configurationObjectsAreEquivalent(u, *response)
		if err != nil {
			return response, client.NewPostStatusEntrySkipped(*response), nil
		}
	}

	if objFound && equal {
		statusEntrySkipped := client.NewPostStatusEntrySkipped(*response)
		return response, statusEntrySkipped, nil
	}

	// If object needs update
	if objFound && !equal {
		response, err = updateConfigurationObject(response, u)
		if err != nil {
			return &u, client.NewPostStatusEntryApplyFailed(*response), err
		}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
			res, err := c.Update(ctx, response, metav1.UpdateOptions{
				DryRun: stages,
			})
			if err != nil {
				return err
			}
			response = res
			return nil
		})

		if err != nil {
			statusEntryFailed := client.NewPostStatusEntryApplyFailed(*response)
			return &u, statusEntryFailed, err
		}

		statusEntryUpdated := client.NewPostStatusEntryUpdated(*response)
		return response, statusEntryUpdated, nil
	}

	response, err = c.Create(ctx, &u, metav1.CreateOptions{
		DryRun: stages,
	})
	if err != nil {
		statusEntryFailed := client.NewPostStatusEntryApplyFailed(u)
		return &u, statusEntryFailed, err
	}

	statusEntryCreated := client.NewStatusEntryCreated(*response)
	return response, statusEntryCreated, nil
}

func updateConfigurationObject(destination *unstructured.Unstructured,
	source unstructured.Unstructured) (*unstructured.Unstructured, error) {
	// Copies elements: `spec`, `metadata/labels`, `metadata/annotations` from source to destination. The rest is unchanged.
	destination.Object["spec"] = source.Object["spec"]

	destinationMetadata, err := getMetadataFromConfigurationObject(*destination)
	if err != nil {
		return nil, err
	}
	sourceMetadata, err := getMetadataFromConfigurationObject(source)
	if err != nil {
		return nil, err
	}

	updateConfigurationElement(&destinationMetadata, sourceMetadata, "labels")
	updateConfigurationElement(&destinationMetadata, sourceMetadata, "annotations")

	return destination, nil
}

func updateConfigurationElement(destination *map[string]interface{}, source map[string]interface{},
	elementName string) {
	sourceElement, ok := source[elementName]
	if ok {
		(*destination)[elementName] = sourceElement
	} else {
		delete(*destination, elementName)
	}
}

func configurationObjectsAreEquivalent(first unstructured.Unstructured, second unstructured.Unstructured) (bool,
	error) {
	// Checks that the elements: `spec`, `metadata/labels`, `metadata/annotations` are equal (semantically).
	specAreEqual := equality.Semantic.DeepEqual(first.Object["spec"], second.Object["spec"])

	firstMetadata, err := getMetadataFromConfigurationObject(first)
	if err != nil {
		return false, err
	}
	secondMetadata, err := getMetadataFromConfigurationObject(second)
	if err != nil {
		return false, err
	}
	labelsAreEqual := configurationElementsAreEqual(firstMetadata, secondMetadata, "labels")
	annotationsAreEqual := configurationElementsAreEqual(firstMetadata, secondMetadata, "annotations")

	return specAreEqual && labelsAreEqual && annotationsAreEqual, nil
}

func getMetadataFromConfigurationObject(configurationObject unstructured.Unstructured) (map[string]interface{}, error) {
	metadata, ok := configurationObject.Object["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("can't cast object for equality checking")
	}
	return metadata, nil
}

func configurationElementsAreEqual(firstMap map[string]interface{}, secondMap map[string]interface{},
	elementName string) bool {
	firstElement, ok := firstMap[elementName]
	if !ok {
		firstElement = map[string]interface{}{}
	}
	secondElement, ok := secondMap[elementName]
	if !ok {
		secondElement = map[string]interface{}{}
	}

	if firstElement == nil && secondElement == nil {
		return true
	} else if firstElement == nil || secondElement == nil {
		return false
	}
	equal := equality.Semantic.DeepEqual(firstElement, secondElement)
	return equal
}

func waitForObject(ctx context.Context, c client.Client, u unstructured.Unstructured) error {
	w, err := c.Watch(ctx, metav1.ListOptions{
		TypeMeta: v1.TypeMeta{
			Kind:       u.GetKind(),
			APIVersion: u.GetAPIVersion(),
		},
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("metadata.name", u.GetName()),
			fields.OneTermEqualSelector("metadata.namespace", u.GetNamespace()),
		).String(),
	})
	if err != nil {
		return err
	}

functionBlock:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-w.ResultChan():
			if !ok {
				break functionBlock
			}
			if event.Type == watch.Added {
				return nil
			}
		}
	}
	return nil
}

func wipeRemoved(ctx context.Context, c client.Client, deletePredicate func(obj map[string]interface{}) (bool, error),
	opts Options) error {
	list, err := c.List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}

	policy := v1.DeletePropagationBackground

	for i := range list.Items {
		match, err := deletePredicate(list.Items[i].Object)
		if err != nil {
			return err
		}

		if !match {
			continue
		}

		if err := fireCallbacks(&list.Items[i], nil, opts.Pre...); err != nil {
			return err
		}
		// delete and delegate flow ctrl to caller
		if err := c.Delete(ctx, list.Items[i].GetName(), v1.DeleteOptions{
			DryRun:            opts.DryRun,
			PropagationPolicy: &policy,
		}); err != nil {
			statusEntryFailed := client.NewPostStatusEntryDeleteFailed(list.Items[i])
			if err := fireCallbacks(statusEntryFailed, err, opts.Post...); err != nil {
				return err
			}
		}
		statusEntryDeleted := client.NewPostStatusEntryDeleted(list.Items[i])
		if err := fireCallbacks(statusEntryDeleted, nil, opts.Post...); err != nil {
			return err
		}
	}

	return nil
}

func deleteObject(ctx context.Context, i client.Client, u unstructured.Unstructured,
	ops DeleteOptions) (client.PostStatusEntry, error) {
	if err := i.Delete(ctx, u.GetName(), metav1.DeleteOptions{
		DryRun:            ops.DryRun,
		PropagationPolicy: &ops.DeletionPropagation,
	}); err != nil {
		statusEntryFailed := client.NewPostStatusEntryDeleteFailed(u)
		return statusEntryFailed, err
	}
	statusEntryDeleted := client.NewPostStatusEntryDeleted(u)
	return statusEntryDeleted, nil
}

func fireCallbacks(v interface{}, err error, cbs ...Callback) error {
	for _, callback := range cbs {
		var callbackErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					callbackErr = fmt.Errorf("%v", r)
				}
			}()
			callbackErr = callback(v, err)
		}()
		if callbackErr != nil {
			return callbackErr
		}
	}
	return err
}

func isOwnerReference(refs []metav1.OwnerReference, ownerName string) bool {
	for i := range refs {
		if refs[i].Kind == "Function" &&
			refs[i].Name == ownerName {
			return true
		}
	}

	return false
}
