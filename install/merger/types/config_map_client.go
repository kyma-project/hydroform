package types

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapClient interface {
	Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ConfigMap, error)
}
