module github.com/kyma-incubator/hydroform/function-examples

go 1.14

replace github.com/kyma-incubator/hydroform/function => github.com/pPrecel/hydroform/function v0.0.0-20210215081157-046f256f8258

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/docker/docker v20.10.3+incompatible
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/kyma-incubator/hydroform/function v0.0.0-20201027094432-8e584f2623f7
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.18.9
)
