package core

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestHTMLReportsEscapeUntrustedText(t *testing.T) {
	attack := `<script>alert("x")</script>`
	var out bytes.Buffer
	r := Report{RunID: attack, Dataset: "demo", ProviderMode: "scripted_replay", Models: []ModelReport{{Model: attack, Count: 2, Passed: 1, SuccessRate: .5}}}
	if err := WriteReportHTML(&out, r); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "<script>") || !strings.Contains(out.String(), "&lt;script&gt;") || !strings.Contains(out.String(), "50.00%") || !strings.Contains(out.String(), "fixture values") {
		t.Fatal(out.String())
	}
	out.Reset()
	comparison, err := Compare(strings.NewReader(comparisonInput("b", []bool{true, false})), strings.NewReader(comparisonInput("a", []bool{false, true})))
	if err != nil {
		t.Fatal(err)
	}
	comparison.Models[0].Changes[0].CaseID = attack
	if err = WriteComparisonHTML(&out, comparison); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Outcome regressions detected", "1 regressions", "1 improvements", "0.00 percentage points", "&lt;script&gt;"} {
		if !strings.Contains(out.String(), text) {
			t.Fatalf("missing %s", text)
		}
	}
	if strings.Contains(out.String(), "<script>") {
		t.Fatal("unescaped markup")
	}
	if WriteReportHTML(brokenWriter{}, r) == nil || WriteComparisonHTML(brokenWriter{}, comparison) == nil {
		t.Fatal("write failure ignored")
	}
}
func TestUnknownUsageExcludedFromTotals(t *testing.T) {
	input := strings.ReplaceAll(comparisonInput("r", []bool{true}), `"usage_known":true`, `"usage_known":false`)
	r, err := Analyze(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if r.Models[0].InputTokens != 0 || r.Models[0].OutputTokens != 0 || r.Models[0].UnknownUsageCount != 1 {
		t.Fatalf("%+v", r)
	}
}
