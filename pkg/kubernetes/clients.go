package kubernetes

import (
	"k8s.io/client-go/dynamic"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

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
