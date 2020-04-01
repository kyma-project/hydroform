package connect

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyma-incubator/hydroform/connect/types"
	"io/ioutil"
	"log"
	"net/http"
)

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

	c.SecureClient, err = c.GetSecureClient()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = writeClientCertificateToFile(*c.Ca)
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
	ioutil.WriteFile(id.ID+".json", serviceDescriptionString, 0644)

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

func (c *KymaConnector) AddEvent(eventDoc string) error {

	file, err := ioutil.ReadFile(eventDoc)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	//	event := types.Event{}

	/*err = json.Unmarshal([]byte(file), &event)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	jsonBytes, err := json.Marshal(event)
	*/
	resp, err := c.SecureClient.Post(c.CsrInfo.API.EventsUrl, "application/json", bytes.NewBuffer([]byte(file)))

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered event")
		//return nil
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

func (c *KymaConnector) GetSubscribedEvents() error {

	resp, err := c.SecureClient.Get(c.CsrInfo.API.EventsInfoUrl)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(err.Error())
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	log.Print(string(response))
	return err
}

func GetBlankKymaConnector() *KymaConnector {
	c := &KymaConnector{
		CsrInfo:      &types.CSRInfo{},
		AppName:      "",
		Ca:           &types.ClientCertificate{},
		SecureClient: nil,
	}

	return c
}

// GetSecureClient is returning an http client with client certificate enabled
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
