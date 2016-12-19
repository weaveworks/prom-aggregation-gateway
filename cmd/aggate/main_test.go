package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	in1 = `
# HELP gauge A gauge
# TYPE gauge gauge
gauge 42
# HELP counter A counter
# TYPE counter counter
counter 31
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{le="1"} 0
histogram_bucket{le="2"} 0
histogram_bucket{le="3"} 3
histogram_bucket{le="4"} 4
histogram_bucket{le="5"} 4
histogram_bucket{le="6"} 4
histogram_bucket{le="7"} 4
histogram_bucket{le="8"} 4
histogram_bucket{le="9"} 4
histogram_bucket{le="10"} 4
histogram_bucket{le="+Inf"} 4
histogram_sum{} 2.5
histogram_count{} 1
`
	in2 = `
# HELP gauge A gauge
# TYPE gauge gauge
gauge 57
# HELP counter A counter
# TYPE counter counter
counter 29
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{le="1"} 0
histogram_bucket{le="2"} 0
histogram_bucket{le="3"} 0
histogram_bucket{le="4"} 4
histogram_bucket{le="5"} 5
histogram_bucket{le="6"} 5
histogram_bucket{le="7"} 5
histogram_bucket{le="8"} 5
histogram_bucket{le="9"} 5
histogram_bucket{le="10"} 5
histogram_bucket{le="+Inf"} 5
histogram_sum{} 4.5
histogram_count{} 1
`
	expected = `
# HELP gauge A gauge
# TYPE gauge gauge
gauge 99
# HELP counter A counter
# TYPE counter counter
counter 60
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{le="1"} 0
histogram_bucket{le="2"} 0
histogram_bucket{le="3"} 3
histogram_bucket{le="4"} 8
histogram_bucket{le="5"} 9
histogram_bucket{le="6"} 9
histogram_bucket{le="7"} 9
histogram_bucket{le="8"} 9
histogram_bucket{le="9"} 9
histogram_bucket{le="10"} 9
histogram_bucket{le="+Inf"} 9
histogram_sum{} 7
histogram_count{} 2
`
)

func TestAggate(t *testing.T) {
	a := newAggate()

	if err := a.parseAndMerge(strings.NewReader(in1)); err != nil {
		t.Fatal(err)
	}
	if err := a.parseAndMerge(strings.NewReader(in2)); err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	a.handler(w, r)

	if w.Body.String() != expected {
		fmt.Println(w.Body.String())
		t.Fatal("Don't match")
	}
}
