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
	"strings"
)

func Connect(configurationUrl string) (string, error) {

	//get CSR Information From Kyma
	resp, err := http.Get(configurationUrl)
	if err != nil {
		return "", fmt.Errorf("error trying to get CSR Information : '%s'", err.Error())
	}

	if resp == nil {
		return "", fmt.Errorf("did not recieve a response from configuration URL : '%s'", configurationUrl)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("recieved non OK status code from configuration URL : '%s'", configurationUrl)
	}

	//unmarshal response json and store in csrInfo
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error trying to parse JSON : '%s'", err.Error())
	}

	csrInfo := types.CSRInfo{}
	err = json.Unmarshal(response, &csrInfo)
	if err != nil {
		return "", fmt.Errorf("error trying to get CSR Information : '%s'", err.Error())
	}

	// gererate RSA key
	keys, err := rsa.GenerateKey(rand.Reader, 2048)
	subject := csrInfo.Certificate.Subject
	parts := strings.Split(subject, ",")

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

	// create CSR
	var csrTemplate = x509.CertificateRequest{
		Subject: pkix.Name{
			Organization:       []string{org},
			OrganizationalUnit: []string{orgUnit},
			Locality:           []string{location},
			StreetAddress:      []string{street},
			Country:            []string{country},
			CommonName:         appName,
			Province:           []string{"Waldorf"},
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	// encode CSR to base64
	encodedCsr := base64.StdEncoding.EncodeToString(csr)

	// store it in file?
	//	fmt.Print(encodedCsr)

	type Payload struct {
		Csr string `json:"csr"`
	}

	// send Csr to Kyma
	data := Payload{
		// fill struct
		Csr: encodedCsr,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		// handle err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", csrInfo.CSRUrl, body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	certificates, err := ioutil.ReadAll(resp.Body)
	crtResponse := types.CRTResponse{}
	err = json.Unmarshal(certificates, &crtResponse)
	if err != nil {
		return "", fmt.Errorf("JSON Error")
	}

	return crtResponse.ClientCRT, err //returning client certificate
}
