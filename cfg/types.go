package cfg

type (
	// Configuration holds a strongly-typed tree of the configuration
	Configuration struct {
		Git     GitConfig
		History HistoryConfig
		Orphan  OrphanConfig
		Log     LogConfig
		Force   bool
	}
	// GitConfig configures git repository
	GitConfig struct {
		CommitLimit  int
		RepoPath     string
		Tag          bool
		SortCriteria string
	}
	// HistoryConfig configures the history command behaviour
	HistoryConfig struct {
		ImageRepository string
		Keep            int
	}
	// OrphanConfig configures the orphans command behaviour
	OrphanConfig struct {
		OlderThan           string
		OrphanDeletionRegex string
	}
	LogConfig struct {
		LogLevel string
		Batch    bool
		Verbose  bool
	}
)

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
