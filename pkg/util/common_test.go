package util

import (
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_PrintNamesAndLabels(t *testing.T) {

	tests := []struct {
		name                    string
		resources               []metav1.ObjectMeta
		resourceNamesWithLabels map[string][]string
	}{
		{
			name:      "GivenListOfObjectMetas_WhenNamesAndLabelsDefined_ThenReturnAMapOfNamesAndLabels",
			resources: generateObjectMetas(),
			resourceNamesWithLabels: map[string][]string{
				"NameA": {
					"LabelKeyA=LabelValueA",
					"LabelKeyB=LabelValueB",
					"LabelKeyC=LabelValueC",
				},
				"NameB": nil,
				"NameC": {"LabelKeyD=LabelValueD"},
			},
		},
		{
			name:                    "GivenEmptyListOfObjectMetas_ThenReturnAnEmptyMapOfNamesAndLabels",
			resources:               []metav1.ObjectMeta{},
			resourceNamesWithLabels: map[string][]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namesAndLabels := GetNamesAndLabels(tt.resources)
			assert.Equal(t, namesAndLabels, tt.resourceNamesWithLabels)
		})
	}
}

func generateObjectMetas() []metav1.ObjectMeta {
	return []metav1.ObjectMeta{
		{
			Name: "NameA",
			Labels: map[string]string{
				"LabelKeyA": "LabelValueA",
				"LabelKeyB": "LabelValueB",
				"LabelKeyC": "LabelValueC",
			},
		},
		{
			Name:   "NameB",
			Labels: map[string]string{},
		},
		{
			Name: "NameC",
			Labels: map[string]string{
				"LabelKeyD": "LabelValueD",
			},
		},
	}
}
