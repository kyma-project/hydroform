package operator

import (
	"context"

	"github.com/kyma-incubator/hydroform/function/pkg/client"
	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type apiRuleOperator struct {
	fnRef           string
	genericOperator genericOperator
}

func NewApiRuleOperator(c client.Client, fnName string, u ...unstructured.Unstructured) Operator {
	return &apiRuleOperator{
		fnRef: fnName,
		genericOperator: genericOperator{
			Client: c,
			items:  u,
		},
	}
}

// buildMatchRemovedApiRulePredicate - creates a predicate to match the objects that should be deleted
func buildMatchRemovedApiRulePredicate(fnName string, items []unstructured.Unstructured) func(map[string]interface{}) (bool, error) {
	return func(obj map[string]interface{}) (bool, error) {
		var apiRule types.ApiRule
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, &apiRule); err != nil {
			return false, err
		}

		if apiRule.Spec.Service.Name != fnName {
			return false, nil
		}

		containsApiRule := contains(items, apiRule.ObjectMeta.Name)
		return !containsApiRule, nil
	}
}

func (o apiRuleOperator) Apply(ctx context.Context, opts ApplyOptions) error {
	predicateFn := buildMatchRemovedApiRulePredicate(o.fnRef, o.genericOperator.items)

	if err := wipeRemoved(ctx, o.genericOperator.Client, predicateFn, opts.Options); err != nil {
		return err
	}
	return o.genericOperator.Apply(ctx, opts)
}

func (o apiRuleOperator) Delete(ctx context.Context, opts DeleteOptions) error {
	return o.genericOperator.Delete(ctx, opts)
}
