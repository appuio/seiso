package cfg

type (
	// Configuration holds a strongly-typed tree of the configuration
	Configuration struct {
		Git     GitConfig     `mapstructure:",squash"`
		History HistoryConfig `mapstructure:",squash"`
		Orphan  OrphanConfig  `mapstructure:",squash"`
		Log     LogConfig
		Force   bool
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
	LogConfig struct {
		LogLevel string
		Batch    bool
		Verbose  bool
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
			Keep: 0,
		},
		Orphan: OrphanConfig{
			OlderThan:           "2mo",
			OrphanDeletionRegex: "^[a-z0-9]{40}$",
		},
		Force: false,
		Log: LogConfig{
			LogLevel: "info",
			Batch:    false,
			Verbose:  false,
		},
	}
}
