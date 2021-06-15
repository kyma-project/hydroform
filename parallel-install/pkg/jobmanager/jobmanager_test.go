package jobmanager_test

import (
	"jobmanager"
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/jobmanager"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestJob(t *testing.T) {
	// k8sMock := fake.NewSimpleClientset(
	// &v1.PersistentVolumeClaim{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "storage-logging-loki-0",
	// 		Namespace: "kyma-system",
	// 	},
	// 	Spec: v1.PersistentVolumeClaimSpec{
	// 		StorageClassName: "default",
	// 		Resources: v1.ResourceRequirements{
	// 			Requests: v1.ResourceList{
	// 				Storage: "1Gi",
	// 			},
	// 		},
	// 		"storage-logging-loki-0",
	// 	},
	// })
	// t.Parallel()

	kubeClient := fake.NewSimpleClientset()
	// i := deployment.newDeployment(t, nil, kubeClient)

	// hc := &mockHelmClient{}
	// provider := &mockProvider{
	// 	hc: hc,
	// }
	// overridesProvider := &mockOverridesProvider{}
	// prerequisitesEng := engine.NewEngine(overridesProvider, provider, engine.Config{
	// 	WorkersCount: 1,
	// 	Log:          logger.NewLogger(true),
	// })
	// componentsEng := engine.NewEngine(overridesProvider, provider, engine.Config{
	// 	WorkersCount: 2,
	// 	Log:          logger.NewLogger(true),
	// })

	// err := i.startKymaDeployment(overridesProvider, prerequisitesEng, componentsEng)

	// assert.NoError(t, err)

	jobmanager.SetConfig()
	jobmanager.SetKubeClient(kubeClient)
	// job := job1{}
	// err := job.execute(&config.Config{}, kubeClient)
	require.NoError(t, err)
}
