package jobmanager

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	istio "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// Register job using implemented interface

type increaseLoggingPvcSize struct{}

// No deprecation planned; 06/2021
var _ = register(increaseLoggingPvcSize{})

func (j increaseLoggingPvcSize) when() (component, executionTime) {
	return component("logging"), Pre
}

func (j increaseLoggingPvcSize) identify() jobName {
	return jobName("increaseLoggingPvcSize")
}

// This job increases the PVC-size of the logging component to 30GB.
// This will be triggered before the deployment of its corresponding component.
func (j increaseLoggingPvcSize) execute(cfg *config.Config, kubeClient kubernetes.Interface, ic istio.Interface, ctx context.Context) error {
	log.Infof("Start of %s", j.identify())

	namespace := "kyma-system"
	pvc := "storage-logging-loki-0"
	statefulset := "logging-loki"
	targetPVCSize := 30

	pvcReturn, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvc, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("PVC %s in namespace %s not found -> Skipping Job", pvc, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			log.Infof("Error getting PVC %s in namespace %s: %v -> Skipping Job",
				pvc, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			log.Infof("Error in get PVC: %s -> Skipping Job", err.Error())
		}
		return err
	} else {
		log.Infof("Found PVC %s in namespace %s", pvc, namespace)

		pvcStorageSize := pvcReturn.Spec.Resources.Requests.Storage().String()
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatch := re.Find([]byte(pvcStorageSize))

		curSize, err := strconv.Atoi(string(submatch))
		if err != nil {
		}
		log.Infof("CurSiyze: %d TarSize %d", curSize, targetPVCSize)
		if curSize != targetPVCSize {
			err := kubeClient.AppsV1().StatefulSets(namespace).Delete(ctx, statefulset, metav1.DeleteOptions{})
			if err != nil {
				log.Warnf("Error deleting StatefulSet: %s", err)
				return err
			} else {
				log.Info("Deleted StatefulSet")
			}

			finalTargetSize := constructSizeString(targetPVCSize)
			log.Infof("Final PVC Size: %s", finalTargetSize)
			jsonPatch := []byte(fmt.Sprintf(`{ "spec": { "resources": { "requests": {"storage": "%s"}}}}`, finalTargetSize))
			res, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Patch(ctx, pvc, types.MergePatchType, jsonPatch, metav1.PatchOptions{})
			log.Infof("Result of Patch %s", res)
			if err != nil {
				log.Warnf("Error patching PVC: %s", err)
				return err
			}

		} else {
			log.Infof("Job %s skipped", j.identify())
		}
		return nil
	}
}

func constructSizeString(szInGibi int) string {
	return strconv.Itoa(szInGibi) + "Gi"
}
