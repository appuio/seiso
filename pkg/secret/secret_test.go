package secret

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	test "k8s.io/client-go/testing"
	"testing"
	"time"
)

type HelperKubernetes struct{}
type HelperKubernetesErr struct{}

func (k *HelperKubernetes) ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error) {
	if "nameA" == value {
		return false, nil
	} else {
		return true, nil
	}
}

func (k *HelperKubernetesErr) ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error) {
	return false, errors.New("error")
}

var testNamespace = "testNamespace"

func Test_PrintNamesAndLabels(t *testing.T) {

	tests := []struct {
		name    string
		secrets []v1.Secret
		err     error
	}{
		{
			name:    "GivenListOfSecrets_WhenListError_ThenReturnError",
			secrets: []v1.Secret{},
			err:     errors.New("error secret"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset()
			clientset.PrependReactor("list", "secrets", func(action test.Action) (handled bool, ret runtime.Object, err error) {
				return true, &v1.SecretList{}, tt.err
			})
			fakeSecretInterface := clientset.CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeSecretInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
			err := service.PrintNamesAndLabels(testNamespace)
			assert.EqualError(t, err, tt.err.Error())
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
		{
			name:    "GivenEmptyListOfSecrets_ThenReturnNothing",
			secrets: []v1.Secret{},
		},
		{
			name:    "GivenListOfSecrets_WhenListError_ThenReturnError",
			secrets: []v1.Secret{},
			err:     errors.New("error secret"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fakeSecretInterface core.SecretInterface
			if len(tt.secrets) != 0 {
				fakeSecretInterface = fake.NewSimpleClientset(&tt.secrets[0], &tt.secrets[1]).CoreV1().Secrets(testNamespace)
			} else {
				clientset := fake.NewSimpleClientset()
				clientset.PrependReactor("list", "secrets", func(action test.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.SecretList{}, tt.err
				})
				fakeSecretInterface = clientset.CoreV1().Secrets(testNamespace)
			}

			service := NewSecretsService(fakeSecretInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})

			list, err := service.List(metav1.ListOptions{})
			if tt.err == nil {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.secrets, list)
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}

func Test_FilterByTime(t *testing.T) {

	tests := []struct {
		name            string
		secrets         []v1.Secret
		filteredSecrets []v1.Secret
		cutOffDate      time.Time
		err             error
	}{
		{
			name:    "GivenListOfSecrets_WhenFilteredByTime_ThenReturnASubsetOfSecrets",
			secrets: generateBaseTestSecrets(),
			filteredSecrets: []v1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nameB",
						Namespace: testNamespace,
						CreationTimestamp: metav1.Time{
							Time: time.Date(2010, 1, 1, 1, 0, 0, 0, time.UTC),
						},
						Labels: map[string]string{"keyB": "valueB", "keyC": "valueC"},
					},
				},
			},
			cutOffDate: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
		},
		{
			name:            "GivenListOfSecrets_WhenFilteredBefore2010_ThenReturnEmptyList",
			secrets:         generateBaseTestSecrets(),
			filteredSecrets: []v1.Secret{},
			cutOffDate:      time.Date(2005, 1, 1, 1, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeSecretInterface := fake.NewSimpleClientset(&tt.secrets[0], &tt.secrets[1]).CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeSecretInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
			filteredSecrets := service.FilterByTime(tt.secrets, tt.cutOffDate)
			assert.ElementsMatch(t, filteredSecrets, tt.filteredSecrets)
		})
	}
}

func Test_FilterByMaxCount(t *testing.T) {

	tests := []struct {
		name            string
		secrets         []v1.Secret
		filteredSecrets []v1.Secret
		keep            int
		err             error
	}{
		{
			name:    "GivenListOfSecrets_FilterByMaxCountOne_ThenReturnOneSecret",
			secrets: generateBaseTestSecrets(),
			filteredSecrets: []v1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nameB",
						Namespace: testNamespace,
						CreationTimestamp: metav1.Time{
							Time: time.Date(2010, 1, 1, 1, 0, 0, 0, time.UTC),
						},
						Labels: map[string]string{"keyB": "valueB", "keyC": "valueC"},
					},
				},
			},
			keep: 1,
		},
		{
			name:            "GivenListOfSecrets_FilterByMaxCountZero_ThenReturnTwoSecrets",
			secrets:         generateBaseTestSecrets(),
			filteredSecrets: generateBaseTestSecrets(),
			keep:            0,
		},
		{
			name:            "GivenListOfSecrets_FilterByMaxCountTwo_ThenReturnEmptyList",
			secrets:         generateBaseTestSecrets(),
			filteredSecrets: []v1.Secret{},
			keep:            2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeSecretInterface := fake.NewSimpleClientset(&tt.secrets[0], &tt.secrets[1]).CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeSecretInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
			filteredSecrets := service.FilterByMaxCount(tt.secrets, tt.keep)
			assert.ElementsMatch(t, filteredSecrets, tt.filteredSecrets)
		})
	}
}

func Test_Delete(t *testing.T) {
	tests := []struct {
		name     string
		secrets  []v1.Secret
		remained int
		err      error
	}{
		{
			name:     "GivenASetOfSecrets_WhenAllPresent_ThenDeleteAllOfThem",
			secrets:  generateBaseTestSecrets(),
			remained: 0,
		},
		{
			name:     "GivenASetOfSecrets_WhenError_ThenReturnError",
			secrets:  generateBaseTestSecrets(),
			remained: 2,
			err:      errors.New("secret error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fakeSecretInterface core.SecretInterface
			if tt.err == nil {
				fakeSecretInterface = fake.NewSimpleClientset(&tt.secrets[0], &tt.secrets[1]).CoreV1().Secrets(testNamespace)
			} else {
				clientset := fake.NewSimpleClientset(&tt.secrets[0], &tt.secrets[1])
				clientset.PrependReactor("delete", "secrets", func(action test.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, tt.err
				})
				fakeSecretInterface = clientset.CoreV1().Secrets(testNamespace)
			}
			service := NewSecretsService(fakeSecretInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
			service.Delete(tt.secrets)
			list, err := fakeSecretInterface.List(metav1.ListOptions{})

			assert.NoError(t, err)
			if tt.err == nil {
				assert.EqualValues(t, tt.remained, len(list.Items))
			} else {
				assert.EqualValues(t, tt.remained, len(list.Items))
			}

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
						Name:      "nameA",
						Namespace: testNamespace,
						CreationTimestamp: metav1.Time{
							Time: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
						},
						Labels: map[string]string{"keyA": "valueA"},
					},
				},
			},
		},
		{
			name:          "GivenASetOfSecrets_WhenError_ThenReturnError",
			allSecrets:    generateBaseTestSecrets(),
			unusedSecrets: generateBaseTestSecrets(),
			err:           errors.New("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				service := NewSecretsService(nil, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
				unused, err := service.GetUnused(testNamespace, tt.allSecrets)
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.unusedSecrets, unused)
			} else {
				service := NewSecretsService(nil, &HelperKubernetesErr{}, ServiceConfiguration{Batch: false})
				unused, err := service.GetUnused(testNamespace, tt.allSecrets)
				assert.Error(t, err)
				assert.ElementsMatch(t, tt.unusedSecrets, unused)
			}
		})
	}
}

func generateBaseTestSecrets() []v1.Secret {
	return []v1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nameA",
				Namespace: testNamespace,
				CreationTimestamp: metav1.Time{
					Time: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
				},
				Labels: map[string]string{"keyA": "valueA"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nameB",
				Namespace: testNamespace,
				CreationTimestamp: metav1.Time{
					Time: time.Date(2010, 1, 1, 1, 0, 0, 0, time.UTC),
				},
				Labels: map[string]string{"keyB": "valueB", "keyC": "valueC"},
			},
		},
	}
}
