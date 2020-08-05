package util

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestFlattenStringMap(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "GivenStringMap_WhenSingleEntry_ThenReturnSingleString",
			labels:   map[string]string{"key": "value"},
			expected: "[key=value]",
		},
		{
			name:     "GivenStringMap_WhenMultipleEntries_ThenReturnMultipleStringsWithinBrackets",
			labels:   map[string]string{"key1": "value", "key2": "value2"},
			expected: "[key1=value, key2=value2]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenStringMap(tt.labels)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareTimestamps(t *testing.T) {
	tests := []struct {
		name           string
		first          metav1.Time
		second         metav1.Time
		expectedResult bool
	}{
		{
			name:           "GivenDifferentTimestamps_WhenFirstIsNewer_ThenReturnTrue",
			first:          parseTime("2000-01-01T12:00:00Z"),
			second:         parseTime("2000-01-01T13:00:00Z"),
			expectedResult: true,
		},
		{
			name:           "GivenDifferentTimestamps_WhenSecondIsNewer_ThenReturnFalse",
			first:          parseTime("2000-01-01T13:00:00Z"),
			second:         parseTime("2000-01-01T12:00:00Z"),
			expectedResult: false,
		},
		{
			name:           "GivenSameTimestamps_WhenBothAreEqual_ThenReturnFalse",
			first:          parseTime("2000-01-01T12:00:00Z"),
			second:         parseTime("2000-01-01T12:00:00Z"),
			expectedResult: false,
		},
		{
			name:           "GivenZeroTimestamps_WhenFirstIsZero_ThenReturnFalse",
			first:          zero(),
			second:         parseTime("2000-01-01T12:00:00Z"),
			expectedResult: false,
		},
		{
			name:           "GivenZeroTimestamps_WhenSecondIsZero_ThenReturnTrue",
			first:          parseTime("2000-01-01T12:00:00Z"),
			second:         zero(),
			expectedResult: true,
		},
		{
			name:           "GivenZeroTimestamps_WhenBothAreZero_ThenReturnFalse",
			first:          zero(),
			second:         zero(),
			expectedResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareTimestamps(tt.first, tt.second)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestIsOlderThan(t *testing.T) {
	tests := []struct {
		name           string
		resource       metav1.Object
		olderThan      time.Time
		expectedResult bool
	}{
		{
			name:           "GivenResourceWithTimestamp_WhenNewer_ThenReturnFalse",
			resource:       &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: parseTime("2000-01-01T13:00:00Z")}},
			olderThan:      parseTime("2000-01-01T12:00:00Z").Time,
			expectedResult: false,
		},
		{
			name:           "GivenResourceWithTimestamp_WhenOlder_ThenReturnTrue",
			resource:       &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: parseTime("2000-01-01T12:00:00Z")}},
			olderThan:      parseTime("2000-01-01T13:00:00Z").Time,
			expectedResult: true,
		},
		{
			name:           "GivenResourceWithoutTimestamp_WhenZero_ThenReturnTrue",
			resource:       &v1.ConfigMap{},
			olderThan:      parseTime("2000-01-01T13:00:00Z").Time,
			expectedResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOlderThan(tt.resource, tt.olderThan)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func parseTime(stamp string) metav1.Time {
	t, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		panic(err)
	}
	return metav1.Time{
		Time: t,
	}
}

func zero() metav1.Time {
	return metav1.Time{
		Time: time.Time{},
	}
}
