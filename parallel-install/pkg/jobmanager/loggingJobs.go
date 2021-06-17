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

type increaseLoggingPvcSize struct{}

var _ = register(increaseLoggingPvcSize{})

func (j increaseLoggingPvcSize) when() (component, executionTime) {
	return component("logging"), Pre
}

func (j increaseLoggingPvcSize) identify() jobName {
	return jobName("increaseLoggingPvcSize")
}

func (j increaseLoggingPvcSize) execute(cfg *config.Config, kubeClient kubernetes.Interface, ctx context.Context) error {
	zapLogger.Infof("Start of %s", j.identify())

	namespace := "kyma-system"
	pvc := "storage-logging-loki-0"
	statefulset := "logging-loki"
	targetPVCSize := 50

	pvcReturn, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvc, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			zapLogger.Debugf("PVC %s in namespace %s not found -> Skipping Job", pvc, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			zapLogger.Debugf("Error getting PVC %s in namespace %s: %v -> Skipping Job",
				pvc, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			zapLogger.Debugf("Error in get PVC: %s -> Skipping Job", err.Error())
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
		zapLogger.Debugf("CurSiyze: %d TarSize %d", curSize, targetPVCSize)
		if curSize != targetPVCSize {
			err := kubeClient.AppsV1().StatefulSets(namespace).Delete(ctx, statefulset, metav1.DeleteOptions{})
			if err != nil {
				zapLogger.Warnf("Error deleting StatefulSet: %s", err)
				return err
			} else {
				zapLogger.Debug("Deleted StatefulSet")
			}

			finalTargetSize := constructSizeString(targetPVCSize)
			zapLogger.Debugf("Final PVC Size: %s", finalTargetSize)
			jsonPatch := []byte(fmt.Sprintf(`{ "spec": { "resources": { "requests": {"storage": "%s"}}}}`, finalTargetSize))
			res, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Patch(ctx, pvc, types.MergePatchType, jsonPatch, metav1.PatchOptions{})

			zapLogger.Debugf("Result of Patch %s", res)
			if err != nil {
				zapLogger.Warnf("Error patching PVC: %s", err)
				return err
			}

		} else {
			zapLogger.Infof("Job %s skipped", j.identify())
		}

		return nil
	}
}

func constructSizeString(szInGibi int) string {
	return strconv.Itoa(szInGibi) + "Gi"
}
