package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	hf "github.com/kyma-incubator/hydroform/provision"
	"github.com/kyma-incubator/hydroform/provision/types"
)

func main() {
	resourceGroup := flag.String("p", "", "Azure resource group")
	machineType := flag.String("m", "Standard_D2_v3", "Azure machine type")
	credentials := flag.String("c", "", "Path to the credentials file")
	persist := flag.Bool("persist", false, "Persistence option. With persistence enabled, hydroform will keep state and configuraion of clusters on the file system.")
	flag.Parse()

	log.SetOutput(ioutil.Discard)

	cluster := &types.Cluster{
		KubernetesVersion: "1.18.8",
		Name:              "hydro",
		DiskSizeGB:        35,
		NodeCount:         2,
		Location:          "westeurope",
		MachineType:       *machineType,
	}
	provider := &types.Provider{
		Type:                types.Azure,
		ProjectName:         *resourceGroup,
		CredentialsFilePath: *credentials,
	}

	var ops []types.Option
	// add persistence option
	if *persist {
		ops = append(ops, types.Persistent())
	}

	fmt.Println("Provisioning...")

	cluster, err := hf.Provision(cluster, provider, ops...)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	fmt.Println("Provisioned successfully")

	fmt.Println("Getting the status")

	status, err := hf.Status(cluster, provider, ops...)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	fmt.Println("Status:", *status)

	fmt.Println("Downloading the kubeconfig")

	content, err := hf.Credentials(cluster, provider, ops...)
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

	// fmt.Println("Deprovisioning...")

	// err = hf.Deprovision(cluster, provider, ops...)
	// if err != nil {
	// 	fmt.Println("Error", err.Error())
	// 	return
	// }

	// fmt.Println("Deprovisioned successfully")
}
