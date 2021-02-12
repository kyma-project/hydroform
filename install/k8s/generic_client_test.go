package k8s

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/hydroform/install/util"
	"k8s.io/apimachinery/pkg/api/meta"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/hydroform/install/scheme"

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

	t.Run("should return nil if pod exists with correct status", func(t *testing.T) {
		// given
		existingPods := []runtime.Object{
			&v1.Pod{
				ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: namespace, Labels: podLabel},
				Status:     v1.PodStatus{Phase: v1.PodRunning},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingPods...)

		client := NewGenericClient(nil, nil, k8sClientSet)

		// when
		err := client.WaitForPodByLabel(namespace, labelSelector, v1.PodRunning, waitForPodTimeout, waitForPodCheckInterval)

		// then
		require.NoError(t, err)
	})

	t.Run("should return nil if pod changed its status to correct one", func(t *testing.T) {
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
			testPod, err := podsClient.Get(context.Background(), "test", v12.GetOptions{})
			require.NoError(t, err)

			time.Sleep(waitForLabelChange)

			testPod.Status = v1.PodStatus{Phase: v1.PodRunning}
			_, err = podsClient.Update(context.Background(), testPod, v12.UpdateOptions{})
			require.NoError(t, err)
		}()

		client := NewGenericClient(nil, nil, k8sClientSet)

		// when
		err := client.WaitForPodByLabel(namespace, labelSelector, v1.PodRunning, waitForPodTimeout, waitForPodCheckInterval)

		// then
		require.NoError(t, err)

		pod, err := podsClient.Get(context.Background(), "test", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, v1.PodRunning, pod.Status.Phase)
	})

	t.Run("should return error if pod does not exist", func(t *testing.T) {
		// given
		k8sClientSet := fake.NewSimpleClientset([]runtime.Object{}...)

		client := NewGenericClient(nil, nil, k8sClientSet)

		// when
		err := client.WaitForPodByLabel(namespace, labelSelector, v1.PodRunning, waitForPodTimeout, waitForPodCheckInterval)

		// then
		require.Error(t, err)
	})

	t.Run("should return error if pod does not have correct status", func(t *testing.T) {
		// given
		existingPods := []runtime.Object{
			&v1.Pod{
				ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: namespace, Labels: podLabel},
				Status:     v1.PodStatus{Phase: v1.PodPending},
			},
		}

		k8sClientSet := fake.NewSimpleClientset(existingPods...)

		client := NewGenericClient(nil, nil, k8sClientSet)

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

		client := NewGenericClient(nil, nil, k8sClientSet)

		// when
		err := client.ApplyConfigMaps(cmsToApply, namespace)

		// then
		require.NoError(t, err)

		cmClient := k8sClientSet.CoreV1().ConfigMaps(namespace)
		cm, err := cmClient.Get(context.Background(), "test1", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, cmsToApply[0].Data, cm.Data)
		cm2, err := cmClient.Get(context.Background(), "test2", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, cmsToApply[1].Data, cm2.Data)
	})

	tests := []struct {
		name     string
		existing map[string]string
		toApply  map[string]string
		labels   map[string]string
		assert   func(t *testing.T, changes, afterUpdate, beforeUpdated map[string]string)
	}{
		{
			"should replace ConfigMap if replace flag is specified",
			map[string]string{"key1": "value1"},
			map[string]string{"key2": "value2", "key3": "value3"},
			map[string]string{OnConflictLabel: ReplaceOnConflict},
			func(t *testing.T, changes, afterUpdate, _ map[string]string) {
				assert.Equal(t, changes, afterUpdate)
			},
		}, {
			"should merge ConfigMap if replace flag is not specified",
			map[string]string{"key1": "value1"},
			map[string]string{"key2": "value2", "key3": "value3"},
			map[string]string{},
			func(t *testing.T, changes, afterUpdate, beforeUpdated map[string]string) {
				assert.Equal(t, util.MergeStringMaps(changes, beforeUpdated), afterUpdate)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given

			existingCMs := []runtime.Object{
				&v1.ConfigMap{
					ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
					Data:       tt.existing,
				},
			}

			k8sClientSet := fake.NewSimpleClientset(existingCMs...)

			client := NewGenericClient(nil, nil, k8sClientSet)

			cmsToApply := []*v1.ConfigMap{
				{
					ObjectMeta: v12.ObjectMeta{
						Name:      "test1",
						Namespace: namespace,
						Labels:    tt.labels,
					},
					Data: tt.toApply,
				},
			}

			// when
			err := client.ApplyConfigMaps(cmsToApply, namespace)

			// then
			require.NoError(t, err)

			cmClient := k8sClientSet.CoreV1().ConfigMaps(namespace)
			cm, err := cmClient.Get(context.Background(), "test1", v12.GetOptions{})
			require.NoError(t, err)

			tt.assert(t, tt.toApply, tt.existing, cm.Data)
		})
	}
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

		client := NewGenericClient(nil, nil, k8sClientSet)

		// when
		err := client.ApplySecrets(secretsToApply, namespace)

		// then
		require.NoError(t, err)

		secretClient := k8sClientSet.CoreV1().Secrets(namespace)
		secret, err := secretClient.Get(context.Background(), "test1", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, secretsToApply[0].Data, secret.Data)
		secret2, err := secretClient.Get(context.Background(), "test2", v12.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, secretsToApply[1].Data, secret2.Data)
	})

	tests := []struct {
		name     string
		existing map[string][]byte
		toApply  map[string][]byte
		labels   map[string]string
		assert   func(t *testing.T, changes, afterUpdate, beforeUpdated map[string][]byte)
	}{
		{
			"should replace Secrets if replace flag is specified",
			map[string][]byte{"key1": []byte("value1")},
			map[string][]byte{"key2": []byte("value2"), "key3": []byte("value3")},
			map[string]string{OnConflictLabel: ReplaceOnConflict},
			func(t *testing.T, changes, afterUpdate, _ map[string][]byte) {
				assert.Equal(t, changes, afterUpdate)
			},
		}, {
			"should replace Secrets if replace flag is not specified",
			map[string][]byte{"key1": []byte("value1")},
			map[string][]byte{"key2": []byte("value2"), "key3": []byte("value3")},
			map[string]string{},
			func(t *testing.T, changes, afterUpdate, beforeUpdated map[string][]byte) {
				assert.Equal(t, util.MergeByteMaps(changes, beforeUpdated), afterUpdate)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			existingCMs := []runtime.Object{
				&v1.Secret{
					ObjectMeta: v12.ObjectMeta{Name: "test1", Namespace: namespace},
					Data:       tt.existing,
				},
			}

			mockClientSet := fake.NewSimpleClientset(existingCMs...)
			genericClient := NewGenericClient(nil, nil, mockClientSet)

			cmsToApply := []*v1.Secret{
				{
					ObjectMeta: v12.ObjectMeta{
						Name:      "test1",
						Namespace: namespace,
						Labels:    tt.labels,
					},
					Data: tt.existing,
				},
			}

			// when
			err := genericClient.ApplySecrets(cmsToApply, namespace)

			// then
			require.NoError(t, err)

			mockClientSet.CoreV1().RESTClient()
			client := mockClientSet.CoreV1().Secrets(namespace)
			secret, err := client.Get(context.Background(), "test1", v12.GetOptions{})
			require.NoError(t, err)

			tt.assert(t, tt.toApply, tt.existing, secret.Data)
		})
	}
}

func TestGenericClient_CreateResources(t *testing.T) {
	resourcesToApply := []K8sObject{
		{
			Object: &v1.Service{
				ObjectMeta: v12.ObjectMeta{Name: "test2", Namespace: namespace},
				Spec:       v1.ServiceSpec{ExternalName: "test2"},
			},
			GVK: &schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		},
	}

	t.Run("should create resource", func(t *testing.T) {
		//given
		restMapper := dummyRestMapper{}

		resourcesScheme, err := scheme.DefaultScheme()
		require.NoError(t, err)
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesScheme)

		k8sClientSet := fake.NewSimpleClientset()

		client := NewGenericClient(restMapper, dynamicClient, k8sClientSet)

		expectedObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resourcesToApply[0].Object)

		//when
		resources, err := client.CreateResources(resourcesToApply)

		//then
		require.NoError(t, err)
		require.Equal(t, expectedObj, resources[0].Object)
	})

	t.Run("should return an error if RESTMapper fails", func(t *testing.T) {
		restMapper := &mocks.RESTMapper{}
		restMapper.On("RESTMapping", schema.GroupKind{Group: "", Kind: "Service"}, "v1").Return(nil, fmt.Errorf("some error"))

		resourcesScheme, err := scheme.DefaultScheme()
		require.NoError(t, err)
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesScheme)

		k8sClientSet := fake.NewSimpleClientset()

		client := NewGenericClient(restMapper, dynamicClient, k8sClientSet)

		// when
		_, err = client.CreateResources(resourcesToApply)

		// then
		require.Error(t, err)
	})
}

func TestGenericClient_ApplyResources(t *testing.T) {
	resourcesToCreate := []K8sObject{
		{
			Object: &v1.Service{
				ObjectMeta: v12.ObjectMeta{Name: "test2", Namespace: namespace},
				Spec:       v1.ServiceSpec{ExternalName: "test2", ClusterIP: "1111", HealthCheckNodePort: 3000},
			},
			GVK: &schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		},
	}

	t.Run("should create resource when not exists", func(t *testing.T) {
		//given
		restMapper := dummyRestMapper{}

		resourcesScheme, err := scheme.DefaultScheme()
		require.NoError(t, err)
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesScheme)

		k8sClientSet := fake.NewSimpleClientset()

		client := NewGenericClient(restMapper, dynamicClient, k8sClientSet)

		expectedObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resourcesToCreate[0].Object)
		require.NoError(t, err)

		//when
		resources, err := client.ApplyResources(resourcesToCreate)

		//then
		require.NoError(t, err)
		require.Equal(t, expectedObj, resources[0].Object)
	})

	t.Run("should update resource when exists", func(t *testing.T) {
		//given
		restMapper := dummyRestMapper{}

		resourcesScheme, err := scheme.DefaultScheme()
		require.NoError(t, err)
		dynamicClient := dynamicFake.NewSimpleDynamicClient(resourcesScheme)

		k8sClientSet := fake.NewSimpleClientset()

		client := NewGenericClient(restMapper, dynamicClient, k8sClientSet)

		//when
		resources, err := client.CreateResources(resourcesToCreate)

		//then
		require.NoError(t, err)
		require.NotEmpty(t, resources)

		//given
		resourcesToApply := []K8sObject{
			{
				Object: &v1.Service{
					ObjectMeta: v12.ObjectMeta{Name: "test2", Namespace: namespace},
					Spec:       v1.ServiceSpec{ExternalName: "test2", ClusterIP: "2222"},
				},
				GVK: &schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
			},
		}

		oldResource, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resourcesToCreate[0].Object)
		require.NoError(t, err)
		newResource, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resourcesToApply[0].Object)
		require.NoError(t, err)

		expectedObj := util.MergeMaps(newResource, oldResource)

		//when
		appliedResources, err := client.ApplyResources(resourcesToApply)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedObj, appliedResources[0].Object)
	})
}

type dummyRestMapper struct{}

func (d dummyRestMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	if len(versions) < 1 {
		return nil, fmt.Errorf("no version provided")
	}

	return &meta.RESTMapping{
		Resource: schema.GroupVersionResource{
			Group:    gk.Group,
			Version:  versions[0],
			Resource: kindToResource(gk.Kind),
		},
		GroupVersionKind: schema.GroupVersionKind{
			Group:   gk.Group,
			Version: versions[0],
			Kind:    gk.Kind,
		},
		Scope: nil,
	}, nil
}

func kindToResource(kind string) string {
	return fmt.Sprintf("%ss", strings.ToLower(kind))
}
