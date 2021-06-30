package jobmanager

import (
	"context"
	"errors"
	"testing"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/require"
	versioned "istio.io/client-go/pkg/clientset/versioned/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLoggingJobs(t *testing.T) {
	t.Run("should increase PVC size", func(t *testing.T) {
		resetFinishedJobsMap()
		setLogger(logger.NewLogger(false))

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
		ic := versioned.NewSimpleClientset()

		config := &installConfig.Config{
			WorkersCount: 1,
		}
		patchErr := increaseLoggingPvcSize{}.execute(config, kubeClient, ic, context.TODO())

		pvcReturn, _ := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvc, metav1.GetOptions{})
		pvcStorageSize := pvcReturn.Spec.Resources.Requests.Storage().String()

		require.Equal(t, "30Gi", pvcStorageSize)
		require.NoError(t, patchErr)
	})

	t.Run("should catch StatefulSet does not exists", func(t *testing.T) {
		resetFinishedJobsMap()
		setLogger(logger.NewLogger(false))

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
		ic := versioned.NewSimpleClientset()

		config := &installConfig.Config{
			WorkersCount: 1,
		}

		err := increaseLoggingPvcSize{}.execute(config, kubeClient, ic, context.TODO())

		require.Error(t, errors.New("statefulsets.apps \"logging-loki\" not found"), err)
	})

	t.Run("should catch PVC does not exists", func(t *testing.T) {
		resetFinishedJobsMap()
		setLogger(logger.NewLogger(false))

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
		ic := versioned.NewSimpleClientset()

		config := &installConfig.Config{
			WorkersCount: 1,
		}

		err := increaseLoggingPvcSize{}.execute(config, kubeClient, ic, context.TODO())

		require.Error(t, errors.New("persistentvolumeclaims \"storage-logging-loki-0\" not found"), err)
	})

}
