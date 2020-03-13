package types

type CRTResponse struct {
	CRT       string `json:"crt"`
	ClientCRT string `json:"clientCrt"`
	CaCRT     string `json:"caCrt"`
}
