package operator

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"
)

type Callback = func(interface{}, error) error

type Operator interface {
	Apply(context.Context, ApplyOptions) error
	Delete(context.Context, DeleteOptions) error
}

func applyObject(ctx context.Context, c client.Client, u unstructured.Unstructured, stages []string) (*unstructured.Unstructured, client.PostStatusEntry, error) {
	// Check if object exists
	response, err := c.Get(ctx, u.GetName(), metav1.GetOptions{})
	objFound := !errors.IsNotFound(err)
	if err != nil && objFound {
		statusEntryFailed := client.NewPostStatusEntryFailed(u)
		return &u, statusEntryFailed, err
	}

	// If object is up to date return
	var equal bool
	if objFound {
		//FIXME this fails for function unstructured - investigate
		equal = equality.Semantic.DeepDerivative(u.Object["spec"], response.Object["spec"])
	}

	if objFound && equal {
		statusEntrySkipped := client.NewPostStatusEntrySkipped(*response)
		return response, statusEntrySkipped, nil
	}

	// If object needs update
	if objFound && !equal {
		response.Object["spec"] = u.Object["spec"]
		err = retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
			response, err = c.Update(ctx, response, metav1.UpdateOptions{
				DryRun: stages,
			})
			return err
		})

		if err != nil {
			statusEntryFailed := client.NewPostStatusEntryFailed(*response)
			return &u, statusEntryFailed, err
		}

		statusEntryUpdated := client.NewPostStatusEntryUpdated(*response)
		return response, statusEntryUpdated, nil
	}

	response, err = c.Create(ctx, &u, metav1.CreateOptions{
		DryRun: stages,
	})
	if err != nil {
		statusEntryFailed := client.NewPostStatusEntryFailed(u)
		return &u, statusEntryFailed, err
	}

	statusEntryCreated := client.NewStatusEntryCreated(*response)
	return response, statusEntryCreated, nil
}

func deleteObject(ctx context.Context, i client.Client, u unstructured.Unstructured, ops DeleteOptions) (client.PostStatusEntry, error) {
	if err := i.Delete(ctx, u.GetName(), metav1.DeleteOptions{
		DryRun:            ops.DryRun,
		PropagationPolicy: &ops.DeletionPropagation,
	}); err != nil {
		statusEntryFailed := client.NewPostStatusEntryFailed(u)
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
