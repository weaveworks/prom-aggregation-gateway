package main

import (
	"sort"
	"strings"

	dto "github.com/prometheus/client_model/go"
)

func strPtr(s string) *string {
	return &s

}

func addLabels(m *dto.Metric, labels map[string]string) {
	found := make(map[string]struct{})

	for _, l := range m.Label {
		name := l.GetName()

		value, ok := labels[name]
		if !ok {
			continue
		}

		l.Value = strPtr(value)
		found[name] = struct{}{}
	}

	for name, value := range labels {
		if _, ok := found[name]; ok {
			continue
		}

		pair := dto.LabelPair{Name: strPtr(name), Value: strPtr(value)}
		m.Label = append(m.Label, &pair)
	}
}

func (a *aggregate) formatLabels(m *dto.Metric, labels map[string]string) {
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
