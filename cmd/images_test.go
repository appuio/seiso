package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SplitNamespaceAndImagestream(t *testing.T) {
	type args struct {
		repo string
	}
	tests := []struct {
		name              string
		args              args
		expectedNamespace string
		expectedImage     string
		wantErr           bool
	}{
		{
			name: "ShouldSplit_NamespaceAndImageName",
			args: args{
				repo: "namespace/image",
			},
			expectedNamespace: "namespace",
			expectedImage:     "image",
		},
		{
			name: "ShouldReturnActiveNamespace_IfRepoDoesNotContainNamespace",
			args: args{
				repo: "image",
			},
			expectedNamespace: "currently-active-ns",
			expectedImage:     "image",
		},
		{
			name: "ShouldThrowError_IfRepoDoesNotContainImage",
			args: args{
				repo: "namespace/",
			},
			wantErr: true,
		},
		{
			name: "ShouldThrowError_IfRepoIsInvalid",
			args: args{
				repo: "/",
			},
			wantErr: true,
		},
		{
			name:    "ShouldThrowError_IfRepoIsEmpty",
			args:    args{},
			wantErr: true,
		},
		{
			name: "ShouldIgnore_Registry",
			args: args{
				repo: "docker.io/namespace/image",
			},
			expectedNamespace: "namespace",
			expectedImage:     "image",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Namespace = "currently-active-ns"
			namespace, image, err := splitNamespaceAndImagestream(tt.args.repo)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, namespace, tt.expectedNamespace)
			assert.Equal(t, image, tt.expectedImage)
		})
	}
}
