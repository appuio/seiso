package util

import (
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamesAndLabels(resources []metav1.ObjectMeta) (resourceNames, labels []string) {
	for _, resource := range resources {
		resourceNames = append(resourceNames, resource.GetName())
		for key, element := range resource.GetLabels() {
			label := key + "=" + element
			if !funk.ContainsString(labels, label) {
				labels = append(labels, label)
			}
		}
	}

	return resourceNames, labels
}
