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
	projectName := flag.String("p", "", "GCP project name")
	machineType := flag.String("m", "n1-standard-4", "GCP machine type")
	credentials := flag.String("c", "", "Path to the credentials file")
	flag.Parse()

	log.SetOutput(ioutil.Discard)

	fmt.Println("Provisioning...")

	cluster := &types.Cluster{
		KubernetesVersion: "1.13",
		Name:              "test-cluster",
		DiskSizeGB:        30,
		NodeCount:         1,
		Location:          "europe-west3-a",
		MachineType:       *machineType,
	}
	provider := &types.Provider{
		Type:                types.GCP,
		ProjectName:         *projectName,
		CredentialsFilePath: *credentials,
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
