package types

type ClientCertificate struct {
	PrivateKey string
	PublicKey  string
	Csr        string
}
