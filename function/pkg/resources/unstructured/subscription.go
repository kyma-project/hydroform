package unstructured

import (
	"fmt"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const apiVersionSubscription = "eventing.kyma-project.io/v1alpha1"

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
	//TODO remove http protocol once it will be fixed in eventing
	sink := fmt.Sprintf("http://%s.%s.svc.cluster.local", cfg.Name, cfg.Namespace)

	for _, subscriptionInfo := range cfg.Subscriptions {
		name := generateSubscriptionName(cfg.Name, subscriptionInfo)
		filter := toTypesFilter(subscriptionInfo.Filter)

		subscription := types.Subscription{
			TypeMeta: v1.TypeMeta{
				APIVersion: apiVersionSubscription,
				Kind:       "Subscription",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: cfg.Namespace,
				Labels:    cfg.Labels,
			},
			Spec: types.SubscriptionSpec{
				Protocol: subscriptionInfo.Protocol,
				Sink:     sink,
				ProtocolSettings: types.ProtocolSettings{
					ExemptHandshake: true,
					Qos:             "AT-LEAST-ONCE",
					WebhookAuth:     types.WebhookAuth{},
				},
				Filter: filter,
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

func generateSubscriptionName(functionName string, s workspace.Subscription) string {
	subscriptionName := s.Name
	subscriptionSources := filterSources(s)

	if subscriptionName == "" {
		elems := append([]string{functionName}, subscriptionSources...)
		subscriptionName = joinNonEmpty(elems, "-")
	}
	return subscriptionName
}

func filterSources(s workspace.Subscription) []string {
	var result []string
	for _, evtFilter := range s.Filter.Filters {
		result = append(result, evtFilter.EventSource.Value)
	}
	return result
}

func toTypesFilter(filter workspace.Filter) types.Filter {
	var filters []types.EventFilter
	for _, evtFilter := range filter.Filters {
		filters = append(filters, types.EventFilter{
			EventSource: types.EventFilterProperty{
				Property: evtFilter.EventSource.Property,
				Type:     evtFilter.EventSource.Type,
				Value:    evtFilter.EventSource.Value,
			},
			EventType: types.EventFilterProperty{
				Property: evtFilter.EventType.Property,
				Type:     evtFilter.EventType.Type,
				Value:    evtFilter.EventType.Value,
			},
		})
	}
	return types.Filter{
		Dialect: filter.Dialect,
		Filters: filters,
	}
}
