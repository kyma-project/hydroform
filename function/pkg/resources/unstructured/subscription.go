package unstructured

import (
	"fmt"

	"github.com/kyma-project/hydroform/function/pkg/resources/types"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const apiVersionSubscriptionV1alpha1 = "eventing.kyma-project.io/v1alpha1"
const apiVersionSubscriptionV1alpha2 = "eventing.kyma-project.io/v1alpha2"

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
	fmt.Sprintf(string(cfg.SchemaVersion))
	switch cfg.SchemaVersion {
	case workspace.SchemaVersionV0:
		return newSubscriptionsV1alpha1(cfg, runtime.DefaultUnstructuredConverter.ToUnstructured)
	case workspace.SchemaVersionV1:
		return newSubscriptionsV1alpha2(cfg, runtime.DefaultUnstructuredConverter.ToUnstructured)
	default:
		return newSubscriptionsV1alpha1(cfg, runtime.DefaultUnstructuredConverter.ToUnstructured)
	}
}

func newSubscriptionsV1alpha1(cfg workspace.Cfg, f toUnstructured) ([]unstructured.Unstructured, error) {
	var list []unstructured.Unstructured
	//TODO remove http protocol once it will be fixed in eventing
	sink := fmt.Sprintf("http://%s.%s.svc.cluster.local", cfg.Name, cfg.Namespace)

	for iterator, subscriptionInfo := range cfg.Subscriptions {
		name := generateSubscriptionName(cfg.Name, subscriptionInfo, iterator)
		filter := toTypesFilter(subscriptionInfo.V0.Filter)

		subscription := types.SubscriptionV1alpha1{
			TypeMeta: v1.TypeMeta{
				APIVersion: apiVersionSubscriptionV1alpha1,
				Kind:       "Subscription",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: cfg.Namespace,
				Labels:    cfg.Labels,
			},
			Spec: types.SubscriptionSpecV1alpha1{
				ProtocolSettings: &types.ProtocolSettings{},
				Protocol:         subscriptionInfo.V0.Protocol,
				Sink:             sink,
				Filter:           filter,
			},
		}

		unstructuredSubscription, err := f(&subscription)
		if err != nil {
			return nil, err
		}

		list = append(list, unstructured.Unstructured{
			Object: unstructuredSubscription,
		})
	}

	return list, nil
}

func newSubscriptionsV1alpha2(cfg workspace.Cfg, f toUnstructured) ([]unstructured.Unstructured, error) {
	var list []unstructured.Unstructured
	//TODO remove http protocol once it will be fixed in eventing
	sink := fmt.Sprintf("http://%s.%s.svc.cluster.local", cfg.Name, cfg.Namespace)

	for iterator, subscriptionInfo := range cfg.Subscriptions {
		name := generateSubscriptionName(cfg.Name, subscriptionInfo, iterator)

		subscription := types.SubscriptionV1alpha2{
			TypeMeta: v1.TypeMeta{
				APIVersion: apiVersionSubscriptionV1alpha2,
				Kind:       "Subscription",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: cfg.Namespace,
				Labels:    cfg.Labels,
			},
			Spec: types.SubscriptionSpecV1alpha2{
				Sink:         sink,
				TypeMatching: subscriptionInfo.V1.TypeMatching,
				EventSource:  subscriptionInfo.V1.Source,
				Types:        subscriptionInfo.V1.Types,
			},
		}

		unstructuredSubscription, err := f(&subscription)
		if err != nil {
			return nil, err
		}

		list = append(list, unstructured.Unstructured{
			Object: unstructuredSubscription,
		})
	}

	return list, nil
}

func generateSubscriptionName(functionName string, s workspace.Subscription, iterator int) string {
	subscriptionName := s.Name

	if subscriptionName == "" {
		subscriptionName = joinNonEmpty([]string{functionName, fmt.Sprint(iterator)}, "-")
	}
	return subscriptionName
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
