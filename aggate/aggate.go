package aggate

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

func lablesLessThan(a, b []*dto.LabelPair) bool {
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
func (a byLabel) Less(i, j int) bool { return lablesLessThan(a[i].Label, a[j].Label) }

// Sort a slice of LabelPairs by name
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
		// No very meaninful way for us to merge gauges.  We'll sum them
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

func mergeFamily(a, b *dto.MetricFamily) (*dto.MetricFamily, error) {
	if *a.Type != *b.Type {
		return nil, fmt.Errorf("Cannot merge metric '%s': type %s != %s",
			*a.Name, a.Type.String(), b.Type.String())
	}

	output := &dto.MetricFamily{
		Name: a.Name,
		Help: a.Help,
		Type: a.Type,
	}

	i, j := 0, 0
	for i < len(a.Metric) && j < len(b.Metric) {
		if lablesLessThan(a.Metric[i].Label, b.Metric[j].Label) {
			output.Metric = append(output.Metric, a.Metric[i])
			i++
		} else if lablesLessThan(b.Metric[j].Label, a.Metric[i].Label) {
			output.Metric = append(output.Metric, b.Metric[j])
			j++
		} else {
			merged := mergeMetric(*a.Type, a.Metric[i], b.Metric[j])
			if merged != nil {
				output.Metric = append(output.Metric, merged)
			}
			i++
			j++
		}
	}
	for ; i < len(a.Metric); i++ {
		output.Metric = append(output.Metric, a.Metric[i])
	}
	for ; j < len(b.Metric); j++ {
		output.Metric = append(output.Metric, b.Metric[j])
	}
	return output, nil
}

type aggate struct {
	familiesLock sync.RWMutex
	families     map[string]*dto.MetricFamily
}

func NewAggate() *aggate {
	return &aggate{
		families: map[string]*dto.MetricFamily{},
	}
}

func validateFamily(f *dto.MetricFamily) error {
	// Map of fingerprints we've seen before in this family
	fingerprints := make(map[model.Fingerprint]struct{}, len(f.Metric))
	for _, m := range f.Metric {
		// Turn protobuf LabelSet into Prometheus model LabelSet
		lset := make(model.LabelSet, len(m.Label)+1)
		for _, p := range m.Label {
			lset[model.LabelName(p.GetName())] = model.LabelValue(p.GetValue())
		}
		lset[model.MetricNameLabel] = model.LabelValue(f.GetName())
		if err := lset.Validate(); err != nil {
			return err
		}
		fingerprint := lset.Fingerprint()
		if _, found := fingerprints[fingerprint]; found {
			return fmt.Errorf("Duplicate labels: %v", lset)
		}
		fingerprints[fingerprint] = struct{}{}
	}
	return nil
}

func appendServerLabels(queryArguments url.Values, labelQueryParam string, metrics []*dto.Metric) {
	for _, m := range metrics {
		if _, foundLabel := queryArguments[labelQueryParam]; foundLabel {
			for _, label := range queryArguments[labelQueryParam] {
				l := strings.Split(label, ":")
				if len(l) == 2 {
					m.Label = append(m.Label, &dto.LabelPair{Name: &l[0], Value: &l[1]})
				}
			}
		}
	}
}

func (a *aggate) ParseAndMerge(r io.Reader, queryArguments url.Values, labelQueryParam string) error {
	var parser expfmt.TextParser
	inFamilies, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return err
	}

	a.familiesLock.Lock()
	defer a.familiesLock.Unlock()
	for name, family := range inFamilies {
		// Append labels server side by provided labels in labelQueryParam
		if labelQueryParam != "" {
			appendServerLabels(queryArguments, labelQueryParam, family.Metric)
		}

		// Sort labels in case source sends them inconsistently
		for _, m := range family.Metric {
			sort.Sort(byName(m.Label))
		}

		if err := validateFamily(family); err != nil {
			return err
		}

		// family must be sorted for the merge
		sort.Sort(byLabel(family.Metric))

		existingFamily, ok := a.families[name]
		if !ok {
			a.families[name] = family
			continue
		}

		merged, err := mergeFamily(existingFamily, family)
		if err != nil {
			return err
		}

		a.families[name] = merged
	}

	return nil
}

func (a *aggate) Handler(w http.ResponseWriter, r *http.Request) {
	contentType := expfmt.Negotiate(r.Header)
	w.Header().Set("Content-Type", string(contentType))
	enc := expfmt.NewEncoder(w, contentType)

	a.familiesLock.RLock()
	defer a.familiesLock.RUnlock()
	metricNames := []string{}
	for name := range a.families {
		metricNames = append(metricNames, name)
	}
	sort.Sort(sort.StringSlice(metricNames))

	for _, name := range metricNames {
		if err := enc.Encode(a.families[name]); err != nil {
			http.Error(w, "An error has occurred during metrics encoding:\n\n"+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// TODO reset gauges
}
