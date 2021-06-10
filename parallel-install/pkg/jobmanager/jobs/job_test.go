package jobmanager

import (
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestJob(t *testing.T) {
	k8sMock := fake.NewSimpleClientset(
		&v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "storage-logging-loki-0",
				Namespace: "kyma-system",
			},
			Spec: v1.PersistentVolumeClaimSpec{
				StorageClassName: "default",
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						Storage: "1Gi",
					},
				},
			},
		})
	job := job1{}
	err := job.execute(&config.Config{}, k8sMock)
	require.NoError(t, err)
}
