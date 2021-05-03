package cmd

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/appuio/seiso/cfg"
	"github.com/stretchr/testify/assert"
)

func Test_validateNamespaceCommandInput(t *testing.T) {
	type args struct {
		args   []string
		config cfg.Configuration
	}
	tests := map[string]struct {
		name    string
		input   args
		wantErr bool
	}{
		"ShouldThrowError_IfNoLabelSelector": {
			input: args{
				config: cfg.Configuration{
					Resource: cfg.ResourceConfig{
						DeleteAfter: "1s",
					},
				},
			},
			wantErr: true,
		},
		"ShouldThrowError_InvalidLabelSelector": {
			input: args{
				config: cfg.Configuration{
					Resource: cfg.ResourceConfig{
						DeleteAfter: "1s",
						Labels:      []string{"invalid"},
					},
				},
			},
			wantErr: true,
		},
		"ShouldThrowError_IfInvalidDeleteAfterFlag": {
			input: args{
				config: cfg.Configuration{
					Resource: cfg.ResourceConfig{
						Labels:      []string{"some=label"},
						DeleteAfter: "invalid",
					},
				},
			},
			wantErr: true,
		},
		"ShouldThrowError_IfNegativeDeleteAfterFlag": {
			input: args{
				config: cfg.Configuration{
					Resource: cfg.ResourceConfig{
						Labels:      []string{"some=label"},
						DeleteAfter: "-1s",
					},
				},
			},
			wantErr: true,
		},
		"Success_IfValidDeleteAfterFlag1d": {
			input: args{
				config: cfg.Configuration{
					Resource: cfg.ResourceConfig{
						Labels:      []string{"some=label"},
						DeleteAfter: "1d",
					},
				},
			},
			wantErr: false,
		},
		"Success_IfValidDeleteAfterFlag1d1y": {
			input: args{
				config: cfg.Configuration{
					Resource: cfg.ResourceConfig{
						Labels:      []string{"some=label"},
						DeleteAfter: "1d1y",
					},
				},
			},
			wantErr: false,
		},
		"Success_IfValidDeleteAfterFlag1w1m": {
			input: args{
				config: cfg.Configuration{
					Resource: cfg.ResourceConfig{
						Labels:      []string{"some=label"},
						DeleteAfter: "1w1m",
					},
				},
			},
			wantErr: false,
		},
		"Success_IfValidDeleteAfterFlag1h1s": {
			input: args{
				config: cfg.Configuration{
					Resource: cfg.ResourceConfig{
						Labels:      []string{"some=label"},
						DeleteAfter: "1h1s",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config = &tt.input.config
			err := validateNsCommandInput(&cobra.Command{}, tt.input.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
