module github.com/kyma-incubator/hydroform/parallel-install

go 1.14

replace (
	//TODO: remove this part as Helm 3.5.4 got released
	//see https://github.com/helm/helm/issues/9354 + https://github.com/helm/helm/pull/9492
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v20.10.6+incompatible
)

require (
	github.com/alcortesm/tgz v0.0.0-20161220082320-9c5fe88206d7
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/blang/semver/v4 v4.0.0
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/docker/docker v20.10.6+incompatible
	github.com/fatih/structs v1.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-git/go-git/v5 v5.3.0
	github.com/google/uuid v1.2.0
	github.com/imdario/mergo v0.3.12
	github.com/kubernetes-sigs/service-catalog v0.3.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	helm.sh/helm/v3 v3.5.3 //Before upgrading: please see TODO comment in replace() section on top!
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.20.2
)
