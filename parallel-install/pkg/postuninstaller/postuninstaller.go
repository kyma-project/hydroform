package postuninstaller

import (
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic"
)

// Config defines configuration values for the PostUninstaller.
type Config struct {
	Log                      logger.Interface        //Logger to be used
	KubeconfigSource         config.KubeconfigSource //KubeconfigSource to be used
	InstallationResourcePath string
}

//go:generate mockery --name PostUninstaller

// PostUninstaller removes leftover resources from k8s cluster, added during Kyma installation.
type PostUninstaller struct {
	cfg             Config
	dynamicClient   dynamic.Interface
	resourceManager preinstaller.ResourceManager
	retryOptions    []retry.Option
}

// Output contains lists of Deleted and NotDeleted resources during PostUninstaller uninstallation.
type Output struct {
	// Installed files during PreInstaller installation.
	Deleted []string
	// NotInstalled files during PreInstaller installation.
	NotDeleted []string
}

// NewPostUninstaller creates a new instance of PostUninstaller.
func NewPostUninstaller(cfg Config, retryOptions []retry.Option) (*PostUninstaller, error) {
	restConfig, err := config.RestConfig(cfg.KubeconfigSource)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	resourceManager, err := preinstaller.NewDefaultResourceManager(cfg.KubeconfigSource, cfg.Log, retryOptions)
	if err != nil {
		return nil, err
	}

	return &PostUninstaller{
		cfg:             cfg,
		dynamicClient:   dynamicClient,
		resourceManager: resourceManager,
		retryOptions:    retryOptions,
	}, nil
}

// UninstallCRDs that belong to Kyma from a k8s cluster.
func (c *PostUninstaller) UninstallCRDs() (Output, error) {
	c.cfg.Log.Infof("Uninstalling CRDs labeled with: %s=%s", preinstaller.LABEL_KEY_ORIGIN, preinstaller.LABEL_VALUE_KYMA)

	selector, err := c.prepareKymaCrdLabelSelector()
	if err != nil {
		return Output{}, err
	}

	crdGvkV1Beta1 := c.crdGvkWith("v1beta1")
	crdGvkV1 := c.crdGvkWith("v1")
	gvks := [2]schema.GroupVersionKind{crdGvkV1Beta1, crdGvkV1}

	for _, gvk := range gvks {
		err = c.resourceManager.DeleteCollectionOfResources(gvk, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			c.cfg.Log.Error(err)
		}
	}

	return Output{}, nil
}

func (c *PostUninstaller) prepareKymaCrdLabelSelector() (selector labels.Selector, err error) {
	kymaCrdReq, err := labels.NewRequirement(preinstaller.LABEL_KEY_ORIGIN, selection.Equals, []string{preinstaller.LABEL_VALUE_KYMA})
	if err != nil {
		return nil, errors.Wrap(err, "Error occurred when preparing Kyma CRD label selector")
	}

	selector = labels.NewSelector()
	selector = selector.Add(*kymaCrdReq)
	return selector, nil
}

func (c *PostUninstaller) crdGvkWith(version string) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: version,
		Kind:    "customresourcedefinition",
	}
}
