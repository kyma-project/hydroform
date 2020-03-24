package main

import "github.com/kyma-incubator/hydroform/connect"

func main() {
	configUrlWithToken := "https://connector-service.kyma.rsnqwxc6j9.i317204kym.shoot.canary.k8s-hana.ondemand.com/v1/applications/signingRequests/info?token=5Q2soU6mQobly1yYRzz_7DwpYeyrB8Ffy91QqFDB1Z1ZmxfBxELIJ_PNaYJMMek5r2SdtfRRsHclSjB_JULTJA=="
	c := connect.KymaConnector{}

	connector, _ := c.Connect(configUrlWithToken)
	connector.RegisterService("apiDocs.json", "", "")
}
