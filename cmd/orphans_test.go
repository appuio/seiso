package cmd

import (
	"github.com/appuio/image-cleanup/cfg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func Test_splitNamespaceAndImagestream(t *testing.T) {
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
			name: "ShouldThrowError_IfRepoDoesNotContainImage",
			args: args{
				repo: "namespace/",
			},
			wantErr: true,
		},
		{
			name: "ShouldThrowError_IfRepoIsInvalid",
			args: args{
				repo: "asdf",
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

func Test_parseOrphanDeletionRegex(t *testing.T) {
	type args struct {
		orphanIncludeRegex string
	}
	tests := []struct {
		name    string
		args    args
		want    *regexp.Regexp
		wantErr bool
	}{
		{
			name: "ShouldParseRegex_IfValidPattern",
			args: args{
				orphanIncludeRegex: ".*",
			},
		},
		{
			name: "ShouldThrowError_IfInvalidPattern",
			args: args{
				orphanIncludeRegex: "*/g",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseOrphanDeletionRegex(tt.args.orphanIncludeRegex)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_validateOrphanCommandInput(t *testing.T) {

	type args struct {
		args   []string
		config cfg.Configuration
	}
	tests := []struct {
		name    string
		input   args
		wantErr bool
	}{
		{
			name: "ShouldThrowError_IfInvalidImageRepository",
			input: args{
				args: []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "ShouldThrowError_IfInvalidSortFlag",
			input: args{
				args: []string{"namespace/image"},
				config: cfg.Configuration{
					Git: cfg.GitConfig{
						SortCriteria: "invalid",
						Tag:          true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ShouldThrowError_IfInvalidDeletionPattern",
			input: args{
				args: []string{"namespace/image"},
				config: cfg.Configuration{
					Orphan:cfg.OrphanConfig{
						OrphanDeletionRegex: "*/g",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "ShouldThrowError_IfInvalidOlderThanFlag",
			input: args{
				args: []string{"namespace/image"},
				config: cfg.Configuration{
					Orphan: cfg.OrphanConfig{
						OlderThan: "invalid",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config = &tt.input.config
			err := validateOrphanCommandInput(tt.input.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
