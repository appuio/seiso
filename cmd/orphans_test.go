package cmd

import (
	"github.com/spf13/cobra"
	"regexp"
	"testing"

	"github.com/appuio/seiso/cfg"
	"github.com/stretchr/testify/assert"
)

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
				args: []string{"/"},
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
					Orphan: cfg.OrphanConfig{
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
			err := validateOrphanCommandInput(&cobra.Command{}, tt.input.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
