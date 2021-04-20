package configmap

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/appuio/seiso/pkg/openshift"
	"github.com/appuio/seiso/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type (
	ConfigMapsService struct {
		configuration ServiceConfiguration
		client        core.ConfigMapInterface
		helper        kubernetes.Kubernetes
	}
	ServiceConfiguration struct {
		Batch bool
	}
)

// NewConfigMapsService creates a new Service instance
func NewConfigMapsService(client core.ConfigMapInterface, helper kubernetes.Kubernetes, configuration ServiceConfiguration) ConfigMapsService {
	return ConfigMapsService{
		client:        client,
		helper:        helper,
		configuration: configuration,
	}
}

func (cms ConfigMapsService) List(ctx context.Context, listOptions metav1.ListOptions) ([]v1.ConfigMap, error) {
	configMaps, err := cms.client.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	return configMaps.Items, nil
}

func (cms ConfigMapsService) GetUnused(ctx context.Context, namespace string, configMaps []v1.ConfigMap) (unusedConfigMaps []v1.ConfigMap, funcErr error) {
	var usedConfigMaps []v1.ConfigMap
	funk.ForEach(openshift.PredefinedResources, func(predefinedResource schema.GroupVersionResource) {
		funk.ForEach(configMaps, func(resource v1.ConfigMap) {

			resourceName := resource.GetName()

			if funk.Contains(usedConfigMaps, resource) {
				// already marked as existing, skip this
				return
			}
			contains, err := cms.helper.ResourceContains(ctx, namespace, resourceName, predefinedResource)
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

func (cms ConfigMapsService) Delete(ctx context.Context, configMaps []v1.ConfigMap) error {
	for _, resource := range configMaps {
		err := cms.client.Delete(ctx, resource.Name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if cms.configuration.Batch {
			fmt.Println(resource.Name)
		} else {
			log.Infof("Deleted ConfigMap %s/%s", resource.Namespace, resource.Name)
		}
	}
	return nil
}

func (cms ConfigMapsService) FilterByTime(configMaps []v1.ConfigMap, olderThan time.Time) (filteredResources []v1.ConfigMap) {
	log.WithFields(log.Fields{
		"olderThan": olderThan,
	}).Debug("Filtering resources older than the specified time")

	for _, resource := range configMaps {
		if util.IsOlderThan(&resource, olderThan) {
			filteredResources = append(filteredResources, resource)
		}
	}
	return filteredResources
}

func (cms ConfigMapsService) FilterByMaxCount(configMaps []v1.ConfigMap, keep int) (filteredResources []v1.ConfigMap) {
	log.WithFields(log.Fields{
		"keep": keep,
	}).Debug("Filtering out oldest resources to a capped amount")

	if len(configMaps) <= keep {
		return []v1.ConfigMap{}
	}

	sort.SliceStable(configMaps, func(i, j int) bool {
		timestampFirst := configMaps[j].GetCreationTimestamp()
		timestampSecond := configMaps[i].GetCreationTimestamp()
		return util.CompareTimestamps(timestampFirst, timestampSecond)
	})

	return configMaps[keep:]
}

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
			log.Infof("Found candidate: %s/%s", resource.Namespace, resource.Name)
		}
	}
}
