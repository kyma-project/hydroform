package main

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/connect"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	storeObj := store{}
	c := connect.GetKymaConnector(storeObj)
	configUrl := "https://connector-service.kyma.cuzqfh0pmp.i317204kym.shoot.canary.k8s-hana.ondemand.com/v1/applications/signingRequests/info?token=VdHUBHKNk-Bjlc3CWhC0PsjVEmYStLx9lxL57NxRHsvHdq8p4EOskWyC58kJU0FWQnDvkRapUY5ZqssWv33Otg=="

	err := c.Connect(configUrl)

	if err != nil {
		log.Print(err.Error())
		return
	}
}

type store struct{}

func (s store) ReadData(filename string) ([]byte, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return ioutil.ReadFile(filename)
}

func (s store) WriteData(fileName string, data []byte) error {
	return ioutil.WriteFile(fileName, data, 0644)
}
