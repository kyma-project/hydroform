package jobmanager

import (
	"context"
	"errors"
	"testing"

	installConfig "github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLoggingJobs(t *testing.T) {
	t.Run("should catch StatefulSet does not exists", func(t *testing.T) {

		kubeClient := fake.NewSimpleClientset(
			&v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "storage-logging-loki-0",
					Namespace: "kyma-system",
				},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: "storage-logging-loki-0",
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
		kubeClient := fake.NewSimpleClientset(
			&appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "logging-loki",
					Namespace: "kyma-system",
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

	t.Run("should increase PVC size", func(t *testing.T) {
		kubeClient := fake.NewSimpleClientset(
			&v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "storage-logging-loki-0",
					Namespace: "kyma-system",
				},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: "storage-logging-loki-0",
				},
				Status: v1.PersistentVolumeClaimStatus{
					Phase: v1.ClaimBound,
				},
			},
			&appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "logging-loki",
					Namespace: "kyma-system",
				},
				Spec:   appsv1.StatefulSetSpec{},
				Status: appsv1.StatefulSetStatus{},
			})

		// TODO
		config := &installConfig.Config{
			WorkersCount: 1,
		}

		err := increaseLoggingPvcSize{}.execute(config, kubeClient, context.TODO())

		logs := getLogs(observedLogs)
		require.Error(t, errors.New("persistentvolumeclaims \"storage-logging-loki-0\" not found"), err)
		require.Contains(t, logs, "PVC storage-logging-loki-0 in namespace kyma-system not found -> Skipping Job")
	})

}
