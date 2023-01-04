package metrics

import (
	"sort"
	"strings"

	dto "github.com/prometheus/client_model/go"
)

func strPtr(s string) *string {
	return &s
}

func addLabels(m *dto.Metric, labels []labelPair) {
	for _, label := range labels {
		pair := dto.LabelPair{Name: strPtr(label.name), Value: strPtr(label.value)}
		m.Label = append(m.Label, &pair)
	}
}

func (a *Aggregate) formatLabels(m *dto.Metric, labels []labelPair) {
	addLabels(m, labels)
	sort.Sort(byName(m.Label))

	if len(a.options.ignoredLabels) > 0 {
		var newLabelList []*dto.LabelPair
		for _, l := range m.Label {
			if !a.options.ignoredLabels.labelInIgnoredList(l) {
				newLabelList = append(newLabelList, l)
			}
		}
		m.Label = newLabelList
	}
}

func (iL ignoredLabels) labelInIgnoredList(l *dto.LabelPair) bool {
	if l == nil || l.Name == nil {
		return true
	}

	for _, label := range iL {
		if l.Name != nil {
			if strings.ToLower(*l.Name) == label {
				return true
			}
		}
	}

	return false
}
