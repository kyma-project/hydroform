package gcp

type GoogleProvider struct {
}

func (g *GoogleProvider) Provision() error {
	return nil
}

func (g *GoogleProvider) Status() error {
	return nil
}

func (g *GoogleProvider) Credentials() error {
	return nil
}

func (g *GoogleProvider) Deprovision() error {
	return nil
}

//Instantiate GCP provider
func New() *GoogleProvider {
	return &GoogleProvider{}
}
