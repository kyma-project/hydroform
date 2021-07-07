module github.com/kyma-incubator/hydroform/install

go 1.13

replace k8s.io/apimachinery => k8s.io/apimachinery v0.21.2

require (
	github.com/kyma-project/kyma/components/kyma-operator v0.0.0-20200817094157-8392259f5be1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)
