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

func (j job1) when() (component, executionTime) {
	return component("logging"), Pre
}

func (j job1) identify() jobName {
	return jobName("exampleJob")
}

func (j job1) execute(cfg *config.Config, kubeClient kubernetes.Interface, ctx context.Context) error {
	ctx.Done()
	fmt.Println("\nStart of sample Job")

	namespace := "kyma-system"
	pvc := "storage-logging-loki-0"
	statefulset := "logging-loki"
	targetPVCSize := 50

	pvcReturn, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvc, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("\nPVC %s in namespace %s not found -> Skipping Job", pvc, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("\nError getting PVC %s in namespace %s: %v -> Skipping Job",
				pvc, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			fmt.Printf("\nError in get PVC: %s -> Skipping Job", err.Error())
		}
		return err
	} else {
		fmt.Printf("\nFound PVC %s in namespace %s\n", pvc, namespace)

		pvcStorageSize := pvcReturn.Spec.Resources.Requests.Storage().String()
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatch := re.Find([]byte(pvcStorageSize))

		curSize, err := strconv.Atoi(string(submatch))
		if err != nil {
		}
		fmt.Printf("CurSiyze: %d TarSize %d\n", curSize, targetPVCSize)
		if curSize != targetPVCSize {
			err := kubeClient.AppsV1().StatefulSets(namespace).Delete(ctx, statefulset, metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("\nError while deleting StatefulSet: %s", err)
			} else {
				fmt.Println("\nDeleted StatefulSet")
			}

			finalTargetSize := constructSizeString(targetPVCSize)
			fmt.Printf("\nFinal PVC Size: %s\n", finalTargetSize)
			jsonPatch := []byte(fmt.Sprintf(`{ "spec": { "resources": { "requests": {"storage": "%s"}}}}`, finalTargetSize))
			res, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Patch(ctx, pvc, types.MergePatchType, jsonPatch, metav1.PatchOptions{})
			fmt.Printf("Result of Patch %s\n", res)
			if errors.IsNotFound(err) {
				fmt.Printf("PVC %s in namespace %s not found\n", pvc, namespace)
			} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
				fmt.Printf("Error adjust PVC size%s in namespace %s: %v\n",
					pvc, namespace, statusError.ErrStatus.Message)
			} else if err != nil {
				panic(err.Error())
			}

		} else {
			fmt.Printf("Following job is skipped, due to current cluster state: %s\n", "exampleJob")
		}

		return nil
	}
}

func constructSizeString(szInGibi int) string {
	return strconv.Itoa(szInGibi) + "Gi"
}
