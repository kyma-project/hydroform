package jobManager

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
)

// Register job using implemented interface

type job1 struct{}

//var _ = register(job1)


func (j job1) execute(cfg *config.Config, kubeClient kubernetes.Interface) {
	namespace := "kyma-system"
	pvc := "storage-logging-loki-0"

	pvc, err = clientset.CoreV1().PersistentVolumeClaim(namespace).Get(context.TODO(), pvc, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %s in namespace %s: %v\n",
			pod, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	}
  return nil
}

func (j job1) when() {
	return ("logging", Pre)
}