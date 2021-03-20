package helm

import (
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/release"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8st "k8s.io/client-go/testing"
)

var expectedLabels = map[string]string{
	"kymaName":         "test",
	"kymaNamespace":    "testNs",
	"kymaComponent":    "true",
	"kymaProfile":      "profile",
	"kymaVersion":      "123",
	"kymaOperationID":  "opsid",
	"kymaCreationTime": "1615831194",
	"kymaPriority":     "1",
	"kymaPrerequisite": "false"}

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

func Test_MetadataGet(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.v1",
					Namespace: "testNs",
					Labels:    expectedLabels,
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
		require.Equal(t, err.Error(), (&helmReleaseNotFoundError{name: "test"}).Error())
	})

	t.Run("No Metadata found for release release found", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "sh.helm.release.v1.test.v1",
					Namespace:   "default",
					Annotations: map[string]string{"foo": "bar"},
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		_, err := metaProv.Get("test")
		require.Error(t, err)
		require.IsType(t, err, (&kymaMetadataUnavailableError{secret: "sh.helm.release.v1.test.v1", err: err}))
	})

	t.Run("Release version not found", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.vx",
					Namespace: "default",
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		_, err := metaProv.Get("test")
		require.Error(t, err)
		require.Equal(t, err.Error(), (&helmSecretNameInvalidError{secret: "sh.helm.release.v1.test.vx", namespace: "default"}).Error())
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
		require.Equal(t, expectedLabels, k8sMock.Fake.Actions()[1].(k8st.UpdateAction).GetObject().(*v1.Secret).GetObjectMeta().GetLabels())
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
		expectedLabelsCopy := make(map[string]string, len(expectedLabels))
		for k, v := range expectedLabels {
			expectedLabelsCopy[k] = v
		}
		expectedLabelsCopy["kymaPriority"] = "2"
		expectedLabelsCopy["kymaPrerequisite"] = "true"

		require.Equal(t, expectedLabelsCopy, k8sMock.Fake.Actions()[1].(k8st.UpdateAction).GetObject().(*v1.Secret).GetObjectMeta().GetLabels())
	})

	t.Run("Release not found", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset()
		metaProv := NewKymaMetadataProvider(k8sMock)
		err := metaProv.Set((&release.Release{Name: "test", Namespace: "default", Version: 1}), (&KymaComponentMetadataTemplate{}))
		require.Error(t, err)
		require.Equal(t, err.Error(), (&helmReleaseNotFoundError{name: "sh.helm.release.v1.test.v1"}).Error())
	})
}

func Test_Version(t *testing.T) {
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
					Labels:    expectedLabels,
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		versionSet, err := metaProv.Versions()
		require.NoError(t, err)
		require.Equal(t, 1, len(versionSet.Versions))
		expectedVersions := []*KymaVersion{
			&KymaVersion{
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
					Labels: map[string]string{
						"kymaName":         "test",
						"kymaNamespace":    "test",
						"kymaComponent":    "true",
						"kymaProfile":      "profile",
						"kymaVersion":      "master",
						"kymaOperationID":  "aaa",
						"kymaCreationTime": "1000000000",
						"kymaPriority":     "1"},
				},
			},
			&v1.Secret{ //installed from "master" by operation "aaa:1000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test2.v2",
					Namespace: "test2",
					Labels: map[string]string{
						"kymaName":         "test2",
						"kymaNamespace":    "test2",
						"kymaComponent":    "true",
						"kymaProfile":      "profile",
						"kymaVersion":      "master",
						"kymaOperationID":  "aaa",
						"kymaCreationTime": "1000000000",
						"kymaPriority":     "2"},
				},
			},
			&v1.Secret{ //installed from "2.0.0" release by operation "bbb:2000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test1.v1",
					Namespace: "test1",
					Labels: map[string]string{
						"kymaName":         "test1",
						"kymaNamespace":    "test1",
						"kymaComponent":    "true",
						"kymaProfile":      "evaluation",
						"kymaVersion":      "2.0.0",
						"kymaOperationID":  "bbb",
						"kymaCreationTime": "2000000000",
						"kymaPriority":     "1"},
				},
			},
			&v1.Secret{ //installed (upgrade) from "2.0.1" release by operation "ccc:3000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test1.v2",
					Namespace: "test1",
					Labels: map[string]string{
						"kymaName":         "test1",
						"kymaNamespace":    "test1",
						"kymaComponent":    "true",
						"kymaProfile":      "production",
						"kymaVersion":      "2.0.1",
						"kymaOperationID":  "ccc",
						"kymaCreationTime": "3000000000",
						"kymaPriority":     "1"},
				},
			},
			&v1.Secret{ //installed from "master" by operation "ddd:4000000000"
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test3.v1",
					Namespace: "test3",
					Labels: map[string]string{
						"kymaName":         "test3",
						"kymaNamespace":    "test3",
						"kymaComponent":    "true",
						"kymaProfile":      "profile",
						"kymaVersion":      "master",
						"kymaOperationID":  "ddd",
						"kymaCreationTime": "4000000000",
						"kymaPriority":     "1"},
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		versionSet, err := metaProv.Versions()
		require.NoError(t, err)
		require.Equal(t, 3, len(versionSet.Versions))
		expectedVersions := []*KymaVersion{
			&KymaVersion{
				Version:      "master",
				Profile:      "profile",
				OperationID:  "aaa",
				CreationTime: 1000000000,
				Components: []*KymaComponentMetadata{
					&KymaComponentMetadata{
						Name:         "test",
						Namespace:    "test",
						Component:    true,
						Profile:      "profile",
						Version:      "master",
						OperationID:  "aaa",
						CreationTime: int64(1000000000),
						Priority:     int64(1),
					},
					&KymaComponentMetadata{
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
			&KymaVersion{
				Version:      "2.0.1",
				Profile:      "production",
				OperationID:  "ccc",
				CreationTime: 3000000000,
				Components: []*KymaComponentMetadata{
					&KymaComponentMetadata{
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
			&KymaVersion{
				Version:      "master",
				Profile:      "profile",
				OperationID:  "ddd",
				CreationTime: 4000000000,
				Components: []*KymaComponentMetadata{
					&KymaComponentMetadata{
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
		for _, version := range versionSet.Versions {
			require.Contains(t, expectedVersions, version)
		}
	})
}
