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

//go:generate mockgen -source=secrets_test.go -destination=mocks/mock_secrets.go -package=mocks -aux_files=github.com/appuio/seiso/cmd=common.go MockSecretResources
type TestSecretResources interface {
	resource.SecretResources
}

func TestExecuteSecretCleanupCommandWithNoLabels(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSecrets := mocks.NewMockTestSecretResources(mockCtrl)

	defaultConfig := cfg.NewDefaultConfig()
	mockSecrets.EXPECT().ListSecrets().Return([]string{"resource_1, resource_2"}, []string{"label_A"}, nil)
	mockSecrets.EXPECT().GetNamespace().Return("test_namespace")

	err := executeSecretCleanupCommand(mockSecrets, defaultConfig)

	assert.True(t, err==nil)
}

func TestExecuteSecretCleanupCommandWithNoLabelsAndError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSecrets := mocks.NewMockTestSecretResources(mockCtrl)

	defaultConfig := cfg.NewDefaultConfig()
	errorString := "test error"
	mockSecrets.EXPECT().ListSecrets().Return([]string{"resource_1, resource_2"}, []string{"label_A"}, errors.New(errorString))
	mockSecrets.EXPECT().GetNamespace().Return("test_namespace")

	err := executeSecretCleanupCommand(mockSecrets, defaultConfig)

	assert.True(t, err != nil)
	assert.True(t, err.Error() == errorString)
}

func TestExecuteSecretCleanupCommandWithLabels(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSecrets := mocks.NewMockTestSecretResources(mockCtrl)

	defaultConfig := cfg.NewDefaultConfig()
	defaultConfig.Resource.Labels = []string{"labelA", "labelB"}
	mockSecrets.EXPECT().GetNamespace().Return("test_namespace")
	mockSecrets.EXPECT().LoadSecrets(gomock.Any()).Return(nil)
	mockSecrets.EXPECT().FilterUsed().Return(nil)
	mockSecrets.EXPECT().FilterByTime(gomock.Any())
	mockSecrets.EXPECT().FilterByMaxCount(gomock.Any())
	mockSecrets.EXPECT().Print(false)

	err := executeSecretCleanupCommand(mockSecrets, defaultConfig)

	assert.True(t, err == nil)
}
