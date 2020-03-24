package types

type ClientCertificate struct {
	PrivateKey string
	PublicKey  string
	Csr        string

	PrivateKeyPath string
	PublicKeyPath  string
	CsrPath        string

	ServerCertPath string
}
