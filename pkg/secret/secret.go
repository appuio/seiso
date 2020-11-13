package secret

import (
	"fmt"
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
	"sort"
	"time"
)

type (
	Service interface {
		// List returns a list of Secrets from a namespace
		List(listOptions metav1.ListOptions) (resources []v1.Secret, err error)
		// GetUnused returns unused Secrets.
		GetUnused(namespace string, secrets []v1.Secret) (unusedSecrets []v1.Secret, funcErr error)
		// Delete removes the given Secrets. Errors are logged only.
		Delete(secrets []v1.Secret) error
		// FilterByTime returns Secrets which are older than specified date
		FilterByTime(secrets []v1.Secret, olderThan time.Time) (filteredSecrets []v1.Secret)
		// FilterByMaxCount returns the latest resources until limited by <keep>. The list of secrets is sorted by
		// CreationTimestamp, with newest entries first.
		FilterByMaxCount(secrets []v1.Secret, keep int) (filteredSecrets []v1.Secret)
		// Print prints the given Secrets line by line. In batch mode, only the Secret name is printed, otherwise default
		// log with info level
		Print(secrets []v1.Secret)
	}

	SecretsService struct {
		configuration ServiceConfiguration
		client        core.SecretInterface
		helper        kubernetes.Kubernetes
	}
	ServiceConfiguration struct {
		Batch bool
	}
)

// NewSecretsService creates a new Service instance
func NewSecretsService(client core.SecretInterface, helper kubernetes.Kubernetes, configuration ServiceConfiguration) Service {
	return SecretsService{
		client:        client,
		helper:        helper,
		configuration: configuration,
	}
}

func (ss SecretsService) List(listOptions metav1.ListOptions) ([]v1.Secret, error) {
	secrets, err := ss.client.List(listOptions)
	if err != nil {
		return nil, err
	}
	return secrets.Items, nil
}

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

func (ss SecretsService) Delete(secrets []v1.Secret) error {
	for _, resource := range secrets {
		err := ss.client.Delete(resource.Name, &metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if ss.configuration.Batch {
			fmt.Println(resource.Name)
		} else {
			log.Infof("Deleted Secret %s/%s", resource.Namespace, resource.Name)
		}
	}
	return nil
}

func (ss SecretsService) FilterByTime(secrets []v1.Secret, olderThan time.Time) (filteredResources []v1.Secret) {
	log.WithFields(log.Fields{
		"olderThan": olderThan,
	}).Debug("Filtering resources older than the specified time.")

	for _, resource := range secrets {
		if util.IsOlderThan(&resource, olderThan) {
			filteredResources = append(filteredResources, resource)
		}
	}
	return filteredResources
}

func (ss SecretsService) FilterByMaxCount(secrets []v1.Secret, keep int) (filteredResources []v1.Secret) {
	log.WithFields(log.Fields{
		"keep": keep,
	}).Debug("Filtering out oldest resources to a capped amount.")

	if len(secrets) <= keep {
		return []v1.Secret{}
	}

	sort.SliceStable(secrets, func(i, j int) bool {
		timestampFirst := secrets[j].GetCreationTimestamp()
		timestampSecond := secrets[i].GetCreationTimestamp()
		return util.CompareTimestamps(timestampFirst, timestampSecond)
	})

	return secrets[keep:]
}

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
			log.Infof("Found candidate: %s/%s", resource.Namespace, resource.Name)
		}
	}
}
