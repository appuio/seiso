package secret

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
		ListNamesAndLabels() (resourceNames, labels []string, err error)
		List(listOptions metav1.ListOptions) (resources []v1.Secret, err error)
		GetUnused(namespace string, secrets []v1.Secret) (unusedSecrets []v1.Secret, funcErr error)
		Delete(secrets []v1.Secret)
		FilterByTime(secrets []v1.Secret, olderThan time.Time) (filteredSecrets []v1.Secret)
		FilterByMaxCount(secrets []v1.Secret, keep int) (filteredSecrets []v1.Secret)
		Print(secrets []v1.Secret)
	}

	SecretsService struct {
		configuration Configuration
		client        core.SecretInterface
		helper        kubernetes.Kubernetes
	}
)

type Configuration struct {
	Batch bool
}

// NewSecretsService creates a new Service instance
func NewSecretsService(client core.SecretInterface, helper kubernetes.Kubernetes, configuration Configuration) Service {
	return SecretsService{
		client:        client,
		helper:        helper,
		configuration: configuration,
	}
}

// ListNamesAndLabels return the names and labels of all secrets
func (ss SecretsService) ListNamesAndLabels() (resourceNames, labels []string, err error) {
	secrets, err := ss.List(metav1.ListOptions{})
	var objectMetas []metav1.ObjectMeta
	for _, s := range secrets {
		objectMetas = append(objectMetas, s.ObjectMeta)
	}
	secretNames, labels := util.GetNamesAndLabels(objectMetas)
	return secretNames, labels, nil
}

// List returns a list of secrets from a namespace
func (ss SecretsService) List(listOptions metav1.ListOptions) ([]v1.Secret, error) {
	secrets, err := ss.client.List(listOptions)
	if err != nil {
		return nil, err
	}
	return secrets.Items, nil
}

// GetUnused returns unused resources
func (ss SecretsService) GetUnused(namespace string, resources []v1.Secret) (unusedResources []v1.Secret, funcErr error) {
	var usedSecrets []v1.Secret
	funk.ForEach(openshift.PredefinedResources, func(predefinedResource schema.GroupVersionResource) {
		funk.ForEach(resources, func(secret v1.Secret) {

			secretName := secret.GetName()

			if funk.Contains(usedSecrets, secret) {
				// already marked as existing, skip this
				return
			}
			contains, err := ss.helper.ResourceContains(namespace, secretName, predefinedResource)
			if err != nil {
				funcErr = err
				return
			}

			if contains {
				usedSecrets = append(usedSecrets, secret)
			}
		})
	})

	for _, resource := range resources {
		if !funk.Contains(usedSecrets, resource) {
			unusedResources = append(unusedResources, resource)
		}
	}

	return unusedResources, funcErr
}

// Delete removes secrets
func (ss SecretsService) Delete(secrets []v1.Secret) {
	for _, resource := range secrets {
		namespace := resource.GetNamespace()
		kind := "Secret"
		name := resource.GetName()

		if ss.configuration.Batch {
			fmt.Println(resource.GetName())
		} else {
			log.Infof("Deleting %s %s/%s", kind, namespace, name)
		}

		err := ss.client.Delete(name, &metav1.DeleteOptions{})

		if err != nil {
			log.WithError(err).Errorf("Failed to delete %s %s/%s", kind, namespace, name)
		}
	}
}

//FilterByTime returns secrets which are older than specified date
func (ss SecretsService) FilterByTime(secrets []v1.Secret, olderThan time.Time) (filteredResources []v1.Secret) {
	log.WithFields(log.Fields{
		"olderThan": olderThan,
	}).Debug("Filtering resources older than the specified time")

	for _, resource := range secrets {
		lastUpdatedDate := resource.GetCreationTimestamp()
		// In case the creation date is null (isZero()) treat as oldest
		if lastUpdatedDate.IsZero() || lastUpdatedDate.Time.Before(olderThan) {
			filteredResources = append(filteredResources, resource)
			log.WithFields(log.Fields{
				"secret": resource.Name,
			}).Debug("Filtering resource")
		} else {
			log.WithField("name", resource.GetName()).Debug("Filtered resource")
		}
	}

	return filteredResources
}

// FilterByMaxCount keep at most n newest resources. The list of secrets is sorted in descending ordered in
func (ss SecretsService) FilterByMaxCount(secrets []v1.Secret, keep int) (filteredResources []v1.Secret) {

	log.WithFields(log.Fields{
		"keep":    keep,
		"secrets": secrets,
	}).Debug("Filtering ordered by time Resources from the n'th number specified")

	sort.SliceStable(secrets, func(i, j int) bool {
		timestampFirst := secrets[j].GetCreationTimestamp()
		timestampSecond := secrets[i].GetCreationTimestamp()
		if timestampFirst.IsZero() || timestampFirst.IsZero() && timestampSecond.IsZero() {
			return true
		} else if timestampSecond.IsZero() {
			return false
		}
		return timestampSecond.Time.Before(timestampSecond.Time)
	})

	if len(secrets) <= keep {
		return []v1.Secret{}
	}

	return secrets[keep:]
}

// Print prints the given resource line by line. In batch mode, only the resource is printed, otherwise default
// log with info level
func (ss SecretsService) Print(resources []v1.Secret) {
	if len(resources) == 0 {
		log.Info("Nothing found to be deleted.")
	}
	if ss.configuration.Batch {
		for _, resource := range resources {
			fmt.Println(resource.GetName())
		}
	} else {
		for _, resource := range resources {
			log.Infof("Found candidate: %s/%s", resource.Namespace, resource.GetName())
		}
	}
}
