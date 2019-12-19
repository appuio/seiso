package cleanup

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

// GetTagsMatchingPrefixes returns all tags matching one of the provided prefixes
func GetTagsMatchingPrefixes(prefixes, tags *[]string) []string {
	var matchingTags []string

	log.Debugf("GetTagsMatchingPrefixes | Prefixes: %s", prefixes)
	log.Debugf("GetTagsMatchingPrefixes | Tags: %s", tags)

	if len(*prefixes) > 0 && len(*tags) > 0 {
		for _, prefix := range *prefixes {
			for _, tag := range *tags {
				if strings.HasPrefix(tag, prefix) {
					matchingTags = append(matchingTags, tag)
					log.Debugf("GetTagsMatchingPrefixes | Tag %s has prefix %s", tag, prefix)
				}
			}
		}
	}
	return matchingTags
}

// GetInactiveTags returns the tags without active tags (unsorted)
func GetInactiveTags(activeTags, tags *[]string) []string {
	var inactiveTags []string

	log.Debugf("GetInactiveTags | Active tags: %s", activeTags)
	log.Debugf("GetInactiveTags | Tags: %s", tags)

	for _, tag := range *tags {
		active := false
		for _, activeTag := range *activeTags {
			if tag == activeTag {
				active = true
				log.Debugf("GetInactiveTags | Tag %s is part of the active tags", tag)
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
func LimitTags(tags *[]string, keep int) []string {
	if len(*tags) > keep {
		limitedTags := make([]string, len(*tags)-keep)
		copy(limitedTags, (*tags)[keep:])
		return limitedTags
	}

	return []string{}
}
