package connect

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"github.com/kyma-incubator/hydroform/connect/types"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
)

func (c *KymaConnector) Connect(configurationUrl string) error {

	if _, err := c.populateCsrInfo(configurationUrl); err != nil {
		return errors.Wrap(err, "Error trying to get CSR Information")
	}

	if _, err := c.populateClientCert(); err != nil {
		return errors.Wrap(err, "Error trying to populate client certificates")
	}

	if c.SecureClient == nil {
		if err := c.populateClient(); err != nil {
			return errors.Wrap(err, "Error trying to populate secure client")
		}
	}

	if _, err := c.populateInfo(); err != nil {
		return errors.Wrap(err, "Error trying to populate client info")
	}

	if err := c.StorageInterface.WriteClientCert(c.Ca); err != nil {
		return errors.Wrap(err, "Error trying to write certificate data")
	}

	return nil
}

func (c *KymaConnector) RegisterService(serviceDescription *Service) (serviceId string, err error) {
	jsonBytes, err := json.Marshal(serviceDescription)
	if err != nil {
		return "", errors.Wrap(err, "Error trying to register service  - Invalid format for Service object")
	}

	if c.CsrInfo == nil || c.CsrInfo.API.MetadataUrl == "" {
		return "", errors.New("Error trying to register service - Client config not populated")
	}

	resp, err := c.SecureClient.Post(c.CsrInfo.API.MetadataUrl, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", errors.Wrap(err, "Error trying to register service - HTTP client is not secure")
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered service with %s", bodyString)
	} else {
		return "", errors.New("Failed to register service - Invalid response from server")
	}

	id := &struct {
		Id string `json: "id"`
	}{}

	if err := json.Unmarshal(bodyBytes, id); err != nil {
		return "", errors.Wrap(err, "Error trying to register service - Invalid response from server")
	}

	return id.Id, nil
}

func (c *KymaConnector) UpdateService(id string, serviceDescription *Service) error {

	jsonBytes, err := json.Marshal(serviceDescription)
	if err != nil {
		return errors.Wrap(err, "Error trying to update service  - Invalid format for Service object")
	}

	if c.CsrInfo == nil || c.CsrInfo.API.MetadataUrl == "" {
		return errors.Wrap(err, "Error trying to update service - Client config not populated")
	}

	req, err := http.NewRequest("PUT", c.CsrInfo.API.MetadataUrl+"/"+id, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.Wrap(err, "Error trying to update service")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.SecureClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error trying to update service - HTTP client is not secure")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully updated service")
	} else {
		return errors.New("Failed to update service - Invalid response from server")
	}

	return nil
}

func (c *KymaConnector) DeleteService(id string) error {
	req, err := http.NewRequest("DELETE", c.CsrInfo.API.MetadataUrl+"/"+id, nil)
	if err != nil {
		return errors.Wrap(err, "Error trying to delete service")
	}
	resp, err := c.SecureClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error trying to delete service - HTTP client is not secure")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		log.Printf("Successfully deleted service")
	} else {
		return errors.New("Error trying to delete service - Invalid response from server")
	}

	return nil
}

func (c *KymaConnector) SendEvent(event types.Event) error {

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "Error trying to send event  - Invalid format for Event object")
	}

	resp, err := c.SecureClient.Post(c.CsrInfo.API.EventsUrl, "application/json", bytes.NewBuffer(eventBytes))
	if err != nil {
		return errors.Wrap(err, "Error trying to send event")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Successfully registered event")
	} else {
		return errors.New("Error trying to send event - Invalid response from server")
	}

	return nil
}

func (c *KymaConnector) GetSubscribedEvents() ([]types.EventResponse, error) {

	resp, err := c.SecureClient.Get(c.CsrInfo.API.EventsInfoUrl)
	if err != nil {
		return nil, errors.Wrap(err, "Error trying to get subscribed events")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Error trying to get subscribed events - Invalid response from server")
	}

	response, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.Wrap(err, "Error trying to get subscribed events")
	}

	type eventsInfoStruct struct {
		EventsInfo []types.EventResponse `json:"eventsInfo"`
	}

	eventsInfoObj := eventsInfoStruct{}
	if err := json.Unmarshal(response, &eventsInfoObj); err != nil {
		return nil, errors.Wrap(err, "Error trying to get subscribed events - Could not parse server response")
	}

	return eventsInfoObj.EventsInfo, nil
}

func GetKymaConnector(writerInterface StorageProvider) *KymaConnector {
	c := &KymaConnector{
		CsrInfo:          &types.CSRInfo{},
		Ca:               &types.ClientCertificate{},
		Info:             &types.Info{},
		StorageInterface: writerInterface,
	}

	c.loadConfig()
	if err := c.populateClient(); err != nil {
		log.Print("GetKymaConnector - Could not populate secure client")
	}
	return c
}

func (c *KymaConnector) GetSecureClient() (*http.Client, error) {
	cert, err := tls.X509KeyPair([]byte(c.Ca.PublicKey), []byte(c.Ca.PrivateKey))
	if err != nil {
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
	keys, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "Error trying to renew certificate - Could not generate new rsa key")
	}

	csr, err := c.getCsr(keys)
	if err != nil {
		return errors.Wrap(err, "Error trying to renew certificate  - Could not generate new CSR")
	}
	type csrStruct struct {
		Csr string `json:"csr"`
	}
	encodedCsr := base64.StdEncoding.EncodeToString(csr)

	requestBody := csrStruct{
		Csr: encodedCsr,
	}

	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return errors.Wrap(err, "Error trying to renew certificate  - Invalid CSR")
	}

	resp, err := c.SecureClient.Post(c.Info.URLs.RenewCertUrl, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.Wrap(err, "Error trying to renew certificate")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		log.Printf("Successfully renewed certificate")
	} else {
		return errors.New("Error trying to renew certificate - Invalid response from server")
	}

	certificates, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Error trying to renew certificate")
	}
	crtResponse := types.CRTResponse{}
	if err := json.Unmarshal(certificates, &crtResponse); err != nil {
		return errors.Wrap(err, "Error trying to renew certificate - Invalid response from server")
	}
	decodedCert, err := base64.StdEncoding.DecodeString(crtResponse.ClientCRT)
	if err != nil {
		return errors.Wrap(err, "Error trying to decode certificate")
	}

	c.Ca.PublicKey = string(decodedCert)

	if err := c.StorageInterface.WriteClientCert(c.Ca); err != nil {
		return errors.Wrap(err, "Error trying to write certificate data")
	}

	return nil
}

func (c *KymaConnector) RevokeCertificate() error {
	resp, err := c.SecureClient.Post(c.Info.URLs.RevokeCertUrl, "application/json", nil)
	if err != nil {
		return errors.Wrap(err, "Error trying to revoke certificate")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		log.Print("Successfully revoked certificate for client")
	} else {
		return errors.New("Error trying to revoke certificate - Invalid response from server")
	}

	return nil
}
