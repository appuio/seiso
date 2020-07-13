package secret

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

type HelperKubernetes struct{}

func (k *HelperKubernetes) ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error) {
	if "nameA" == value {
		return false, nil
	} else {
		return true, nil
	}
}

var now = metav1.Now()
var testNamespace = "testNamespace"

func Test_ListNamesAndLabels(t *testing.T) {
	tests := []struct {
		name        string
		secrets     []v1.Secret
		secretNames []string
		labels      []string
		err         error
	}{
		{
			name:        "GivenListOfSecrets_WhenLabelsAndNamesDefined_ThenReturnNamesAndLabels",
			secrets:     generateBaseTestSecrets(),
			secretNames: []string{"nameA", "nameB"},
			labels:      []string{"keyA=valueA", "keyB=valueB", "keyC=valueC"},
		},
		{
			name: "GivenListOfSecrets_WhenOnlyNamesDefined_ThenReturnNamesWithNoLabels",
			secrets: []v1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "nameA",
						Namespace:         testNamespace,
						CreationTimestamp: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "nameB",
						Namespace:         testNamespace,
						CreationTimestamp: now,
					},
				},
			},
			secretNames: []string{"nameA", "nameB"},
			labels:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeSecretInterface := fake.NewSimpleClientset(&tt.secrets[0], &tt.secrets[1]).CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeSecretInterface, &HelperKubernetes{}, Configuration{Batch: false})
			resources, labels, err := service.ListNamesAndLabels()
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.secretNames, resources)
			assert.ElementsMatch(t, tt.labels, labels)
		})
	}
}

func Test_List(t *testing.T) {

	tests := []struct {
		name    string
		secrets []v1.Secret
		err     error
	}{
		{
			name:    "GivenListOfSecrets_WhenAllPresent_ThenReturnAllOfThem",
			secrets: generateBaseTestSecrets(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeSecretInterface := fake.NewSimpleClientset(&tt.secrets[0], &tt.secrets[1]).CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeSecretInterface, &HelperKubernetes{}, Configuration{Batch: false})
			list, err := service.List(metav1.ListOptions{})
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.secrets, list)
		})
	}
}

func Test_Delete(t *testing.T) {
	tests := []struct {
		name    string
		secrets []v1.Secret
		err     error
	}{
		{
			name:    "GivenASetOfSecrets_WhenAllPresent_ThenDeleteAllOfThem",
			secrets: generateBaseTestSecrets(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeSecretInterface := fake.NewSimpleClientset(&tt.secrets[0], &tt.secrets[1]).CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeSecretInterface, &HelperKubernetes{}, Configuration{Batch: false})
			service.Delete(tt.secrets)
			list, err := fakeSecretInterface.List(metav1.ListOptions{})
			assert.NoError(t, err)
			assert.EqualValues(t, 0, len(list.Items))
		})
	}
}

func Test_GetUnused(t *testing.T) {
	tests := []struct {
		name          string
		allSecrets    []v1.Secret
		unusedSecrets []v1.Secret
		err           error
	}{
		{
			name:       "GivenASetOfSecrets_WhenOneSecretIsUsed_ThenFilterItOut",
			allSecrets: generateBaseTestSecrets(),
			unusedSecrets: []v1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "nameA",
						Namespace:         testNamespace,
						CreationTimestamp: now,
						Labels:            map[string]string{"keyA": "valueA"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSecretsService(nil, &HelperKubernetes{}, Configuration{Batch: false})
			unused, err := service.GetUnused(testNamespace, tt.allSecrets)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.unusedSecrets, unused)
		})
	}
}

func generateBaseTestSecrets() []v1.Secret {
	return []v1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nameA",
				Namespace:         testNamespace,
				CreationTimestamp: now,
				Labels:            map[string]string{"keyA": "valueA"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nameB",
				Namespace:         testNamespace,
				CreationTimestamp: now,
				Labels:            map[string]string{"keyB": "valueB", "keyC": "valueC"},
			},
		},
	}
}
