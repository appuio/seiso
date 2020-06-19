package kubernetes

import (
	"k8s.io/client-go/dynamic"
	apps "k8s.io/client-go/kubernetes/typed/apps/v1"
	batch "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type GenericClient interface {
	NewAppsV1Client() (*apps.AppsV1Client, error)
	NewBatchV1beta1Client() (*batch.BatchV1beta1Client, error)
	NewDynamicClient() (dynamic.Interface, error)
	NewCoreV1Client() (*core.CoreV1Client, error)
}

// CoreV1Client is used to interact with features provided by the  group.
type CoreV1ClientInt interface {
	ComponentStatuses() core.ComponentStatusInterface
	ConfigMaps(namespace string) core.ConfigMapInterface
	Endpoints(namespace string) core.EndpointsInterface
	Events(namespace string) core.EventInterface
	LimitRanges(namespace string) core.LimitRangeInterface
	Namespaces() core.NamespaceInterface
	Nodes() core.NodeInterface
	PersistentVolumes() core.PersistentVolumeInterface
	PersistentVolumeClaims(namespace string) core.PersistentVolumeClaimInterface
	Pods(namespace string) core.PodInterface
	PodTemplates(namespace string) core.PodTemplateInterface
	ReplicationControllers(namespace string) core.ReplicationControllerInterface
	ResourceQuotas(namespace string) core.ResourceQuotaInterface
	Secrets(namespace string) core.SecretInterface
	Services(namespace string) core.ServiceInterface
	ServiceAccounts(namespace string) core.ServiceAccountInterface
}

type client struct {}

func Init() *client {
	return &client{}
}


// NewAppsV1Client for current kubeconfig
func (c *client) NewAppsV1Client() (*apps.AppsV1Client, error) {
	restConfig, err := RestConfig()
	if err != nil {
		return nil, err
	}

	return apps.NewForConfig(restConfig)
}

// NewBatchV1beta1Client creates a new BatchV1beta1Client for the current kubeconfig
func (c *client) NewBatchV1beta1Client() (*batch.BatchV1beta1Client, error) {
	restConfig, err := RestConfig()
	if err != nil {
		return nil, err
	}

	return batch.NewForConfig(restConfig)
}

// NewDynamicClient creates a new dynamic client
func (c *client) NewDynamicClient() (dynamic.Interface, error) {
	restConfig, err := RestConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(restConfig)
}

// NewCoreV1Client creates a new dynamic client
func (c *client) NewCoreV1Client() (*core.CoreV1Client, error) {
	restConfig, err := RestConfig()
	if err != nil {
		return nil, err
	}

	return core.NewForConfig(restConfig)
}
