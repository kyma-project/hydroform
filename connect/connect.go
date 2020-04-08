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

type KymaInterface interface {
	getCsrInfo(string) error
	getCertSigningRequest() error
	getClientCert() error
	populateClient() error
	writeClientCertificateToFile(writerInterface) error
	writeToFile(string, []byte) error
	getRawJsonFromDoc(string) (json.RawMessage, error)
	readService(string, *Service) error
}

func (c *KymaConnector) Connect(configurationUrl string) error {

	err := c.getCsrInfo(configurationUrl)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = c.getCertSigningRequest()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = c.getClientCert()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	//err = c.populateClient()

	/*	if err != nil {
		return fmt.Errorf(err.Error())
	}*/

	err = c.writeClientCertificateToFile(c)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return err
}

func (c *KymaConnector) RegisterService(apiDocs string, eventDocs string, serviceConfig string) (err error) {
	serviceDescription := new(Service)

	serviceDescription.Documentation = new(ServiceDocumentation)
	serviceDescription.Documentation.DisplayName = "Kavya's Service"
	serviceDescription.Documentation.Description = "Kavya's decsription"
	serviceDescription.Documentation.Tags = []string{"Tag0", "Tag1"}
	serviceDescription.Documentation.Type = "Kavya's Type"

	serviceDescription.Description = "Kavya's API Description"
	serviceDescription.ShortDescription = "Kavya's API Short Description"

	serviceDescription.Provider = "Kavya provider"
	serviceDescription.Name = "Kavya name"

	if serviceConfig != "" {
		log.Println("Read Service Config")
		err := c.readService(serviceConfig, serviceDescription)
		if err != nil {
			log.Printf("Failed to read service config: %s", serviceConfig)
			return err
		}
	}

	if apiDocs != "" {
		if serviceDescription.API == nil {
			log.Println("No Service Description")
			serviceDescription.API = new(ServiceAPI)
			serviceDescription.API.TargetURL = "http://localhost:8080/"
		}

		serviceDescription.API.Spec, err = c.getRawJsonFromDoc(apiDocs)
		if err != nil {
			return err
		}
	}

	if eventDocs != "" {
		serviceDescription.Events = new(ServiceEvent)
		serviceDescription.Events.Spec, err = c.getRawJsonFromDoc(eventDocs)
		if err != nil {
			return err
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
		return fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	if err != nil {
		log.Printf("could not dump response: %v", err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered service with")
		log.Printf(bodyString)
	} else {
		log.Printf("Status: %d >%s< \n on URL: %s", resp.StatusCode, bodyString, c.CsrInfo.API.MetadataUrl)
		return errors.New("Failed to register")
	}

	id := &struct {
		ID string `json: "id"`
	}{}

	err = json.Unmarshal(bodyBytes, id)
	if err != nil {
		log.Println("Failed to parse registration response")
		return err
	}

	log.Printf("%v", id)
	serviceDescription.id = id.ID
	serviceDescriptionString, err := json.Marshal(serviceDescription)
	c.writeToFile(id.ID+".json", serviceDescriptionString)
	return err
}

func (c *KymaConnector) UpdateService(id string, apiDocs string, eventDocs string) error {
	serviceDescription := new(Service)
	err := c.readService(id+".json", serviceDescription)
	if err != nil {
		log.Printf("Failed to read service config: %s", id+".json")
		return err
	}

	if apiDocs != "" {
		if serviceDescription.API == nil {
			serviceDescription.API = new(ServiceAPI)
			serviceDescription.API.TargetURL = "http://localhost:8080/"
		}

		serviceDescription.API.Spec, err = c.getRawJsonFromDoc(apiDocs)
		if err != nil {
			return err
		}

	}

	if eventDocs != "" {

		serviceDescription.Events = new(ServiceEvent)

		serviceDescription.Events.Spec, err = c.getRawJsonFromDoc(eventDocs)
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
	log.Println(string(jsonBytes[:]))
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.SecureClient.Do(req)
	if err != nil {
		log.Printf("Couldn't register service: %s", err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered service with")
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

func (c *KymaConnector) GetSubscribedEvents() ([]types.Event, error) {

	resp, err := c.SecureClient.Get(c.CsrInfo.API.EventsInfoUrl)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(err.Error())
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	var eventsArr []types.Event
	err = json.Unmarshal(response, &eventsArr)
	if err != nil {
		log.Println("Failed to parse registration response")
		return nil, err
	}
	log.Print(string(response))
	return eventsArr, err
}

func GetBlankKymaConnector() *KymaConnector {
	c := &KymaConnector{
		CsrInfo:      &types.CSRInfo{},
		AppName:      "",
		Ca:           &types.ClientCertificate{},
		SecureClient: nil,
	}

	c.loadConfig()

	c.populateClient()
	// are we supposed to load config here -- maybe certificates need to be populated from file??
	return c
}

func GetKymaConnector(kymaInterface KymaInterface) *KymaConnector {
	//c  := kymaInterface.(*KymaConnector)
	switch kymaInterface.(type) {
	case *KymaConnector:
		c := kymaInterface.(*KymaConnector)
		c.CsrInfo = &types.CSRInfo{}
		c.AppName = ""
		c.Ca = &types.ClientCertificate{}
		return c
		/*case *MockConnect :
		return kymaInterface.(*MockConnect)*/
	}
	return nil
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

	resp, err := c.SecureClient.Post(c.CsrInfo.API.CertificatesUrl, "application/json", bytes.NewBuffer(jsonBytes))
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
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	c.Ca.PublicKey = string(decodedCert)
	return err

}

func (c *KymaConnector) RevokeCertificate() error {
	resp, err := c.SecureClient.Post(c.CsrInfo.API.CertificatesUrl+"/revocations", "application/json", nil)
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
