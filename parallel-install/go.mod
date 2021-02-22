module github.com/kyma-incubator/hydroform/parallel-install

go 1.14

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/imdario/mergo v0.3.8
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	helm.sh/helm/v3 v3.4.2
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/cli-runtime v0.19.4
	k8s.io/client-go v0.19.4
	rsc.io/letsencrypt v0.0.3 // indirect
)
