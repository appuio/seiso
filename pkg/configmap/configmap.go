package configmap

import (
	"fmt"
	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/appuio/seiso/pkg/openshift"
	"github.com/appuio/seiso/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"sort"
	"time"
)

type (
	Service interface {
		ListNamesAndLabels() (configMapNames, labels []string, err error)
		List(listOptions metav1.ListOptions) (configMaps []v1.ConfigMap, err error)
		GetUnused(namespace string, configMaps []v1.ConfigMap) (unusedConfigMaps []v1.ConfigMap, funcErr error)
		Delete(configMaps []v1.ConfigMap)
		FilterByTime(configMaps []v1.ConfigMap, olderThan time.Time) (filteredConfigMaps []v1.ConfigMap)
		FilterByMaxCount(configMaps []v1.ConfigMap, keep int) (filteredConfigMaps []v1.ConfigMap)
		Print(configMaps []v1.ConfigMap)
	}
	ConfigMapsService struct {
		configuration Configuration
		client        core.ConfigMapInterface
		helper        kubernetes.Kubernetes
	}
)

type Configuration struct {
	Batch bool
}

// NewConfigMapsService creates a new Service instance
func NewConfigMapsService(client core.ConfigMapInterface, helper kubernetes.Kubernetes, configuration Configuration) ConfigMapsService {
	return ConfigMapsService{
		client:        client,
		helper:        helper,
		configuration: configuration,
	}
}

// ListNamesAndLabels return names and labels of Config Maps
func (cms ConfigMapsService) ListNamesAndLabels() (resourceNames, labels []string, err error) {
	configMaps, err := cms.List(metav1.ListOptions{})
	var objectMetas []metav1.ObjectMeta
	for _, cm := range configMaps {
		objectMetas = append(objectMetas, cm.ObjectMeta)
	}
	configMapNames, labels := util.GetNamesAndLabels(objectMetas)
	return configMapNames, labels, nil
}

// List returns a list of ConfigMaps from a namespace
func (cms ConfigMapsService) List(listOptions metav1.ListOptions) ([]v1.ConfigMap, error) {
	configMaps, err := cms.client.List(listOptions)
	if err != nil {
		return nil, err
	}

	return configMaps.Items, nil
}

// GetUnused return unused Config Maps
func (cms ConfigMapsService) GetUnused(namespace string, configMaps []v1.ConfigMap) (unusedConfigMaps []v1.ConfigMap, funcErr error) {
	var usedConfigMaps []v1.ConfigMap
	funk.ForEach(openshift.PredefinedResources, func(predefinedResource schema.GroupVersionResource) {
		funk.ForEach(configMaps, func(resource v1.ConfigMap) {

			resourceName := resource.GetName()

			if funk.Contains(usedConfigMaps, resource) {
				// already marked as existing, skip this
				return
			}
			contains, err := cms.helper.ResourceContains(namespace, resourceName, predefinedResource)
			if err != nil {
				funcErr = err
				return
			}

			if contains {
				usedConfigMaps = append(usedConfigMaps, resource)
			}
		})
	})

	for _, resource := range configMaps {
		if !funk.Contains(usedConfigMaps, resource) {
			unusedConfigMaps = append(unusedConfigMaps, resource)
		}
	}

	return unusedConfigMaps, funcErr
}

// Delete removes Config Maps
func (cms ConfigMapsService) Delete(configMaps []v1.ConfigMap) {
	for _, resource := range configMaps {
		namespace := resource.GetNamespace()
		name := resource.GetName()
		kind := "ConfigMaps"

		if cms.configuration.Batch {
			fmt.Println(resource.GetName())
		} else {
			log.Infof("Deleting %s %s/%s", kind, namespace, name)
		}

		err := cms.client.Delete(name, &metav1.DeleteOptions{})

		if err != nil {
			log.WithError(err).Errorf("Failed to delete config map %s/%s", namespace, name)
		}
	}
}

//FilterByTime returns config maps which are older than specified date
func (cms ConfigMapsService) FilterByTime(configMaps []v1.ConfigMap, olderThan time.Time) (filteredResources []v1.ConfigMap) {
	log.WithFields(log.Fields{
		"olderThan": olderThan,
	}).Debug("Filtering resources older than the specified time")

	for _, resource := range configMaps {
		lastUpdatedDate := resource.GetCreationTimestamp()
		// In case the creation date is null (isZero()) treat as oldest
		if lastUpdatedDate.IsZero() || lastUpdatedDate.Time.Before(olderThan) {
			filteredResources = append(filteredResources, resource)
			log.WithFields(log.Fields{
				"configMap": resource.Name,
			}).Debug("Filtering resource")

		} else {
			log.WithField("name", resource.GetName()).Debug("Filtered resource")
		}
	}

	return filteredResources
}

// FilterByMaxCount keep at most n newest resources. The list of config maps is sorted in descending ordered in
func (cms ConfigMapsService) FilterByMaxCount(configMaps []v1.ConfigMap, keep int) (filteredResources []v1.ConfigMap) {

	log.WithFields(log.Fields{
		"keep":       keep,
		"configMaps": configMaps,
	}).Debug("Filtering ordered by time Resources from the n'th number specified")

	sort.SliceStable(configMaps, func(i, j int) bool {
		timestampFirst := configMaps[j].GetCreationTimestamp()
		timestampSecond := configMaps[i].GetCreationTimestamp()
		if timestampFirst.IsZero() || timestampFirst.IsZero() && timestampSecond.IsZero() {
			return true
		} else if timestampSecond.IsZero() {
			return false
		}
		return configMaps[j].GetCreationTimestamp().Time.Before(configMaps[i].GetCreationTimestamp().Time)
	})

	if len(configMaps) <= keep {
		return []v1.ConfigMap{}
	}

	return configMaps[keep:]
}

// Print prints the given resource line by line. In batch mode, only the resource is printed, otherwise default
// log with info level
func (cms ConfigMapsService) Print(resources []v1.ConfigMap) {
	if len(resources) == 0 {
		log.Info("Nothing found to be deleted.")
	}
	if cms.configuration.Batch {
		for _, resource := range resources {
			fmt.Println(resource.GetName())
		}
	} else {
		for _, resource := range resources {
			log.Infof("Found candidate: %s/%s", resource.Namespace, resource.GetName())
		}
	}
}
