package configmap

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
		name       string
		configMaps []v1.ConfigMap
		err        error
	}{
		{
			name:       "GivenListOfConfigMaps_WhenListError_ThenReturnError",
			configMaps: []v1.ConfigMap{},
			err:        errors.New("error config map"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset()
			clientset.PrependReactor("list", "configmaps", func(action test.Action) (handled bool, ret runtime.Object, err error) {
				return true, &v1.ConfigMapList{}, tt.err
			})
			fakeConfigMapsInterface := clientset.CoreV1().ConfigMaps(testNamespace)
			service := NewConfigMapsService(fakeConfigMapsInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
			err := service.PrintNamesAndLabels(testNamespace)
			assert.EqualError(t, err, tt.err.Error())
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
		{
			name:       "GivenEmptyListOfConfigMaps_ThenReturnNothing",
			configMaps: []v1.ConfigMap{},
		},
		{
			name:       "GivenListOfConfigMap_WhenListError_ThenReturnError",
			configMaps: []v1.ConfigMap{},
			err:        errors.New("error configmap"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fakeConfigMapInterface core.ConfigMapInterface
			if len(tt.configMaps) != 0 {
				fakeConfigMapInterface = fake.NewSimpleClientset(&tt.configMaps[0], &tt.configMaps[1]).CoreV1().ConfigMaps(testNamespace)
			} else {
				clientset := fake.NewSimpleClientset()
				clientset.PrependReactor("list", "configmaps", func(action test.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.ConfigMapList{}, tt.err
				})
				fakeConfigMapInterface = clientset.CoreV1().ConfigMaps(testNamespace)
			}

			service := NewConfigMapsService(fakeConfigMapInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})

			list, err := service.List(metav1.ListOptions{})
			if tt.err == nil {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.configMaps, list)
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}

func Test_FilterByTime(t *testing.T) {

	tests := []struct {
		name               string
		configMaps         []v1.ConfigMap
		filteredConfigMaps []v1.ConfigMap
		cutOffDate         time.Time
		err                error
	}{
		{
			name:       "GivenListOfConfigMaps_WhenFilteredByTime_ThenReturnASubsetOfConfigMaps",
			configMaps: generateBaseTestConfigMaps(),
			filteredConfigMaps: []v1.ConfigMap{
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
			name:               "GivenListOfConfigMaps_WhenFilteredBefore2010_ThenReturnEmptyList",
			configMaps:         generateBaseTestConfigMaps(),
			filteredConfigMaps: []v1.ConfigMap{},
			cutOffDate:         time.Date(2005, 1, 1, 1, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeConfigMapInterface := fake.NewSimpleClientset(&tt.configMaps[0], &tt.configMaps[1]).CoreV1().ConfigMaps(testNamespace)
			service := NewConfigMapsService(fakeConfigMapInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
			filteredConfigMaps := service.FilterByTime(tt.configMaps, tt.cutOffDate)
			assert.ElementsMatch(t, filteredConfigMaps, tt.filteredConfigMaps)
		})
	}
}

func Test_FilterByMaxCount(t *testing.T) {

	tests := []struct {
		name               string
		configMaps         []v1.ConfigMap
		filteredConfigMaps []v1.ConfigMap
		keep               int
		err                error
	}{
		{
			name:       "GivenListOfConfigMaps_FilterByMaxCountOne_ThenReturnOneConfigMap",
			configMaps: generateBaseTestConfigMaps(),
			filteredConfigMaps: []v1.ConfigMap{
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
			name:               "GivenListOfConfigMaps_FilterByMaxCountZero_ThenReturnTwoConfigMaps",
			configMaps:         generateBaseTestConfigMaps(),
			filteredConfigMaps: generateBaseTestConfigMaps(),
			keep:               0,
		},
		{
			name:               "GivenListOfConfigMaps_FilterByMaxCountTwo_ThenReturnEmptyList",
			configMaps:         generateBaseTestConfigMaps(),
			filteredConfigMaps: []v1.ConfigMap{},
			keep:               2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeConfigMapInterface := fake.NewSimpleClientset(&tt.configMaps[0], &tt.configMaps[1]).CoreV1().ConfigMaps(testNamespace)
			service := NewConfigMapsService(fakeConfigMapInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
			filteredConfigMaps := service.FilterByMaxCount(tt.configMaps, tt.keep)
			assert.ElementsMatch(t, filteredConfigMaps, tt.filteredConfigMaps)
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
		{
			name:       "GivenASetOfConfigMaps_WhenError_ThenReturnError",
			configMaps: generateBaseTestConfigMaps(),
			err:        errors.New("ConfigMap error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fakeConfigMapInterface core.ConfigMapInterface
			if tt.err == nil {
				fakeConfigMapInterface = fake.NewSimpleClientset(&tt.configMaps[0], &tt.configMaps[1]).CoreV1().ConfigMaps(testNamespace)
			} else {
				clientset := fake.NewSimpleClientset(&tt.configMaps[0], &tt.configMaps[1])
				clientset.PrependReactor("delete", "configmaps", func(action test.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, tt.err
				})
				fakeConfigMapInterface = clientset.CoreV1().ConfigMaps(testNamespace)
			}
			service := NewConfigMapsService(fakeConfigMapInterface, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
			service.Delete(tt.configMaps)
			list, err := fakeConfigMapInterface.List(metav1.ListOptions{})

			assert.NoError(t, err)
			if tt.err == nil {
				assert.EqualValues(t, 0, len(list.Items))
			} else {
				assert.EqualValues(t, 2, len(list.Items))
			}
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
			name:          "GivenASetOfConfigMaps_WhenOneConfigMapIsUsed_ThenFilterItOut",
			allConfigMaps: generateBaseTestConfigMaps(),
			unusedConfigMaps: []v1.ConfigMap{
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
			name:             "GivenASetOfConfigMaps_WhenError_ThenReturnError",
			allConfigMaps:    generateBaseTestConfigMaps(),
			unusedConfigMaps: generateBaseTestConfigMaps(),
			err:              errors.New("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				service := NewConfigMapsService(nil, &HelperKubernetes{}, ServiceConfiguration{Batch: false})
				unused, err := service.GetUnused(testNamespace, tt.allConfigMaps)
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.unusedConfigMaps, unused)
			} else {
				service := NewConfigMapsService(nil, &HelperKubernetesErr{}, ServiceConfiguration{Batch: false})
				unused, err := service.GetUnused(testNamespace, tt.allConfigMaps)
				assert.Error(t, err)
				assert.ElementsMatch(t, tt.unusedConfigMaps, unused)
			}
		})
	}
}

func generateBaseTestConfigMaps() []v1.ConfigMap {
	return []v1.ConfigMap{
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
