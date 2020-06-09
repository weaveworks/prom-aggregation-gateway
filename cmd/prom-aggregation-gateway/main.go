package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
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

func mergeMetric(ty dto.MetricType, count int, a, b *dto.Metric) *dto.Metric {
	switch ty {
	case dto.MetricType_COUNTER:
		return &dto.Metric{
			Label: a.Label,
			Counter: &dto.Counter{
				Value: float64ptr(*a.Counter.Value + *b.Counter.Value),
			},
		}

	case dto.MetricType_GAUGE:
		// Average out value
		return &dto.Metric{
			Label: a.Label,
			Gauge: &dto.Gauge{
				Value: float64ptr((*a.Gauge.Value*(float64(count)-1) + *b.Gauge.Value) / float64(count)),
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

// Takes a new family (nf) and adds it to an existing family (nf)
func (a *aggate) mergeFamily(nf *dto.MetricFamily) error {

	metrics := make(map[model.Fingerprint]*dto.Metric)

	// Add exiting metrics
	ef, ok := a.families[*nf.Name]
	if ok {

		// Check the metric types
		if *ef.Type != *nf.Type {
			return fmt.Errorf("Cannot merge metric '%s': type %s != %s",
				*ef.Name, ef.Type.String(), nf.Type.String())
		}

		for _, m := range ef.Metric {
			fp, err := fingerprint(*ef.Name, m)
			if err != nil {
				return err
			}
			metrics[fp] = m
		}
	}

	// Merge or add new Metrics
	for _, m := range nf.Metric {

		fp, err := fingerprint(*nf.Name, m)
		if err != nil {
			return err
		}
		// Add count to fingerprints
		a.fingerprintCounts[fp]++

		oldMetric, ok := metrics[fp]
		if ok {
			metrics[fp] = mergeMetric(*nf.Type, a.fingerprintCounts[fp], oldMetric, m)
		} else {
			metrics[fp] = m
		}
	}

	// Add the metrics back
	nf.Metric = []*dto.Metric{}
	for _, m := range metrics {

		sort.Sort(byName(m.Label)) // Sort metrics labels
		nf.Metric = append(nf.Metric, m)
	}

	sort.Sort(byLabel(nf.Metric)) // Sort metrics
	a.families[*nf.Name] = nf

	return nil
}

func fingerprint(name string, m *dto.Metric) (f model.Fingerprint, err error) {
	lset := make(model.LabelSet, len(m.Label)+1)
	for _, p := range m.Label {
		lset[model.LabelName(p.GetName())] = model.LabelValue(p.GetValue())
	}
	lset[model.MetricNameLabel] = model.LabelValue(name)
	if err := lset.Validate(); err != nil {
		return f, err
	}
	return lset.Fingerprint(), nil
}

func validateFamily(f *dto.MetricFamily) error {
	// Map of fingerprints we've seen before in this family
	fingerprints := make(map[model.Fingerprint]struct{}, len(f.Metric))
	for _, m := range f.Metric {

		fingerprint, err := fingerprint(f.GetName(), m)
		if err != nil {
			return err
		}
		if _, found := fingerprints[fingerprint]; found {
			return fmt.Errorf("Duplicate labels: %v", m)
		}
		fingerprints[fingerprint] = struct{}{}
	}
	return nil
}

type aggate struct {
	sync.RWMutex
	families          map[string]*dto.MetricFamily
	fingerprintCounts map[model.Fingerprint]int
}

func newAggate() *aggate {
	return &aggate{
		families:          map[string]*dto.MetricFamily{},
		fingerprintCounts: make(map[model.Fingerprint]int),
	}
}

func (a *aggate) parseAndMerge(r io.Reader) error {
	var parser expfmt.TextParser
	inFamilies, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return err
	}

	a.Lock()
	defer a.Unlock()
	for _, family := range inFamilies {

		if err := validateFamily(family); err != nil {
			return err
		}

		if err := a.mergeFamily(family); err != nil {
			return err
		}
	}

	return nil
}

func (a *aggate) handler(w http.ResponseWriter, r *http.Request) {
	contentType := expfmt.Negotiate(r.Header)
	w.Header().Set("Content-Type", string(contentType))
	enc := expfmt.NewEncoder(w, contentType)

	a.Lock()
	defer a.Unlock()

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

	// reset gauges
	a.families = map[string]*dto.MetricFamily{}
	a.fingerprintCounts = make(map[model.Fingerprint]int)
}

func main() {
	listen := flag.String("listen", ":80", "Address and port to listen on.")
	cors := flag.String("cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")
	pushPath := flag.String("push-path", "/metrics/", "HTTP path to accept pushed metrics.")
	flag.Parse()

	a := newAggate()
	http.HandleFunc("/metrics", a.handler)
	http.HandleFunc(*pushPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", *cors)
		if err := a.parseAndMerge(r.Body); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})
	log.Fatal(http.ListenAndServe(*listen, nil))
}
