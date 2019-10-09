module github.com/kyma-incubator/hydroform

go 1.13

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

require (
	cloud.google.com/go v0.46.3
	cloud.google.com/go/bigtable v1.0.0 // indirect
	cloud.google.com/go/storage v1.1.0 // indirect
	github.com/dustinkirkland/golang-petname v0.0.0-20190613200456-11339a705ed2 // indirect
	github.com/gardener/gardener v0.0.0-20190906111529-f9ad04069615
	github.com/hashicorp/terraform v0.11.14
	github.com/kyma-incubator/terraform-provider-gardener v0.0.0-20191009110559-03aaeb836c35
	github.com/pkg/errors v0.8.1
	github.com/stoewer/go-strcase v1.0.2 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/terraform-providers/terraform-provider-google v1.20.0
	github.com/terraform-providers/terraform-provider-null v1.0.0
	github.com/terraform-providers/terraform-provider-random v2.0.0+incompatible // indirect
	google.golang.org/api v0.11.0
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
