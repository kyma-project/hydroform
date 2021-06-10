package jobmanager

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// Register job using implemented interface

type job1 struct{}

var _ = register(job1{})

func (j job1) execute(cfg *config.Config, kubeClient kubernetes.Interface) error {
	ctx := context.TODO()

	namespace := "kyma-system"
	pvc := "storage-logging-loki-0"
	statefulset := "logging-loki"
	targetPVCSize := 50

	pvcReturn, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvc, metav1.GetOptions{})
	pvcStorageSize := pvcReturn.Spec.Resources.Requests.Storage().String()
	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)

	submatch := re.Find([]byte(pvcStorageSize))

	curSize, err := strconv.Atoi(string(submatch))
	fmt.Printf("\nCurrent PVC Size: %s\n", pvcStorageSize)
	if err != nil {
	}

	if curSize != targetPVCSize {
		// kubectl delete statefulsets.apps -n kyma-system logging-loki
		kubeClient.AppsV1().StatefulSets(namespace).Delete(ctx, statefulset, metav1.DeleteOptions{})

		// kubectl patch persistentvolumeclaims -n kyma-system storage-logging-loki-0 --patch '{"spec": {"resources": {"requests": {"storage": "$(PVC_SIZE)"}}}}' || echo "true"
		jsonPatch := []byte(fmt.Sprintf(`{ "spec": { "resources": { "requests": {"storage": "%s"}}}}`, constructSizeString(targetPVCSize)))
		res, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Patch(ctx, pvc, types.JSONPatchType, jsonPatch, metav1.PatchOptions{})
		fmt.Printf("Result of Patch %s", res)
		if errors.IsNotFound(err) {
			fmt.Printf("PVC %s in namespace %s not found\n", pvc, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting PVC %s in namespace %s: %v\n",
				pvc, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found PVC %s in namespace %s\n", pvc, namespace)
		}

	} else {
		fmt.Printf("Following job is skipped, due to current cluster state: %s", "jobname")
	}

	return nil
}

func (j job1) when() (component, executionTime) {
	return component("logging"), Pre
}

func (j job1) identify() jobName {
	return jobName("exampleJob")
}

func constructSizeString(szInGibi int) string {
	return strconv.Itoa(szInGibi) + "Gi"
}
