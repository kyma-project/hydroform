package main

import (
	"github.com/kyma-incubator/hydroform/connect"
	"log"
)

func main() {

	c := connect.GetBlankKymaConnector()

	c.RenewCertificateSigningRequest()

	c.RevokeCertificate()

	/*configUrl := "https://connector-service.kyma.kum1oij6gw.i317204kym.shoot.canary.k8s-hana.ondemand.com/v1/applications/signingRequests/info?token=r_rFrXgzgQfPSoVETTsxXaPo7sYgVPlmVB3tOlvee0KoVq6HFEQE4_5_jzfIc9EOlHl1n4EMAmO49cvBzmQJFA=="

	err := c.Connect(configUrl)
	if err != nil {
		log.Println(err.Error())
		return
	}
	*/
	//c.RegisterService("api-docs.json", "event-docs.json", "")
	//err = c.UpdateService("ff54f2d0-99b2-414a-ba50-dc025e2a9d5f", "api-docs.json", "event-docs.json")
	/*
		if err != nil {
			log.Println(err.Error())
			return
		}*/

	log.Println("Success.")
}
