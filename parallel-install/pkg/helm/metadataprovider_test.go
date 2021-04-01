package helm

import (
	"encoding/json"
	"fmt"
	"testing"

	"encoding/base64"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8st "k8s.io/client-go/testing"
)

var b64 = base64.StdEncoding

var expectedKymaLabels = map[string]string{
	KymaLabelPrefix + "name":         "test",
	KymaLabelPrefix + "namespace":    "testNs",
	KymaLabelPrefix + "component":    "true",
	KymaLabelPrefix + "profile":      "profile",
	KymaLabelPrefix + "version":      "123",
	KymaLabelPrefix + "operationID":  "opsid",
	KymaLabelPrefix + "creationTime": "1615831194",
	KymaLabelPrefix + "priority":     "1",
	KymaLabelPrefix + "prerequisite": "false"}

var expectedKymaCompMetadata = &KymaComponentMetadata{
	Name:         "test",
	Namespace:    "testNs",
	Component:    true,
	Version:      "123",
	Profile:      "profile",
	OperationID:  "opsid",
	CreationTime: int64(1615831194),
	Priority:     int64(1),
}

var kymaCompMetaTpl = &KymaComponentMetadataTemplate{
	Component:    true,
	Version:      "123",
	Profile:      "profile",
	OperationID:  "opsid",
	CreationTime: int64(1615831194),
}

func marshalRelease(rls *release.Release) []byte {
	b, err := json.Marshal(rls)
	if err != nil {
		panic(err)
	}
	result := make([]byte, b64.EncodedLen(len(b)))
	b64.Encode(result, b)
	return result
}

func addHelmLabels(labels map[string]string) map[string]string {
	//copy expected labels
	result := make(map[string]string, len(expectedKymaLabels)+2)
	for k, v := range labels {
		result[k] = v
	}
	result["owner"] = "helm"
	result["name"] = "test"
	return result
}

func Test_MetadataGet(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		//prepare mock
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.v1",
					Namespace: "testNs",
					Labels:    addHelmLabels(expectedKymaLabels),
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "test",
						Version:   1,
						Namespace: "testNs",
					}),
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		metadata, err := metaProv.Get("test")
		require.NoError(t, err)
		require.Equal(t, metadata, expectedKymaCompMetadata)
	})

	t.Run("No Helm release found", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset()
		metaProv := NewKymaMetadataProvider(k8sMock)
		_, err := metaProv.Get("test")
		require.Error(t, err)
		require.Equal(t, "release: not found", err.Error())
	})

	t.Run("No Metadata found for release release found", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.v1",
					Namespace: "default",
					Labels:    map[string]string{"owner": "helm", "name": "test"},
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "test",
						Version:   1,
						Namespace: "default",
					}),
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		_, err := metaProv.Get("test")
		require.Error(t, err)
		require.IsType(t, (&kymaMetadataUnavailableError{secret: "sh.helm.release.v1.test.v1", err: err}), err)
	})

	t.Run("Release version not found", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.testx.v1",
					Namespace: "default",
					Labels:    map[string]string{"owner": "helm", "name": "testx"},
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "testx",
						Version:   1,
						Namespace: "default",
					}),
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		_, err := metaProv.Get("test")
		require.Error(t, err)
		require.Equal(t, "release: not found", err.Error())
	})

}

func Test_MetadataSet(t *testing.T) {
	t.Run("Happy path - for components", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.v1",
					Namespace: "testNs",
					Labels:    make(map[string]string),
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		err := metaProv.Set((&release.Release{Name: "test", Namespace: "testNs", Version: 1}), kymaCompMetaTpl.ForComponents())
		require.NoError(t, err)
		require.Equal(t, expectedKymaLabels, k8sMock.Fake.Actions()[1].(k8st.UpdateAction).GetObject().(*v1.Secret).GetObjectMeta().GetLabels())
	})

	t.Run("Happy path - for prerequisites", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.v1",
					Namespace: "testNs",
					Labels:    make(map[string]string),
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		//test for prerequisites
		err := metaProv.Set((&release.Release{Name: "test", Namespace: "testNs", Version: 1}), kymaCompMetaTpl.ForPrerequisites())
		require.NoError(t, err)

		//align expected values
		expectedKymaLabelsCopy := make(map[string]string, len(expectedKymaLabels))
		for k, v := range expectedKymaLabels {
			expectedKymaLabelsCopy[k] = v
		}
		expectedKymaLabelsCopy[KymaLabelPrefix+"priority"] = "2"
		expectedKymaLabelsCopy[KymaLabelPrefix+"prerequisite"] = "true"

		require.Equal(t, expectedKymaLabelsCopy, k8sMock.Fake.Actions()[1].(k8st.UpdateAction).GetObject().(*v1.Secret).GetObjectMeta().GetLabels())
	})

	t.Run("Release not found", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset()
		metaProv := NewKymaMetadataProvider(k8sMock)
		err := metaProv.Set((&release.Release{Name: "test", Namespace: "default", Version: 1}), (&KymaComponentMetadataTemplate{}))
		require.Error(t, err)
		require.Equal(t, err.Error(), (&helmReleaseNotFoundError{name: "sh.helm.release.v1.test.v1"}).Error())
	})
}

func Test_Versions(t *testing.T) {
	t.Run("No Kyma installed", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset()
		metaProv := NewKymaMetadataProvider(k8sMock)
		versionSet, err := metaProv.Versions()
		require.NoError(t, err)
		require.Equal(t, 0, versionSet.Count())
	})
	t.Run("One version of Kyma installed", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.v1",
					Namespace: "testNs",
					Labels:    addHelmLabels(expectedKymaLabels),
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "test",
						Version:   1,
						Namespace: "testNs",
						Info: &release.Info{
							Status: release.StatusDeployed,
						},
					}),
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		versionSet, err := metaProv.Versions()
		require.NoError(t, err)
		require.Equal(t, 1, len(versionSet.Versions))
		expectedVersions := []*KymaVersion{
			{
				Version:      "123",
				Profile:      "profile",
				OperationID:  "opsid",
				CreationTime: 1615831194,
				Components: []*KymaComponentMetadata{
					expectedKymaCompMetadata,
				},
			},
		}
		require.Equal(t, expectedVersions, versionSet.Versions)
	})
	t.Run("Different versions of Kyma installed", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{ //installed from "master" by operation "aaa:1000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.v1",
					Namespace: "test",
					Labels: addHelmLabels(map[string]string{
						KymaLabelPrefix + "name":         "test",
						KymaLabelPrefix + "namespace":    "test",
						KymaLabelPrefix + "component":    "true",
						KymaLabelPrefix + "profile":      "profile",
						KymaLabelPrefix + "version":      "master",
						KymaLabelPrefix + "operationID":  "aaa",
						KymaLabelPrefix + "creationTime": "1000000000",
						KymaLabelPrefix + "priority":     "1"}),
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "test",
						Version:   1,
						Namespace: "test",
						Info: &release.Info{
							Status: release.StatusDeployed,
						},
					}),
				},
			},
			&v1.Secret{ //installed from "master" by operation "aaa:1000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test2.v2",
					Namespace: "test2",
					Labels: addHelmLabels(map[string]string{
						KymaLabelPrefix + "name":         "test2",
						KymaLabelPrefix + "namespace":    "test2",
						KymaLabelPrefix + "component":    "true",
						KymaLabelPrefix + "profile":      "profile",
						KymaLabelPrefix + "version":      "master",
						KymaLabelPrefix + "operationID":  "aaa",
						KymaLabelPrefix + "creationTime": "1000000000",
						KymaLabelPrefix + "priority":     "2"}),
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "test2",
						Version:   2,
						Namespace: "test2",
						Info: &release.Info{
							Status: release.StatusDeployed,
						},
					}),
				},
			},
			&v1.Secret{ //installed from "2.0.0" release by operation "bbb:2000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test1.v1",
					Namespace: "test1",
					Labels: addHelmLabels(map[string]string{
						KymaLabelPrefix + "name":         "test1",
						KymaLabelPrefix + "namespace":    "test1",
						KymaLabelPrefix + "component":    "true",
						KymaLabelPrefix + "profile":      "evaluation",
						KymaLabelPrefix + "version":      "2.0.0",
						KymaLabelPrefix + "operationID":  "bbb",
						KymaLabelPrefix + "creationTime": "2000000000",
						KymaLabelPrefix + "priority":     "1"}),
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "test1",
						Version:   1,
						Namespace: "test1",
						Info: &release.Info{
							Status: release.StatusSuperseded,
						},
					}),
				},
			},
			&v1.Secret{ //installed (upgrade) from "2.0.1" release by operation "ccc:3000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test1.v2",
					Namespace: "test1",
					Labels: addHelmLabels(map[string]string{
						KymaLabelPrefix + "name":         "test1",
						KymaLabelPrefix + "namespace":    "test1",
						KymaLabelPrefix + "component":    "true",
						KymaLabelPrefix + "profile":      "production",
						KymaLabelPrefix + "version":      "2.0.1",
						KymaLabelPrefix + "operationID":  "ccc",
						KymaLabelPrefix + "creationTime": "3000000000",
						KymaLabelPrefix + "priority":     "1"}),
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "test1",
						Version:   2,
						Namespace: "test1",
						Info: &release.Info{
							Status: release.StatusDeployed,
						},
					}),
				},
			},
			&v1.Secret{ //installed from "master" by operation "ddd:4000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test3.v1",
					Namespace: "test3",
					Labels: addHelmLabels(map[string]string{
						KymaLabelPrefix + "name":         "test3",
						KymaLabelPrefix + "namespace":    "test3",
						KymaLabelPrefix + "component":    "true",
						KymaLabelPrefix + "profile":      "profile",
						KymaLabelPrefix + "version":      "master",
						KymaLabelPrefix + "operationID":  "ddd",
						KymaLabelPrefix + "creationTime": "4000000000",
						KymaLabelPrefix + "priority":     "1"}),
				},
				Data: map[string][]byte{
					"release": marshalRelease(&release.Release{
						Name:      "test3",
						Version:   1,
						Namespace: "test3",
						Info: &release.Info{
							Status: release.StatusDeployed,
						},
					}),
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		versionSet, err := metaProv.Versions()
		require.NoError(t, err)
		require.Equal(t, 3, len(versionSet.Versions))
		expectedVersions := []*KymaVersion{
			{
				Version:      "master",
				Profile:      "profile",
				OperationID:  "aaa",
				CreationTime: 1000000000,
				Components: []*KymaComponentMetadata{
					{
						Name:         "test",
						Namespace:    "test",
						Component:    true,
						Profile:      "profile",
						Version:      "master",
						OperationID:  "aaa",
						CreationTime: int64(1000000000),
						Priority:     int64(1),
					},
					{
						Name:         "test2",
						Namespace:    "test2",
						Component:    true,
						Profile:      "profile",
						Version:      "master",
						OperationID:  "aaa",
						CreationTime: int64(1000000000),
						Priority:     int64(2),
					},
				},
			},
			{
				Version:      "2.0.1",
				Profile:      "production",
				OperationID:  "ccc",
				CreationTime: 3000000000,
				Components: []*KymaComponentMetadata{
					{
						Name:         "test1",
						Namespace:    "test1",
						Component:    true,
						Profile:      "production",
						Version:      "2.0.1",
						OperationID:  "ccc",
						CreationTime: int64(3000000000),
						Priority:     int64(1),
					},
				},
			},
			{
				Version:      "master",
				Profile:      "profile",
				OperationID:  "ddd",
				CreationTime: 4000000000,
				Components: []*KymaComponentMetadata{
					{
						Name:         "test3",
						Namespace:    "test3",
						Component:    true,
						Profile:      "profile",
						Version:      "master",
						OperationID:  "ddd",
						CreationTime: int64(4000000000),
						Priority:     int64(1),
					},
				},
			},
		}
		//compare different versions (distinguished by their operationID)
		versionMap := make(map[string]*KymaVersion, len(versionSet.Versions))
		for _, version := range versionSet.Versions {
			versionMap[version.OperationID] = version
		}
		for _, kymaVersion := range expectedVersions {
			version, ok := versionMap[kymaVersion.OperationID]
			require.True(t, ok, fmt.Sprintf("Version with name is '%s' missing in version set", kymaVersion.Version))
			require.Equal(t, kymaVersion.Version, version.Version)
			require.Equal(t, len(kymaVersion.Components), len(version.Components))
			for _, comp := range kymaVersion.Components {
				require.Equal(t, comp.Name, comp.Name)
				require.Equal(t, comp.Priority, comp.Priority)
				require.Equal(t, comp.CreationTime, comp.CreationTime)
			}
		}
	})
}
