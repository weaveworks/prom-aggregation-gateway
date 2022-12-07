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

type metricFamily struct {
	*dto.MetricFamily
	lock sync.RWMutex
}

type aggregate struct {
	familiesLock sync.RWMutex
	families     map[string]*metricFamily
}

func newAggregate() *aggregate {
	return &aggregate{
		families: map[string]*metricFamily{},
	}
}

// setFamilyOrGetExistingFamily either sets a new family or returns an existing family
func (a *aggregate) setFamilyOrGetExistingFamily(familyName string, family *dto.MetricFamily) *metricFamily {
	a.familiesLock.Lock()
	defer a.familiesLock.Unlock()
	existingFamily, ok := a.families[familyName]
	if !ok {
		a.families[familyName] = &metricFamily{MetricFamily: family}
		return nil
	}
	return existingFamily
}

func (a *aggregate) saveFamily(familyName string, family *dto.MetricFamily) error {
	existingFamily := a.setFamilyOrGetExistingFamily(familyName, family)
	if existingFamily != nil {
		err := existingFamily.mergeFamily(family)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *aggregate) parseAndMerge(r io.Reader) error {
	var parser expfmt.TextParser
	inFamilies, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return err
	}

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

		if err := a.saveFamily(name, family); err != nil {
			return err
		}

	}

	return nil
}

func (a *aggregate) handler(c *gin.Context) {
	contentType := expfmt.Negotiate(c.Request.Header)
	c.Header("Content-Type", string(contentType))
	enc := expfmt.NewEncoder(c.Writer, contentType)

	a.familiesLock.RLock()
	metricNames := []string{}
	for name := range a.families {
		metricNames = append(metricNames, name)
	}
	a.familiesLock.RUnlock()
	sort.Strings(metricNames)

	for _, name := range metricNames {
		a.families[name].lock.RLock()
		defer a.families[name].lock.RUnlock()
		if err := enc.Encode(a.families[name].MetricFamily); err != nil {
			log.Printf("An error has occurred during metrics encoding:\n\n%s\n", err.Error())
			return
		}
	}

	// TODO reset gauges
}
