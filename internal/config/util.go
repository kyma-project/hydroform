package config

func ExtendConfig(sourceCfg map[string]interface{}, extensions map[string]interface{}) {
	for k, v := range extensions {
		sourceCfg[k] = v
	}
}
