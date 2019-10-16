module github.com/kyma-incubator/hydroform

go 1.13

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

replace github.com/hashicorp/terraform v0.12.8 => github.com/hashicorp/terraform v0.11.14

require (
	cloud.google.com/go v0.46.3
	cloud.google.com/go/bigtable v1.0.0 // indirect
	cloud.google.com/go/storage v1.1.0 // indirect
	git.apache.org/thrift.git v0.12.0 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-dump v0.0.0-20190214190832-042adf3cf4a0 // indirect
	github.com/armon/circbuf v0.0.0-20190214190532-5111143e8da2 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/dustinkirkland/golang-petname v0.0.0-20190613200456-11339a705ed2 // indirect
	github.com/dylanmei/winrmtest v0.0.0-20190225150635-99b7fe2fddf1 // indirect
	github.com/gardener/gardener v0.0.0-20190906111529-f9ad04069615
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.4 // indirect
	github.com/hashicorp/go-rootcerts v1.0.0 // indirect
	github.com/hashicorp/go-tfe v0.3.16 // indirect
	github.com/hashicorp/memberlist v0.1.0 // indirect
	github.com/hashicorp/terraform v0.12.8
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kyma-incubator/terraform-provider-gardener v0.0.0-20191009110559-03aaeb836c35
	github.com/masterzen/winrm v0.0.0-20190223112901-5e5c9a7fe54b // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-shellwords v1.0.4 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/go-linereader v0.0.0-20190213213312-1b945b3263eb // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/panicwrap v0.0.0-20190213213626-17011010aaa4 // indirect
	github.com/mitchellh/prefixedio v0.0.0-20190213213902-5733675afd51 // indirect
	github.com/openzipkin/zipkin-go v0.1.6 // indirect
	github.com/pkg/errors v0.8.1
	github.com/posener/complete v1.2.1 // indirect
	github.com/satori/uuid v1.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/terraform-providers/terraform-provider-azurerm v1.32.0
	github.com/terraform-providers/terraform-provider-google v1.20.1-0.20190430222256-f9a9636be7cd
	github.com/terraform-providers/terraform-provider-null v1.0.0
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/vmihailenco/msgpack v4.0.1+incompatible // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	google.golang.org/api v0.11.0
	gopkg.in/ini.v1 v1.44.0 // indirect
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
