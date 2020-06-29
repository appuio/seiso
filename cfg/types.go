package cfg

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/appuio/seiso/pkg/kubernetes"
)

type (
	// Configuration holds a strongly-typed tree of the configuration
	Configuration struct {
		Namespace string
		Git       GitConfig      `koanf:",squash"`
		History   HistoryConfig  `koanf:",squash"`
		Orphan    OrphanConfig   `koanf:",squash"`
		Resource  ResourceConfig `koanf:",squash"`
		Log       LogConfig
		Delete    bool
	}
	// GitConfig configures git repository
	GitConfig struct {
		CommitLimit  int    `koanf:"commit-limit"`
		RepoPath     string `koanf:"repo-path"`
		Tag          bool   `koanf:"tags"`
		SortCriteria string `koanf:"sort"`
	}
	// HistoryConfig configures the history command behaviour
	HistoryConfig struct {
		Keep int
	}
	// OrphanConfig configures the orphans command behaviour
	OrphanConfig struct {
		OlderThan           string `koanf:"older-than"`
		OrphanDeletionRegex string `koanf:"deletion-pattern"`
	}
	// LogConfig configures the log
	LogConfig struct {
		LogLevel string `koanf:"level"`
		Batch    bool
		Verbose  bool
	}
	// ResourceConfig configures the resources and secrets
	ResourceConfig struct {
		Labels    []string `koanf:"label"`
		OlderThan string   `koanf:"older-than"`
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
