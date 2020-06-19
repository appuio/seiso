package resource

import (
	"fmt"
	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/appuio/seiso/pkg/openshift"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sort"
	"time"
)

type Resources interface {
	Delete(resourceSelectorFunc cfg.ResourceNamespaceSelector) error
	Print(batch bool)
	GetNamespace() string
	FilterByTime(olderThan time.Time)
	FilterUsed() error
	FilterByMaxCount(keep int)
}

type SecretResources interface {
	Resources
	LoadSecrets(listOptions metav1.ListOptions) error
	ListSecrets() ([]string, []string, error)
}

type ConfigMapResources interface {
	Resources
	LoadConfigMaps(listOptions metav1.ListOptions) error
	ListConfigMaps() ([]string, []string, error)
}

type Secrets struct {
	*GenericResources
}

type ConfigMaps struct {
	*GenericResources
}

type GenericResources struct {
	KubernetesResource *[]cfg.KubernetesResource
	Client kubernetes.CoreV1ClientInt
	Namespace string
}

type ImagesTagsInterface interface {
	DeleteImageTags()
	PrintImageTags()
}


var (
	predefinedResources = []schema.GroupVersionResource{
		{Version: "v1", Resource: "pods"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps.openshift.io", Version: "v1", Resource: "deploymentconfigs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},
		{Group: "extensions", Version: "v1beta1", Resource: "daemonsets"},
		{Group: "extensions", Version: "v1beta1", Resource: "deployments"},
		{Group: "extensions", Version: "v1beta1", Resource: "replicasets"},
	}
	helper = kubernetes.New()
)

// Load updates the structure with Secret resources
func (sr Secrets) LoadSecrets(listOptions metav1.ListOptions) (err error) {
	resources, err := listSecrets(sr.Client, sr.Namespace, listOptions)
	if err != nil {
		return err
	}

	sr.KubernetesResource = &resources

	return nil
}

// List returns a list of Secrets from a namespace
func (sr Secrets) ListSecrets() ([]string, []string, error) {
	secretResources, err := listSecrets(sr.Client, sr.Namespace, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	secretNames, labels := getNamesAndLabels(secretResources)
	return secretNames, labels, nil
}

// Load updates the structure with ConfigMap resources
func (cm ConfigMaps) LoadConfigMaps(listOptions metav1.ListOptions) (err error) {
	resources, err := listConfigMaps(cm.Client, cm.Namespace, listOptions)
	if err != nil {
		return err
	}

	cm.KubernetesResource = &resources

	return nil
}

// List returns a list of ConfigMaps from a namespace
func (cm ConfigMaps) ListConfigMaps() ([]string, []string, error) {
	configMapResources, err := listConfigMaps(cm.Client, cm.Namespace, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	configMapNames, labels := getNamesAndLabels(configMapResources)
	return configMapNames, labels, nil
}

// Delete deletes a list of ConfigMaps or secrets
func (r *GenericResources) Delete(resourceSelectorFunc cfg.ResourceNamespaceSelector) error {
	for _, resource := range *r.KubernetesResource {
		namespace := resource.GetNamespace()
		kind := resource.GetKind()
		name := resource.GetName()

		log.Infof("Deleting %s %s/%s", kind, namespace, name)

		if err := openshift.DeleteResource(name, &r.Client, resourceSelectorFunc); err != nil {
			log.WithError(err).Errorf("Failed to delete %s %s/%s", kind, namespace, name)
			return err
		}
	}
	return nil
}

// Print prints the given resource line by line. In batch mode, only the resource is printed, otherwise default
// log with info level
func (r *GenericResources) Print(batch bool) {
	if len(*r.KubernetesResource) == 0 {
		log.Info("Nothing found to be deleted.")
	}
	if batch {
		for _, resource := range *r.KubernetesResource {
			fmt.Println(resource.GetKind() + ": " + resource.GetName())
		}
	} else {
		for _, resource := range *r.KubernetesResource {
			log.Infof("Found %s candidate: %s/%s", resource.GetKind(), r.Namespace, resource.GetName())
		}
	}
}

func (r *GenericResources) GetNamespace() string {
	return r.Namespace
}

//FilterByTime returns resources which are older than specified date
func (r *GenericResources) FilterByTime(olderThan time.Time) {
	log.WithFields(log.Fields{
		"olderThan": olderThan,
		"resources": *r.KubernetesResource,
	}).Debug("Filtering resources older than the specified time")

	var filteredResources []cfg.KubernetesResource
	for _, resource := range *r.KubernetesResource {

		lastUpdatedDate := resource.GetCreationTimestamp()
		if lastUpdatedDate.Before(olderThan) {
			filteredResources = append(filteredResources, resource)
		} else {
			kind := resource.GetKind()
			name := resource.GetName()
			log.WithField("resource "+kind, name).Debug("Filtered resource " + kind + " " + name)
		}
	}

	r.KubernetesResource = &filteredResources
}

// FilterByMaxCount keep at most n newest resources. The list of resources is sorted in descending ordered in
func (r *GenericResources) FilterByMaxCount(keep int) {

	log.WithFields(log.Fields{
		"keep":      keep,
		"resources": *r.KubernetesResource,
	}).Debug("Filtering ordered by time ConfigMaps from the n'th number specified")

	resources := *r.KubernetesResource
	sort.SliceStable(resources, func(i, j int) bool {
		return resources[j].GetCreationTimestamp().Before(resources[i].GetCreationTimestamp())
	})

	if len(resources) <= keep {
		r.KubernetesResource = &[]cfg.KubernetesResource{}
	} else {
		*r.KubernetesResource = resources[keep:]
	}
}

// FilterUsed lists resources that are unused
func (r *GenericResources) FilterUsed() (funcErr error) {
	var usedResources []cfg.KubernetesResource
	var unusedResources []cfg.KubernetesResource
	funk.ForEach(predefinedResources, func(predefinedResource schema.GroupVersionResource) {
		funk.ForEach(*r.KubernetesResource, func(resource cfg.KubernetesResource) {

			resourceName := resource.GetName()

			if funk.Contains(usedResources, resource) {
				// already marked as existing, skip this
				return
			}
			contains, err := helper.ResourceContains(r.Namespace, resourceName, predefinedResource)
			if err != nil {
				funcErr = err
				return
			}

			if contains {
				usedResources = append(usedResources, resource)
			}
		})
	})

	for _, resource := range *r.KubernetesResource {
		if !funk.Contains(usedResources, resource) {
			unusedResources = append(unusedResources, resource)
		}
	}

	r.KubernetesResource = &unusedResources

	return funcErr
}

func listSecrets(coreClient kubernetes.CoreV1ClientInt, namespace string, listOptions metav1.ListOptions) ([]cfg.KubernetesResource, error) {
	secrets, err := coreClient.Secrets(namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	var resources []cfg.KubernetesResource
	for _, secret := range secrets.Items {
		resource := cfg.NewSecretResource(
			secret.GetName(),
			secret.GetNamespace(),
			secret.GetCreationTimestamp().Time,
			secret.GetLabels())
		resources = append(resources, resource)
	}
	return resources, err
}

func listConfigMaps(coreClient kubernetes.CoreV1ClientInt, namespace string, listOptions metav1.ListOptions) ([]cfg.KubernetesResource, error) {
	configMaps, err := coreClient.ConfigMaps(namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	var resources []cfg.KubernetesResource
	for _, configMap := range configMaps.Items {
		resource := cfg.NewConfigMapResource(
			configMap.GetName(),
			configMap.GetNamespace(),
			configMap.GetCreationTimestamp().Time,
			configMap.GetLabels())
		resources = append(resources, resource)
	}
	return resources, err
}


func getNamesAndLabels(resources []cfg.KubernetesResource) (resourceNames, labels []string) {
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
