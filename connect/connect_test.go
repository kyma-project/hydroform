package connect

import (
	"encoding/json"
	"github.com/kyma-incubator/hydroform/connect/types"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConnect(t *testing.T) {

	sendCsrToKymaServer := sendCsrToKymaServer(t)
	defer sendCsrToKymaServer.Close()

	getCsrInfoServer := getCsrInfoServer(t, sendCsrToKymaServer.URL)
	defer getCsrInfoServer.Close()

	type args struct {
		configurationUrl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "correctURL",
			args: args{
				configurationUrl: getCsrInfoServer.URL,
			},
			wantErr: false,
			want:    "clientCrtEncoded",
		},
		{
			name: "blankURL",
			args: args{
				configurationUrl: "",
			},
			wantErr: true,
		},
		{
			name: "incorrectURL",
			args: args{
				configurationUrl: "incorrectUrl",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := Connect(tt.args.configurationUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Connect() got = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			ClientCRT: "clientCrtEncoded",
			CaCRT:     "CaCrtEncoded",
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
				MetadataUrl:     "456",
				EventsUrl:       "789",
				EventsInfoUrl:   "012",
				InfoUrl:         "345",
				CertificatesUrl: "678",
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
