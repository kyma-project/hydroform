package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/kyma-incubator/hydroform/provision/action"

	hf "github.com/kyma-incubator/hydroform/provision"
	"github.com/kyma-incubator/hydroform/provision/types"
)

func main() {
	projectName := flag.String("p", "", "Gardener project name")
	machineType := flag.String("m", "m4.2xlarge", "AWS machine type")
	credentials := flag.String("c", "", "Path to the credentials file")
	secret := flag.String("s", "", "Name of the secret to access the underlying provider of gardener")
	persist := flag.Bool("persist", false, "Persistence option. With persistence enabled, hydroform will keep state and configuraion of clusters on the file system.")
	flag.Parse()

	log.SetOutput(ioutil.Discard)

	cluster := &types.Cluster{
		CPU:               1,
		KubernetesVersion: "1.15.4",
		Name:              "hydro-aws",
		DiskSizeGB:        35,
		NodeCount:         2,
		Location:          "eu-west-1",
		MachineType:       *machineType,
	}
	provider := &types.Provider{
		Type:                types.Gardener,
		ProjectName:         *projectName,
		CredentialsFilePath: *credentials,
		CustomConfigurations: map[string]interface{}{
			"target_provider": "aws",
			"target_secret":   *secret,
			"disk_type":       "gp2",
			"workercidr":      "10.250.0.0/19",
			"publicscidr":     "10.250.96.0/22",
			"vpccidr":         "10.250.0.0/16",
			"internalscidr":   "10.250.112.0/22",
			"zone":            "eu-west-1b",
			"autoscaler_min":  2,
			"autoscaler_max":  4,
			"max_surge":       4,
			"max_unavailable": 1,
		},
	}

	var ops []types.Option
	// add persistence option
	if *persist {
		ops = append(ops, types.Persistent())
	}

	action.SetArgs(cluster.Name, provider.Type)

	action.SetBefore(action.FuncAction(func(args ...interface{}) (interface{}, error) {
		fmt.Printf("Provisioning %s on %s...\n", args[0], args[1])
		return nil, nil
	}))

	action.SetAfter(action.FuncAction(func(args ...interface{}) (interface{}, error) {
		fmt.Printf("Provisioned %s successfully\n", args[0])
		return nil, nil
	}))
	cluster, err := hf.Provision(cluster, provider, ops...)
	if err != nil {
		fmt.Println("Error", err.Error())
		return
	}

	action.SetBefore(action.FuncAction(func(args ...interface{}) (interface{}, error) {
		fmt.Printf("Getting the status of %s\n", args[0])
		return nil, nil
	}))
	status, err := hf.Status(cluster, provider, ops...)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	fmt.Println("Status:", status.Phase)

	action.SetBefore(action.FuncAction(func(args ...interface{}) (interface{}, error) {
		fmt.Println("Downloading the kubeconfig")
		return nil, nil
	}))

	action.SetAfter(action.FuncAction(func(args ...interface{}) (interface{}, error) {
		fmt.Println("Kubeconfig downloaded")
		return nil, nil
	}))
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

	//fmt.Println("Deprovisioning...")
	//
	//err = hf.Deprovision(cluster, provider, ops...)
	//if err != nil {
	//	fmt.Println("Error", err.Error())
	//	return
	//}
	//
	//fmt.Println("Deprovisioned successfully")
}
