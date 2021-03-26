module github.com/kyma-incubator/hydroform/parallel-install

go 1.14

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/blang/semver/v4 v4.0.0
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/imdario/mergo v0.3.11
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	helm.sh/helm/v3 v3.5.1
	k8s.io/api v0.20.1
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/cli-runtime v0.20.1
	k8s.io/client-go v0.20.1
	rsc.io/letsencrypt v0.0.3 // indirect
)
