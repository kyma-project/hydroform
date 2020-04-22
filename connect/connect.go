package connect

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
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

	_, err := c.populateCsrInfo(configurationUrl)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	_, err = c.populateClientCert()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if c.SecureClient == nil {
		err = c.populateClient()
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	_, err = c.populateInfo()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = c.StorageInterface.WriteClientCert(c.Ca)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return err
}

func (c *KymaConnector) RegisterService(serviceDescription Service) (serviceId string, err error) {
	jsonBytes, err := json.Marshal(serviceDescription)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	if c.CsrInfo == nil || c.CsrInfo.API.MetadataUrl == "" {
		return "", fmt.Errorf(err.Error())
	}

	resp, err := c.SecureClient.Post(c.CsrInfo.API.MetadataUrl, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered service with %s", bodyString)
	} else {
		return "", errors.New("Failed to register service")
	}

	id := &struct {
		Id string `json: "id"`
	}{}

	err = json.Unmarshal(bodyBytes, id)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return id.Id, err
}

func (c *KymaConnector) UpdateService(id string, serviceDescription Service) error {

	jsonBytes, err := json.Marshal(serviceDescription)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if c.CsrInfo == nil || c.CsrInfo.API.MetadataUrl == "" {
		return fmt.Errorf(err.Error())
	}

	url := c.CsrInfo.API.MetadataUrl + "/" + id
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.SecureClient.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully updated service")
	} else {
		return errors.New("failed to update service")
	}

	return err
}

func (c *KymaConnector) DeleteService(id string) error {
	url := c.CsrInfo.API.MetadataUrl + "/" + id
	req, _ := http.NewRequest("DELETE", url, nil)

	resp, err := c.SecureClient.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		log.Printf("Successfully deleted service")
		return nil
	} else {
		return errors.New("failed to delete")
	}
}

func (c *KymaConnector) SendEvent(event types.Event) error {

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
		return fmt.Errorf(err.Error())
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	log.Print(string(bodyBytes))
	if err != nil {
		return fmt.Errorf(err.Error())
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
		return nil, fmt.Errorf(err.Error())
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{Transport: transport}, nil

}

func (c *KymaConnector) RenewCertificateSigningRequest() error {
	keys, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	csr, err := c.getCsr(keys)
	type csrStruct struct {
		Csr string `json:"csr"`
	}
	encodedCsr := base64.StdEncoding.EncodeToString(csr)

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
		return errors.New("Failed to renew certificate")
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

	c.StorageInterface.WriteClientCert(c.Ca)

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
		return errors.New("Error in trying to revoke certificate")
	}
	return err
}
