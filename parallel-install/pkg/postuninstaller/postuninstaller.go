package postuninstaller

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/config"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/logger"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/preinstaller"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	c.cfg.Log.Infof("Uninstalling CRDs labeled with: %s=%s", preinstaller.KYMA_CRD_LABEL_KEY, preinstaller.KYMA_LABEL_VALUE)

	selector, err := c.prepareKymaCrdLabelSelector()
	if err != nil {
		return Output{}, err
	}

	crdsList, err := c.fetchCrdsLabeledWith(selector)
	if err != nil {
		return Output{}, err
	}

	return c.deleteCrdsFrom(crdsList)
}

func (c *PostUninstaller) prepareKymaCrdLabelSelector() (selector labels.Selector, err error) {
	kymaCrdReq, err := labels.NewRequirement(preinstaller.KYMA_CRD_LABEL_KEY, selection.Equals, []string{preinstaller.KYMA_LABEL_VALUE})
	if err != nil {
		return nil, errors.Wrap(err, "Error occurred when preparing Kyma CRD label selector")
	}

	selector = labels.NewSelector()
	selector = selector.Add(*kymaCrdReq)
	return selector, nil
}

func (c *PostUninstaller) fetchCrdsLabeledWith(selector labels.Selector) (crds []*unstructured.Unstructured, err error) {
	crdsMap := make(map[*unstructured.Unstructured]bool)
	crdGvrV1Beta1 := c.crdGvkWith("v1beta1")
	crdGvrV1 := c.crdGvkWith("v1")
	gvrs := [2]schema.GroupVersionResource{crdGvrV1Beta1, crdGvrV1}

	for _, gvr := range gvrs {
		list, err := c.listResourcesUsing(gvr, selector)
		if err != nil {
			return nil, err
		}

		for _, obj := range list.Items {
			crdsMap[&obj] = true
		}
	}

	for obj, _ := range crdsMap {
		crds = append(crds, obj)
	}

	return
}

func (c *PostUninstaller) crdGvkWith(version string) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  version,
		Resource: "customresourcedefinitions",
	}
}

func (c *PostUninstaller) listResourcesUsing(gvr schema.GroupVersionResource, selector labels.Selector) (resourcesList *unstructured.UnstructuredList, err error) {
	err = retry.Do(func() error {
		resourcesList, err = c.dynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			c.cfg.Log.Warnf("Error occurred when retrieving resource: %s", err.Error())
			return nil
		}

		return nil
	}, c.retryOptions...)

	if err != nil {
		return nil, err
	}

	return
}

func (c *PostUninstaller) deleteCrdsFrom(crdsList []*unstructured.Unstructured) (o Output, err error) {
	for _, crd := range crdsList {
		crdName := crd.GetName()
		c.cfg.Log.Infof("Deleting resource: %s", crdName)
		err := c.resourceManager.DeleteResource(crdName, crd.GroupVersionKind())
		if err != nil {
			c.cfg.Log.Warnf("Error occurred when deleting a CRD %s : %s", crdName, err.Error())
			o.NotDeleted = append(o.NotDeleted, crdName)
			continue
		}

		o.Deleted = append(o.NotDeleted, crdName)
	}

	return o, nil
}
