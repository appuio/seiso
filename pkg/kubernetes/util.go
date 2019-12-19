package kubernetes

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceContains evaluates if a given resource contains a given string
func ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error) {
	dynamicclient, err := NewDynamicClient()
	if err != nil {
		return false, err
	}

	objectlist, err := dynamicclient.Resource(resource).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, item := range objectlist.Items {
		return ObjectContains(item.Object, value), nil
	}

	return false, nil
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
