package metadata

// StatusEnum describes deployment / uninstallation status
type StatusEnum string

const (
	//DeploymentInProgress means deployment of kyma is in progress
	DeploymentInProgress StatusEnum = "DeploymentInProgress"

	//UninstallationInProgress means uninstallation of kyma is in progress
	UninstallationInProgress StatusEnum = "Uninstallation in progress"

	//DeploymentError means error occurred during kyma deployment
	DeploymentError StatusEnum = "DeploymentError"

	//UninstallationError means error occurred during kyma uninstallation
	UninstallationError StatusEnum = "UninstallationError"

	//Deployed means kyma deployed successfully
	Deployed StatusEnum = "Deployed"
)

type KymaMetadata struct {
	Profile           string
	Version           string
	ComponentListData []byte
	ComponentListFile string
	Status            StatusEnum
	Reason            string
}

func (km *KymaMetadata) withAttributes(attr *Attributes) *KymaMetadata {
	km.Version = attr.version
	km.Profile = attr.profile
	km.ComponentListData = attr.componentListData
	km.ComponentListFile = attr.componentListFile
	return km
}

func (km *KymaMetadata) withError(err StatusEnum, reason string) *KymaMetadata {
	km.Status = err
	km.Reason = reason
	return km
}

func (km *KymaMetadata) withStatus(status StatusEnum) *KymaMetadata {
	km.Status = status
	return km
}
