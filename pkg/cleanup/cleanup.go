package cleanup

import (
	"strings"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
	imagev1 "github.com/openshift/api/image/v1"
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

// GetInactiveImageTags returns the tags without active tags (unsorted)
func GetInactiveImageTags(activeTags, tags *[]string) []string {
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

// GetOrphanImageTags returns the tags that do not have any git commit match
func GetOrphanImageTags(gitValues, imageTags *[]string, matchOption MatchOption) []string {
	orphans := []string{}

	log.WithField("gitValues", gitValues).Debug("Git commits/tags")
	log.WithField("imageTags", imageTags).Debug("Image stream tags")

	for _, tag := range *imageTags {
		found := false
		for _, value := range *gitValues {
			if match(tag, value, matchOption) {
				found = true
				break
			}
		}
		if !found {
			orphans = append(orphans, tag)
		}
	}

	return orphans
}

// FilterByRegex returns the tags that match the regexp
func FilterByRegex(imageTags *[]string, regexp *regexp.Regexp) []string {
	var matchedTags []string

	log.WithField("pattern:", regexp).Debug("Filtering image tags with regex...")
	
	for _, tag := range *imageTags {
		imageTagMatched := regexp.MatchString(tag)
		log.WithField("imageTag:", tag).WithField("match:", imageTagMatched).Debug("Matching image tag")
		if imageTagMatched {
			matchedTags = append(matchedTags, tag)
		}
	}
	return matchedTags
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

// FilterImageTagsByTime returns the tags which are older than the specified time
func FilterImageTagsByTime(imageStreamObjectTags *[]imagev1.NamedTagEventList, olderThan time.Time) []string {
	var imageStreamTags []string

	for _, imageStreamTag := range *imageStreamObjectTags {
		lastUpdatedDate := imageStreamTag.Items[0].Created.Time
		for _, tagEvent := range imageStreamTag.Items {
			if lastUpdatedDate.Before(tagEvent.Created.Time) {
				lastUpdatedDate = tagEvent.Created.Time
			}
		}

		if lastUpdatedDate.Before(olderThan) {
			imageStreamTags = append(imageStreamTags, imageStreamTag.Tag)
		}
	}

	return imageStreamTags
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
