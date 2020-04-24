package main

import (
	"encoding/json"
	"github.com/kyma-incubator/hydroform/connect"
	"github.com/kyma-incubator/hydroform/connect/types"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	storeObj := store{}
	c := connect.GetKymaConnector(storeObj)
	configUrl := "https://connector-service.kyma.local/v1/applications/signingRequests/info?token=x0sEfJZFGxiJzpzEN86ENmOxCAoO2nhbOy6N8cxb6o8paWQgs5wZSG30X0y2DcpS7_dL0nH5SQcnWEDOqQOgTA=="

	if err := c.Connect(configUrl); err != nil {
		log.Print(err.Error())
		return
	}

	testService := &connect.Service{
		Id:               "testService",
		Provider:         "testProvider",
		Name:             "servTest",
		Description:      "desc",
		ShortDescription: "",
		Labels:           nil,
		API: &connect.ServiceAPI{
			TargetURL: "localhost:8080",
			Spec:      []byte("{\"swagger\":\"2.0\",\"info\":{\"version\":\"1.0.0\",\"title\":\"Default title here\",\"description\":\"A simple test API\",\"contact\":{\"name\":\"Kavya Kathuria\"},\"license\":{\"name\":\"Apache 2.0\"}},\"host\":\"localhost\",\"basePath\":\"/\",\"schemes\":[\"http\"],\"consumes\":[\"application/json\"],\"produces\":[\"application/json\"],\"paths\":{\"/start\":{\"post\":{\"description\":\"Start tells driver to get ready to do work\",\"operationId\":\"startDrone\",\"responses\":{\"204\":{\"description\":\"Drone started\"},\"default\":{\"description\":\"unexpected error\",\"schema\":{\"$ref\":\"#/definitions/ErrorModel\"}}}}}},\"definitions\":{\"ValueModel\":{\"type\":\"object\",\"required\":[\"value\"],\"properties\":{\"value\":{\"type\":\"integer\",\"format\":\"int32\",\"minimum\":0,\"maximum\":100}}}}}"),
		},
		Events: &connect.ServiceEvent{
			Spec: []byte("{\"asyncapi\":\"1.0.0\",\"info\":{\"title\":\"PetStore Events\",\"version\":\"1.0.0\",\"description\":\"Description of all the events\"},\"baseTopic\":\"stage.com.some.company.system\",\"topics\":{\"petCreated.v1\":{\"subscribe\":{\"summary\":\"test event\",\"payload\":{\"type\":\"object\",\"properties\":{\"pet\":{\"type\":\"object\",\"required\":[\"id\",\"name\"],\"example\":{\"id\":\"4caad296-e0c5-491e-98ac-0ed118f9474e\",\"category\":\"mammal\",\"name\":\"doggie\"},\"properties\":{\"id\":{\"title\":\"Id\",\"description\":\"Resource identifier\",\"type\":\"string\"},\"name\":{\"title\":\"Name\",\"description\":\"Pet name\",\"type\":\"string\"},\"category\":{\"title\":\"Category\",\"description\":\"Animal category\",\"type\":\"string\"}}}}}}}}}"),
		},
		Documentation: nil,
	}

	serviceId, err := c.RegisterService(testService)
	if err != nil {
		log.Print(err.Error())
		return
	}

	if err := c.DeleteService(serviceId); err != nil {
		log.Print(err.Error())
		return
	}
}

type store struct{}

func (s store) ReadConfig() (*types.CSRInfo, error) {
	config := &types.CSRInfo{}
	if _, err := os.Stat("config.json"); err != nil {
		return nil, err
	}
	configBytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(configBytes, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (s store) WriteConfig(config *types.CSRInfo) error {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("config.json", configBytes, 0644)
}

func (s store) ReadInfo() (*types.Info, error) {
	info := &types.Info{}
	if _, err := os.Stat("info.json"); err != nil {
		return nil, err
	}
	infoBytes, err := ioutil.ReadFile("info.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(infoBytes, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (s store) WriteInfo(info *types.Info) error {
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("info.json", infoBytes, 0644)
}

func (s store) ReadClientCert() (*types.ClientCertificate, error) {
	cert := &types.ClientCertificate{}

	if _, err := os.Stat("generated.crt"); err != nil {
		return nil, err
	}
	publicKey, err := ioutil.ReadFile("generated.crt")
	if err != nil {
		return nil, err
	}
	cert.PublicKey = string(publicKey[:])

	if _, err = os.Stat("generated.key"); err != nil {
		return nil, err
	}
	privateKey, err := ioutil.ReadFile("generated.key")
	if err != nil {
		return nil, err
	}
	cert.PrivateKey = string(privateKey[:])

	return cert, nil
}

func (s store) WriteClientCert(cert *types.ClientCertificate) error {
	if err := ioutil.WriteFile("generated.crt", []byte(cert.PublicKey), 0644); err != nil {
		return err
	}

	if err := ioutil.WriteFile("generated.key", []byte(cert.PrivateKey), 0644); err != nil {
		return err
	}

	return nil
}
