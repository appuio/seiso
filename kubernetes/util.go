package kubernetes

import (
	"k8s.io/client-go/dynamic"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type (
	Kubernetes interface {
		ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error)
	}
	// KubernetesHelper is an implementation of the interface. (Better name? introduced for better testing support)
	KubernetesHelper struct {
		client dynamic.Interface
	}
)

func New() Kubernetes {
	return &KubernetesHelper{}
}

// ResourceContains evaluates if a given resource contains a given string
func (k *KubernetesHelper) ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error) {
	err := k.initClient()
	if err != nil {
		return false, err
	}
	objectlist, err := k.client.Resource(resource).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, item := range objectlist.Items {
		return ObjectContains(item.Object, value), nil
	}

	return false, nil
}

func (k *KubernetesHelper) initClient() error {
	if k.client == nil {
		client, err := NewDynamicClient()
		if err != nil {
			return err
		}
		k.client = client
	}
	return nil
}

// ObjectContains evaluates if a Kubernetes object contains a certain string
func ObjectContains(genericObject interface{}, value string) bool {
	switch (genericObject).(type) {
	case map[string]interface{}:
		objects := (genericObject).(map[string]interface{})

		for key := range objects {
			object := objects[key]
			if ObjectContains(object, value) {
				return true
			}
		}

		return false

	case []interface{}:
		for _, object := range (genericObject).([]interface{}) {
			if ObjectContains(object, value) {
				return true
			}
		}

		return false

	case string:
		return strings.Contains(genericObject.(string), value)

	default:
		return false
	}
}
