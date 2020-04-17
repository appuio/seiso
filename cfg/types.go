package cfg

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type (
	// Configuration holds a strongly-typed tree of the configuration
	Configuration struct {
		Git      GitConfig      `mapstructure:",squash"`
		History  HistoryConfig  `mapstructure:",squash"`
		Orphan   OrphanConfig   `mapstructure:",squash"`
		Resource ResourceConfig `mapstructure:",squash"`
		Log      LogConfig
		Force    bool
	}
	// GitConfig configures git repository
	GitConfig struct {
		CommitLimit  int    `mapstructure:"commit-limit"`
		RepoPath     string `mapstructure:"repo-path"`
		Tag          bool   `mapstructure:"tags"`
		SortCriteria string `mapstructure:"sort"`
	}
	// HistoryConfig configures the history command behaviour
	HistoryConfig struct {
		Keep int
	}
	// OrphanConfig configures the orphans command behaviour
	OrphanConfig struct {
		OlderThan           string `mapstructure:"older-than"`
		OrphanDeletionRegex string `mapstructure:"deletion-pattern"`
	}
	// LogConfig configures the log
	LogConfig struct {
		LogLevel string
		Batch    bool
		Verbose  bool
	}
	// ResourceConfig configures the resources and secrets
	ResourceConfig struct {
		Labels    []string `mapstructure:"label"`
		OlderThan string   `mapstructure:"older-than"`
	}
)

// NewDefaultConfig retrieves the hardcoded configs with sane defaults
func NewDefaultConfig() *Configuration {
	return &Configuration{
		Git: GitConfig{
			CommitLimit:  0,
			RepoPath:     ".",
			Tag:          false,
			SortCriteria: "version",
		},
		History: HistoryConfig{
			Keep: 10,
		},
		Orphan: OrphanConfig{
			OlderThan:           "2mo",
			OrphanDeletionRegex: "^[a-z0-9]{40}$",
		},
		Resource: ResourceConfig{
			Labels:    []string{},
			OlderThan: "2mo",
		},
		Force: false,
		Log: LogConfig{
			LogLevel: "info",
			Batch:    false,
			Verbose:  false,
		},
	}
}

//CoreObjectInterface defines interface for core kubernetes resources
type CoreObjectInterface interface {
	Delete(name string, options *metav1.DeleteOptions) error
}

//ResourceNamespaceSelector gets resource from client
type ResourceNamespaceSelector func(*core.CoreV1Client) CoreObjectInterface
