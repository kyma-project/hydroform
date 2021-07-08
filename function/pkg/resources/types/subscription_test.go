package types

import (
	"encoding/json"
	"os"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_subscription(t *testing.T) {
	subscription := Subscription{
		TypeMeta: v1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-subscription",
			Namespace: "default",
		},
		Spec: SubscriptionSpec{
			Filter: Filter{
				Dialect: "silesian",
				Filters: []EventFilter{
					{
						EventSource: EventFilterProperty{
							Property: "test",
							Value:    "test2",
						},
						EventType: EventFilterProperty{
							Property: "test3",
							Value:    "test4",
						},
					},
				},
			},
			Protocol: "tcp",
			ProtocolSettings: &ProtocolSettings{
				ContentMode:     "lol",
				ExemptHandshake: false,
				Qos:             "lol",
				WebhookAuth: WebhookAuth{
					ClientID:     "123",
					ClientSecret: "test",
					GrantType:    "lol",
					Scope: []string{
						"wtf",
					},
					Type:     "t123",
					TokenURL: "localhost123",
				},
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
