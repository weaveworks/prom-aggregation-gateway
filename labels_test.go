package main

import (
	"fmt"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestFormatLabels(t *testing.T) {
	a := newAggregate()
	a.options.ignoredLabels = []string{"ignore_me"}

	m := &dto.Metric{
		Label: []*dto.LabelPair{
			{Name: strPtr("thing2"), Value: strPtr("value2")},
			{Name: strPtr("ignore_me"), Value: strPtr("ignored_value")},
			{Name: strPtr("thing1"), Value: strPtr("value1")},
			{},
		},
	}
	a.formatLabels(m, "test")

	assert.Equal(t, &dto.LabelPair{Name: strPtr("job"), Value: strPtr("test")}, m.Label[0])
	assert.Equal(t, &dto.LabelPair{Name: strPtr("thing1"), Value: strPtr("value1")}, m.Label[1])
	assert.Equal(t, &dto.LabelPair{Name: strPtr("thing2"), Value: strPtr("value2")}, m.Label[2])
	assert.Len(t, m.Label, 3)

}

var testLabelTable = []struct {
	inputName     string
	m             *dto.Metric
	ignoredLabels []string
}{
	{"no_labels", &dto.Metric{Label: []*dto.LabelPair{}}, []string{}},
	{"no_labels_1_ignored_label", &dto.Metric{Label: []*dto.LabelPair{}},
		[]string{"ignore_me"}},
	{"no_ignored_labels", &dto.Metric{Label: []*dto.LabelPair{
		{Name: strPtr("l1"), Value: strPtr("v1")},
	}},
		[]string{},
	},
	{"no_ignored_labels_with_3_ignored_labels", &dto.Metric{Label: []*dto.LabelPair{
		{Name: strPtr("l1"), Value: strPtr("v1")},
	}},
		[]string{"ignore_me", "ignore_me_1", "ignore_me_2"},
	},
	{"1_ignored_labels_with_1_ignores_set", &dto.Metric{Label: []*dto.LabelPair{
		{Name: strPtr("l1"), Value: strPtr("v1")},
		{Name: strPtr("ignore_me"), Value: strPtr("ignore")},
	}},
		[]string{"ignore_me"},
	},
	{"1_ignored_labels_with_3_ignores_set", &dto.Metric{Label: []*dto.LabelPair{
		{Name: strPtr("l1"), Value: strPtr("v1")},
		{Name: strPtr("ignore_me"), Value: strPtr("ignore")},
	}},
		[]string{"ignore_me", "ignore_me_1", "ignore_me_2"},
	},
	{"2_ignored_labels", &dto.Metric{Label: []*dto.LabelPair{
		{Name: strPtr("l1"), Value: strPtr("v1")},
		{Name: strPtr("ignore_me"), Value: strPtr("ignore")},
		{Name: strPtr("ignore_me_1"), Value: strPtr("ignore1")},
	}},
		[]string{"ignore_me", "ignore_me_1", "ignore_me_2"},
	},
	{"2_ignored_labels_with_lots_of_labels_ignored_labels", &dto.Metric{Label: []*dto.LabelPair{
		{Name: strPtr("l1"), Value: strPtr("v1")},
		{Name: strPtr("ignore_me"), Value: strPtr("ignore")},
		{Name: strPtr("ignore_me_1"), Value: strPtr("ignore1")},
		{Name: strPtr("l3"), Value: strPtr("v3")},
		{Name: strPtr("l2"), Value: strPtr("v2")},
		{Name: strPtr("l5"), Value: strPtr("v5")},
		{Name: strPtr("l4"), Value: strPtr("v5")},
	}},
		[]string{"ignore_me", "ignore_me_1"},
	},
}

func BenchmarkFormatLabels(b *testing.B) {
	for _, v := range testLabelTable {
		a := newAggregate(AddIgnoredLabels(v.ignoredLabels...))
		b.Run(fmt.Sprintf("metric_type_%s", v.inputName), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				a.formatLabels(v.m, "test")
			}
		})
	}
}
