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
	zapLogger.Info("Start of sample Job")

	namespace := "kyma-system"
	pvc := "storage-logging-loki-0"
	statefulset := "logging-loki"
	targetPVCSize := 50

	pvcReturn, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvc, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			zapLogger.Infof("PVC %s in namespace %s not found -> Skipping Job", pvc, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			zapLogger.Infof("Error getting PVC %s in namespace %s: %v -> Skipping Job",
				pvc, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			zapLogger.Infof("Error in get PVC: %s -> Skipping Job", err.Error())
		}
		return err
	} else {
		zapLogger.Infof("Found PVC %s in namespace %s", pvc, namespace)

		pvcStorageSize := pvcReturn.Spec.Resources.Requests.Storage().String()
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatch := re.Find([]byte(pvcStorageSize))

		curSize, err := strconv.Atoi(string(submatch))
		if err != nil {
		}
		zapLogger.Infof("CurSiyze: %d TarSize %d", curSize, targetPVCSize)
		if curSize != targetPVCSize {
			err := kubeClient.AppsV1().StatefulSets(namespace).Delete(ctx, statefulset, metav1.DeleteOptions{})
			if err != nil {
				zapLogger.Infof("Error while deleting StatefulSet: %s", err)
			} else {
				zapLogger.Info("Deleted StatefulSet")
			}

			finalTargetSize := constructSizeString(targetPVCSize)
			zapLogger.Infof("Final PVC Size: %s", finalTargetSize)
			jsonPatch := []byte(fmt.Sprintf(`{ "spec": { "resources": { "requests": {"storage": "%s"}}}}`, finalTargetSize))
			res, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Patch(ctx, pvc, types.MergePatchType, jsonPatch, metav1.PatchOptions{})
			zapLogger.Infof("Result of Patch %s", res)
			if errors.IsNotFound(err) {
				zapLogger.Infof("PVC %s in namespace %s not found", pvc, namespace)
			} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
				zapLogger.Infof("Error adjust PVC size%s in namespace %s: %v",
					pvc, namespace, statusError.ErrStatus.Message)
			} else if err != nil {
				panic(err.Error())
			}

		} else {
			zapLogger.Infof("Following job is skipped, due to current cluster state: %s", "exampleJob")
		}

		return nil
	}
}

func constructSizeString(szInGibi int) string {
	return strconv.Itoa(szInGibi) + "Gi"
}
