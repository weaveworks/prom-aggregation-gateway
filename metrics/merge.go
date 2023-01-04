package metrics

import (
	"fmt"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
)

func labelsLessThan(a, b []*dto.LabelPair) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if *a[i].Name != *b[j].Name {
			return *a[i].Name < *b[j].Name
		}
		if *a[i].Value != *b[j].Value {
			return *a[i].Value < *b[j].Value
		}
		i++
		j++
	}
	return len(a) < len(b)
}

type byLabel []*dto.Metric

func (a byLabel) Len() int           { return len(a) }
func (a byLabel) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLabel) Less(i, j int) bool { return labelsLessThan(a[i].Label, a[j].Label) }

// Sort a slice of labelPairs by name
type byName []*dto.LabelPair

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].GetName() < a[j].GetName() }

func uint64ptr(a uint64) *uint64 {
	return &a
}

func float64ptr(a float64) *float64 {
	return &a
}

func mergeBuckets(a, b []*dto.Bucket) []*dto.Bucket {
	output := []*dto.Bucket{}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if *a[i].UpperBound < *b[j].UpperBound {
			output = append(output, a[i])
			i++
		} else if *a[i].UpperBound > *b[j].UpperBound {
			output = append(output, b[j])
			j++
		} else {
			output = append(output, &dto.Bucket{
				CumulativeCount: uint64ptr(*a[i].CumulativeCount + *b[j].CumulativeCount),
				UpperBound:      a[i].UpperBound,
			})
			i++
			j++
		}
	}
	for ; i < len(a); i++ {
		output = append(output, a[i])
	}
	for ; j < len(b); j++ {
		output = append(output, b[j])
	}
	return output
}

func mergeMetric(ty dto.MetricType, a, b *dto.Metric) *dto.Metric {
	switch ty {
	case dto.MetricType_COUNTER:
		return &dto.Metric{
			Label: a.Label,
			Counter: &dto.Counter{
				Value: float64ptr(*a.Counter.Value + *b.Counter.Value),
			},
		}

	case dto.MetricType_GAUGE:
		// No very meaningful way for us to merge gauges.  We'll sum them
		// and clear out any gauges on scrape, as a best approximation, but
		// this relies on client pushing with the same interval as we scrape.
		return &dto.Metric{
			Label: a.Label,
			Gauge: &dto.Gauge{
				Value: float64ptr(*a.Gauge.Value + *b.Gauge.Value),
			},
		}

	case dto.MetricType_HISTOGRAM:
		return &dto.Metric{
			Label: a.Label,
			Histogram: &dto.Histogram{
				SampleCount: uint64ptr(*a.Histogram.SampleCount + *b.Histogram.SampleCount),
				SampleSum:   float64ptr(*a.Histogram.SampleSum + *b.Histogram.SampleSum),
				Bucket:      mergeBuckets(a.Histogram.Bucket, b.Histogram.Bucket),
			},
		}

	case dto.MetricType_UNTYPED:
		return &dto.Metric{
			Label: a.Label,
			Untyped: &dto.Untyped{
				Value: float64ptr(*a.Untyped.Value + *b.Untyped.Value),
			},
		}

	case dto.MetricType_SUMMARY:
		// No way of merging summaries, abort.
		return nil
	}

	return nil
}

func (mf *metricFamily) mergeFamily(b *dto.MetricFamily) error {
	if *mf.Type != *b.Type {
		return fmt.Errorf("cannot merge metric '%s': type %s != %s",
			*mf.Name, mf.Type.String(), b.Type.String())
	}

	newMetric := []*dto.Metric{}

	i, j := 0, 0
	mf.lock.Lock()
	defer mf.lock.Unlock()
	for i < len(mf.Metric) && j < len(b.Metric) {
		if labelsLessThan(mf.Metric[i].Label, b.Metric[j].Label) {
			newMetric = append(newMetric, mf.Metric[i])
			i++
		} else if labelsLessThan(b.Metric[j].Label, mf.Metric[i].Label) {
			newMetric = append(newMetric, b.Metric[j])
			j++
		} else {
			merged := mergeMetric(*mf.Type, mf.Metric[i], b.Metric[j])
			if merged != nil {
				newMetric = append(newMetric, merged)
			}
			i++
			j++
		}
	}

	for ; i < len(mf.Metric); i++ {
		newMetric = append(newMetric, mf.Metric[i])
	}
	for ; j < len(b.Metric); j++ {
		newMetric = append(newMetric, b.Metric[j])
	}

	mf.Metric = newMetric
	return nil
}

func validateFamily(f *dto.MetricFamily) error {
	// Map of fingerprints we've seen before in this family
	fingerprints := make(map[model.Fingerprint]struct{}, len(f.Metric))
	for _, m := range f.Metric {
		// Turn protobuf LabelSet into Prometheus model LabelSet
		lSet := make(model.LabelSet, len(m.Label)+1)
		for _, p := range m.Label {
			lSet[model.LabelName(p.GetName())] = model.LabelValue(p.GetValue())
		}
		lSet[model.MetricNameLabel] = model.LabelValue(f.GetName())
		if err := lSet.Validate(); err != nil {
			return err
		}
		fingerprint := lSet.Fingerprint()
		if _, found := fingerprints[fingerprint]; found {
			return fmt.Errorf("duplicate labels: %v", lSet)
		}
		fingerprints[fingerprint] = struct{}{}
	}
	return nil
}
