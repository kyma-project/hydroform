module github.com/kyma-incubator/hydroform

go 1.13

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

require (
	cloud.google.com/go v0.46.3
	cloud.google.com/go/bigtable v1.0.0 // indirect
	cloud.google.com/go/storage v1.1.0 // indirect
	github.com/Azure/azure-sdk-for-go v34.1.0+incompatible // indirect
	github.com/chzyer/logex v1.1.11-0.20160617073814-96a4d311aa9b // indirect
	github.com/dustinkirkland/golang-petname v0.0.0-20190613200456-11339a705ed2 // indirect
	github.com/gardener/gardener v0.0.0-20190906111529-f9ad04069615
	github.com/gopherjs/gopherjs v0.0.0-20181103185306-d547d1d9531e // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-sockaddr v1.0.0 // indirect
	github.com/hashicorp/serf v0.8.2-0.20171022020050-c20a0b1b1ea9 // indirect
	github.com/hashicorp/terraform v0.12.8
	github.com/kyma-incubator/terraform-provider-gardener v0.0.0-20191009110559-03aaeb836c35
	github.com/miekg/dns v1.0.14 // indirect
	github.com/pkg/errors v0.8.1
	github.com/smartystreets/assertions v0.0.0-20190116191733-b6c0e53d7304 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/terraform-providers/terraform-provider-azurerm v1.35.0
	github.com/terraform-providers/terraform-provider-google v1.20.1-0.20190430222256-f9a9636be7cd
	github.com/terraform-providers/terraform-provider-null v1.0.0
	github.com/ugorji/go/codec v1.1.7 // indirect
	google.golang.org/api v0.11.0
	gopkg.in/ini.v1 v1.44.0 // indirect
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
