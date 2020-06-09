package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
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
histogram_sum 4.5
histogram_count 1
`
	want = `# HELP counter A counter
# TYPE counter counter
counter 60
# HELP gauge A gauge
# TYPE gauge gauge
gauge 49.5
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
histogram_sum 7
histogram_count 2
`

	multilabel1 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 1
`
	multilabel2 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 2
`
	multilabelResult = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 3
`
	labelFields1 = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{path="/org/:orgId"} 1
ui_page_render_errors{path="/prom/:orgId"} 1
`
	labelFields2 = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{path="/prom/:orgId"} 1
`
	labelFieldResult = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{path="/org/:orgId"} 1
ui_page_render_errors{path="/prom/:orgId"} 2
`
	gaugeInput = `
# HELP ui_external_lib_loaded A gauge with entries in un-sorted order
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{name="ga",loaded="true"} 1
ui_external_lib_loaded{name="Intercom",loaded="true"} 1
ui_external_lib_loaded{name="mixpanel",loaded="true"} 1
`
	gaugeOutput = `# HELP ui_external_lib_loaded A gauge with entries in un-sorted order
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{loaded="true",name="Intercom"} 1
ui_external_lib_loaded{loaded="true",name="ga"} 1
ui_external_lib_loaded{loaded="true",name="mixpanel"} 1
`
	duplicateLabels = `
# HELP ui_external_lib_loaded Test with duplicate values
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{name="Munchkin",loaded="true"} 15171
ui_external_lib_loaded{name="Munchkin",loaded="true"} 1
`
	duplicateErrorPrefix = `Duplicate labels`

	reorderedLabels1 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 1
`
	reorderedLabels2 = `# HELP counter A counter
# TYPE counter counter
counter{b="b",a="a"} 2
`
	reorderedLabelsResult = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 3
`
)

func TestAggate(t *testing.T) {
	for _, c := range []struct {
		a, b        string
		want        string
		errorPrefix string
		err2        error
	}{
		{gaugeInput, gaugeInput, gaugeOutput, "", nil},
		{in1, in2, want, "", nil},
		{multilabel1, multilabel2, multilabelResult, "", nil},
		{labelFields1, labelFields2, labelFieldResult, "", nil},
		{duplicateLabels, "", "", duplicateErrorPrefix, nil},
		{reorderedLabels1, reorderedLabels2, reorderedLabelsResult, "", nil},
	} {
		a := newAggate()

		if err := a.parseAndMerge(strings.NewReader(c.a)); err != nil {
			if c.errorPrefix == "" {
				t.Logf("Unexpected error: %s", err)
				t.Fail()
			} else if !strings.HasPrefix(err.Error(), c.errorPrefix) {
				t.Logf("Expected error starting with %s, got %s", c.errorPrefix, err.Error())
				t.Fail()
			}
		}
		if err := a.parseAndMerge(strings.NewReader(c.b)); err != c.err2 {
			t.Logf("Expected %s, got %s", c.err2, err)
			t.Fail()
		}

		r := httptest.NewRequest("GET", "http://example.com/foo", nil)
		w := httptest.NewRecorder()
		a.handler(w, r)

		if have := w.Body.String(); have != c.want {
			text, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(c.want),
				B:        difflib.SplitLines(have),
				FromFile: "want",
				ToFile:   "have",
				Context:  3,
			})
			t.Log(text)
			t.Fail()
		}
	}
}
