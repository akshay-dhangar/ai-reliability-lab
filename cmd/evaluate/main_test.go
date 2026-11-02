package main

import (
	"bytes"
	"context"
	core "github.com/akshay-dhangar/ai-reliability-lab"
	"testing"
)

func TestReplayPipeline(t *testing.T) {
	var out, errout bytes.Buffer
	err := run(context.Background(), []string{"-dataset", "../../examples/tasks.json", "-replay", "../../examples/replay.json", "-run-id", "demo"}, &out, &errout)
	if err != nil {
		t.Fatal(err)
	}
	report, err := core.Analyze(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	m := report.Models[0]
	if m.Count != 3 || m.Passed != 2 || m.Failed != 1 || m.InputTokens != 39 || m.OutputTokens != 5 || report.ProviderMode != "scripted_replay" || len(report.DatasetSHA256) != 64 {
		t.Fatalf("%+v", report)
	}
}
func TestBadArguments(t *testing.T) {
	var out bytes.Buffer
	if run(context.Background(), nil, &out, &out) == nil {
		t.Fatal("missing arguments accepted")
	}
}
