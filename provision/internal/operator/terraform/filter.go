package terraform

import "github.com/kyma-incubator/hydroform/provision/types"

// varFilter defines the signature to filter variables out of the tfvars file.
// varFilter should return true if the given key and/or value should be added to the tfvars file. False if it should not be in the tfvars file.
// This is meant to keep tfvars files clean and only with the strictly necessary variables for each provider.
type varFilter func(key string, value interface{}) bool

func gcpFilter(key string, value interface{}) bool {
	// by default all keys stay in the vars for GCP
	return true
}

func azureFilter(key string, value interface{}) bool {
	excludedKeys := []string{"project", "create_timeout", "update_timeout", "delete_timeout"}

	for _, e := range excludedKeys {
		if key == e {
			return false
		}
	}
	return true
}

func gardenerFilter(key string, value interface{}) bool {
	// by default all keys stay in the vars for GCP
	return true
}

func kindFilter(key string, value interface{}) bool {
	// by default all keys stay in the vars for GCP
	return true
}

// filterVars takes the full hydroform configuration map and given a provider, it fetches its filter function and removes the keys that should not be there.
// Each provider should implement varFilter to control which vars it should have in its tfvars file.
func filterVars(cfg map[string]interface{}, p types.ProviderType) map[string]interface{} {
	vars := make(map[string]interface{})
	var f varFilter
	switch p {
	case types.GCP:
		f = gcpFilter
	case types.Gardener:
		f = gardenerFilter
	case types.Azure:
		f = azureFilter
	case types.AWS:
		f = nil // not supported yet
	case types.Kind:
		f = kindFilter
	}

	for key, value := range cfg {
		if f(key, value) {
			vars[key] = value
		}
	}
	return vars
}
