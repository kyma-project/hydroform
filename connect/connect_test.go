package connect

import (
	"encoding/json"
	"github.com/kyma-incubator/hydroform/connect/types"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
func Test_getCsrInfo(t *testing.T) {

	getCsrInfoServer := getCsrInfoServer(t, "test.com/csrurl")
	defer getCsrInfoServer.Close()

	type args struct {
		configurationUrl string
	}

	tests := []struct {
		name    string
		args    args
		want    *types.CSRInfo
		wantErr bool
	}{
		{
			name: "correctUrl",
			args: args{configurationUrl: getCsrInfoServer.URL},
			want: &types.CSRInfo{
				CSRUrl: "test.com/csrurl",
				API: &types.API{
					MetadataUrl:     "test.com/metadataurl",
					EventsUrl:       "test.com/eventsurl",
					EventsInfoUrl:   "test.com/eventsinfourl",
					InfoUrl:         "test.com/infourl",
					CertificatesUrl: "test.com/certificatesurl",
				},
				Certificate: &types.Certificate{
					Subject:      "O=Organization,OU=OrgUnit,L=Waldorf,ST=Waldorf,C=DE,CN=testApplication",
					Extensions:   "",
					KeyAlgorithm: "rsa2048",
				},
			},
			wantErr: false,
		},
		{
			name:    "incorrectURL",
			args:    args{configurationUrl: "configurl"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "blankURL",
			args:    args{configurationUrl: ""},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := getCsrInfo(tt.args.configurationUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCsrInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCsrInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCertSigningRequest(t *testing.T) {
	type args struct {
		subject string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "correct",
			args:    args{subject: "O=Organization,OU=OrgUnit,L=Waldorf,ST=Waldorf,C=DE,CN=myTestApplication"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getCertSigningRequest(tt.args.subject)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCertSigningRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.HasPrefix(got, "-----BEGIN CERTIFICATE REQUEST-----") {
				t.Errorf("getCertSigningRequest() Invalid CSR: %v", got)
			}
			if !strings.HasPrefix(got1, "-----BEGIN RSA PRIVATE KEY-----") {
				t.Errorf("getCertSigningRequest() Invalid key: %v", got1)
			}
		})
	}
}

func Test_getClientCert(t *testing.T) {

	sendCsrToKymaServer := sendCsrToKymaServer(t)
	defer sendCsrToKymaServer.Close()

	getCsrInfoServer := getCsrInfoServer(t, sendCsrToKymaServer.URL)
	defer getCsrInfoServer.Close()

	type args struct {
		csrUrl string
		csr    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "correct",
			args: args{
				csrUrl: sendCsrToKymaServer.URL,
				csr:    "",
			},
			want: "-----BEGIN CERTIFICATE-----\n" +
				"MIIEHzCCAgegAwIBAgIBAjANBgkqhkiG9w0BAQsFADAPMQ0wCwYDVQQDEwRLeW1h\n" +
				"MB4XDTIwMDMyMDEwMDAwOFoXDTIwMDYyMDEwMDAwOFowbjELMAkGA1UEBhMCREUx\n" +
				"EDAOBgNVBAgTB1dhbGRvcmYxEDAOBgNVBAcTB1dhbGRvcmYxFTATBgNVBAoTDE9y\n" +
				"Z2FuaXphdGlvbjEQMA4GA1UECxMHT3JnVW5pdDESMBAGA1UEAxMJbXl0ZXN0YXBw\n" +
				"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9HE+5tVADtjeHBaOqIn4\n" +
				"Lf/LW+ZGGY7DB700kdkP7H0JaFIOJx1RAX+deVIDXX2/WtPVJ4M+cwuFjCR+OPme\n" +
				"em4dG9suY2oh/qcwZgXIq5mMvtWamyDxu49TXXAyLNqIaCu0T12D5SaTFxeZf45U\n" +
				"NkXWi0l0OXNps9qkvbjWAuawy95n0l8gCFfCDRZHvl96TJRzu+SHK7nfp1rw7kZ1\n" +
				"An4KkwKwuk8lMT9nWpbHelrjdO8sXa1qjkA67Di/4QarMuEMG1BxeC/GTYRq37VB\n" +
				"Hp5iEpbjOCsxDBpBYZy90lZVNzy7LO+TAN8a0Ogh7Qhn6RmshOgv/cNW3OXNlzDY\n" +
				"zQIDAQABoycwJTAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0lBAwwCgYIKwYBBQUHAwIw\n" +
				"DQYJKoZIhvcNAQELBQADggIBADUAYmNMaofpV0n/aqtw07XZ1DAeyyuR43EBEbMy\n" +
				"XBfmzUt8qK+bRxL1ipHxcpR92QPceqtXEairlpI+CqwCIt7zE6oNR4Jabpp0iFpN\n" +
				"rwkxTbGVk9u+uMqUk8fYgUGJET2AEKag7WT9zTrKhfBSx1gBfAvHLWiqfSExOtfZ\n" +
				"PJTJy0Y3BMt9WO8T13yvf2wzAW3aJZsSmKYA6nwkEg0p5kIV2lw35ANnYcnlhOW6\n" +
				"G/2JxersUSNzhYDfgJhILk/t7huFM3RJhcfD5Y5+QG/agTFT6Qilw8frGsSwdWYJ\n" +
				"gwH5hFYQNBV71Ss7xZ/9S5qzH84Ow+ylBxomO756hc2z1K8I9ANYbnR+6H2XjFd2\n" +
				"6p5zV/5qsRJ2pnQefPw50O1Q9zKR0UIdw/9yY3sjQf9kqvVYCzpWOCRVAYpZvty2\n" +
				"hPP5Ew0o7raqHGOYSajZ5j2h3QDIB3v13IdaIdZm16GW4St9w5uDZkEqBVro5Nep\n" +
				"buFXqLZDXP3LARWhP3SM71/0Ya15DNNyOKsCPrxa50M9L3Wqj1pv8cwohVv2OPZi\n" +
				"f5azLp6oF1tS5JCu8QAF4CSAGcnPM/g5AJb3ze6aawahQuq8flxbG4Pq61WL6s52\n" +
				"nngI0TIlO4g+fIH9IVQtor5EMMEnpIzfIc8VRiI3t+mEP0Isb7vjMTICrZFxQGhJ\n" +
				"vO5Z\n" +
				"-----END CERTIFICATE-----\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getClientCert(tt.args.csrUrl, tt.args.csr)
			if (err != nil) != tt.wantErr {
				t.Errorf("getClientCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getClientCert() got = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
func sendCsrToKymaServer(t *testing.T) *httptest.Server {
	sendCsrToKymaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}
		reqJson, err := ioutil.ReadAll(r.Body)

		csrResponse := types.CSRResponse{}
		err = json.Unmarshal(reqJson, &csrResponse)

		if err != nil {
			t.Errorf("Unexpected error in parsing JSON ")
		}
		//fmt.Print(csrResponse)

		crtResponse := types.CRTResponse{
			CRT:       "crtEncoded",
			CaCRT:     "caCrtEncoded",
			ClientCRT: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVIekNDQWdlZ0F3SUJBZ0lCQWpBTkJna3Foa2lHOXcwQkFRc0ZBREFQTVEwd0N3WURWUVFERXdSTGVXMWgKTUI0WERUSXdNRE15TURFd01EQXdPRm9YRFRJd01EWXlNREV3TURBd09Gb3diakVMTUFrR0ExVUVCaE1DUkVVeApFREFPQmdOVkJBZ1RCMWRoYkdSdmNtWXhFREFPQmdOVkJBY1RCMWRoYkdSdmNtWXhGVEFUQmdOVkJBb1RERTl5CloyRnVhWHBoZEdsdmJqRVFNQTRHQTFVRUN4TUhUM0puVlc1cGRERVNNQkFHQTFVRUF4TUpiWGwwWlhOMFlYQncKTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUE5SEUrNXRWQUR0amVIQmFPcUluNApMZi9MVytaR0dZN0RCNzAwa2RrUDdIMEphRklPSngxUkFYK2RlVklEWFgyL1d0UFZKNE0rY3d1RmpDUitPUG1lCmVtNGRHOXN1WTJvaC9xY3daZ1hJcTVtTXZ0V2FteUR4dTQ5VFhYQXlMTnFJYUN1MFQxMkQ1U2FURnhlWmY0NVUKTmtYV2kwbDBPWE5wczlxa3ZialdBdWF3eTk1bjBsOGdDRmZDRFJaSHZsOTZUSlJ6dStTSEs3bmZwMXJ3N2taMQpBbjRLa3dLd3VrOGxNVDluV3BiSGVscmpkTzhzWGExcWprQTY3RGkvNFFhck11RU1HMUJ4ZUMvR1RZUnEzN1ZCCkhwNWlFcGJqT0NzeERCcEJZWnk5MGxaVk56eTdMTytUQU44YTBPZ2g3UWhuNlJtc2hPZ3YvY05XM09YTmx6RFkKelFJREFRQUJveWN3SlRBT0JnTlZIUThCQWY4RUJBTUNCNEF3RXdZRFZSMGxCQXd3Q2dZSUt3WUJCUVVIQXdJdwpEUVlKS29aSWh2Y05BUUVMQlFBRGdnSUJBRFVBWW1OTWFvZnBWMG4vYXF0dzA3WFoxREFleXl1UjQzRUJFYk15ClhCZm16VXQ4cUsrYlJ4TDFpcEh4Y3BSOTJRUGNlcXRYRWFpcmxwSStDcXdDSXQ3ekU2b05SNEphYnBwMGlGcE4KcndreFRiR1ZrOXUrdU1xVWs4ZllnVUdKRVQyQUVLYWc3V1Q5elRyS2hmQlN4MWdCZkF2SExXaXFmU0V4T3RmWgpQSlRKeTBZM0JNdDlXTzhUMTN5dmYyd3pBVzNhSlpzU21LWUE2bndrRWcwcDVrSVYybHczNUFOblljbmxoT1c2CkcvMkp4ZXJzVVNOemhZRGZnSmhJTGsvdDdodUZNM1JKaGNmRDVZNStRRy9hZ1RGVDZRaWx3OGZyR3NTd2RXWUoKZ3dINWhGWVFOQlY3MVNzN3haLzlTNXF6SDg0T3creWxCeG9tTzc1NmhjMnoxSzhJOUFOWWJuUis2SDJYakZkMgo2cDV6Vi81cXNSSjJwblFlZlB3NTBPMVE5ektSMFVJZHcvOXlZM3NqUWY5a3F2VllDenBXT0NSVkFZcFp2dHkyCmhQUDVFdzBvN3JhcUhHT1lTYWpaNWoyaDNRRElCM3YxM0lkYUlkWm0xNkdXNFN0OXc1dURaa0VxQlZybzVOZXAKYnVGWHFMWkRYUDNMQVJXaFAzU003MS8wWWExNUROTnlPS3NDUHJ4YTUwTTlMM1dxajFwdjhjd29oVnYyT1BaaQpmNWF6THA2b0YxdFM1SkN1OFFBRjRDU0FHY25QTS9nNUFKYjN6ZTZhYXdhaFF1cThmbHhiRzRQcTYxV0w2czUyCm5uZ0kwVElsTzRnK2ZJSDlJVlF0b3I1RU1NRW5wSXpmSWM4VlJpSTN0K21FUDBJc2I3dmpNVElDclpGeFFHaEoKdk81WgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==",
		}
		js, err := json.Marshal(crtResponse)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	}))
	return sendCsrToKymaServer
}

func getCsrInfoServer(t *testing.T, csrUrl string) *httptest.Server {
	getCsrInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got '%s", r.Method)
		}

		csrInfo := types.CSRInfo{
			CSRUrl: csrUrl,
			API: &types.API{
				MetadataUrl:     "test.com/metadataurl",
				EventsUrl:       "test.com/eventsurl",
				EventsInfoUrl:   "test.com/eventsinfourl",
				InfoUrl:         "test.com/infourl",
				CertificatesUrl: "test.com/certificatesurl",
			},
			Certificate: &types.Certificate{
				Subject:      "O=Organization,OU=OrgUnit,L=Waldorf,ST=Waldorf,C=DE,CN=testApplication",
				Extensions:   "",
				KeyAlgorithm: "rsa2048",
			},
		}
		js, err := json.Marshal(csrInfo)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}))

	return getCsrInfoServer
}
