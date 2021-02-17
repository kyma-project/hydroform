package preinstaller

type preInstallerResource struct {
	name      string
	validator func(data string) bool
}

func newCrdPreInstallerResource() *preInstallerResource {
	return &preInstallerResource{
		name: "crds",
		validator: func(data string) bool {
			if len(data) < 1 {
				return false
			}

			return true
		},
	}
}

func newNamespacePreInstallerResource() *preInstallerResource {
	return &preInstallerResource{
		name: "namespaces",
		validator: func(data string) bool {
			if len(data) < 1 {
				return false
			}

			return true
		},
	}
}
