package util

import (
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamesAndLabels(resources []metav1.ObjectMeta) map[string][]string {
	resourceNamesWithLabels := make(map[string][]string)
	for _, resource := range resources {
		var labels []string
		for key, element := range resource.Labels {
			label := key + "=" + element
			if !funk.ContainsString(labels, label) {
				labels = append(labels, label)
			}
		}
		resourceNamesWithLabels[resource.Name] = labels
	}

	return resourceNamesWithLabels
}
