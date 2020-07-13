package configmap

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
		name           string
		configMaps     []v1.ConfigMap
		configMapNames []string
		labels         []string
		err            error
	}{
		{
			name:           "GivenListOfConfigMaps_WhenLabelsAndNamesDefined_ThenReturnNamesAndLabels",
			configMaps:     generateBaseTestConfigMaps(),
			configMapNames: []string{"nameA", "nameB"},
			labels:         []string{"keyA=valueA", "keyB=valueB", "keyC=valueC"},
		},
		{
			name: "GivenListOfConfigMaps_WhenOnlyNamesDefined_ThenReturnNamesWithNoLabels",
			configMaps: []v1.ConfigMap{
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
			configMapNames: []string{"nameA", "nameB"},
			labels:         []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeConfigMapInterface := fake.NewSimpleClientset(&tt.configMaps[0], &tt.configMaps[1]).CoreV1().ConfigMaps(testNamespace)
			service := NewConfigMapsService(fakeConfigMapInterface, &HelperKubernetes{}, Configuration{Batch: false})
			resources, labels, err := service.ListNamesAndLabels()
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.configMapNames, resources)
			assert.ElementsMatch(t, tt.labels, labels)
		})
	}
}

func Test_List(t *testing.T) {
	tests := []struct {
		name       string
		configMaps []v1.ConfigMap
		err        error
	}{
		{
			name:       "GivenListOfConfigMaps_WhenAllPresent_ThenReturnAllOfThem",
			configMaps: generateBaseTestConfigMaps(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeConfigMapInterface := fake.NewSimpleClientset(&tt.configMaps[0], &tt.configMaps[1]).CoreV1().ConfigMaps(testNamespace)
			service := NewConfigMapsService(fakeConfigMapInterface, &HelperKubernetes{}, Configuration{Batch: false})
			list, err := service.List(metav1.ListOptions{})
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.configMaps, list)
		})
	}
}

func Test_Delete(t *testing.T) {
	tests := []struct {
		name       string
		configMaps []v1.ConfigMap
		err        error
	}{
		{
			name:       "GivenASetOfConfigMaps_WhenAllPresent_ThenDeleteAllOfThem",
			configMaps: generateBaseTestConfigMaps(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeConfigMapInterface := fake.NewSimpleClientset(&tt.configMaps[0], &tt.configMaps[1]).CoreV1().ConfigMaps(testNamespace)
			service := NewConfigMapsService(fakeConfigMapInterface, &HelperKubernetes{}, Configuration{Batch: false})
			service.Delete(tt.configMaps)
			list, err := fakeConfigMapInterface.List(metav1.ListOptions{})
			assert.NoError(t, err)
			assert.EqualValues(t, 0, len(list.Items))
		})
	}
}

func Test_GetUnused(t *testing.T) {
	tests := []struct {
		name             string
		allConfigMaps    []v1.ConfigMap
		unusedConfigMaps []v1.ConfigMap
		err              error
	}{
		{
			name:          "GivenASetOfConfigMaps_WhenOneSecretIsUsed_ThenFilterItOut",
			allConfigMaps: generateBaseTestConfigMaps(),
			unusedConfigMaps: []v1.ConfigMap{
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
			service := NewConfigMapsService(nil, &HelperKubernetes{}, Configuration{Batch: false})
			unused, err := service.GetUnused(testNamespace, tt.allConfigMaps)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.unusedConfigMaps, unused)
		})
	}
}

func generateBaseTestConfigMaps() []v1.ConfigMap {
	return []v1.ConfigMap{
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
