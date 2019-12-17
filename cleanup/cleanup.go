package cleanup

import (
	"strings"
)

// GetTagsMatchingPrefixes returns all tags matching one of the provided prefixes
func GetTagsMatchingPrefixes(prefixes []string, tags []string) []string {
	var matchingTags []string

	if len(prefixes) > 0 && len(tags) > 0 {
		for _, prefix := range prefixes {
			for _, tag := range tags {
				if strings.HasPrefix(tag, prefix) {
					matchingTags = append(matchingTags, tag)
				}
			}
		}
	}
	return matchingTags
}

// GetInactiveTags returns the tags without active tags (unsorted)
func GetInactiveTags(activeTags, tags []string) []string {
	var inactiveTags []string

	for _, tag := range tags {
		active := false
		for _, activeTag := range activeTags {
			if tag == activeTag {
				active = true
				break
			}
		}
		if !active {
			inactiveTags = append(inactiveTags, tag)
		}
	}

	return inactiveTags
}

// LimitTags returns the tags which should not be kept by removing the first n tags
func LimitTags(tags []string, keep int) []string {
	if len(tags) > keep {
		tags = tags[keep:]
	} else {
		tags = []string{}
	}

	return tags
}
