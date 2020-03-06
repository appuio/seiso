package git

import (
	"errors"
	"sort"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

// SortOption type for sort order
type SortOption string

const (
	// SortOptionVersion sorts by version
	SortOptionVersion SortOption = "version"
	// SortOptionAlphabetic sorts in alphabetical order
	SortOptionAlphabetic SortOption = "alphabetic"
)

// IsValidSortValue function tries to cast the string to SortedTagBy type
func IsValidSortValue(sortValue string) bool {
	return SortOption(sortValue) == SortOptionVersion || SortOption(sortValue) == SortOptionAlphabetic
}

// Sort function sorts the slice according to the sort type
func sortTags(tags []string, sortTagBy SortOption) ([]string, error) {
	switch sortTagBy {

	case SortOptionVersion:
		var versionTags []*version.Version
		for _, raw := range tags {
			version, err := version.NewVersion(raw)
			if err != nil {
				log.WithError(err).WithField("tag", raw).Warn("Skipped invalid version")
			} else {
				versionTags = append(versionTags, version)
			}
		}

		sort.Sort(sort.Reverse(version.Collection(versionTags)))

		sortedTags := make([]string, len(versionTags))
		for i, sortedVersion := range versionTags {
			sortedTags[i] = sortedVersion.Original()
		}

		return sortedTags, nil

	case SortOptionAlphabetic:
		sort.Strings(tags)
		return tags, nil

	default:
		return nil, errors.New("Undefined sort type")
	}
}
