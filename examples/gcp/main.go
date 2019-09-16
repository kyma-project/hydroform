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
		CPU:               "1",
		KubernetesVersion: "1.12",
		Name:              "test-cluster",
		DiskSizeGB:        30,
		NodeCount:         2,
		Location:          "europe-west3-a",
		MachineType:       *machineType,
	}
	provider := &types.Provider{
		Type:        types.GCP,
		ProjectName: *projectName,
		CustomConfigurations: map[string]interface{}{
			"credentials_file_path": *credentials,
		},
	}

	_, err := hydroform.Provision(cluster, provider)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	fmt.Println("Provisioned successfully")
	//fmt.Printf("Cluster status: %s, IP: %s\r\n", clusterInfo.Status, clusterInfo.IP)
}
