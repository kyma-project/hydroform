package types

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretClient interface {
	Update(ctx context.Context, configMap *v1.Secret, opts metav1.UpdateOptions) (*v1.Secret, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Secret, error)
}
