package k8s

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/install/k8s/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	namespace               = "test"
	labelSelector           = "name=test"
	waitForPodTimeout       = 200 * time.Millisecond
	waitForPodCheckInterval = 20 * time.Millisecond
	waitForLabelChange      = 100 * time.Millisecond
)

var (
	podLabel = map[string]string{"name": "test"}
)

func TestGenericClient_WaitForPodByLabel(t *testing.T) {

	t.Run("should return nil if pod exists with correct label", func(t *testing.T) {
		// given
		existingPods := []runtime.Object{
			&v1.Pod{
				ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: namespace, Labels: podLabel},
				Status:     v1.PodStatus{Phase: v1.PodRunning},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingPods...)

		client := NewGenericClient(nil, nil, k8sClientSet, nil)

		// when
		err := client.WaitForPodByLabel(namespace, labelSelector, v1.PodRunning, waitForPodTimeout, waitForPodCheckInterval)

		// then
		require.NoError(t, err)
	})

	t.Run("should return nil if pod changed its label to correct one", func(t *testing.T) {
		// given
		existingPods := []runtime.Object{
			&v1.Pod{
				ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: namespace, Labels: podLabel},
				Status:     v1.PodStatus{Phase: v1.PodPending},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingPods...)
		podsClient := k8sClientSet.CoreV1().Pods(namespace)

		go func() {
			testPod, err := podsClient.Get("test", v12.GetOptions{})
			require.NoError(t, err)

			time.Sleep(waitForLabelChange)

			testPod.Status = v1.PodStatus{Phase: v1.PodRunning}
			_, err = podsClient.Update(testPod)
			require.NoError(t, err)
		}()

		client := NewGenericClient(nil, nil, k8sClientSet, nil)

		// when
		err := client.WaitForPodByLabel(namespace, labelSelector, v1.PodRunning, waitForPodTimeout, waitForPodCheckInterval)

		// then
		require.NoError(t, err)

		pod, err := podsClient.Get("test", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, v1.PodRunning, pod.Status.Phase)
	})

	t.Run("should return error if pod does not exist", func(t *testing.T) {
		// given
		k8sClientSet := fake.NewSimpleClientset([]runtime.Object{}...)

		client := NewGenericClient(nil, nil, k8sClientSet, nil)

		// when
		err := client.WaitForPodByLabel(namespace, labelSelector, v1.PodRunning, waitForPodTimeout, waitForPodCheckInterval)

		// then
		require.Error(t, err)
	})

	t.Run("should return error if pod does not have correct label", func(t *testing.T) {
		// given
		existingPods := []runtime.Object{
			&v1.Pod{
				ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: namespace, Labels: podLabel},
				Status:     v1.PodStatus{Phase: v1.PodPending},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingPods...)

		client := NewGenericClient(nil, nil, k8sClientSet, nil)

		// when
		err := client.WaitForPodByLabel(namespace, labelSelector, v1.PodRunning, waitForPodTimeout, waitForPodCheckInterval)

		// then
		require.Error(t, err)
	})
}

func TestGenericClient_ApplyConfigMaps(t *testing.T) {

	t.Run("should apply config maps", func(t *testing.T) {
		// given
		existingCMs := []runtime.Object{
			&v1.ConfigMap{
				ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
				Data:       map[string]string{"key1": "value1"},
			},
		}

		cmsToApply := []*v1.ConfigMap{
			{
				ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
				Data:       map[string]string{"key1": "value1", "key2": "value2"},
			},
			{
				ObjectMeta: v12.ObjectMeta{Name: "test2", Namespace: namespace},
				Data:       map[string]string{"key1": "value1"},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingCMs...)

		client := NewGenericClient(nil, nil, k8sClientSet, nil)

		// when
		err := client.ApplyConfigMaps(cmsToApply, namespace)

		// then
		require.NoError(t, err)

		cmClient := k8sClientSet.CoreV1().ConfigMaps(namespace)
		cm, err := cmClient.Get("test1", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, cmsToApply[0].Data, cm.Data)
		cm2, err := cmClient.Get("test2", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, cmsToApply[1].Data, cm2.Data)
	})
}

func TestGenericClient_ApplySecrets(t *testing.T) {

	t.Run("should apply config maps", func(t *testing.T) {
		// given
		existingSecrets := []runtime.Object{
			&v1.Secret{
				ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
				Data:       map[string][]byte{"key1": []byte("value1")},
			},
		}

		secretsToApply := []*v1.Secret{
			{
				ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
				Data:       map[string][]byte{"key1": []byte("value1"), "key2": []byte("value2")},
			},
			{
				ObjectMeta: v12.ObjectMeta{Name: "test2", Namespace: namespace},
				Data:       map[string][]byte{"key1": []byte("value1")},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingSecrets...)

		client := NewGenericClient(nil, nil, k8sClientSet, nil)

		// when
		err := client.ApplySecrets(secretsToApply, namespace)

		// then
		require.NoError(t, err)

		secretClient := k8sClientSet.CoreV1().Secrets(namespace)
		secret, err := secretClient.Get("test1", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, secretsToApply[0].Data, secret.Data)
		secret2, err := secretClient.Get("test2", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, secretsToApply[1].Data, secret2.Data)
	})
}

func TestGenericClient_ApplyResources(t *testing.T) {

	t.Run("should return an error if RESTMapper fails", func(t *testing.T) {
		// given
		resourcesToApply := []K8sObject{
			{
				Object: &v1.Service{
					ObjectMeta: v12.ObjectMeta{Name: "test2", Namespace: namespace},
					Spec:       v1.ServiceSpec{ExternalName: "test2"},
				},
				GVK: &schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
			},
		}

		restMapper := &mocks.RESTMapper{}
		restMapper.On("RESTMapping", schema.GroupKind{Group: "", Kind: "Service"}, "v1").Return(nil, fmt.Errorf("some error"))

		resourcesScheme, err := DefaultScheme()
		require.NoError(t, err)
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesScheme)

		k8sClientSet := fake.NewSimpleClientset()

		client := NewGenericClient(restMapper, dynamicClient, k8sClientSet, nil)

		// when
		err = client.ApplyResources(resourcesToApply)

		// then
		require.Error(t, err)
	})
}
