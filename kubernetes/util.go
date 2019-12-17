package kubernetes

import (
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceContains evaluates if a given resource contains a given string
func ResourceContains(value string, resource schema.GroupVersionResource) bool {
	dynamicclient := NewDynamicClient()

	objectlist, err := dynamicclient.Resource(resource).Namespace(Namespace()).List(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).WithField("resource", resource).Fatal("Could not load objects.")
	}
	for _, item := range objectlist.Items {
		return ObjectContains(value, item.Object)
	}

	return false
}

// ObjectContains evaluates if a Kubernetes object contains a certain string
func ObjectContains(value string, genericObject interface{}) bool {
	switch genericObject.(type) {
	case map[string]interface{}:
		object := genericObject.(map[string]interface{})

		for key := range object {
			object := object[key]
			if ObjectContains(value, object) {
				return true
			}
		}

		return false

	case []interface{}:
		for _, object := range genericObject.([]interface{}) {
			if ObjectContains(value, object) {
				return true
			}
		}

		return false

	case string:
		if strings.Contains(genericObject.(string), value) {
			return true
		}

		return false

	default:
		return false
	}
}
