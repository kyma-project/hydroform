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
		Subscriber struct {
			Reference struct {
				Kind      string `json:"kind"`
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"ref"`
		} `json:"subscriber"`
	} `json:"spec"`
}

func (t Trigger) IsReference(name, namespace string) bool {
	return t.Spec.Subscriber.Reference.Kind == "Service" &&
		t.Spec.Subscriber.Reference.Name == name &&
		t.Spec.Subscriber.Reference.Namespace == namespace
}
