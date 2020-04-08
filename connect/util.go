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
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type writerInterface interface {
	writeToFile(string, []byte) error
}

func (c *KymaConnector) getCsrInfo(configurationUrl string) error {
	url, _ := url.Parse(configurationUrl)

	resp, err := http.Get(url.String())

	if err != nil {
		return fmt.Errorf("error trying to get CSR Information : '%s'", err.Error())
	}
	if resp == nil {
		return fmt.Errorf("did not recieve a response from configuration URL : '%s'", url)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("recieved non OK status code from configuration URL : '%s'", url)
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error trying to parse JSON : '%s'", err.Error())
	}

	csrInfo := types.CSRInfo{}
	err = json.Unmarshal(response, &csrInfo)
	if err != nil {
		return fmt.Errorf("error trying to get CSR Information : '%s'", err.Error())
	}

	c.CsrInfo = &csrInfo
	err = c.writeToFile("config.json", response)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return err
}

func (c *KymaConnector) getCertSigningRequest() error {
	keys, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

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

	c.AppName = appName

	pkixName := pkix.Name{
		Organization:       []string{org},
		OrganizationalUnit: []string{orgUnit},
		Locality:           []string{location},
		StreetAddress:      []string{street},
		Country:            []string{country},
		CommonName:         appName,
		Province:           []string{"Waldorf"}, // KAVYA - gives error if empty string provided / string not provided, should be returned in subject field ideally with other data?
	}

	// create CSR
	var csrTemplate = x509.CertificateRequest{
		Subject:            pkixName,
		SignatureAlgorithm: x509.SHA256WithRSA, // KAVYA - add extensions
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	var privateKey bytes.Buffer
	err = pem.Encode(&privateKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keys)})

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	c.Ca.PrivateKey = privateKey.String()
	c.Ca.Csr = string(csr)
	return err
}

func (c *KymaConnector) getClientCert() error {

	// encode CSR to base64
	encodedCsr := base64.StdEncoding.EncodeToString([]byte(c.Ca.Csr))
	type Payload struct {
		Csr string `json:"csr"`
	}

	data := Payload{
		Csr: encodedCsr,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", c.CsrInfo.CSRUrl, body)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	defer resp.Body.Close()
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

func (c *KymaConnector) writeClientCertificateToFile(w writerInterface) error {
	if c.Ca.Csr != "" {
		err := w.writeToFile("generated.csr", []byte(c.Ca.Csr))
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	if c.Ca.PublicKey != "" {
		err := w.writeToFile("generated.crt", []byte(c.Ca.PublicKey))
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	if c.Ca.PrivateKey != "" {
		err := w.writeToFile("generated.key", []byte(c.Ca.PrivateKey))
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func (c *KymaConnector) readService(path string, s *Service) error {
	_, err := os.Stat(path)
	if err != nil {
		log.Println("No service config available")
		return err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Failed to read file")
		return err
	}

	err = json.Unmarshal(b, s)
	if err != nil {
		log.Println("Failed to parse json")
		return err
	}

	return nil
}

func (c *KymaConnector) getRawJsonFromDoc(doc string) (m json.RawMessage, err error) {
	bytes, err := ioutil.ReadFile(doc)
	if err != nil {
		log.Println("Read error on API Docs")
		return
	}
	m = json.RawMessage(string(bytes[:]))
	return
}

func (c *KymaConnector) writeToFile(fileName string, data []byte) error {
	return ioutil.WriteFile(fileName, data, 0644)
}

func (c *KymaConnector) populateClient() (err error) {
	c.SecureClient, err = c.GetSecureClient()
	return err
}

func (c *KymaConnector) populateAppName() {
	parts := strings.Split(c.CsrInfo.Certificate.Subject, ",")
	for i := range parts {
		subjectTitle := strings.Split(parts[i], "=")
		switch subjectTitle[0] {
		case "CN":
			c.AppName = subjectTitle[1]
		}
	}
}
func (c *KymaConnector) loadConfig() error {

	_, err := os.Stat("config.json")
	if err == nil {
		config, err := ioutil.ReadFile("config.json")
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		csrInfo := &types.CSRInfo{}
		json.Unmarshal(config, csrInfo)
		c.CsrInfo = csrInfo
	}

	_, err = os.Stat("generated.csr")
	if err == nil {
		csr, err := ioutil.ReadFile("generated.csr")
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		c.Ca.Csr = string(csr[:])
	}

	_, err = os.Stat("generated.crt")
	if err == nil {
		crt, err := ioutil.ReadFile("generated.crt")
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		c.Ca.PublicKey = string(crt[:])
	}

	_, err = os.Stat("generated.key")
	if err == nil {
		key, err := ioutil.ReadFile("generated.key")
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		c.Ca.PrivateKey = string(key[:])
	}

	c.populateAppName()
	return err
}
