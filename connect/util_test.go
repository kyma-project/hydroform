package connect

import (
	"github.com/kyma-incubator/hydroform/connect/types"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestKymaConnector_populateCsrInfo(t *testing.T) {
	mockWriter := &MockWriter{}
	getCsrInfoServer := getCsrInfoServer(t, "test.com/csrurl", "test.com/infourl")
	type fields struct {
		CsrInfo          *types.CSRInfo
		AppName          string
		Ca               *types.ClientCertificate
		SecureClient     *http.Client
		StorageInterface WriterInterface
	}
	type args struct {
		configurationUrl string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.CSRInfo
		wantErr bool
	}{
		{
			name: "testPopulateCsrInfo",
			fields: fields{
				CsrInfo: &types.CSRInfo{},
				AppName: "testClient",
				Ca: &types.ClientCertificate{
					PrivateKey: "",
					PublicKey:  "",
				},
				SecureClient:     nil,
				StorageInterface: mockWriter,
			},
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
			name: "testPopulateCsrInfo_incorrectUrl",
			fields: fields{
				CsrInfo: &types.CSRInfo{
					CSRUrl:      "",
					API:         nil,
					Certificate: nil,
				},
				AppName:          "",
				Ca:               nil,
				SecureClient:     nil,
				StorageInterface: mockWriter,
			},
			args: args{configurationUrl: "incorrectConfigurationUrl"},
			want: &types.CSRInfo{
				CSRUrl:      "",
				API:         nil,
				Certificate: nil,
			},
			wantErr: true,
		},
		{
			name: "testPopulateCsrInfo_blankUrl",
			fields: fields{
				CsrInfo: &types.CSRInfo{
					CSRUrl:      "",
					API:         nil,
					Certificate: nil,
				},
				AppName:          "",
				Ca:               nil,
				SecureClient:     nil,
				StorageInterface: mockWriter,
			},
			args: args{configurationUrl: ""},
			want: &types.CSRInfo{
				CSRUrl:      "",
				API:         nil,
				Certificate: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &KymaConnector{
				CsrInfo:          tt.fields.CsrInfo,
				Ca:               tt.fields.Ca,
				SecureClient:     tt.fields.SecureClient,
				StorageInterface: tt.fields.StorageInterface,
			}

			_, err := c.populateCsrInfo(tt.args.configurationUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("populateCsrInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if c.CsrInfo != nil && !reflect.DeepEqual(c.CsrInfo, tt.want) {
				t.Errorf("populateCsrInfo() got = %v, want %v", c.CsrInfo, tt.want)
			}
		})
	}
}

func TestKymaConnector_populateCertSigningRequest(t *testing.T) {
	sendCsrToKymaServer := sendCsrToKymaServer(t)
	defer sendCsrToKymaServer.Close()

	type fields struct {
		CsrInfo      *types.CSRInfo
		AppName      string
		Ca           *types.ClientCertificate
		SecureClient *http.Client
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "testPopulateCert",
			fields: fields{
				CsrInfo: &types.CSRInfo{
					CSRUrl: sendCsrToKymaServer.URL,
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
				AppName: "testApplication",
				Ca: &types.ClientCertificate{
					PrivateKey: "",
					PublicKey:  "",
				},
				SecureClient: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &KymaConnector{
				CsrInfo:      tt.fields.CsrInfo,
				Ca:           tt.fields.Ca,
				SecureClient: tt.fields.SecureClient,
			}
			if _, err := c.populateClientCert(); (err != nil) != tt.wantErr {
				t.Errorf("populateClientCert() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !strings.HasPrefix(c.Ca.PrivateKey, "-----BEGIN RSA PRIVATE KEY-----") {
				t.Errorf("populateClientCert() Invalid key: %v", c.Ca.PrivateKey)
			}
		})
	}
}

func TestKymaConnector_populateClientCert(t *testing.T) {

	sendCsrToKymaServer := sendCsrToKymaServer(t)
	defer sendCsrToKymaServer.Close()

	getInfoServer := getInfoServer(t)
	defer getInfoServer.Close()

	getCsrInfoServer := getCsrInfoServer(t, sendCsrToKymaServer.URL, getInfoServer.URL)
	defer getCsrInfoServer.Close()

	type fields struct {
		CsrInfo      *types.CSRInfo
		AppName      string
		Ca           *types.ClientCertificate
		SecureClient *http.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "testPopulateClientCert",
			fields: fields{
				CsrInfo: &types.CSRInfo{
					CSRUrl: sendCsrToKymaServer.URL,
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
				AppName: "testApplication",
				Ca: &types.ClientCertificate{
					PrivateKey: "",
					PublicKey:  "",
				},
				SecureClient: nil,
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
			c := &KymaConnector{
				CsrInfo:      tt.fields.CsrInfo,
				Ca:           tt.fields.Ca,
				SecureClient: tt.fields.SecureClient,
			}
			if _, err := c.populateClientCert(); (err != nil) != tt.wantErr {
				t.Errorf("populateClientCert() error = %v, wantErr %v", err, tt.wantErr)
			}

			if c.Ca.PublicKey != tt.want {
				t.Errorf("populateClientCert() got = %v, want = %v", c.Ca.PublicKey, tt.want)
			}
		})
	}
}
