package main

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/hydroform/connect"
	"github.com/kyma-incubator/hydroform/connect/types"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	storeObj := store{}
	c := connect.GetKymaConnector(storeObj)
	configUrl := "https://connector-service.kyma.cuzqfh0pmp.i317204kym.shoot.canary.k8s-hana.ondemand.com/v1/applications/signingRequests/info?token=bdK9ob26ZHQN_C9byn6uk2ELUKSWORc8A6BniEmFZes60sP6BOLoArMCE8DbhHpkDDhOwpT8jisd3yUvtexoug=="

	err := c.Connect(configUrl)

	if err != nil {
		log.Print(err.Error())
		return
	}

}

type store struct{}

func (s store) ReadConfig() (*types.CSRInfo, error) {

	config := &types.CSRInfo{}
	_, err := os.Stat("config.json")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	configBytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	err = json.Unmarshal(configBytes, config)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return config, err
}

func (s store) WriteConfig(config *types.CSRInfo) error {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	return ioutil.WriteFile("config.json", configBytes, 0644)
}

func (s store) ReadInfo() (*types.Info, error) {
	info := &types.Info{}
	_, err := os.Stat("info.json")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	infoBytes, err := ioutil.ReadFile("info.json")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	err = json.Unmarshal(infoBytes, info)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return info, err
}

func (s store) WriteInfo(info *types.Info) error {
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	return ioutil.WriteFile("info.json", infoBytes, 0644)
}

func (s store) ReadClientCert() (*types.ClientCertificate, error) {

	cert := &types.ClientCertificate{}
	_, err := os.Stat("generated.crt")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	publicKey, err := ioutil.ReadFile("generated.crt")
	cert.PublicKey = string(publicKey[:])

	_, err = os.Stat("generated.key")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	privateKey, err := ioutil.ReadFile("generated.crt")
	cert.PrivateKey = string(privateKey[:])

	return cert, err

}

func (s store) WriteClientCert(cert *types.ClientCertificate) error {
	err := ioutil.WriteFile("generated.crt", []byte(cert.PublicKey), 0644)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ioutil.WriteFile("generated.key", []byte(cert.PrivateKey), 0644)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	return err
}
