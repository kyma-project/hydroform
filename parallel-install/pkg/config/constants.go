package config

const (
	//LABEL_KEY_ORIGIN is used for marking where resource comes from.
	LABEL_KEY_ORIGIN = "origin"

	//LABEL_VALUE_KYMA indicates that resource is managed by Kyma.
	//Used for marking CRDs, so they can be deleted during uninstallation.
	LABEL_VALUE_KYMA = "kyma"
)

