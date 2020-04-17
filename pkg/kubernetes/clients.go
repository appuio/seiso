package kubernetes

import (
	"k8s.io/client-go/dynamic"
	apps "k8s.io/client-go/kubernetes/typed/apps/v1"
	batch "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

// NewAppsV1Client for current kubeconfig
func NewAppsV1Client() (*apps.AppsV1Client, error) {
	restConfig, err := RestConfig()
	if err != nil {
		return nil, err
	}

	return apps.NewForConfig(restConfig)
}

// NewBatchV1beta1Client creates a new BatchV1beta1Client for the current kubeconfig
func NewBatchV1beta1Client() (*batch.BatchV1beta1Client, error) {
	restConfig, err := RestConfig()
	if err != nil {
		return nil, err
	}

	return batch.NewForConfig(restConfig)
}

// NewDynamicClient creates a new dynamic client
func NewDynamicClient() (dynamic.Interface, error) {
	restConfig, err := RestConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(restConfig)
}

// NewCoreV1Client creates a new dynamic client
func NewCoreV1Client() (*core.CoreV1Client, error) {
	restConfig, err := RestConfig()
	if err != nil {
		return nil, err
	}

	return core.NewForConfig(restConfig)
}
