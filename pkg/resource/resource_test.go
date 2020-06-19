package resource

import (
	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/pkg/kubernetes"
	"github.com/appuio/seiso/pkg/resource/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"testing"
	"time"
)

//go:generate mockgen -source=resource_test.go -destination=mocks/mock_resources.go -package=mocks MockResources
type MockCoreClient interface {
	kubernetes.CoreV1ClientInt
}

type MockSecretInterface interface {
	core.SecretInterface
}

type KubernetesMock interface {

}

func TestExecuteLoad(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockResources := mocks.NewMockMockCoreClient(mockCtrl)
	mockSecretInterface := mocks.NewMockMockSecretInterface(mockCtrl)

	name := "SecretName"
	genericName := "SecretGenericName"
	namespace := "namespaceA"
	resourceVersion := "10"
	generation := 1
	labelKey := "A"
	labelValue := "LabelA"
	secret := v1.Secret {
		ObjectMeta: metav1.ObjectMeta{
			Name:                       name,
			GenerateName:               genericName,
			Namespace:                  namespace,
			ResourceVersion:            resourceVersion,
			Generation:                 int64(generation),
			ClusterName:                "",
			Labels: map[string]string{labelKey:labelValue},
		},
		Data:       map[string][]byte{},
		StringData: map[string]string{},
	Type:           v1.SecretTypeOpaque,
	}

	mockResources.EXPECT().Secrets(gomock.Any()).Return(mockSecretInterface)
	mockSecretInterface.EXPECT().List(gomock.Any()).Return(&v1.SecretList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items:    []v1.Secret{secret},
	}, nil)

	secrets := Secrets{
		GenericResources: &GenericResources{
			Client: mockResources,
			Namespace: namespace,
		},
	}
	err := secrets.Load(metav1.ListOptions{})

	assert.True(t, err==nil, "no errors to be expected")
	assert.True(t, len(*secrets.KubernetesResource) == 1, "one resource should be present")
	assert.True(t, (*secrets.KubernetesResource)[0].GetNamespace() == namespace, "one resource should be present")
	assert.True(t, (*secrets.KubernetesResource)[0].GetLabels()[labelKey] == labelValue, "label key A should be present")
}

func TestExecuteList(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockResources := mocks.NewMockMockCoreClient(mockCtrl)
	mockSecretInterface := mocks.NewMockMockSecretInterface(mockCtrl)

	name := "SecretName"
	genericName := "SecretGenericName"
	namespace := "namespaceA"
	resourceVersion := "10"
	generation := 1
	labelKeyA := "A"
	labelValueA := "LabelA"
	labelKeyB := "B"
	labelValueB := "LabelB"
	secret := v1.Secret {
		ObjectMeta: metav1.ObjectMeta{
			Name:                       name,
			GenerateName:               genericName,
			Namespace:                  namespace,
			ResourceVersion:            resourceVersion,
			Generation:                 int64(generation),
			ClusterName:                "",
			Labels: map[string]string{labelKeyA:labelValueA,labelKeyB:labelValueB},
		},
		Data:       map[string][]byte{},
		StringData: map[string]string{},
		Type:           v1.SecretTypeOpaque,
	}

	mockResources.EXPECT().Secrets(gomock.Any()).Return(mockSecretInterface)
	mockSecretInterface.EXPECT().List(gomock.Any()).Return(&v1.SecretList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items:    []v1.Secret{secret},
	}, nil)

	secrets := Secrets{
		GenericResources: &GenericResources{
			Client: mockResources,
			Namespace: namespace,
		},
	}
	list, labels, err := secrets.List()

	assert.True(t, err==nil, "no errors to be expected")
	assert.True(t, len(labels) == 2, "two labels should be present")
	assert.Contains(t, labels, labelKeyA + "=" + labelValueA, "expecting label A")
	assert.Contains(t, labels, labelKeyB + "=" + labelValueB, "expecting label B")
	assert.True(t, len(list) == 1, "one resource should be present")
	assert.Contains(t, list, name, "the resource name should be as expected")
}

func TestExecuteDelete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockResources := mocks.NewMockMockCoreClient(mockCtrl)
	mockSecretInterface := mocks.NewMockMockSecretInterface(mockCtrl)

	name := "SecretName"
	kind := "Secret"
	namespace := "namespaceA"

	mockResources.EXPECT().Secrets(gomock.Any()).Return(mockSecretInterface)
	mockSecretInterface.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)

	resource := cfg.KubernetesResource{}
	resource.SetName(name)
	resource.SetNamespace(namespace)
	resource.SetKind(kind)

	secrets := Secrets{
		GenericResources: &GenericResources{
			Client:             mockResources,
			Namespace:          namespace,
			KubernetesResource: &[]cfg.KubernetesResource{resource},
		},
	}
	err := secrets.Delete(func(client kubernetes.CoreV1ClientInt) cfg.CoreObjectInterface {
		return client.Secrets(namespace)
	})

	assert.True(t, err==nil, "no errors to be expected")
}

func TestExecuteFilterByTime(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockResources := mocks.NewMockMockCoreClient(mockCtrl)

	nameA := "SecretNameA"
	kindA := "SecretA"
	namespaceA := "namespaceA"
	timeA := time.Date(2010,1, 1,0,0,0,0, time.UTC)

	nameB := "SecretNameB"
	kindB := "SecretB"
	namespaceB := "namespaceB"
	timeB := time.Date(2020,1, 1,0,0,0,0, time.UTC)

	resourceA := cfg.KubernetesResource{}
	resourceA.SetName(nameA)
	resourceA.SetNamespace(namespaceA)
	resourceA.SetKind(kindA)
	resourceA.SetCreationTimestamp(timeA)

	resourceB := cfg.KubernetesResource{}
	resourceB.SetName(nameB)
	resourceB.SetNamespace(namespaceB)
	resourceB.SetKind(kindB)
	resourceB.SetCreationTimestamp(timeB)

	secrets := Secrets{
		GenericResources: &GenericResources{
			Client:             mockResources,
			Namespace:          namespaceA,
			KubernetesResource: &[]cfg.KubernetesResource{resourceA, resourceB},
		},
	}
	secrets.FilterByTime(time.Date(2015,1, 1,0,0,0,0, time.UTC))

	assert.True(t, len(*secrets.KubernetesResource) == 1, "expecting one resource to be filtered")
	assert.True(
		t,
		(*secrets.KubernetesResource)[0].GetCreationTimestamp() == timeA,
		"expecting the newest resource to be preserved")
}

func TestExecuteFilterByMaXCountWithResources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockResources := mocks.NewMockMockCoreClient(mockCtrl)

	nameA := "SecretNameA"
	kindA := "SecretA"
	namespaceA := "namespaceA"
	timeA := time.Date(2010,1, 1,0,0,0,0, time.UTC)

	nameB := "SecretNameB"
	kindB := "SecretB"
	namespaceB := "namespaceB"
	timeB := time.Date(2020,1, 1,0,0,0,0, time.UTC)

	resourceA := cfg.KubernetesResource{}
	resourceA.SetName(nameA)
	resourceA.SetNamespace(namespaceA)
	resourceA.SetKind(kindA)
	resourceA.SetCreationTimestamp(timeA)

	resourceB := cfg.KubernetesResource{}
	resourceB.SetName(nameB)
	resourceB.SetNamespace(namespaceB)
	resourceB.SetKind(kindB)
	resourceB.SetCreationTimestamp(timeB)

	secrets := Secrets{
		GenericResources: &GenericResources{
			Client:             mockResources,
			Namespace:          namespaceA,
			KubernetesResource: &[]cfg.KubernetesResource{resourceA, resourceB},
		},
	}
	secrets.FilterByMaxCount(1)

	assert.True(t, len(*secrets.KubernetesResource) == 1, "expecting one resource to be filtered")
	assert.True(
		t,
		(*secrets.KubernetesResource)[0].GetCreationTimestamp() == timeA,
		"expecting the newest resource to be preserved")
}

func TestExecuteFilterByMacCountWithoutResources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockResources := mocks.NewMockMockCoreClient(mockCtrl)

	secrets := Secrets{
		GenericResources: &GenericResources{
			Client:             mockResources,
			Namespace:          "namespaceA",
			KubernetesResource: &[]cfg.KubernetesResource{},
		},
	}
	secrets.FilterByMaxCount(1)

	assert.True(t, len(*secrets.KubernetesResource) == 0, "expecting one resource to be filtered")
}
