package connect

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/kyma-incubator/hydroform/connect/types"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func (c *KymaConnector) populateCsrInfo(configurationUrl string) (*types.CSRInfo, error) {
	url, _ := url.Parse(configurationUrl)

	resp, err := http.Get(url.String())

	if err != nil {
		return nil, fmt.Errorf("error trying to get CSR Information : '%s'", err.Error())
	}
	if resp == nil {
		return nil, fmt.Errorf("did not recieve a response from configuration URL : '%s'", url)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recieved non OK status code from configuration URL : '%s'", url)
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error trying to parse JSON : '%s'", err.Error())
	}

	csrInfo := types.CSRInfo{}
	err = json.Unmarshal(response, &csrInfo)
	if err != nil {
		return nil, fmt.Errorf("error trying to get CSR Information : '%s'", err.Error())
	}

	c.CsrInfo = &csrInfo
	err = c.StorageInterface.WriteConfig(&csrInfo)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return &csrInfo, err
}

func (c *KymaConnector) populateInfo() (*types.Info, error) {

	resp, err := c.SecureClient.Get(c.CsrInfo.API.InfoUrl)

	if err != nil {
		return nil, fmt.Errorf("error trying to get info : '%s'", err.Error())
	}
	if resp == nil {
		return nil, fmt.Errorf("did not recieve a response from info URL")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recieved non OK status code from info URL")
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error trying to parse JSON : '%s'", err.Error())
	}

	info := types.Info{}
	err = json.Unmarshal(response, &info)
	if err != nil {
		return nil, fmt.Errorf("error trying to get CSR Information : '%s'", err.Error())
	}

	c.Info = &info

	err = c.StorageInterface.WriteInfo(&info)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return &info, err
}

func (c *KymaConnector) populateClientCert() (*types.ClientCertificate, error) {

	certificate := &types.ClientCertificate{}

	keys, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	csr, err := c.getCsr(keys)
	var privateKey bytes.Buffer
	err = pem.Encode(&privateKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keys)})

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	c.Ca.PrivateKey = privateKey.String()
	certificate.PrivateKey = privateKey.String()

	encodedCsr := base64.StdEncoding.EncodeToString(csr)
	type Payload struct {
		Csr string `json:"csr"`
	}

	data := Payload{
		Csr: encodedCsr,
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", c.CsrInfo.CSRUrl, body)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resp.Body.Close()
	certificates, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	crtResponse := types.CRTResponse{}
	err = json.Unmarshal(certificates, &crtResponse)
	if err != nil {
		return nil, fmt.Errorf("JSON Error")
	}
	decodedCert, err := base64.StdEncoding.DecodeString(crtResponse.ClientCRT)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	c.Ca.PublicKey = string(decodedCert)

	certificate.PublicKey = string(decodedCert)
	return certificate, err
}

func (c *KymaConnector) populateClient() (err error) {
	c.SecureClient, err = c.GetSecureClient()
	return err
}

func (c *KymaConnector) loadConfig() (err error) {
	c.CsrInfo, err = c.StorageInterface.ReadConfig()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	c.Info, err = c.StorageInterface.ReadInfo()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	c.Ca, err = c.StorageInterface.ReadClientCert()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return err
}

func (c *KymaConnector) getCsr(keys *rsa.PrivateKey) ([]byte, error) {
	parts := strings.Split(c.CsrInfo.Certificate.Subject, ",")

	var org, orgUnit, location, street, country, appName string

	for i := range parts {
		subjectTitle := strings.Split(parts[i], "=")
		switch subjectTitle[0] {
		case "O":
			org = subjectTitle[1]
		case "OU":
			orgUnit = subjectTitle[1]
		case "L":
			location = subjectTitle[1]
		case "ST":
			street = subjectTitle[1]
		case "C":
			country = subjectTitle[1]
		case "CN":
			appName = subjectTitle[1]
		}
	}

	pkixName := pkix.Name{
		Organization:       []string{org},
		OrganizationalUnit: []string{orgUnit},
		Locality:           []string{location},
		StreetAddress:      []string{street},
		Country:            []string{country},
		CommonName:         appName,
		Province:           []string{"Waldorf"},
	}

	// create CSR
	var csrTemplate = x509.CertificateRequest{
		Subject:            pkixName,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	}), err

}
