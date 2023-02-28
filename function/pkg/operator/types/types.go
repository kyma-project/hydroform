package types

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GVRFunction = schema.GroupVersionResource{
		Group:    "serverless.kyma-project.io",
		Version:  "v1alpha2",
		Resource: "functions",
	}
	GVRSubscriptionV1alpha1 = schema.GroupVersionResource{
		Group:    "eventing.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "subscriptions",
	}
	GVRSubscriptionV1alpha2 = schema.GroupVersionResource{
		Group:    "eventing.kyma-project.io",
		Version:  "v1alpha2",
		Resource: "subscriptions",
	}
	GVRApiRule = schema.GroupVersionResource{
		Group:    "gateway.kyma-project.io",
		Version:  "v1beta1",
		Resource: "apirules",
	}
)
