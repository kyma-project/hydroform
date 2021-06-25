package jobmanager

// import (
// 	"context"

// 	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/rest"
// )

// // Register job using implemented interface

// type annotateCertificatesGateway struct{}

// // No deprecation planned; 06/2021
// var _ = register(annotateCertificatesGateway{})

// func (j annotateCertificatesGateway) when() (component, executionTime) {
// 	return component("certificates"), Pre
// }

// func (j annotateCertificatesGateway) identify() jobName {
// 	return jobName("annotateCertificatesGateway")
// }

// // This job increases the PVC-size of the logging component to 30GB.
// // This will be triggered before the deployment of its corresponding component.
// func (j annotateCertificatesGateway) execute(cfg *config.Config, kubeClient kubernetes.Interface, rc *rest.Config, ctx context.Context) error {
// 	log.Infof("Start of %s", j.identify())

// 	// 	kubectl -n kyma-system annotate gateway kyma-gateway meta.helm.sh/release-name=certificates --overwrite=true
// 	// kubectl -n kyma-system annotate gateway kyma-gateway meta.helm.sh/release-namespace=istio-system --overwrite=true

// 	kubeClient.CoreV1().RESTClient().Get().Resource("customresourcedefinitions.apiextensions.k8s.io").AbsPath("").DoRaw(context.Background())
// 	// namespace := "kyma-system"
// 	// gatewayName := "kyma-gateway"
// 	// overwrite := true
// 	// annotationOne := "meta.helm.sh/release-name=certificates"
// 	// annotationTwo := "meta.helm.sh/release-namespace=istio-system"

// 	// c := client.New()

// 	// if err != nil {
// 	// 	if errors.IsNotFound(err) {
// 	// 		log.Infof("PVC %s in namespace %s not found -> Skipping Job", pvc, namespace)
// 	// 	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
// 	// 		log.Infof("Error getting PVC %s in namespace %s: %v -> Skipping Job",
// 	// 			pvc, namespace, statusError.ErrStatus.Message)
// 	// 	} else if err != nil {
// 	// 		log.Infof("Error in get PVC: %s -> Skipping Job", err.Error())
// 	// 	}
// 	// 	return err
// 	// } else {
// 	// 	log.Infof("Found PVC %s in namespace %s", pvc, namespace)

// 	// 	pvcStorageSize := pvcReturn.Spec.Resources.Requests.Storage().String()
// 	// 	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
// 	// 	submatch := re.Find([]byte(pvcStorageSize))

// 	// 	curSize, err := strconv.Atoi(string(submatch))
// 	// 	if err != nil {
// 	// 	}
// 	// 	log.Infof("CurSiyze: %d TarSize %d", curSize, targetPVCSize)
// 	// 	if curSize != targetPVCSize {
// 	// 		err := kubeClient.AppsV1().StatefulSets(namespace).Delete(ctx, statefulset, metav1.DeleteOptions{})
// 	// 		if err != nil {
// 	// 			log.Warnf("Error deleting StatefulSet: %s", err)
// 	// 			return err
// 	// 		} else {
// 	// 			log.Info("Deleted StatefulSet")
// 	// 		}

// 	// 		finalTargetSize := constructSizeString(targetPVCSize)
// 	// 		log.Infof("Final PVC Size: %s", finalTargetSize)
// 	// 		jsonPatch := []byte(fmt.Sprintf(`{ "spec": { "resources": { "requests": {"storage": "%s"}}}}`, finalTargetSize))
// 	// 		res, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Patch(ctx, pvc, types.MergePatchType, jsonPatch, metav1.PatchOptions{})
// 	// 		log.Infof("Result of Patch %s", res)
// 	// 		if err != nil {
// 	// 			log.Warnf("Error patching PVC: %s", err)
// 	// 			return err
// 	// 		}

// 	// 	} else {
// 	// 		log.Infof("Job %s skipped", j.identify())
// 	// 	}
// 	// 	return nil
// 	// }
// 	return nil
// }
