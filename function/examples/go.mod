module github.com/kyma-incubator/hydroform/function-examples

go 1.14

replace github.com/kyma-incubator/hydroform/function => github.com/m00g3n/hydroform/function v0.0.0-20201021195651-c9b70f93d49d

require (
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/kyma-incubator/hydroform/function v0.0.0-20201013144143-a2b21fbd1824
	github.com/kyma-project/kyma/components/function-controller v0.0.0-20201014113918-b2b508b70d2e
	github.com/sirupsen/logrus v1.7.0
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.18.9
)
