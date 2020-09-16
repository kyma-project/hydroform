package operator

import (
	"context"
	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
)

type Parent interface {
	Resource
	Delete(client.Client, DeleteOptions) (client.Status, error)
}

type parent struct {
	unstructured.Unstructured
	schema.GroupVersionResource
}

func NewParent(u unstructured.Unstructured, r schema.GroupVersionResource) Parent {
	return parent{
		Unstructured:         u,
		GroupVersionResource: r,
	}
}

func NewParentFunction(u unstructured.Unstructured) Parent {
	return NewParent(u, schema.GroupVersionResource{
		Group:    "serverless.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "functions",
	})
}

func applyObject(c client.Client, u unstructured.Unstructured, stages []string) (client.StatusEntry, error) {
	// Check if object exists
	ctx := context.Background()
	response, err := c.Get(ctx, u.GetName(), metav1.GetOptions{})
	objFound := !errors.IsNotFound(err)
	if err != nil && objFound {
		statusEntryFailed := client.NewStatusEntryFailed(u)
		return statusEntryFailed, err
	}

	// If object is up to date return
	var equal bool
	if objFound {
		equal = equality.Semantic.DeepDerivative(response.Object["spec"], u.Object["spec"])
	}

	if objFound && equal {
		statusEntrySkipped := client.NewStatusEntrySkipped(*response)
		return statusEntrySkipped, nil
	}

	// If object needs update
	if objFound && !equal {
		response.Object["spec"] = u.Object["spec"]
		err = retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
			response, err = c.Update(ctx, response, v1.UpdateOptions{
				DryRun: stages,
			})
			return err
		})

		if err != nil {
			statusEntryFailed := client.NewStatusEntryFailed(*response)
			return statusEntryFailed, err
		}

		statusEntryUpdated := client.NewStatusEntryUpdated(*response)
		return statusEntryUpdated, nil
	}

	response, err = c.Create(ctx, &u, v1.CreateOptions{
		DryRun: stages,
	})
	if err != nil {
		statusEntryFailed := client.NewStatusEntryFailed(u)
		return statusEntryFailed, err
	}

	statusEntryCreated := client.NewStatusEntryCreated(*response)
	return statusEntryCreated, nil
}

func (p parent) Apply(c client.Client, opts ApplyOptions) (client.Status, error) {
	p.Unstructured.SetOwnerReferences(opts.OwnerReferences)
	statusEntry, err := applyObject(c, p.Unstructured, opts.DryRun)
	return []client.StatusEntry{statusEntry}, err
}

func (p parent) GetGroupVersionResource() schema.GroupVersionResource {
	return p.GroupVersionResource
}

func (p parent) Delete(c client.Client, opts DeleteOptions) (client.Status, error) {
	if err := c.Delete(context.Background(), p.Unstructured.GetName(), v1.DeleteOptions{
		DryRun:            opts.DryRun,
		PropagationPolicy: &opts.DeletionPropagation,
	}); err != nil {
		statusEntryFailed := client.NewStatusEntryFailed(p.Unstructured)
		return []client.StatusEntry{statusEntryFailed}, err
	}

	statusEntryDeleted := client.NewStatusEntryDeleted(p.Unstructured)
	return []client.StatusEntry{statusEntryDeleted}, nil
}
