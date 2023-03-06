module github.com/kyma-project/hydroform/function

go 1.18

replace golang.org/x/net v0.2.0 => golang.org/x/net v0.7.0

require (
	github.com/docker/cli v20.10.14+incompatible
	github.com/docker/docker v20.10.21+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/golang/mock v1.6.0
	github.com/invopop/jsonschema v0.4.0
	github.com/moby/moby v20.10.21+incompatible
	github.com/onsi/gomega v1.20.1
	github.com/opencontainers/image-spec v1.0.3-0.20220303224323-02efb9a75ee1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.25.1
	k8s.io/apimachinery v0.25.1
	k8s.io/client-go v0.24.1
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gotest.tools/v3 v3.1.0 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.1-0.20220625051128-279f4069feb5
	github.com/emicklei/go-restful/v3 => github.com/emicklei/go-restful/v3 v3.8.0
)
