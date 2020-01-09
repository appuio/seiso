package cleanup

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// MatchOption type defines how the tags should be matched
type MatchOption int8

const (
	MatchOptionDefault MatchOption = iota
	MatchOptionExact
	MatchOptionPrefix
)

// GetMatchingTags returns all tags matching one of the provided prefixes
func GetMatchingTags(values, tags *[]string, matchOption MatchOption) []string {
	var matchingTags []string

	log.WithField("values", values).Debug("values")
	log.WithField("tags", tags).Debug("tags")

	if len(*values) > 0 && len(*tags) > 0 {
		for _, value := range *values {
			for _, tag := range *tags {
				if match(tag, value, matchOption) {
					matchingTags = append(matchingTags, tag)
					log.WithFields(log.Fields{
						"tag":   tag,
						"value": value}).
						Debug("Tag matched")
				}
			}
		}
	}
	return matchingTags
}

// GetInactiveTags returns the tags without active tags (unsorted)
func GetInactiveTags(activeTags, tags *[]string) []string {
	var inactiveTags []string

	log.WithField("activeTags", activeTags).Debug("Active tags")
	log.WithField("tags", tags).Debug("Tags")

	for _, tag := range *tags {
		active := false
		for _, activeTag := range *activeTags {
			if tag == activeTag {
				active = true
				log.WithField("tag", tag).Debug("The tag is part of the active tags")
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

func match(tag, value string, matchOption MatchOption) bool {
	switch matchOption {
	case MatchOptionDefault, MatchOptionPrefix:
		return strings.HasPrefix(tag, value)
	case MatchOptionExact:
		return tag == value
	}
	return false
}
