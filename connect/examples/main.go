package main

import (
	"github.com/kyma-incubator/hydroform/connect"
	"log"
)

func main() {

	c := connect.GetBlankKymaConnector()

	configUrl := "https://connector-service.kyma.rsnqwxc6j9.i317204kym.shoot.canary.k8s-hana.ondemand.com/v1/applications/signingRequests/info?token=aWbirf2KE2fuqF9wEYTFkyjD-QtIyPN_Oz3CJ9C73CTtGYBVvZ4_2vp7OOzgMOqyPLwwP4M5o8ROVQLPdU_LEQ=="

	err := c.Connect(configUrl)
	if err != nil {
		log.Println(err.Error())
		return
	}
	//c.RegisterService("api-docs.json", "event-docs.json", "")
	err = c.UpdateService("ff54f2d0-99b2-414a-ba50-dc025e2a9d5f", "api-docs.json", "event-docs.json")
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println("Success.")
}
