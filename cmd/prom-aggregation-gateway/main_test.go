package main

import (
	"fmt"
	"net/http/httptest"
	"net/url"
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
gauge 99
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
	wantWithServerLabels = `# HELP counter A counter
# TYPE counter counter
counter{appversion="3.2",osversion="2"} 60
# HELP gauge A gauge
# TYPE gauge gauge
gauge{appversion="3.2",osversion="2"} 99
# HELP histogram A histogram
# TYPE histogram histogram
histogram_bucket{appversion="3.2",osversion="2",le="1"} 0
histogram_bucket{appversion="3.2",osversion="2",le="2"} 0
histogram_bucket{appversion="3.2",osversion="2",le="3"} 3
histogram_bucket{appversion="3.2",osversion="2",le="4"} 8
histogram_bucket{appversion="3.2",osversion="2",le="5"} 9
histogram_bucket{appversion="3.2",osversion="2",le="6"} 9
histogram_bucket{appversion="3.2",osversion="2",le="7"} 9
histogram_bucket{appversion="3.2",osversion="2",le="8"} 9
histogram_bucket{appversion="3.2",osversion="2",le="9"} 9
histogram_bucket{appversion="3.2",osversion="2",le="10"} 9
histogram_bucket{appversion="3.2",osversion="2",le="+Inf"} 9
histogram_sum{appversion="3.2",osversion="2"} 7
histogram_count{appversion="3.2",osversion="2"} 2
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
	multilabelResultWithServerLabels = `# HELP counter A counter
# TYPE counter counter
counter{a="a",appversion="3.2",b="b",osversion="2"} 3
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
	labelFieldResultWithServerLabels = `# HELP ui_page_render_errors A counter
# TYPE ui_page_render_errors counter
ui_page_render_errors{appversion="3.2",osversion="2",path="/org/:orgId"} 1
ui_page_render_errors{appversion="3.2",osversion="2",path="/prom/:orgId"} 2
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
ui_external_lib_loaded{loaded="true",name="Intercom"} 2
ui_external_lib_loaded{loaded="true",name="ga"} 2
ui_external_lib_loaded{loaded="true",name="mixpanel"} 2
`
	gaugeOutputWithServerLabels = `# HELP ui_external_lib_loaded A gauge with entries in un-sorted order
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{appversion="3.2",loaded="true",name="Intercom",osversion="2"} 2
ui_external_lib_loaded{appversion="3.2",loaded="true",name="ga",osversion="2"} 2
ui_external_lib_loaded{appversion="3.2",loaded="true",name="mixpanel",osversion="2"} 2
`
	duplicateLabels = `
# HELP ui_external_lib_loaded Test with duplicate values
# TYPE ui_external_lib_loaded gauge
ui_external_lib_loaded{name="Munchkin",loaded="true"} 15171
ui_external_lib_loaded{name="Munchkin",loaded="true"} 1
`
	duplicateError = `Duplicate labels: {__name__="ui_external_lib_loaded", loaded="true", name="Munchkin"}`

	duplicateErrorWithServerLabels = `Duplicate labels: {__name__="ui_external_lib_loaded", appversion="3.2", loaded="true", name="Munchkin", osversion="2"}`

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
	reorderedLabelsResultWithServerLabels = `# HELP counter A counter
# TYPE counter counter
counter{a="a",appversion="3.2",b="b",osversion="2"} 3
`
)

func TestAggate(t *testing.T) {
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
		a := NewAggate()

		if err := a.ParseAndMerge(strings.NewReader(c.a), url.Values{}, ""); err != nil {
			if c.err1 == nil {
				t.Fatalf("Unexpected error: %s", err)
			} else if c.err1.Error() != err.Error() {
				t.Fatalf("Expected %s, got %s", c.err1, err)
			}
		}
		if err := a.ParseAndMerge(strings.NewReader(c.b), url.Values{}, ""); err != c.err2 {
			t.Fatalf("Expected %s, got %s", c.err2, err)
		}

		r := httptest.NewRequest("GET", "http://example.com/foo", nil)
		w := httptest.NewRecorder()
		a.Handler(w, r)

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

func TestAggateWithServerLabels(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want string
		err1 error
		err2 error
	}{
		{gaugeInput, gaugeInput, gaugeOutputWithServerLabels, nil, nil},
		{in1, in2, wantWithServerLabels, nil, nil},
		{multilabel1, multilabel2, multilabelResultWithServerLabels, nil, nil},
		{labelFields1, labelFields2, labelFieldResultWithServerLabels, nil, nil},
		{duplicateLabels, "", "", fmt.Errorf("%s", duplicateErrorWithServerLabels), nil},
		{reorderedLabels1, reorderedLabels2, reorderedLabelsResultWithServerLabels, nil, nil},
	} {
		a := NewAggate()

		query := url.Values{}
		query.Add("_label", "appversion:3.2")
		query.Add("_label", "osversion:2")
		if err := a.ParseAndMerge(strings.NewReader(c.a), query, "_label"); err != nil {
			if c.err1 == nil {
				t.Fatalf("Unexpected error: %s", err)
			} else if c.err1.Error() != err.Error() {
				t.Fatalf("Expected %s, got %s", c.err1, err)
			}
		}
		if err := a.ParseAndMerge(strings.NewReader(c.b), query, "_label"); err != c.err2 {
			t.Fatalf("Expected %s, got %s", c.err2, err)
		}

		r := httptest.NewRequest("GET", "http://example.com/foo?_label=appversion:3.2&_label=osversion:2", nil)
		w := httptest.NewRecorder()
		a.Handler(w, r)

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
