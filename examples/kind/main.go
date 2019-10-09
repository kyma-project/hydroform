package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/kyma-incubator/hydroform"
	"github.com/kyma-incubator/hydroform/types"
)

func main() {
	projectName := flag.String("p", "", "kind project name")
	kubernetesVersion := flag.String("k", "", "kubernetes version of the cluster")
	flag.Parse()

	log.SetOutput(ioutil.Discard)

	fmt.Println("Provisioning...")

	cluster := &types.Cluster{
		KubernetesVersion: *kubernetesVersion,
		Name:              "test-cluster",
		NodeCount:         1,
	}
	provider := &types.Provider{
		Type:        types.Kind,
		ProjectName: *projectName,
	}

	cluster, err := hydroform.Provision(cluster, provider)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	fmt.Println("Provisioned successfully")

	fmt.Println("Getting the status")

	status, err := hydroform.Status(cluster, provider)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	fmt.Println("Status:", *status)

	fmt.Println("Downloading the kubeconfig")

	content, err := hydroform.Credentials(cluster, provider)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	err = ioutil.WriteFile("kubeconfig.yaml", content, 0600)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	fmt.Println("Kubeconfig downloaded")

	//fmt.Println("Deprovisioning...")
	//
	//err = hydroform.Deprovision(cluster, provider)
	//if err != nil {
	//	fmt.Println("Error", err.Error())
	//	return
	//}
	//
	//fmt.Println("Deprovisioned successfully")
	//fmt.Printf("Cluster status: %s, IP: %s\r\n", clusterInfo.Status, clusterInfo.IP)
}
