package types

type CSRInfo struct {
	CSRUrl      string       `json:"csrUrl"`
	API         *API         `json:"api"`
	Certificate *Certificate `json:"certificate"`
}

type API struct {
	MetadataUrl     string `json:"metadataUrl"`
	EventsUrl       string `json:"eventsUrl"`
	EventsInfoUrl   string `json:"eventsInfoUrl"`
	InfoUrl         string `json:"infoUrl"`
	CertificatesUrl string `json:"certificatesUrl"`
}

type Certificate struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}
