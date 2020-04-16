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
	configUrl := "https://connector-service.kyma.cuzqfh0pmp.i317204kym.shoot.canary.k8s-hana.ondemand.com/v1/applications/signingRequests/info?token=E3dRK_QDz2V4MRphKgBJleNWlyZmzjbKuRX6CetSsNgSFKo_4HgX2MpqJWHFhuK1JcalPVQcmdb3dgbNbAcjmQ=="

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

func (s store) ReadService(serviceId string) ([]byte, error) {
	path := serviceId + ".json"
	_, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return ioutil.ReadFile(path)
}

func (s store) ReadCSR() ([]byte, error) {
	_, err := os.Stat("generated.csr")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return ioutil.ReadFile("generated.csr")
}

func (s store) WriteCSR(data []byte) error {
	return ioutil.WriteFile("generated.csr", data, 0644)
}

func (s store) ReadCert() ([]byte, error) {
	_, err := os.Stat("generated.crt")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return ioutil.ReadFile("generated.crt")
}

func (s store) WriteCert(data []byte) error {
	return ioutil.WriteFile("generated.crt", data, 0644)
}

func (s store) ReadPrivateKey() ([]byte, error) {
	_, err := os.Stat("generated.key")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return ioutil.ReadFile("generated.key")
}

func (s store) WritePrivateKey(data []byte) error {
	return ioutil.WriteFile("generated.key", data, 0644)
}

func (s store) ReadConfig() ([]byte, error) {
	_, err := os.Stat("config.json")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return ioutil.ReadFile("config.json")
}

func (s store) WriteConfig(data []byte) error {
	return ioutil.WriteFile("config.json", data, 0644)
}

func (s store) ReadInfo() ([]byte, error) {
	_, err := os.Stat("info.json")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return ioutil.ReadFile("info.json")
}

func (s store) WriteInfo(data []byte) error {
	return ioutil.WriteFile("info.json", data, 0644)
}
