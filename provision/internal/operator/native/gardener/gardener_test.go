package gardener

import (
	"context"
	"testing"
	"time"

	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerFake "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sTesting "k8s.io/client-go/testing"
)

const (
	testNamespace  = "someNamespace"
	testShootsName = "someCluster"
)

type testCase struct {
	name        string
	shootObject func(name string, namespace string) *gardenerTypes.Shoot
	assertErr   func(*testing.T, error)
}

func TestWaitForShoot(t *testing.T) {
	t.Parallel()
	tests := []testCase{
		{
			name:        "With proper LastOperation",
			shootObject: stubForShootWithLastOperation(100, gardenerTypes.LastOperationStateSucceeded),
			assertErr:   func(t *testing.T, err error) { require.NoError(t, err) },
		},
		{
			name:        "Without LastOperation",
			shootObject: stubForShootWithoutLastOperation(),
			assertErr: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Equal(t, "Provisioning timed out", err.Error())
			},
		},
	}

	for _, tst := range tests {
		func(tcase testCase) {
			t.Run(tcase.name,
				func(t *testing.T) {
					t.Parallel()
					//given
					reactor := func(action k8sTesting.Action) (bool, runtime.Object, error) {
						getAction := action.(k8sTesting.GetActionImpl)
						testShoot := tcase.shootObject(getAction.Name, getAction.Namespace)
						return true, testShoot, nil
					}

					f := &k8sTesting.Fake{}
					f.AddReactor("get", "shoots", reactor)

					ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Millisecond)
					defer cancelFunc()
					fakeShootsGetter := gardenerFake.FakeCoreV1beta1{
						Fake: f,
					}

					//when
					err := waitForShoot(ctx, &fakeShootsGetter, testShootsName, testNamespace, 3*time.Millisecond)

					//then
					if tcase.assertErr != nil {
						tcase.assertErr(t, err)
					}
				})
		}(tst)
	}
}

func stubForShootWithLastOperation(progress int32, state gardenerTypes.LastOperationState) func(name, namespace string) *gardenerTypes.Shoot {

	return func(name, namespace string) *gardenerTypes.Shoot {
		res := stubForShootWithoutLastOperation()(name, namespace)
		res.Status.LastOperation = &gardenerTypes.LastOperation{
			Progress: progress,
			State:    state,
		}
		return res
	}
}

func stubForShootWithoutLastOperation() func(name, namespace string) *gardenerTypes.Shoot {
	return func(name, namespace string) *gardenerTypes.Shoot {
		return &gardenerTypes.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: gardenerTypes.ShootStatus{
				LastOperation: nil,
			},
		}
	}

}
