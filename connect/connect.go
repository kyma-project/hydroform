package connect

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyma-incubator/hydroform/connect/types"
	"io/ioutil"
	"log"
	"net/http"
)

func (c *KymaConnector) Connect(configurationUrl string) error {

	err := c.populateCsrInfo(configurationUrl)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = c.populateCertSigningRequest()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = c.populateClientCert()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if c.SecureClient == nil {
		err = c.populateClient()
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	err = c.PopulateInfo()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = c.persistCertificate()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return err
}

func (c *KymaConnector) RegisterService(apiDocs string, eventDocs string, serviceConfig string) (serviceId string, err error) {
	serviceDescription := new(Service)

	serviceDescription.Documentation = new(ServiceDocumentation)
	serviceDescription.Documentation.DisplayName = "Kyma Service"
	serviceDescription.Documentation.Description = "Default description"
	serviceDescription.Documentation.Tags = []string{"Tag0", "Tag1"}
	serviceDescription.Documentation.Type = "Service Type"

	serviceDescription.Description = "Default API Description"
	serviceDescription.ShortDescription = "Default API Short Description"

	serviceDescription.Provider = "Default provider"
	serviceDescription.Name = "Default service name"

	if serviceConfig != "" {
		log.Println("Read Service Config")
		err := c.ReadService(serviceConfig, serviceDescription)
		if err != nil {
			log.Printf("Failed to read service config: %s", serviceConfig)
			return "", err
		}
	}

	if apiDocs != "" {
		if serviceDescription.API == nil {
			log.Println("No Service Description")
			serviceDescription.API = new(ServiceAPI)
			serviceDescription.API.TargetURL = "http://localhost:8080/"
		}

		serviceDescription.API.Spec, err = c.GetRawJsonFromDoc(apiDocs)
		if err != nil {
			return "", err
		}
	}

	if eventDocs != "" {
		serviceDescription.Events = new(ServiceEvent)
		serviceDescription.Events.Spec, err = c.GetRawJsonFromDoc(eventDocs)
		if err != nil {
			return "", err
		}
	}

	jsonBytes, err := json.Marshal(serviceDescription)
	if err != nil {
		log.Printf("JSON marshal failed: %s", err)
		return
	}

	if c.CsrInfo == nil || c.CsrInfo.API.MetadataUrl == "" {
		log.Printf("%s", fmt.Errorf("metadata url is missing, cannot proceed"))
		return
	}

	//	client, err := c.GetSecureClient()
	resp, err := c.SecureClient.Post(c.CsrInfo.API.MetadataUrl, "application/json", bytes.NewBuffer(jsonBytes))

	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	if err != nil {
		log.Printf("could not dump response: %v", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered service with %s", bodyString)
	} else {
		log.Printf("Status: %d >%s< \n on URL: %s", resp.StatusCode, bodyString, c.CsrInfo.API.MetadataUrl)
		return "", errors.New("Failed to register service")
	}

	id := &struct {
		Id string `json: "id"`
	}{}

	err = json.Unmarshal(bodyBytes, id)
	if err != nil {
		log.Println("Failed to parse registration response")
		return "", err
	}
	return id.Id, err
}

func (c *KymaConnector) UpdateService(id string, apiDocs string, eventDocs string) error {
	serviceDescription := new(Service)
	err := c.ReadService(id, serviceDescription)
	if err != nil {
		log.Printf("Failed to read service config: %s", id+".json")
		return err
	}

	if apiDocs != "" {
		if serviceDescription.API == nil {
			serviceDescription.API = new(ServiceAPI)
			serviceDescription.API.TargetURL = "http://localhost:8080/"
		}

		serviceDescription.API.Spec, err = c.GetRawJsonFromDoc(apiDocs)
		if err != nil {
			return err
		}

	}

	if eventDocs != "" {
		serviceDescription.Events = new(ServiceEvent)
		serviceDescription.Events.Spec, err = c.GetRawJsonFromDoc(eventDocs)
		if err != nil {
			return err
		}
	}

	jsonBytes, err := json.Marshal(serviceDescription)
	if err != nil {
		log.Printf("JSON marshal failed: %s", err)
		return err
	}

	if c.CsrInfo == nil || c.CsrInfo.API.MetadataUrl == "" {
		log.Printf("%s", fmt.Errorf("metadata url is missing, cannot proceed"))
		return err
	}

	url := c.CsrInfo.API.MetadataUrl + "/" + id
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.SecureClient.Do(req)
	if err != nil {
		log.Printf("Couldn't register service: %s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered service")
	} else {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		log.Printf("Status: %d >%s<\n on URL: %s", resp.StatusCode, bodyString, url)
		return errors.New("Failed to Update")
	}
	return err
}

func (c *KymaConnector) DeleteService(id string) error {

	url := c.CsrInfo.API.MetadataUrl + "/" + id
	req, _ := http.NewRequest("DELETE", url, nil)

	resp, err := c.SecureClient.Do(req)
	if err != nil {
		log.Printf("Couldn't delete service: %s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		log.Printf("Successfully deleted service")
		return nil
	} else {
		return errors.New("Failed to delete")
	}
}

func (c *KymaConnector) AddEvent(event types.Event) error {

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	resp, err := c.SecureClient.Post(c.CsrInfo.API.EventsUrl, "application/json", bytes.NewBuffer(eventBytes))

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered event")
	} else {
		log.Print("Incorrect response")
		return err
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	log.Print(string(bodyBytes))
	if err != nil {
		log.Println("Failed to parse registration response")
		return err
	}

	return err
}

func (c *KymaConnector) GetSubscribedEvents() ([]types.EventResponse, error) {

	resp, err := c.SecureClient.Get(c.CsrInfo.API.EventsInfoUrl)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(err.Error())
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	type eventsInfoStruct struct {
		EventsInfo []types.EventResponse `json:"eventsInfo"`
	}

	eventsInfoObj := eventsInfoStruct{}
	err = json.Unmarshal(response, &eventsInfoObj)
	return eventsInfoObj.EventsInfo, err
}

func GetKymaConnector(writerInterface WriterInterface) *KymaConnector {
	c := &KymaConnector{
		CsrInfo:          &types.CSRInfo{},
		Ca:               &types.ClientCertificate{},
		Info:             &types.Info{},
		StorageInterface: writerInterface,
	}

	c.loadConfig()
	c.populateClient()
	return c
}

func (c *KymaConnector) GetSecureClient() (*http.Client, error) {
	cert, err := tls.X509KeyPair([]byte(c.Ca.PublicKey), []byte(c.Ca.PrivateKey))
	if err != nil {
		log.Println("Can't load certificates")
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{Transport: transport}, nil

}

func (c *KymaConnector) RenewCertificateSigningRequest() error {

	type csrStruct struct {
		Csr string `json:"csr"`
	}

	encodedCsr := base64.StdEncoding.EncodeToString([]byte(c.Ca.Csr))

	requestBody := csrStruct{
		Csr: encodedCsr,
	}

	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	resp, err := c.SecureClient.Post(c.Info.URLs.RenewCertUrl, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		log.Printf("Successfully renewed certificate")
	} else {
		log.Printf("error in renewing csr")
		return errors.New("Failed to renew")
	}

	certificates, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	crtResponse := types.CRTResponse{}
	err = json.Unmarshal(certificates, &crtResponse)
	if err != nil {
		return fmt.Errorf("JSON Error")
	}
	decodedCert, err := base64.StdEncoding.DecodeString(crtResponse.ClientCRT)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	c.Ca.PublicKey = string(decodedCert)

	c.persistCertificate()

	return err
}

func (c *KymaConnector) RevokeCertificate() error {
	resp, err := c.SecureClient.Post(c.Info.URLs.RevokeCertUrl, "application/json", nil)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		log.Print("Successfully revoked certificate for client")
	} else {
		log.Print("Error in trying to revoke certificate")
		return errors.New("error in trying to revoke certificate")
	}
	return err
}
