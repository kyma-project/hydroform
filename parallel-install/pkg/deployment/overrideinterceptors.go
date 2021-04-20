package deployment

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/pkg/errors"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	localKymaDevDomain    = "local.kyma.dev"
	defaultTLSCrtEnc      = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURQVENDQWlXZ0F3SUJBZ0lSQVByWW0wbGhVdUdkeVNCTHo4d3g5VGd3RFFZSktvWklodmNOQVFFTEJRQXcKTURFVk1CTUdBMVVFQ2hNTVkyVnlkQzF0WVc1aFoyVnlNUmN3RlFZRFZRUURFdzVzYjJOaGJDMXJlVzFoTFdSbApkakFlRncweU1EQTNNamt3T1RJek5UTmFGdzB6TURBM01qY3dPVEl6TlROYU1EQXhGVEFUQmdOVkJBb1RER05sCmNuUXRiV0Z1WVdkbGNqRVhNQlVHQTFVRUF4TU9iRzlqWVd3dGEzbHRZUzFrWlhZd2dnRWlNQTBHQ1NxR1NJYjMKRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFDemE4VEV5UjIyTFRKN3A2aXg0M2E3WTVVblovRkNicGNOQkdEbQpxaDRiRGZLcjFvMm1CYldWdUhDbTVBdTBkeHZnbUdyd0tvZzJMY0N1bEd5UXVlK1JLQ0RIVFBJVjdqZEJwZHJhCkNZMXQrNjlJMkJWV0xiblFNVEZmOWw3Vy8yZFFFU0ExZHZQajhMZmlrcEQvUEQ5ekdHR0FQa2hlenVNRU80dUwKaUxXSloyYmpYK1dtaGZXb0lrOG5oak5YNVBFN2l4alMvNnB3QU56eXk2NW95NDJPaHNuYXlDR1grbmhFVk5SRApUejEraEMvdjJaOS9lRG1OdHdjT1hJSk4relZtUTJ4VHh2Sm0rbDUwYzlnenZTY3YzQXg0dUJsOTk3UnVlcUszCmdZMVRmVklFQ0FOTE9hb29jRG5kcW1FY1lBb25SeGJKK0M2U1RJYlhuUVAyMmYxQkFnTUJBQUdqVWpCUU1BNEcKQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WSFNVRUREQUtCZ2dyQmdFRkJRY0RBVEFNQmdOVkhSTUJBZjhFQWpBQQpNQnNHQTFVZEVRUVVNQktDRUNvdWJHOWpZV3d1YTNsdFlTNWtaWFl3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCCkFBUnVOd0VadW1PK2h0dDBZSWpMN2VmelA3UjllK2U4NzJNVGJjSGtyQVhmT2hvQWF0bkw5cGhaTHhKbVNpa1IKY0tJYkJneDM3RG5ka2dPY3doNURTT2NrdHBsdk9sL2NwMHMwVmFWbjJ6UEk4Szk4L0R0bEU5bVAyMHRLbE90RwpaYWRhdkdrejhXbDFoRzhaNXdteXNJNWlEZHNpajVMUVJ6Rk04YmRGUUJiRGkxbzRvZWhIRTNXbjJjU3NTUFlDCkUxZTdsM00ySTdwQ3daT2lFMDY1THZEeEszWFExVFRMR2oxcy9hYzRNZUxCaXlEN29qb25MQmJNYXRiaVJCOUIKYlBlQS9OUlBaSHR4TDArQ2Nvb1JndmpBNEJMNEtYaFhxZHZzTFpiQWlZc0xTWk0yRHU0ZWZ1Q25SVUh1bW1xNQpVNnNOOUg4WXZxaWI4K3B1c0VpTUttND0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	defaultTLSKeyEnc      = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBczJ2RXhNa2R0aTB5ZTZlb3NlTjJ1Mk9WSjJmeFFtNlhEUVJnNXFvZUd3M3lxOWFOCnBnVzFsYmh3cHVRTHRIY2I0SmhxOENxSU5pM0FycFJza0xudmtTZ2d4MHp5RmU0M1FhWGEyZ21OYmZ1dlNOZ1YKVmkyNTBERXhYL1plMXY5blVCRWdOWGJ6NC9DMzRwS1EvencvY3hoaGdENUlYczdqQkR1TGk0aTFpV2RtNDEvbApwb1gxcUNKUEo0WXpWK1R4TzRzWTB2K3FjQURjOHN1dWFNdU5qb2JKMnNnaGwvcDRSRlRVUTA4OWZvUXY3OW1mCmYzZzVqYmNIRGx5Q1RmczFaa05zVThieVp2cGVkSFBZTTcwbkw5d01lTGdaZmZlMGJucWl0NEdOVTMxU0JBZ0QKU3ptcUtIQTUzYXBoSEdBS0owY1d5Zmd1a2t5RzE1MEQ5dG45UVFJREFRQUJBb0lCQUJwVmYvenVFOWxRU3UrUgpUUlpHNzM5VGYybllQTFhtYTI4eXJGSk90N3A2MHBwY0ZGQkEyRVVRWENCeXFqRWpwa2pSdGlobjViUW1CUGphCnVoQ0g2ZHloU2laV2FkWEVNQUlIcU5hRnZtZGRJSDROa1J3aisvak5yNVNKSWFSbXVqQXJRMUgxa3BockZXSkEKNXQwL1o0U3FHRzF0TnN3TGk1QnNlTy9TOGVvbnJ0Q3gzSmxuNXJYdUIzT1hSQnMyVGV6dDNRRlBEMEJDY2c3cgpBbEQrSDN6UjE0ZnBLaFVvb0J4S0VacmFHdmpwVURFeThSSy9FemxaVzBxMDB1b2NhMWR0c0s1V1YxblB2aHZmCjBONGRYaUxuOE5xY1k0c0RTMzdhMWhYV3VJWWpvRndZa0traFc0YS9LeWRKRm5acmlJaDB0ZU81Q0I1ZnpaVnQKWklOYndyMENnWUVBd0gzeksvRTdmeTVpd0tJQnA1M0YrUk9GYmo1a1Y3VUlkY0RIVjFveHhIM2psQzNZUzl0MQo3Wk9UUHJ6eGZ4VlB5TVhnOEQ1clJybkFVQk43cE5xeWxHc3FMOFA1dnZlbVNwOGNKU0REQWN4RFlqeEJLams5CldtOXZnTGpnaERSUFN1Um50QXNxQVVqcWhzNmhHUzQ4WUhMOVI2QlI5dmY2U2xWLzN1NWMvTXNDZ1lFQTdwM1UKRDBxcGNlM1liaiszZmppVWFjcTRGcG9rQmp1MTFVTGNvREUydmZFZUtEQldsM3BJaFNGaHYvbnVqaUg2WWJpWApuYmxKNVRlSnI5RzBGWEtwcHNLWW9vVHFkVDlKcFp2QWZGUzc2blZZaUJvMHR3VzhwMGVCS3QyaUFyejRYRmxUCnpRSnNOS1dsRzBzdGJmSzNqdUNzaWJjYTBUd09lbTdSdjdHV0dLTUNnWUJjZmFoVVd1c2RweW9vS1MvbVhEYisKQVZWQnJaVUZWNlVpLzJoSkhydC9FSVpEY3V2Vk56UW8zWm9Jc1R6UXRXcktxOW56VmVxeDV4cnkzd213SXExZwpCMFlVQVhTRlAvV1ZNWEtTbkhWVzdkRUs2S3pmSHZYTitIRjVSbHdLNmgrWGVyd2hsS093VGxyeVAyTEUrS1JtCks1cHJ5aXJZSWpzUGNKbXFncG9IbFFLQmdCVWVFcTVueFNjNERYZDBYQ0Rua1BycjNlN2lKVjRIMnNmTTZ3bWkKVVYzdUFPVTlvZXcxL2tVSjkwU3VNZGFTV3o1YXY5Qk5uYVNUamJQcHN5NVN2NERxcCtkNksrWEVmQmdUK0swSQpNcmxGT1ZpU09TZ1pjZUM4QzBwbjR2YXJFcS9abC9rRXhkN0M2aUhJUFhVRmpna3ZDUllIQm5DT0NCbjl4TUphClRSWlJBb0dBWS9QYSswMFo1MHYrUU40cVhlRHFrS2FLZU80OFUzOHUvQUJMeGFMNHkrZkJpUStuaXh5ZFUzOCsKYndBR3JtMzUvSU5VRTlkWE44d21XRUlXVUZ3YVR2dHY5NXBpcWNKL25QZkFiY2pDeU8wU3BJWCtUYnFRSkljbgpTVjlrKzhWUFNiRUJ5YXRKVTdIQ3FaNUNTWEZuUnRNanliaWNYYUFKSWtBQm4zVjJ3OFk9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
	externalKymaDevDomain = "k8s.example.com"
	externalTLSCrtEnc     = "xxx"
	externalTLSKeyEnc     = "xxx"
)

// OverrideInterceptor is controlling access to override values
type OverrideInterceptor interface {
	//String shows the value of the override
	String(value interface{}, key string) string
	//Intercept is executed when the override is retrieved
	Intercept(value interface{}, key string) (interface{}, error)
	//Undefined is executed when the override is not defined
	Undefined(overrides map[string]interface{}, key string) error
}

func NewDomainNameOverrideInterceptor(kubeClient kubernetes.Interface, log logger.Interface) *DomainNameOverrideInterceptor {
	retryOptions := []retry.Option{
		retry.Delay(2 * time.Second),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	}

	return &DomainNameOverrideInterceptor{
		kubeClient:   kubeClient,
		retryOptions: retryOptions,
		log:          log,
		findClusterHost: func() string {
			return kubeClient.Discovery().RESTClient().Get().URL().Host
		},
	}
}

//DomainNameOverrideInterceptor resolves the domain name for the cluster
type DomainNameOverrideInterceptor struct {
	kubeClient      kubernetes.Interface
	retryOptions    []retry.Option
	log             logger.Interface
	findClusterHost func() string
}

func (i *DomainNameOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (i *DomainNameOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	//on gardener domain provided by user should be ignored
	domainName, err := i.getGardenerDomain()
	if err != nil {
		return nil, err
	}

	if domainName != "" {
		return domainName, nil
	}

	return value, nil
}

func (i *DomainNameOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	// gardener? y -> getgardenerdomain ,
	// isk3d? y -> getlocaldomain ,
	// externalkymadomain
	domain, err := i.getDomainName()
	if err != nil {
		return err
	}

	return NewFallbackOverrideInterceptor(domain).Undefined(overrides, key)
}

func (i *DomainNameOverrideInterceptor) getDomainName() (string, error) {
	var domainName string
	var err error

	domainName = os.Getenv("CUSTOMDOMAIN")
	if domainName != "" {
		return domainName, nil
	}

	domainName, err = i.getGardenerDomain()
	if err != nil {
		return "", err
	}
	if domainName != "" {
		return domainName, nil
	}

	domainName, err = i.getLocalDomain()
	if err != nil {
		return "", err
	}
	if domainName != "" {
		return domainName, nil
	}

	return externalKymaDevDomain, nil
}

func (i *DomainNameOverrideInterceptor) getGardenerDomain() (domainName string, err error) {
	err = retry.Do(func() error {
		configMap, err := i.kubeClient.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "shoot-info", metav1.GetOptions{})

		if err != nil {
			if apierr.IsNotFound(err) {
				return nil
			}
			return err
		}

		domainName = configMap.Data["domain"]
		if domainName == "" {
			return fmt.Errorf("domain is empty in %s configmap", "shoot-info")
		}

		return nil
	}, i.retryOptions...)

	if err != nil {
		return "", err
	}

	return domainName, nil
}

func (i *DomainNameOverrideInterceptor) getLocalDomain() (domainName string, err error) {
	err = retry.Do(func() error {
		clusterHost := i.findClusterHost()

		isLocalCluster := strings.Contains(clusterHost, localKymaDevDomain)
		if isLocalCluster {
			domainName = localKymaDevDomain
			return nil
		}

		return nil
	}, i.retryOptions...)

	if err != nil {
		return "", err
	}

	return domainName, nil
}

//CertificateOverrideInterceptor handles certificates
type CertificateOverrideInterceptor struct {
	tlsCrtOverrideKey string
	tlsKeyOverrideKey string
	tlsCrtEnc         string
	tlsKeyEnc         string
}

func (i *CertificateOverrideInterceptor) String(value interface{}, key string) string {
	return "<masked>"
}

func (i *CertificateOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	switch key {
	case i.tlsCrtOverrideKey:
		i.tlsCrtEnc = value.(string)
	case i.tlsKeyOverrideKey:
		i.tlsKeyEnc = value.(string)
	}
	if err := i.validate(); err != nil {
		return nil, err
	}
	return value, nil
}

func (i *CertificateOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	var fbInterc *FallbackOverrideInterceptor
	switch key {
	case i.tlsCrtOverrideKey:
		fbInterc = NewFallbackOverrideInterceptor(defaultTLSCrtEnc)
		i.tlsCrtEnc = defaultTLSCrtEnc
	case i.tlsKeyOverrideKey:
		fbInterc = NewFallbackOverrideInterceptor(defaultTLSKeyEnc)
		i.tlsKeyEnc = defaultTLSKeyEnc
	default:
		return fmt.Errorf("certificate interceptor can not handle overrides-key '%s'", key)
	}
	if err := fbInterc.Undefined(overrides, key); err != nil {
		return err
	}
	return i.validate()
}

func (i *CertificateOverrideInterceptor) validate() error {
	if i.tlsCrtEnc != "" && i.tlsKeyEnc != "" {
		//decode tls crt and key
		crt, err := base64.StdEncoding.DecodeString(i.tlsCrtEnc)
		if err != nil {
			return err
		}
		key, err := base64.StdEncoding.DecodeString(i.tlsKeyEnc)
		if err != nil {
			return err
		}
		//ensure that crt and key are fitting together
		_, err = tls.X509KeyPair(crt, key)
		if err != nil {
			return errors.Wrap(err,
				fmt.Sprintf("Provided TLS certificate (passed in keys '%s' and '%s') is invalid", i.tlsCrtOverrideKey, i.tlsKeyOverrideKey))
		}
	}
	return nil
}

func NewCertificateOverrideInterceptor(tlsCrtOverrideKey, tlsKeyOverrideKey string) *CertificateOverrideInterceptor {
	return &CertificateOverrideInterceptor{
		tlsCrtOverrideKey: tlsCrtOverrideKey,
		tlsKeyOverrideKey: tlsKeyOverrideKey,
	}
}

//FallbackOverrideInterceptor sets a default value for an undefined overwrite
type FallbackOverrideInterceptor struct {
	fallback interface{}
}

func (i *FallbackOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (i *FallbackOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (i *FallbackOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	subKeys := strings.Split(key, ".")
	maxDepth := len(subKeys)
	lastProcessedEntry := overrides

	for depth, subKey := range subKeys {
		if _, ok := lastProcessedEntry[subKey]; !ok {
			//sub-element does not exist - add map
			lastProcessedEntry[subKey] = make(map[string]interface{})
		}
		if _, ok := lastProcessedEntry[subKey].(map[string]interface{}); !ok {
			//ensure existing sub-element is map otherwise fail
			return fmt.Errorf("override '%s' cannot be set with default value as sub-key '%s' is not a map", key, strings.Join(subKeys[:depth+1], "."))
		}

		if depth == (maxDepth - 1) {
			//we are in the last loop, set default value
			lastProcessedEntry[subKey] = i.fallback
		} else {
			//continue processing the next sub-entry
			lastProcessedEntry = lastProcessedEntry[subKey].(map[string]interface{})
		}
	}

	return nil
}

func (i *FallbackOverrideInterceptor) Fallback() interface{} {
	return i.fallback
}

func NewFallbackOverrideInterceptor(fallback interface{}) *FallbackOverrideInterceptor {
	return &FallbackOverrideInterceptor{
		fallback: fallback,
	}
}

// This struct is introduced to ensure backward compatibility of Kyma 2.0 with 1.x
// It can be removed when Kyma 2.0 is released
type InstallLegacyCRDsInterceptor struct{}

func (i *InstallLegacyCRDsInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (i *InstallLegacyCRDsInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	// We should never install CRDs in the legacy way with Kyma 2.0
	return false, nil
}

func (i *InstallLegacyCRDsInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	// We should never install CRDs in the legacy way with Kyma 2.0
	return NewFallbackOverrideInterceptor(false).Undefined(overrides, key)
}

func NewInstallLegacyCRDsInterceptor() *InstallLegacyCRDsInterceptor {
	return &InstallLegacyCRDsInterceptor{}
}
