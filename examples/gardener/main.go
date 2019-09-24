package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	hf "github.com/kyma-incubator/hydroform"
	"github.com/kyma-incubator/hydroform/types"
)

func main() {
	projectName := flag.String("p", "", "Gardener project name")
	machineType := flag.String("m", "n1-standard-4", "GCP machine type")
	credentials := flag.String("c", "", "Path to the credentials file")
	secret := flag.String("s", "", "Name of the secret to access the underlying provider of gardener")
	flag.Parse()

	log.SetOutput(ioutil.Discard)

	cluster := &types.Cluster{
		CPU:               "1",
		KubernetesVersion: "1.15.4",
		Name:              "hydro-cluster",
		DiskSizeGB:        30,
		NodeCount:         2,
		Location:          "europe-west4",
		MachineType:       *machineType,
	}
	provider := &types.Provider{
		Type:                types.Gardener,
		ProjectName:         *projectName,
		CredentialsFilePath: *credentials,
		CustomConfigurations: map[string]interface{}{
			"target_provider": "gcp",
			"target_secret":   *secret,
			"disk_type":       "pd-standard",
			"zone":            "europe-west4-b",
			"cidr":            "10.250.0.0/19",
			"autoscaler_min":  2,
			"autoscaler_max":  4,
			"max_surge":       4,
			"max_unavailable": 1,
		},
	}

	fmt.Println("Provisioning...")
	cluster, err := hf.Provision(cluster, provider)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}
	fmt.Println("Provisioned successfully")

	fmt.Println("Getting the status")
	status, err := hf.Status(cluster, provider)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	fmt.Println("Status:", *status)

	fmt.Println("Downloading the kubeconfig")

	content, err := hf.Credentials(cluster, provider)
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
	//err = hf.Deprovision(cluster, provider)
	//if err != nil {
	//	fmt.Println("Error", err.Error())
	//	return
	//}
	//
	//fmt.Println("Deprovisioned successfully")
}
