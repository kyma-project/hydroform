package deployment

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// interceptor which is replacing a value
type replaceOverrideInterceptor struct {
}

func (roi *replaceOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *replaceOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return "intercepted", nil
}

func (roi *replaceOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

//stringerOverrideInterceptor hides the value of an override when the value is converted to a string
type stringerOverrideInterceptor struct {
}

func (i *stringerOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("string-%v", value)
}

func (i *stringerOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (i *stringerOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

// interceptor which is failing
type failingOverrideInterceptor struct {
}

func (roi *failingOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *failingOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return nil, fmt.Errorf("Interceptor failed")
}

func (roi *failingOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

// interceptor which is returning a value for an undefined key
type undefinedOverrideInterceptor struct {
}

func (roi *undefinedOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *undefinedOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (roi *undefinedOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return fmt.Errorf("This value was missing")
}

func Test_InterceptValue(t *testing.T) {
	t.Run("Test interceptor without failures", func(t *testing.T) {
		builder := OverridesBuilder{}
		err := builder.AddFile("../test/data/deployment-overrides-intercepted.yaml")
		require.NoError(t, err)
		builder.AddInterceptor([]string{"chart.key2.key2-1", "chart.key4"}, &replaceOverrideInterceptor{})

		// read expected result
		data, err := ioutil.ReadFile("../test/data/deployment-overrides-intercepted-result.yaml")
		require.NoError(t, err)
		var expected map[string]interface{}
		err = yaml.Unmarshal(data, &expected)
		require.NoError(t, err)

		// verify merge result with expected data
		overrides, err := builder.Build()
		require.NoError(t, err)
		require.Equal(t, expected, overrides.Map())
	})

	t.Run("Test interceptor with failure", func(t *testing.T) {
		builder := OverridesBuilder{}
		err := builder.AddFile("../test/data/deployment-overrides-intercepted.yaml")
		require.NoError(t, err)
		builder.AddInterceptor([]string{"chart.key1"}, &failingOverrideInterceptor{})
		overrides, err := builder.Build()
		require.Empty(t, overrides.Map())
		require.Error(t, err)
	})
}

func Test_InterceptStringer(t *testing.T) {
	builder := OverridesBuilder{}
	err := builder.AddFile("../test/data/deployment-overrides-intercepted.yaml")
	require.NoError(t, err)
	builder.AddInterceptor([]string{"chart.key1", "chart.key3"}, &stringerOverrideInterceptor{})
	overrides, err := builder.Build()
	require.NoError(t, err)
	require.Equal(t,
		"map[chart:map[key1:string-value1yaml key2:map[key2-1:value2.1yaml key2-2:value2.2yaml] key3:string-value3yaml key4:value4yaml]]",
		fmt.Sprint(overrides))
}

func Test_InterceptUndefined(t *testing.T) {
	builder := OverridesBuilder{}
	err := builder.AddFile("../test/data/deployment-overrides-intercepted.yaml")
	require.NoError(t, err)
	builder.AddInterceptor([]string{"I.dont.exist"}, &undefinedOverrideInterceptor{})
	overrides, err := builder.Build()
	require.Empty(t, overrides.Map())
	require.Error(t, err)
	require.Equal(t, "This value was missing", err.Error())
}

func Test_FallbackInterceptor(t *testing.T) {
	builder := OverridesBuilder{}
	err := builder.AddFile("../test/data/deployment-overrides-intercepted.yaml")
	require.NoError(t, err)

	t.Run("Test FallbackInterceptor happy path", func(t *testing.T) {
		builder.AddInterceptor([]string{"I.dont.exist"}, NewFallbackOverrideInterceptor("I am the fallback"))
		overrides, err := builder.Build()
		require.NotEmpty(t, overrides)
		require.NoError(t, err)
		require.Equal(t, "I am the fallback", overrides.Map()["I"].(map[string]interface{})["dont"].(map[string]interface{})["exist"])
	})

	t.Run("Test FallbackInterceptor with sub-key which is not a map", func(t *testing.T) {
		builder.AddInterceptor([]string{"chart.key3.xyz"}, NewFallbackOverrideInterceptor("Use me as fallback"))
		overrides, err := builder.Build()
		require.Empty(t, overrides.Map())
		require.Error(t, err)
	})
}

func Test_GlobalOverridesInterceptionForLocalCluster(t *testing.T) {
	ob := OverridesBuilder{}
	kubeClient := fake.NewSimpleClientset()
	log := logger.NewLogger(true)

	newDomainNameOverrideInterceptor := NewDomainNameOverrideInterceptor(kubeClient, log)
	newDomainNameOverrideInterceptor.isLocalCluster = isLocalClusterFunc(true)

	newCertificateOverrideInterceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
	newCertificateOverrideInterceptor.isLocalCluster = isLocalClusterFunc(true)

	ob.AddInterceptor([]string{"global.isLocalEnv", "global.environment.gardener"}, NewFallbackOverrideInterceptor(false))
	ob.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, newDomainNameOverrideInterceptor)
	ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, newCertificateOverrideInterceptor)

	// read expected result
	data, err := ioutil.ReadFile("../test/data/deployment-global-overrides.yaml")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)

	// verify global overrides
	overrides, err := ob.Build()
	require.NotEmpty(t, overrides)
	require.NoError(t, err)
	require.Equal(t, expected, overrides.Map())
}

func Test_GlobalOverridesInterceptionForNonGardenerCluster(t *testing.T) {
	ob := OverridesBuilder{}
	kubeClient := fake.NewSimpleClientset()
	log := logger.NewLogger(true)

	newDomainNameOverrideInterceptor := NewDomainNameOverrideInterceptor(kubeClient, log)
	newDomainNameOverrideInterceptor.isLocalCluster = isLocalClusterFunc(false)

	newCertificateOverrideInterceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
	newCertificateOverrideInterceptor.isLocalCluster = isLocalClusterFunc(false)

	ob.AddInterceptor([]string{"global.isLocalEnv", "global.environment.gardener"}, NewFallbackOverrideInterceptor(false))
	ob.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, newDomainNameOverrideInterceptor)
	ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, newCertificateOverrideInterceptor)

	// read expected result
	data, err := ioutil.ReadFile("../test/data/deployment-global-overrides-for-remote-cluster.yaml")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)

	// verify global overrides
	overrides, err := ob.Build()
	require.NotEmpty(t, overrides)
	require.NoError(t, err)
	require.Equal(t, expected, overrides.Map())
}

func Test_DomainNameOverrideInterceptor(t *testing.T) {
	ob := OverridesBuilder{}
	log := logger.NewLogger(true)

	domainData := make(map[string]string)
	domainData["domain"] = "gardener.domain"

	gardenerCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shoot-info",
			Namespace: "kube-system",
		},
		Data: domainData,
	}

	mockNewDomainNameOverrideInterceptor := func(kubeClient kubernetes.Interface, log logger.Interface, isLocal bool) *DomainNameOverrideInterceptor {
		newDomainNameOverrideInterceptor := NewDomainNameOverrideInterceptor(kubeClient, log)
		newDomainNameOverrideInterceptor.isLocalCluster = isLocalClusterFunc(isLocal)
		return newDomainNameOverrideInterceptor
	}

	t.Run("test default domain for local cluster", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset()
		ob.AddInterceptor([]string{"global.domainName"}, mockNewDomainNameOverrideInterceptor(kubeClient, log, true))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Contains(t, overrides.String(), localKymaDevDomain)
	})

	t.Run("test default domain for remote non-gardener cluster", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset()
		ob.AddInterceptor([]string{"global.domainName"}, mockNewDomainNameOverrideInterceptor(kubeClient, log, false))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Contains(t, overrides.String(), defaultRemoteKymaDomain)
	})

	t.Run("test valid domain for a gardener cluster", func(t *testing.T) {
		//given
		kubeClient := fake.NewSimpleClientset(gardenerCM)
		ob.AddInterceptor([]string{"global.domainName"}, NewDomainNameOverrideInterceptor(kubeClient, log))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Contains(t, overrides.String(), "gardener.domain")
	})

	t.Run("test user-provided domain is overriden on gardener cluster", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset(gardenerCM)

		ob := OverridesBuilder{}
		domainNameOverrides := make(map[string]interface{})
		domainNameOverrides["domainName"] = "user.domain"
		err := ob.AddOverrides("global", domainNameOverrides)
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.domainName"}, NewDomainNameOverrideInterceptor(kubeClient, log))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Contains(t, overrides.String(), "gardener.domain")
		require.NotContains(t, overrides.String(), "user.domain")
	})

	t.Run("test user-provided domain is not overriden on local cluster", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset()

		ob := OverridesBuilder{}
		domainNameOverrides := make(map[string]interface{})
		domainNameOverrides["domainName"] = "user.domain"
		err := ob.AddOverrides("global", domainNameOverrides)
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.domainName"}, mockNewDomainNameOverrideInterceptor(kubeClient, log, true))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Contains(t, overrides.String(), "user.domain")
	})

	t.Run("test user-provided domain is not overriden on remote cluster", func(t *testing.T) {
		// given
		kubeClient := fake.NewSimpleClientset()

		ob := OverridesBuilder{}
		domainNameOverrides := make(map[string]interface{})
		domainNameOverrides["domainName"] = "user.domain"
		err := ob.AddOverrides("global", domainNameOverrides)
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.domainName"}, mockNewDomainNameOverrideInterceptor(kubeClient, log, false))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Contains(t, overrides.String(), "user.domain")
	})
}

func Test_CertificateOverridesInterception(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	newCertificateOverrideInterceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
	newCertificateOverrideInterceptor.isLocalCluster = isLocalClusterFunc(true)

	t.Run("CertificateInterceptor using fallbacks", func(t *testing.T) {
		ob := OverridesBuilder{}

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, newCertificateOverrideInterceptor)
		// verify cert overrides
		overrides, err := ob.Build()
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
	})

	t.Run("CertificateInterceptor using existing certs", func(t *testing.T) {
		ob := OverridesBuilder{}

		tlsOverrides := make(map[string]interface{})
		tlsOverrides["tlsCrt"] = defaultLocalTLSCrtEnc
		tlsOverrides["tlsKey"] = defaultLocalTLSKeyEnc
		err := ob.AddOverrides("global", tlsOverrides)
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, newCertificateOverrideInterceptor)
		// verify cert overrides
		overrides, err := ob.Build()
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
	})

	t.Run("CertificateInterceptor using invalid certs", func(t *testing.T) {
		ob := OverridesBuilder{}

		tlsOverrides := make(map[string]interface{})
		tlsOverrides["tlsCrt"] = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZSRENDQXl3Q0NRQ2pOdWF5a2xVZGRqQU5CZ2txaGtpRzl3MEJBUXNGQURCa01Rc3dDUVlEVlFRR0V3SkUKUlRFUU1BNEdBMVVFQ0F3SFFtRjJZWEpwWVRFUE1BMEdBMVVFQnd3R1RYVnVhV05vTVE4d0RRWURWUVFLREFaVApRVkFnVTBVeERUQUxCZ05WQkFzTUJFdDViV0V4RWpBUUJnTlZCQU1NQ1hSbGMzUXVZMkZ6WlRBZUZ3MHlNVEF5Ck1UZ3hNVEl3TkRaYUZ3MHlNakF5TVRneE1USXdORFphTUdReEN6QUpCZ05WQkFZVEFrUkZNUkF3RGdZRFZRUUkKREFkQ1lYWmhjbWxoTVE4d0RRWURWUVFIREFaTmRXNXBZMmd4RHpBTkJnTlZCQW9NQmxOQlVDQlRSVEVOTUFzRwpBMVVFQ3d3RVMzbHRZVEVTTUJBR0ExVUVBd3dKZEdWemRDNWpZWE5sTUlJQ0lqQU5CZ2txaGtpRzl3MEJBUUVGCkFBT0NBZzhBTUlJQ0NnS0NBZ0VBMFFBM1BPOFlWY2NFbVVvYkppQzZQZjN0eHBNWFJlRmNObUZiVDgvY1ArcDcKT2hIVVZzMUE4YWxRS2VXVy8yMTU2bm83clpsMUtVUXlBYVVNL054cTdhNWJaRUF1WmdtcjhWSVJNUDlnME14aQoxb3NGcXJiaE05cVMrL29adjFURlg1M2pZVHFvZkxselVnbWZPeHFyby9Wb1RWZS9mMTR2TG5EQkF5UG0vRXdOCkZxOWtqblhnaERxNnJpSTJ4T1c2YVpaR3lGVHN3ZHpzbm5CK3B4L3dqc21nTGlTSVRXbDA3ekRTa0RaRjlIY3kKbFhGWGIxeGJNZUpySWtYTCtqRVE2T3hTbWw0QjZMeGszc3o3L0JFb0JVaG1zMDR2T2paNjgzQi9zd1FZeEdzRgpGcU4vcnRXTHBGQmUxckdDM2NKYXFuVUIwTVRBK0dGL2dTcGdPV0g5Q3JScnc0RXhQMFNTWkwvUWZRaGc2Nmw1CmZxNUdGNzVEYWx2ckNlOVpWYzN5TDFJWUJHY2cxMUlPQk9ZaGFYUElEWEx1K3pFbERCaTlLZzdnYTkxYmJoQ3cKUXpXZE5wZzVJby9wRnEvT2pPMitxU1pDdVdITFZrNCtVeTVySS9IVWtmWmFWSXplVDkwTzRMT2VkZEFaamd0WQoveFppMWxXQVcrZjZhQWZLRUdpWXg3MlE5NkJ1cUtGc0Y0MlpoWmp5czY3OWRrbE5pMy9Ta2dlR1Qzb2lHZUNPCjFHZmx3R2tBWUxkZ3hxZTBMOURXRUxWcy8vTjFkMUF3VFZ0RUtncHd0cDJzSkg5b1laZEZ3eVJiemczRW44NmwKc05DNlNLTENHcDdYc0ZMZ3VHcDdRNFhmVWZsT0ZwSVVycFZhQ00xUThEbTBlTnZyaTAzWVlCUzJuSkVNY0JNQwpBd0VBQVRBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQWdFQWZzN0dQRFVqN1BXTE4rTkVYY0NvbExwbFoxTjE0emZJCnJhWTJ0c1VQcTNGeGxjMUpsa0R3QUlLcGxoTVVIY0Iya2Q1YTVHOUlNSFpyZ29nVWVWTjlLUklIL1pTMDAydloKRktPeDd5M1owYWZ0Q2Z1aEZKTk1pV09DV2UxVFBuUUJod082eElOWjZWZktoa3dZRGNtRXIxQnJTdi85c2RJUwovT0czbU91Mi9VcnNLdkZmN1d0NFVQUjhONnphUjFDUFIxUytOWFhKeXNjZ2RoNC80UVRwZm1hRUFKWnRxQ1NNCmpUUk5DVlVTZnZGK21Kem4yVnJ3YjFKSUkwWVhQVi9VSng1WTdLeFFFV2JGdkp2b1ZYaG15RVZ5dllxNWVPTWQKbDc1VHZhbTJ0ek0vQnIvMUpkNkdxNFhhU1pZL08wbmg2MlVVZlVJMXdPNG5OVlBwTEs4d1Z4SElnWng2ZUIyYwpncW83NDJZQ3JyVXZ5Y080VTlJaWNQTEcyVmduNzlnTVRJZDdRL0o5WjFFakFvbmIwL0tuTFFaZVlReks2T09MCndyQlVBaEtrbnI4MXU0R3BabGU2eVVPd0Q0ZDRhTGJQM08zTG5LUVF3Y1M0andCVDFGRlpyeEFoUEVBRGZveXEKemNKeS9SU2t3WU9NaWpoZ3RXR3cxdU5FSnFXekw5MExKOXVLRWJrODN6c2h3MFFHQ0ROb3hmNStMVmtPWVBURwppaTdxdE8xYUczSFJnQWdRKytzOXdreGdaRjNYeGxISUlaRmEvRHZuaUJxaGxEOXJLQzl0eFVUN01SU3dvcWN6CldJZEdqeW9RZ2hKbTJSS3F5REQxQjE4SEJ2MXdCbStMdzRSbHJDOVREeWM5OTFqMEgxZEViWDRpMCtRZUZnYXQKWThxd2hqeStjME09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
		tlsOverrides["tlsKey"] = defaultLocalTLSKeyEnc
		err := ob.AddOverrides("global", tlsOverrides)
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, newCertificateOverrideInterceptor)
		// verify cert overrides
		overrides, err := ob.Build()
		require.Error(t, err)
		require.Empty(t, overrides.Map())
	})

	t.Run("CertificateInterceptor using existing certs for external domain", func(t *testing.T) {
		ob := OverridesBuilder{}
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(false)

		tlsOverrides := make(map[string]interface{})
		tlsOverrides["tlsCrt"] = defaultRemoteTLSCrtEnc
		tlsOverrides["tlsKey"] = defaultRemoteTLSKeyEnc
		err := ob.AddOverrides("global", tlsOverrides)
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)
		// verify cert overrides
		overrides, err := ob.Build()
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
	})
}

func isLocalClusterFunc(val bool) func() (bool, error) {
	return func() (bool, error) {
		return val, nil
	}
}
