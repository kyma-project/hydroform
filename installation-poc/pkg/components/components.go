package components

import (
	"path"

	"github.com/kyma-incubator/hydroform/installation-poc/pkg/helm"
	"github.com/kyma-incubator/hydroform/installation-poc/pkg/overrides"
)

type Provider struct {
	overridesProvider overrides.OverridesProvider
	path              string
}

func NewComponents(overridesProvider overrides.OverridesProvider, path string) *Provider {
	return &Provider{
		overridesProvider: overridesProvider,
		path:              path,
	}
}

type ComponentsProvider interface {
	GetComponents() ([]Component, error)
}

func (p *Provider) GetComponents() ([]Component, error) {
	helmClient := &helm.Client{}

	err := p.overridesProvider.ReadOverridesFromCluster()
	if err != nil {
		return nil, err
	}

	return []Component{
		Component{
			Name:       "istio-kyma-patch",
			Namespace:  "istio-system",
			Overrides:  p.overridesProvider.OverridesFor("istio-kyma-patch"),
			ChartDir:   path.Join(p.path, "istio-kyma-patch"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "knative-eventing",
			Namespace:  "knative-eventing",
			Overrides:  p.overridesProvider.OverridesFor("knative-eventing"),
			ChartDir:   path.Join(p.path, "knative-eventing"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "dex",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("dex"),
			ChartDir:   path.Join(p.path, "dex"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "ory",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("ory"),
			ChartDir:   path.Join(p.path, "ory"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "api-gateway",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("api-gateway"),
			ChartDir:   path.Join(p.path, "api-gateway"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "rafter",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("rafter"),
			ChartDir:   path.Join(p.path, "rafter"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "service-catalog",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("service-catalog"),
			ChartDir:   path.Join(p.path, "service-catalog"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "service-catalog-addons",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("service-catalog-addons"),
			ChartDir:   path.Join(p.path, "service-catalog-addons"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "nats-streaming",
			Namespace:  "natss",
			Overrides:  p.overridesProvider.OverridesFor("nats-streaming"),
			ChartDir:   path.Join(p.path, "nats-streaming"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "core",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("core"),
			ChartDir:   path.Join(p.path, "core"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "cluster-users",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("cluster-users"),
			ChartDir:   path.Join(p.path, "cluster-users"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "permission-controller",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("permission-controller"),
			ChartDir:   path.Join(p.path, "permission-controller"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "apiserver-proxy",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("apiserver-proxy"),
			ChartDir:   path.Join(p.path, "apiserver-proxy"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "iam-kubeconfig-service",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("iam-kubeconfig-service"),
			ChartDir:   path.Join(p.path, "iam-kubeconfig-service"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "serverless",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("serverless"),
			ChartDir:   path.Join(p.path, "serverless"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "knative-provisioner-natss",
			Namespace:  "knative-eventing",
			Overrides:  p.overridesProvider.OverridesFor("knative-provisioner-natss"),
			ChartDir:   path.Join(p.path, "knative-provisioner-natss"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "event-sources",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("event-sources"),
			ChartDir:   path.Join(p.path, "event-sources"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "application-connector",
			Namespace:  "kyma-integration",
			Overrides:  p.overridesProvider.OverridesFor("application-connector"),
			ChartDir:   path.Join(p.path, "application-connector"),
			HelmClient: helmClient,
		},
		Component{
			Name:       "console",
			Namespace:  "kyma-system",
			Overrides:  p.overridesProvider.OverridesFor("console"),
			ChartDir:   path.Join(p.path, "console"),
			HelmClient: helmClient,
		},
	}, nil
}
