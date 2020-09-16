package operator

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"
)

type Callback = func(entry client.StatusEntry, err error) error

type Operator interface {
	Apply(ApplyOptions, ...Callback) error
	Delete(DeleteOptions, ...Callback) error
}

func applyObject(c client.Client, u unstructured.Unstructured, stages []string) (*unstructured.Unstructured, client.StatusEntry, error) {
	// Check if object exists
	response, err := c.Get(u.GetName(), metav1.GetOptions{})
	objFound := !errors.IsNotFound(err)
	if err != nil && objFound {
		statusEntryFailed := client.NewStatusEntryFailed(u)
		return &u, statusEntryFailed, err
	}

	// If object is up to date return
	var equal bool
	if objFound {
		//FIXME this fails for function unstructured - investigate
		equal = equality.Semantic.DeepDerivative(u.Object["spec"], response.Object["spec"])
	}

	if objFound && equal {
		statusEntrySkipped := client.NewStatusEntrySkipped(*response)
		return response, statusEntrySkipped, nil
	}

	// If object needs update
	if objFound && !equal {
		response.Object["spec"] = u.Object["spec"]
		err = retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
			response, err = c.Update(response, metav1.UpdateOptions{
				DryRun: stages,
			})
			return err
		})

		if err != nil {
			statusEntryFailed := client.NewStatusEntryFailed(*response)
			return &u, statusEntryFailed, err
		}

		statusEntryUpdated := client.NewStatusEntryUpdated(*response)
		return response, statusEntryUpdated, nil
	}

	response, err = c.Create(&u, metav1.CreateOptions{
		DryRun: stages,
	})
	if err != nil {
		statusEntryFailed := client.NewStatusEntryFailed(u)
		return &u, statusEntryFailed, err
	}

	statusEntryCreated := client.NewStatusEntryCreated(*response)
	return response, statusEntryCreated, nil
}

func deleteObject(i client.Client, u unstructured.Unstructured, ops DeleteOptions) (client.StatusEntry, error) {
	if err := i.Delete(u.GetName(), &metav1.DeleteOptions{
		DryRun:            ops.DryRun,
		PropagationPolicy: &ops.DeletionPropagation,
	}); err != nil {
		statusEntryFailed := client.NewStatusEntryFailed(u)
		return statusEntryFailed, err
	}
	statusEntryDeleted := client.NewStatusEntryDeleted(u)
	return statusEntryDeleted, nil
}

func fireCallbacks(e client.StatusEntry, err error, c []Callback) error {
	for _, callback := range c {
		var callbackErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					callbackErr = fmt.Errorf("%v", r)
				}
			}()
			callbackErr = callback(e, err)
		}()
		if callbackErr != nil {
			return callbackErr
		}
	}
	return nil
}
