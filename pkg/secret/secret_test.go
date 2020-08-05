package secret

import (
	"errors"
	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	test "k8s.io/client-go/testing"
	"testing"
	"time"
)

type HelperKubernetes struct{}
type HelperKubernetesErr struct{}

func (k HelperKubernetes) ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error) {
	if "nameA" == value {
		return false, nil
	} else {
		return true, nil
	}
}

func (k HelperKubernetesErr) ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error) {
	return false, errors.New("error")
}

var testNamespace = "testNamespace"

func Test_PrintNamesAndLabels(t *testing.T) {

	tests := []struct {
		name      string
		secrets   []v1.Secret
		expectErr bool
		reaction  test.ReactionFunc
	}{
		{
			name:      "GivenListOfSecrets_WhenListError_ThenReturnError",
			secrets:   []v1.Secret{},
			reaction:  createErrorReactor(),
			expectErr: true,
		},
		// TODO: Add test case that asserts for correct lines printed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(convertToRuntime(tt.secrets)[:]...)
			clientset.PrependReactor("list", "secrets", tt.reaction)
			fakeClient := clientset.CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeClient, &HelperKubernetes{}, ServiceConfiguration{})
			err := service.PrintNamesAndLabels(testNamespace)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_List(t *testing.T) {

	tests := []struct {
		name      string
		secrets   []v1.Secret
		reaction  test.ReactionFunc
		expectErr bool
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
			name:      "GivenListOfSecrets_WhenListError_ThenReturnError",
			secrets:   []v1.Secret{},
			reaction:  createErrorReactor(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(convertToRuntime(tt.secrets)[:]...)
			if tt.reaction != nil {
				clientset.PrependReactor("list", "secrets", tt.reaction)
			}
			fakeClient := clientset.CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeClient, &HelperKubernetes{}, ServiceConfiguration{})

			list, err := service.List(metav1.ListOptions{})
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.secrets, list)
		})
	}
}

func Test_FilterByTime(t *testing.T) {

	tests := []struct {
		name            string
		secrets         []v1.Secret
		filteredSecrets []v1.Secret
		cutOffDate      time.Time
	}{
		{
			name:    "GivenListOfSecrets_WhenOlder_ThenReturnASubsetOfSecrets",
			secrets: generateBaseTestSecrets(),
			filteredSecrets: []v1.Secret{
				generateBaseTestSecrets()[1],
			},
			cutOffDate: time.Date(2015, 1, 1, 1, 0, 0, 0, time.UTC),
		},
		{
			name:            "GivenListOfSecrets_WhenNewer_ThenReturnEmptyList",
			secrets:         generateBaseTestSecrets(),
			filteredSecrets: []v1.Secret{},
			cutOffDate:      time.Date(2005, 1, 1, 1, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset(convertToRuntime(tt.secrets)[:]...).CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeClient, &HelperKubernetes{}, ServiceConfiguration{})
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
	}{
		{
			name:            "GivenListOfSecrets_FilterByMaxCountOne_ThenReturnOneSecret",
			secrets:         generateBaseTestSecrets(),
			filteredSecrets: []v1.Secret{generateBaseTestSecrets()[1]},
			keep:            1,
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
			fakeClient := fake.NewSimpleClientset(convertToRuntime(tt.secrets)[:]...).CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeClient, &HelperKubernetes{}, ServiceConfiguration{})
			filteredSecrets := service.FilterByMaxCount(tt.secrets, tt.keep)
			assert.ElementsMatch(t, filteredSecrets, tt.filteredSecrets)
		})
	}
}

func Test_Delete(t *testing.T) {
	tests := []struct {
		name              string
		secrets           []v1.Secret
		expectedRemaining []v1.Secret
		reaction          test.ReactionFunc
		expectErr         bool
	}{
		{
			name:              "GivenSetOfSecrets_WhenAllPresent_ThenDeleteAllOfThem",
			secrets:           generateBaseTestSecrets(),
			expectedRemaining: []v1.Secret{},
		},
		{
			name:              "GivenSetOfSecrets_WhenDeletionError_ThenReturnError",
			secrets:           generateBaseTestSecrets(),
			expectedRemaining: generateBaseTestSecrets(),
			reaction:          createErrorReactor(),
			expectErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(convertToRuntime(tt.secrets)[:]...)
			if tt.reaction != nil {
				clientset.PrependReactor("delete", "secrets", tt.reaction)
			}
			fakeClient := clientset.CoreV1().Secrets(testNamespace)
			service := NewSecretsService(fakeClient, &HelperKubernetes{}, ServiceConfiguration{})
			err := service.Delete(tt.secrets)
			if tt.expectErr {
				assert.Error(t, err)
			}
			list, err := fakeClient.List(metav1.ListOptions{})
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedRemaining, list.Items)
		})
	}
}

func Test_GetUnused(t *testing.T) {
	tests := []struct {
		name          string
		allSecrets    []v1.Secret
		unusedSecrets []v1.Secret
		expectErr     bool
	}{
		{
			name:          "GivenASetOfSecrets_WhenOneSecretIsUsed_ThenFilterItOut",
			allSecrets:    generateBaseTestSecrets(),
			unusedSecrets: []v1.Secret{generateBaseTestSecrets()[0]},
		},
		{
			name:          "GivenASetOfSecrets_WhenError_ThenReturnError",
			allSecrets:    generateBaseTestSecrets(),
			unusedSecrets: generateBaseTestSecrets(),
			expectErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var helper kubernetes.Kubernetes = HelperKubernetes{}
			if tt.expectErr {
				helper = HelperKubernetesErr{}
			}
			service := NewSecretsService(nil, helper, ServiceConfiguration{})
			unused, err := service.GetUnused(testNamespace, tt.allSecrets)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.unusedSecrets, unused)
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

func convertToRuntime(secrets []v1.Secret) (objects []runtime.Object) {
	for _, s := range secrets {
		objects = append(objects, s.DeepCopyObject())
	}
	return objects
}

func createErrorReactor() test.ReactionFunc {
	return func(action test.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("error")
	}
}
