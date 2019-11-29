package k8s

import (
	"io/ioutil"
	"testing"

	"github.com/kyma-incubator/hydroform/install/scheme"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_ParseYaml(t *testing.T) {

	decoder, err := scheme.DefaultDecoder()
	require.NoError(t, err)

	t.Run("should parse valid yaml file", func(t *testing.T) {
		// given
		yamlBytes, err := ioutil.ReadFile("testdata/k8s-resources.yaml")
		require.NoError(t, err)
		yamlContent := string(yamlBytes)

		// when
		k8sObjects, err := ParseYamlToK8sObjects(decoder, yamlContent)

		// then
		require.NoError(t, err)

		require.Equal(t, 9, len(k8sObjects))

		assertK8sObject(t, k8sObjects[0], schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition"})
		assertK8sObject(t, k8sObjects[1], schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"})
		assertK8sObject(t, k8sObjects[2], schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "ClusterRoleBinding"})
		assertK8sObject(t, k8sObjects[3], schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "Deployment"})
		assertK8sObject(t, k8sObjects[4], schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"})
		assertK8sObject(t, k8sObjects[5], schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"})
		assertK8sObject(t, k8sObjects[6], schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"})
		assertK8sObject(t, k8sObjects[7], schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"})
		assertK8sObject(t, k8sObjects[8], schema.GroupVersionKind{Group: "installer.kyma-project.io", Version: "v1alpha1", Kind: "Installation"})
	})

	t.Run("should return error if invalid k8s object in file", func(t *testing.T) {
		// given
		yamlContent := `apiVersion: v1
metadata:
  name: service-account
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: cluster-role-binding
subjects:
  - kind: ServiceAccount
    name: service-account
    namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
---`

		// when
		k8sObjects, err := ParseYamlToK8sObjects(decoder, yamlContent)

		// then
		require.Error(t, err)
		assert.Nil(t, k8sObjects)
	})

}

func assertK8sObject(t *testing.T, k8sObject K8sObject, gvk schema.GroupVersionKind) {
	assert.NotNil(t, k8sObject.Object)

	assert.Equal(t, gvk.Kind, k8sObject.GVK.Kind)
	assert.Equal(t, gvk.Version, k8sObject.GVK.Version)
	assert.Equal(t, gvk.Group, k8sObject.GVK.Group)
}
