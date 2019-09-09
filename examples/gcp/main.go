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
	credentials := flag.String("c", "/Users/i504462/gcp-test.json", "Credentials file")
	flag.Parse()

	log.SetOutput(ioutil.Discard)
	provider := hydroform.NewGoogleProvider("terraform")
	fmt.Println("Provisioning...")

	clusterConfig := &types.Cluster{
		CPU:               "1",
		KubernetesVersion: "1.12",
		Name:              "test-cluster",
		DiskSizeGB:        30,
	}
	platformConfig := &types.Platform{
		NodesCount: 2,
		Location:   "europe-west3-a",
		Configuration: map[string]interface{}{
			"credentials_file_path": *credentials,
		},
		ProjectName: *projectName,
		MachineType: *machineType,
	}

	err := provider.Provision(clusterConfig, platformConfig)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}
	fmt.Println("Provisioned successfully")
}
