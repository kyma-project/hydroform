package types

import (
	"encoding/json"
	"os"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_subscriptionV1alpha2(t *testing.T) {
	subscription := SubscriptionV1alpha2{
		TypeMeta: v1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-subscription",
			Namespace: "default",
		},
		Spec: SubscriptionSpecV1alpha2{
			TypeMatching: "matching",
			EventSource:  "source",
			Types: []string{
				"type1",
				"type2",
				"type3",
			},
		},
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(subscription); err != nil {
		t.Error(err)
		t.Fail()
	}
}
