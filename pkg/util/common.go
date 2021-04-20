package util

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TimeFormat defines the format to be used for dates
	TimeFormat = time.RFC3339
)

// IsOlderThan returns true if the given resource is older than the specified timestamp. If the resource does not have
// a timestamp or is zero, it returns true.
func IsOlderThan(resource metav1.Object, olderThan time.Time) bool {
	lastUpdatedDate := resource.GetCreationTimestamp()
	return lastUpdatedDate.IsZero() || lastUpdatedDate.Time.Before(olderThan)
}

// CompareTimestamps compares whether the first timestamp is newer than the second. If both timestamps share the same
// time down to the second, the nano second will be compared. If the time is zero, it will be treated as older than the other.
func CompareTimestamps(first, second metav1.Time) bool {
	if first.IsZero() {
		return false
	}
	if second.IsZero() {
		return true
	}
	return first.Time.Before(second.Time)
}
