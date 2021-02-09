package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pmezard/go-difflib/difflib"
)

var (
	in1 = `
# HELP gauge A gauge
# TYPE gauge gauge
gauge 42 %[1]d
# HELP counter A counter
# TYPE counter counter
counter 31 %[1]d
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{le="1"} 0 %[1]d
histogram_bucket{le="2"} 0 %[1]d
histogram_bucket{le="3"} 3 %[1]d
histogram_bucket{le="4"} 4 %[1]d
histogram_bucket{le="5"} 4 %[1]d
histogram_bucket{le="6"} 4 %[1]d
histogram_bucket{le="7"} 4 %[1]d
histogram_bucket{le="8"} 4 %[1]d
histogram_bucket{le="9"} 4 %[1]d
histogram_bucket{le="10"} 4 %[1]d
histogram_bucket{le="+Inf"} 4 %[1]d
histogram_sum{} 2.5 %[1]d
histogram_count{} 1 %[1]d
`
	in2 = `
# HELP gauge A gauge
# TYPE gauge gauge
gauge 57 %[1]d
# HELP counter A counter
# TYPE counter counter
counter 29 %[1]d
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{le="1"} 0 %[1]d
histogram_bucket{le="2"} 0 %[1]d
histogram_bucket{le="3"} 0 %[1]d
histogram_bucket{le="4"} 4 %[1]d
histogram_bucket{le="5"} 5 %[1]d
histogram_bucket{le="6"} 5 %[1]d
histogram_bucket{le="7"} 5 %[1]d
histogram_bucket{le="8"} 5 %[1]d
histogram_bucket{le="9"} 5 %[1]d
histogram_bucket{le="10"} 5 %[1]d
histogram_bucket{le="+Inf"} 5 %[1]d
histogram_sum 4.5 %[1]d
histogram_count 1 %[1]d
`
	want = `# HELP counter A counter
# TYPE counter counter
counter 60 %[1]d
# HELP gauge A gauge
# TYPE gauge gauge
gauge 99 %[1]d
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{le="1"} 0 %[1]d
histogram_bucket{le="2"} 0 %[1]d
histogram_bucket{le="3"} 3 %[1]d
histogram_bucket{le="4"} 8 %[1]d
histogram_bucket{le="5"} 9 %[1]d
histogram_bucket{le="6"} 9 %[1]d
histogram_bucket{le="7"} 9 %[1]d
histogram_bucket{le="8"} 9 %[1]d
histogram_bucket{le="9"} 9 %[1]d
histogram_bucket{le="10"} 9 %[1]d
histogram_bucket{le="+Inf"} 9 %[1]d
histogram_sum 7 %[1]d
histogram_count 2 %[1]d
`

	multilabel1 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 1 %[1]d
`
	multilabel2 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 2 %[1]d
`
	multilabelResult = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 3 %[1]d
`
	labelFields1 = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{path="/org/:orgId"} 1 %[1]d
ui_page_render_errors{path="/prom/:orgId"} 1 %[1]d
`
	labelFields2 = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{path="/prom/:orgId"} 1 %[1]d
`
	labelFieldResult = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{path="/org/:orgId"} 1 %[1]d
ui_page_render_errors{path="/prom/:orgId"} 2 %[1]d
`
	gaugeInput = `
# HELP ui_external_lib_loaded A gauge with entries in un-sorted order
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{name="ga",loaded="true"} 1 %[1]d
ui_external_lib_loaded{name="Intercom",loaded="true"} 1 %[1]d
ui_external_lib_loaded{name="mixpanel",loaded="true"} 1 %[1]d
`
	gaugeOutput = `# HELP ui_external_lib_loaded A gauge with entries in un-sorted order
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{loaded="true",name="Intercom"} 2 %[1]d
ui_external_lib_loaded{loaded="true",name="ga"} 2 %[1]d
ui_external_lib_loaded{loaded="true",name="mixpanel"} 2 %[1]d
`
	duplicateLabels = `
# HELP ui_external_lib_loaded Test with duplicate values
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{name="Munchkin",loaded="true"} 15171
ui_external_lib_loaded{name="Munchkin",loaded="true"} 1
`
	duplicateError = `Duplicate labels: {__name__="ui_external_lib_loaded", loaded="true", name="Munchkin"}`

	reorderedLabels1 = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 1 %[1]d
`
	reorderedLabels2 = `# HELP counter A counter
# TYPE counter counter
counter{b="b",a="a"} 2 %[1]d
`
	reorderedLabelsResult = `# HELP counter A counter
# TYPE counter counter
counter{a="a",b="b"} 3 %[1]d
`
)

func TestAggate(t *testing.T) {
	now := time.Now().UnixNano() / int64(time.Millisecond)

	for _, c := range []struct {
		a, b string
		want string
		err1 error
		err2 error
	}{
		{gaugeInput, gaugeInput, gaugeOutput, nil, nil},
		{in1, in2, want, nil, nil},
		{multilabel1, multilabel2, multilabelResult, nil, nil},
		{labelFields1, labelFields2, labelFieldResult, nil, nil},
		{duplicateLabels, "", "", fmt.Errorf("%s", duplicateError), nil},
		{reorderedLabels1, reorderedLabels2, reorderedLabelsResult, nil, nil},
	} {
		a := newAggate(3600000)
		if c.b != "" {
			c.a, c.b, c.want = fmt.Sprintf(c.a, now), fmt.Sprintf(c.b, now), fmt.Sprintf(c.want, now)
		}

		if err := a.parseAndMerge(strings.NewReader(c.a)); err != nil {
			if c.err1 == nil {
				t.Fatalf("Unexpected error: '%s'", err)
			} else if c.err1.Error() != err.Error() {
				t.Fatalf("Expected '%s', got '%s'", c.err1, err)
			}
		}
		if err := a.parseAndMerge(strings.NewReader(c.b)); err != c.err2 {
			t.Fatalf("Expected '%s', got '%s'", c.err2, err)
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
			t.Fatal(text)
		}
	}
}
