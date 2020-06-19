package cfg

import (
	"github.com/appuio/seiso/pkg/kubernetes"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	// Configuration holds a strongly-typed tree of the configuration
	Configuration struct {
		Namespace string
		Git       GitConfig      `mapstructure:",squash"`
		History   HistoryConfig  `mapstructure:",squash"`
		Orphan    OrphanConfig   `mapstructure:",squash"`
		Resource  ResourceConfig `mapstructure:",squash"`
		Log       LogConfig
		Delete    bool
		Force     bool // deprecated! remove by June 30, 2020
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
	namespace, err := kubernetes.Namespace()
	if err != nil {
		log.Warning("Unable to determine default namespace. Falling back to: default")
		namespace = "default"
	}
	return &Configuration{
		Namespace: namespace,
		Git: GitConfig{
			CommitLimit:  0,
			RepoPath:     ".",
			Tag:          false,
			SortCriteria: "version",
		},
		History: HistoryConfig{
			Keep: 3,
		},
		Orphan: OrphanConfig{
			OlderThan:           "1w",
			OrphanDeletionRegex: "^[a-z0-9]{40}$",
		},
		Resource: ResourceConfig{
			Labels:    []string{},
			OlderThan: "1w",
		},
		Delete: false,
		Force:  false,
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
type ResourceNamespaceSelector func(kubernetes.CoreV1ClientInt) CoreObjectInterface
