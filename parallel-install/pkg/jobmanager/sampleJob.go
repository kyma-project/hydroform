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
	fmt.Println("\nStart of sample Job")
	ctx := context.TODO()

	namespace := "kyma-system"
	pvc := "storage-logging-loki-0"
	statefulset := "logging-loki"
	targetPVCSize := 50

	pvcReturn, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvc, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("\nPVC %s in namespace %s not found", pvc, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("\nError getting PVC %s in namespace %s: %v",
				pvc, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			fmt.Printf("\nError in get PVC: %s", err.Error())
		}
		fmt.Printf("\nFollowing job is skipped: exampleJob")
		return err
	} else {
		fmt.Printf("\nFound PVC %s in namespace %s", pvc, namespace)

		pvcStorageSize := pvcReturn.Spec.Resources.Requests.Storage().String()
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)

		submatch := re.Find([]byte(pvcStorageSize))

		curSize, err := strconv.Atoi(string(submatch))
		fmt.Printf("\nCurrent PVC Size: %s", pvcStorageSize)
		if err != nil {
		}

		if curSize != targetPVCSize {
			// kubectl delete statefulsets.apps -n kyma-system logging-loki
			err := kubeClient.AppsV1().StatefulSets(namespace).Delete(ctx, statefulset, metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("\nError while deleting StatefulSet: %s", err)
			} else {
				fmt.Println("\nDeleted StatefulSet")
			}

			// kubectl patch persistentvolumeclaims -n kyma-system storage-logging-loki-0 --patch '{"spec": {"resources": {"requests": {"storage": "$(PVC_SIZE)"}}}}' || echo "true"
			finalTargetSize := constructSizeString(targetPVCSize)
			fmt.Printf("\nFinal PVC Size: %s\n", finalTargetSize)
			jsonPatch := []byte(fmt.Sprintf(`{ "spec": { "resources": { "requests": {"storage": "%s"}}}}`, finalTargetSize))
			res, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Patch(ctx, pvc, types.MergePatchType, jsonPatch, metav1.PatchOptions{})
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
