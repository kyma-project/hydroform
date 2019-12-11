module github.com/kyma-incubator/hydroform/provision

go 1.13

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

require (
	github.com/hashicorp/terraform v0.12.13
	github.com/hashicorp/terraform-svchost v0.0.0-20191011084731-65d371908596
	github.com/kyma-incubator/hydroform/provision v0.0.0-20191209064640-97160354ef3e //kavya
	github.com/mitchellh/cli v1.0.0
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	k8s.io/apimachinery v0.0.0-20191004115701-31ade1b30762 // tag kubernetes-1.15.6
)
