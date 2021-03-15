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
	"kymaComponent":    "test",
	"kymaProfile":      "profile",
	"kymaVersion":      "123",
	"kymaOperationID":  "opsid",
	"kymaCreationTime": "1615831194"}

var expectedStruct = &KymaMetadata{
	Component:    "test",
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
					Namespace: "default",
					Labels:    expectedLabels,
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		metadata, err := metaProv.Get("test")
		require.NoError(t, err)
		require.Equal(t, metadata, expectedStruct)
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
		require.Equal(t, err.Error(), (&kymaMetadataUnavailableError{secret: "sh.helm.release.v1.test.v1"}).Error())
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
	t.Run("Happy path", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset(
			&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sh.helm.release.v1.test.v1",
					Namespace: "default",
					Labels:    make(map[string]string),
				},
			},
		)
		metaProv := NewKymaMetadataProvider(k8sMock)
		err := metaProv.Set((&release.Release{Name: "test", Namespace: "default", Version: 1}), expectedStruct)
		require.NoError(t, err)
		require.Equal(t, expectedLabels, k8sMock.Fake.Actions()[1].(k8st.UpdateAction).GetObject().(*v1.Secret).GetObjectMeta().GetLabels())
	})

	t.Run("Release not found", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset()
		metaProv := NewKymaMetadataProvider(k8sMock)
		err := metaProv.Set((&release.Release{Name: "test", Namespace: "default", Version: 1}), (&KymaMetadata{}))
		require.Error(t, err)
		require.Equal(t, err.Error(), (&helmReleaseNotFoundError{name: "sh.helm.release.v1.test.v1"}).Error())
	})
}

func Test_Version(t *testing.T) {
	t.Run("No Kyma installed", func(t *testing.T) {
		k8sMock := fake.NewSimpleClientset()
		metaProv := NewKymaMetadataProvider(k8sMock)
		versions, err := metaProv.Version()
		require.NoError(t, err)
		require.Empty(t, versions)
	})
}
