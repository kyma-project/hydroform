package types

type Attributes struct {
	Eventtypeversion string `json:"eventtypeversion"`
	Source           string `json:"source"`
	Type             string `json:"type"`
}

type Trigger struct {
	Spec struct {
		Filter struct {
			Attributes Attributes `json:"attributes"`
		} `json:"filter"`
	} `json:"spec"`
}
