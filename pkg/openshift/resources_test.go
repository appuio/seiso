package openshift

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thoas/go-funk"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type (
	MockHelper struct {
		mock.Mock
	}
)

func (m *MockHelper) ResourceContains(namespace, value string, resource schema.GroupVersionResource) (bool, error) {
	args := m.Called(namespace, value, resource)
	return args.Bool(0), args.Error(1)
}

func TestGetActiveImageStreamTags(t *testing.T) {
	type args struct {
		namespace       string
		imageStream     string
		imageStreamTags []string
	}
	tests := []struct {
		name                      string
		args                      args
		wantActiveImageStreamTags []string
		wantErr                   bool
		helperMock                *MockHelper
		resources                 []schema.GroupVersionResource
	}{
		{
			name: "ShouldFilter_ActiveImageStream_FromAPod",
			args: args{
				namespace:       "namespace",
				imageStream:     "image",
				imageStreamTags: []string{"inactive", "active"},
			},
			wantActiveImageStreamTags: []string{"active"},
			helperMock:                new(MockHelper),
			resources: []schema.GroupVersionResource{
				{Version: "v1", Resource: "pods"},
			},
		},
		{
			name: "ShouldFilter_ActiveImageStream_FromPodAndDeployment",
			args: args{
				namespace:       "namespace",
				imageStream:     "image",
				imageStreamTags: []string{"inactive", "active"},
			},
			wantActiveImageStreamTags: []string{"active"},
			helperMock:                new(MockHelper),
			resources: []schema.GroupVersionResource{
				{Version: "v1", Resource: "pods"},
				{Group: "apps", Version: "v1", Resource: "deployments"},
			},
		},
		{
			name: "ShouldThrowError_IfClientFails",
			args: args{
				namespace:       "namespace",
				imageStream:     "image",
				imageStreamTags: []string{"inactive", "active"},
			},
			helperMock: new(MockHelper),
			resources: []schema.GroupVersionResource{
				{Version: "v1", Resource: "pods"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper = tt.helperMock
			for _, resource := range PredefinedResources {
				for _, tag := range tt.args.imageStreamTags {
					value := funk.ContainsString(tt.wantActiveImageStreamTags, tag)
					var err error = nil
					if tt.wantErr {
						err = errors.New("client error")
					}
					tt.helperMock.
						On("ResourceContains", tt.args.namespace, "image:"+tag, resource).
						Return(value, err)
				}
			}
			result, err := GetActiveImageStreamTags(tt.args.namespace, tt.args.imageStream, tt.args.imageStreamTags)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantActiveImageStreamTags, result)
		})
	}
}
