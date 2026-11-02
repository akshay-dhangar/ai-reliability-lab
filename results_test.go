package core

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"
)

const valid = `{"run_id":"r1","dataset":"d1","case_id":"c1","model":"m1","success":false,"latency_ms":100,"input_tokens":0,"output_tokens":2}`

func TestAnalyze(t *testing.T) {
	second := strings.ReplaceAll(valid, `"c1"`, `"c2"`)
	second = strings.ReplaceAll(second, `false`, `true`)
	second = strings.ReplaceAll(second, `:100`, `:300`)
	report, err := Analyze(strings.NewReader(valid + "\n\n" + second + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	m := report.Models[0]
	if m.Count != 2 || m.Passed != 1 || m.Failed != 1 || m.SuccessRate != .5 || m.MeanLatencyMS != 200 || m.P95LatencyMS != 300 || m.OutputTokens != 4 {
		t.Fatalf("%+v", m)
	}
}
func TestRejectInvalidResults(t *testing.T) {
	cases := []string{"", "null", "{}", valid + " {}", strings.Replace(valid, `"success":false,`, "", 1), strings.Replace(valid, `"latency_ms":100`, `"latency_ms":null`, 1), strings.Replace(valid, `"latency_ms":100`, `"latency_ms":-1`, 1), strings.Replace(valid, `"latency_ms":100`, `"latency_ms":1.5`, 1), strings.Replace(valid, `"success":false`, `"unexpected":false`, 1), valid + "\n" + valid, valid + "\n" + strings.Replace(valid, `"r1"`, `"r2"`, 1), valid + "\n" + strings.Replace(valid, `"d1"`, `"d2"`, 1)}
	for _, input := range cases {
		if _, err := Analyze(strings.NewReader(input)); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	if _, err := LoadResults(bytes.NewReader(bytes.Repeat([]byte(" "), MaxResultsBytes+1))); err == nil {
		t.Fatal("accepted oversized input")
	}
}
func TestSummaryOverflow(t *testing.T) {
	for _, samples := range [][]Sample{{{Latency: time.Duration(math.MaxInt64)}, {Latency: 1}}, {{InputTokens: math.MaxInt}, {InputTokens: 1}}, {{OutputTokens: math.MaxInt}, {OutputTokens: 1}}} {
		if _, err := Summarize(samples); err == nil {
			t.Fatal("accepted overflow")
		}
	}
}
func TestStableModelOrder(t *testing.T) {
	other := strings.Replace(valid, `"m1"`, `"a-model"`, 1)
	report, err := Analyze(strings.NewReader(valid + "\n" + other))
	if err != nil || report.Models[0].Model != "a-model" {
		t.Fatalf("%+v %v", report, err)
	}
}
