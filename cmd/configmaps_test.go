package cmd

import (
	"errors"
	"github.com/appuio/seiso/cfg"
	"github.com/appuio/seiso/cmd/mocks"
	"github.com/appuio/seiso/pkg/resource"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:generate mockgen -source=configmaps_test.go -destination=mocks/mock_configmaps.go -package=mocks -aux_files=github.com/appuio/seiso/cmd=common.go MockConfigMapsResources
type TestConfigMapResources interface {
	resource.ConfigMapResources
}

func TestExecuteSConfigMapCleanupCommandWithNoLabels(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfigMaps := mocks.NewMockTestConfigMapResources(mockCtrl)

	defaultConfig := cfg.NewDefaultConfig()
	mockConfigMaps.EXPECT().ListConfigMaps().Return([]string{"resource_1, resource_2"}, []string{"label_A"}, nil)
	mockConfigMaps.EXPECT().GetNamespace().Return("test_namespace")

	err := executeConfigMapCleanupCommand(mockConfigMaps, defaultConfig)

	assert.True(t, err==nil)
}

func TestExecuteConfigMapsCleanupCommandWithNoLabelsAndError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfigMaps := mocks.NewMockTestConfigMapResources(mockCtrl)

	defaultConfig := cfg.NewDefaultConfig()
	errorString := "test error"
	mockConfigMaps.EXPECT().ListConfigMaps().Return([]string{"resource_1, resource_2"}, []string{"label_A"}, errors.New(errorString))
	mockConfigMaps.EXPECT().GetNamespace().Return("test_namespace")

	err := executeConfigMapCleanupCommand(mockConfigMaps, defaultConfig)

	assert.True(t, err != nil)
	assert.True(t, err.Error() == errorString)
}

func TestExecuteConfigMapsCleanupCommandWithLabels(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfigMaps := mocks.NewMockTestConfigMapResources(mockCtrl)

	defaultConfig := cfg.NewDefaultConfig()
	defaultConfig.Resource.Labels = []string{"labelA", "labelB"}
	mockConfigMaps.EXPECT().GetNamespace().Return("test_namespace")
	mockConfigMaps.EXPECT().LoadConfigMaps(gomock.Any()).Return(nil)
	mockConfigMaps.EXPECT().FilterUsed().Return(nil)
	mockConfigMaps.EXPECT().FilterByTime(gomock.Any())
	mockConfigMaps.EXPECT().FilterByMaxCount(gomock.Any())
	mockConfigMaps.EXPECT().Print(false)

	err := executeConfigMapCleanupCommand(mockConfigMaps, defaultConfig)

	assert.True(t, err == nil)
}
