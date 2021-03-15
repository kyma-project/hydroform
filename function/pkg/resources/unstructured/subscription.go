package unstructured

import (
	"fmt"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	apiVersionSubscription = "eventing.kyma-project.io/v1alpha1"
)

func joinNonEmpty(elems []string, sep string) string {
	length := len(elems)
	if length == 0 {
		return ""
	}

	if length == 1 {
		return elems[0]
	}

	result := elems[0]
	for _, elem := range elems[1:] {
		if elem == "" {
			continue
		}
		result = fmt.Sprintf("%s%s%s", result, sep, elem)
	}
	return result
}

type toUnstructured func(obj interface{}) (map[string]interface{}, error)

func NewSubscriptions(cfg workspace.Cfg) ([]unstructured.Unstructured, error) {
	return newSubscriptions(cfg, runtime.DefaultUnstructuredConverter.ToUnstructured)
}

func newSubscriptions(cfg workspace.Cfg, f toUnstructured) ([]unstructured.Unstructured, error) {
	var list []unstructured.Unstructured
	sink := fmt.Sprintf("%s.%s.svc.cluster.local", cfg.Name, cfg.Namespace)

	for _, subscriptionInfo := range cfg.Triggers {
		subscriptionName := subscriptionInfo.Name
		if subscriptionName == "" {
			subscriptionName = joinNonEmpty([]string{
				cfg.Name,
				subscriptionInfo.Source,
			}, "-")
		}

		subscription := types.Subscription{
			TypeMeta: v1.TypeMeta{
				APIVersion: apiVersionSubscription,
				Kind:       "Subscription",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      subscriptionName,
				Namespace: cfg.Namespace,
				Labels:    cfg.Labels,
			},
			Spec: types.SubscriptionSpec{
				Protocol: "NATS",
				Sink:     sink,
				ProtocolSettings: types.ProtocolSettings{
					ExemptHandshake: true,
					Qos:             "AT-LEAST-ONCE",
					WebhookAuth:     types.WebhookAuth{},
				},
				Filter: types.Filter{
					Filters: []types.EventFilter{
						{
							EventSource: types.EventFilterProperty{
								Property: "source",
								Type:     "exact",
								Value:    subscriptionInfo.Source,
							},
							EventType: types.EventFilterProperty{
								Property: "type",
								Type:     "exact",
								Value: joinNonEmpty([]string{
									subscriptionInfo.Type,
									subscriptionInfo.EventTypeVersion,
								}, "."),
							},
						},
					},
				},
			},
		}

		unstructuredStubscription, err := f(&subscription)
		if err != nil {
			return nil, err
		}

		list = append(list, unstructured.Unstructured{
			Object: unstructuredStubscription,
		})
	}

	return list, nil
}
