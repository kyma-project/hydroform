module github.com/kyma-incubator/hydroform/installation-poc

go 1.14

require (
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/kyma-project/kyma/components/kyma-operator v0.0.0-20201020070353-8d6c1b9037cc
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	helm.sh/helm/v3 v3.3.4
	k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime v0.18.8
	k8s.io/client-go v0.18.8
)
