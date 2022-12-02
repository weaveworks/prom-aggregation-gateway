package main

import (
	"io"
	"log"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type aggregate struct {
	familiesLock sync.RWMutex
	families     map[string]*dto.MetricFamily
}

func newAggregate() *aggregate {
	return &aggregate{
		families: map[string]*dto.MetricFamily{},
	}
}

var Aggregate *aggregate

func init() {
	Aggregate = newAggregate()
}

func (a *aggregate) parseAndMerge(r io.Reader) error {
	var parser expfmt.TextParser
	inFamilies, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return err
	}

	a.familiesLock.Lock()
	defer a.familiesLock.Unlock()
	for name, family := range inFamilies {
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

func aggregateHandler(c *gin.Context) {
	contentType := expfmt.Negotiate(c.Request.Header)
	c.Header("Content-Type", string(contentType))
	enc := expfmt.NewEncoder(c.Writer, contentType)

	Aggregate.familiesLock.RLock()
	defer Aggregate.familiesLock.RUnlock()
	metricNames := []string{}
	for name := range Aggregate.families {
		metricNames = append(metricNames, name)
	}
	sort.Strings(metricNames)

	for _, name := range metricNames {
		if err := enc.Encode(Aggregate.families[name]); err != nil {
			log.Println("An error has occurred during metrics encoding:\n\n" + err.Error())
			return
		}
	}

	// TODO reset gauges
}
