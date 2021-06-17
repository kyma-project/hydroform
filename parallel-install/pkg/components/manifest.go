package components

type ManifestType string

const (
	CRD       ManifestType = "crd"
	HelmChart ManifestType = "helmChart"
)

type Manifest struct {
	Type      ManifestType
	Name      string
	Manifest  string
	Component string
}
