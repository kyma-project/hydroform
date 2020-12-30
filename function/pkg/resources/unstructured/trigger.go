package unstructured

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/hydroform/function/pkg/resources/types"
	"github.com/kyma-incubator/hydroform/function/pkg/workspace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	triggerApiVersion = "eventing.knative.dev/v1alpha1"
	triggerNameFormat = "%s-%s"
)

func NewTriggers(cfg workspace.Cfg) ([]unstructured.Unstructured, error) {
	var list []unstructured.Unstructured

	for _, triggerInfo := range cfg.Triggers {
		triggerName := fmt.Sprintf(triggerNameFormat, cfg.Name, triggerInfo.Source)
		if triggerInfo.Name != "" {
			triggerName = triggerInfo.Name
		}

		t := types.Trigger{
			ApiVersion: triggerApiVersion,
			Kind:       "Trigger",
			ObjectMeta: metav1.ObjectMeta{
				Name:      triggerName,
				Namespace: cfg.Namespace,
				Labels:    cfg.Labels,
			},
			Spec: types.TriggerSpec{
				Filter: types.TriggerFilter{
					Attributes: types.Attributes{
						EventTypeVersion: triggerInfo.EventTypeVersion,
						Source:           triggerInfo.Source,
						Type:             triggerInfo.Type,
					},
				},
				Subscriber: types.TriggerSubscriber{
					Reference: types.TriggerReference{
						ApiVersion: "v1",
						Kind:       "Service",
						Name:       cfg.Name,
						Namespace:  cfg.Namespace,
					},
				},
				Broker: "default",
			},
		}

		unstructuredTrigger, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&t)
		if err != nil {
			return nil, err
		}
		list = append(list, unstructured.Unstructured{Object: unstructuredTrigger})

	}

	return list, nil
}
