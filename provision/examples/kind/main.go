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
	projectName := flag.String("p", "", "kind project name")
	nodeImage := flag.String("n", "", "kind node image of the cluster")
	persist := flag.Bool("persist", false, "Persistence option. With persistence enabled, hydroform will keep state and configuraion of clusters on the file system.")
	deprovision := flag.Bool("d", false, "Deprovision option. Deletes the cluster if it exists.")
	flag.Parse()

	log.SetOutput(ioutil.Discard)

	cluster := &types.Cluster{
		Name: "test-cluster",
	}
	provider := &types.Provider{
		Type:        types.Kind,
		ProjectName: *projectName,
		CustomConfigurations: map[string]interface{}{
			"node_image": *nodeImage,
		},
	}

	var ops []types.Option
	// add persistence option
	if *persist {
		ops = append(ops, types.Persistent())
	}

	if !*deprovision {
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
	} else {

		fmt.Println("Deprovisioning...")

		err := hf.Deprovision(cluster, provider, ops...)
		if err != nil {
			fmt.Println("Error", err.Error())
			return
		}

		fmt.Println("Deprovisioned successfully")
	}
}
