package jobmanager

import (
	"context"
	"errors"
	"testing"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLoggingJobs(t *testing.T) {
	t.Run("should increase PVC size", func(t *testing.T) {
		requestedBytes := 100
		namespace := "kyma-system"
		pvc := "storage-logging-loki-0"
		statefuleset := "logging-loki"

		kubeClient := fake.NewSimpleClientset(
			&v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvc,
					Namespace: namespace,
					UID:       "testid",
				},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: pvc,
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceName(v1.ResourceStorage): *resource.NewQuantity(int64(requestedBytes), resource.BinarySI),
						},
					},
				},
				Status: v1.PersistentVolumeClaimStatus{
					Phase: v1.ClaimBound,
				},
			},
			&appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      statefuleset,
					Namespace: namespace,
				},
				Spec:   appsv1.StatefulSetSpec{},
				Status: appsv1.StatefulSetStatus{},
			})

		config := &installConfig.Config{
			WorkersCount: 1,
		}
		patchErr := increaseLoggingPvcSize{}.execute(config, kubeClient, context.TODO())
		// logs := getLogs(observedLogs)

		pvcReturn, _ := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvc, metav1.GetOptions{})
		pvcStorageSize := pvcReturn.Spec.Resources.Requests.Storage().String()

		require.Equal(t, "50Gi", pvcStorageSize)
		require.NoError(t, patchErr)
	})

	t.Run("should catch StatefulSet does not exists", func(t *testing.T) {
		namespace := "kyma-system"
		pvc := "storage-logging-loki-0"

		kubeClient := fake.NewSimpleClientset(
			&v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvc,
					Namespace: namespace,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: pvc,
				},
				Status: v1.PersistentVolumeClaimStatus{
					Phase: v1.ClaimBound,
				},
			})

		config := &installConfig.Config{
			WorkersCount: 1,
		}

		err := increaseLoggingPvcSize{}.execute(config, kubeClient, context.TODO())

		logs := getLogs(observedLogs)
		require.Error(t, errors.New("statefulsets.apps \"logging-loki\" not found"), err)
		require.Contains(t, logs, "Error deleting StatefulSet: statefulsets.apps \"logging-loki\" not found")
	})

	t.Run("should catch PVC does not exists", func(t *testing.T) {
		namespace := "kyma-system"
		statefuleset := "logging-loki"

		kubeClient := fake.NewSimpleClientset(
			&appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      statefuleset,
					Namespace: namespace,
				},
				Spec:   appsv1.StatefulSetSpec{},
				Status: appsv1.StatefulSetStatus{},
			})

		config := &installConfig.Config{
			WorkersCount: 1,
		}

		err := increaseLoggingPvcSize{}.execute(config, kubeClient, context.TODO())

		logs := getLogs(observedLogs)
		require.Error(t, errors.New("persistentvolumeclaims \"storage-logging-loki-0\" not found"), err)
		require.Contains(t, logs, "PVC storage-logging-loki-0 in namespace kyma-system not found -> Skipping Job")
	})

}
