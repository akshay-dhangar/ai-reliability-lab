package core

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func comparisonInput(run string, passes []bool) string {
	var out bytes.Buffer
	for i, pass := range passes {
		lat := int64(100 + i*100)
		tokens := 10
		known := true
		r := Record{RunID: run, Dataset: "d", DatasetSHA256: strings.Repeat("a", 64), ProviderMode: "test", CaseID: []string{"a", "b"}[i], Model: "m", Success: &pass, LatencyMS: &lat, InputTokens: &tokens, OutputTokens: &tokens, UsageKnown: &known}
		_ = json.NewEncoder(&out).Encode(r)
	}
	return out.String()
}
func TestCompareCatchesRegressionDespiteEqualSuccessRate(t *testing.T) {
	b := comparisonInput("base", []bool{true, false})
	a := comparisonInput("next", []bool{false, true})
	report, err := Compare(strings.NewReader(b), strings.NewReader(a))
	if err != nil {
		t.Fatal(err)
	}
	m := report.Models[0]
	if !report.HasRegressions || m.Regressions != 1 || m.Improvements != 1 || m.SuccessRateDelta != 0 || len(m.Changes) != 2 || m.Changes[0].CaseID != "a" || !m.TokenComparisonAvailable || *m.InputTokensDelta != 0 {
		t.Fatalf("%+v", report)
	}
}
func TestCompareRejectsIncompatibleInputs(t *testing.T) {
	b := comparisonInput("base", []bool{true, false})
	a := comparisonInput("next", []bool{true, false})
	for _, candidate := range []string{b, strings.ReplaceAll(a, strings.Repeat("a", 64), strings.Repeat("b", 64)), strings.ReplaceAll(a, `"provider_mode":"test"`, `"provider_mode":"live"`), strings.ReplaceAll(a, `"case_id":"b"`, `"case_id":"c"`), strings.ReplaceAll(a, `"model":"m"`, `"model":"different"`), strings.Split(a, "\n")[0], strings.ReplaceAll(a, `"dataset_sha256":"`+strings.Repeat("a", 64)+`",`, "")} {
		if _, err := Compare(strings.NewReader(b), strings.NewReader(candidate)); err == nil {
			t.Fatal("accepted incompatible input")
		}
	}
}
func TestCompareUnknownUsageAndOrdering(t *testing.T) {
	b := comparisonInput("base", []bool{false, false})
	a := comparisonInput("next", []bool{true, true})
	a = strings.ReplaceAll(a, `"usage_known":true`, `"usage_known":false`)
	lines := strings.Split(strings.TrimSpace(a), "\n")
	a = lines[1] + "\n" + lines[0]
	r, err := Compare(strings.NewReader(b), strings.NewReader(a))
	if err != nil {
		t.Fatal(err)
	}
	m := r.Models[0]
	if r.HasRegressions || m.Improvements != 2 || m.TokenComparisonAvailable || m.InputTokensDelta != nil || m.OutputTokensDelta != nil {
		t.Fatalf("%+v", r)
	}
	legacy := strings.ReplaceAll(a, `"usage_known":false,`, "")
	r, err = Compare(strings.NewReader(b), strings.NewReader(legacy))
	if err != nil || r.Models[0].TokenComparisonAvailable {
		t.Fatal("missing provenance accepted for token delta")
	}
}
