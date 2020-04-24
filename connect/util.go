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
	"github.com/kyma-incubator/hydroform/connect/types"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func (c *KymaConnector) populateCsrInfo(configurationUrl string) (*types.CSRInfo, error) {
	url, err := url.Parse(configurationUrl)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.New("Did not receive a response from the config URL")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Received invalid response from config URL")
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	csrInfo := types.CSRInfo{}
	if err := json.Unmarshal(response, &csrInfo); err != nil {
		return nil, err
	}

	c.CsrInfo = &csrInfo
	if err := c.Storage.WriteConfig(&csrInfo); err != nil {
		return nil, err
	}

	return &csrInfo, nil
}

func (c *KymaConnector) populateInfo() (*types.Info, error) {

	resp, err := c.SecureClient.Get(c.CsrInfo.API.InfoUrl)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.New("Did not receive a response from info URL")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Received invalid response from info URL")
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	info := types.Info{}
	if err := json.Unmarshal(response, &info); err != nil {
		return nil, err
	}

	c.Info = &info
	if err := c.Storage.WriteInfo(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (c *KymaConnector) populateClientCert() (*types.ClientCertificate, error) {

	certificate := &types.ClientCertificate{}
	if c.Ca == nil {
		c.Ca = certificate
	}

	keys, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	csr, err := c.getCsr(keys)
	if err != nil {
		return nil, err
	}
	var privateKey bytes.Buffer
	if err := pem.Encode(&privateKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keys)}); err != nil {
		return nil, err
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
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", c.CsrInfo.CSRUrl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	certificates, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	crtResponse := types.CRTResponse{}

	if err = json.Unmarshal(certificates, &crtResponse); err != nil {
		return nil, err
	}

	decodedCert, err := base64.StdEncoding.DecodeString(crtResponse.ClientCRT)
	if err != nil {
		return nil, err
	}

	c.Ca.PublicKey = string(decodedCert)
	certificate.PublicKey = string(decodedCert)

	return certificate, nil
}

func (c *KymaConnector) populateClient() (err error) {
	if c.Ca != nil {
		c.SecureClient, err = c.GetSecureClient()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *KymaConnector) loadConfig() {
	err := errors.New("")

	c.CsrInfo, err = c.Storage.ReadConfig()
	if err != nil {
		log.Printf("Could not find existing config stored")
	}
	c.Info, err = c.Storage.ReadInfo()
	if err != nil {
		log.Printf("Could not find existing info stored")
	}

	c.Ca, err = c.Storage.ReadClientCert()
	if err != nil {
		log.Printf("Could not find existing certificates stored")
	}
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
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	}), nil

}
