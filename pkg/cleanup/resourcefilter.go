package cleanup

import (
	"sort"
	"time"

	"github.com/appuio/seiso/cfg"
	log "github.com/sirupsen/logrus"
)

type Filter interface {
	FilterResourcesByTime()
	FilterResourcesByMaxCount()
}

type FilterResources struct {
	resources []cfg.KubernetesResource
}

//FilterResourcesByTime returns resources which are older than specified date
func FilterResourcesByTime(resources []cfg.KubernetesResource, olderThan time.Time) (filteredResources []cfg.KubernetesResource) {
	log.WithFields(log.Fields{
		"olderThan": olderThan,
		"resources": resources,
	}).Debug("Filtering resources older than the specified time")

	for _, resource := range resources {

		lastUpdatedDate := resource.GetCreationTimestamp()
		if lastUpdatedDate.Before(olderThan) {
			filteredResources = append(filteredResources, resource)
		} else {
			kind := resource.GetKind()
			name := resource.GetName()
			log.WithField("resource "+kind, name).Debug("Filtered resource " + kind + " " + name)
		}
	}

	return filteredResources
}

// FilterResourcesByMaxCount keep at most n newest resources. The list of resources is sorted in descending ordered in
func FilterResourcesByMaxCount(resources []cfg.KubernetesResource, keep int) (filteredResources []cfg.KubernetesResource) {

	log.WithFields(log.Fields{
		"keep":      keep,
		"resources": resources,
	}).Debug("Filtering ordered by time ConfigMaps from the n'th number specified")

	sort.SliceStable(resources, func(i, j int) bool {
		return resources[j].GetCreationTimestamp().Before(resources[i].GetCreationTimestamp())
	})

	if len(resources) <= keep {
		return []cfg.KubernetesResource{}
	}

	return resources[keep:]
}
