package k8s

//MergeMaps copies the keys that don't exist in the new map from the original map
func MergeMaps(new, original map[string]interface{}) map[string]interface{} {
	for key, originalValue := range original {
		newValue, exists := new[key]
		if !exists {
			new[key] = originalValue
		} else {
			nextNew, ok1 := newValue.(map[string]interface{})
			nextOriginal, ok2 := originalValue.(map[string]interface{})
			if ok1 && ok2 {
				MergeMaps(nextNew, nextOriginal)
			}
		}
	}
	return new
}
