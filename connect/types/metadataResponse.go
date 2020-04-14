package types

type Info struct {
	ClientIdentity *ClientIdentity `json:"clientIdentity"`
	URLs           *URLs           `json:"urls"`
}

type ClientIdentity struct {
	AppName string `json:"application"`
}

type URLs struct {
	MetadataUrl   string `json:"metadataUrl"`
	EventsUrl     string `json:"eventsUrl"`
	RenewCertUrl  string `json:"renewCertUrl"`
	RevokeCertUrl string `json:"revokeCertUrl"`
}
