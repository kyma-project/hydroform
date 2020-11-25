module github.com/kyma-incubator/hydroform/parallel-install

go 1.14

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/kyma-project/kyma/components/kyma-operator v0.0.0-20201020070353-8d6c1b9037cc
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	helm.sh/helm/v3 v3.3.4
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime v0.18.8
	k8s.io/client-go v0.18.8
)
