package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	core "github.com/akshay-dhangar/ai-reliability-lab"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCompareCLIAndGate(t *testing.T) {
	raw, err := os.Open("../../examples/tasks.json")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	d, err := core.LoadDataset(raw)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	paths := []string{}
	for i, name := range []string{"replay.json", "replay-candidate.json"} {
		file, err := os.Open("../../examples/" + name)
		if err != nil {
			t.Fatal(err)
		}
		p, err := core.LoadReplay(file, d)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		records, err := core.Evaluate(context.Background(), d, p, core.EvaluationOptions{RunID: []string{"baseline", "candidate"}[i], Model: "scripted", ProviderMode: "scripted_replay", Timeout: time.Second})
		if err != nil {
			t.Fatal(err)
		}
		var data bytes.Buffer
		for _, r := range records {
			if err = json.NewEncoder(&data).Encode(r); err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(dir, name+".jsonl")
		if err = os.WriteFile(path, data.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	args := []string{"-baseline", paths[0], "-candidate", paths[1]}
	var out, diag bytes.Buffer
	if err = run(args, &out, &diag); err != nil {
		t.Fatal(err)
	}
	var report core.Comparison
	if err = json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if !report.HasRegressions || report.Models[0].SuccessRateDelta != 0 {
		t.Fatal("lost paired regression")
	}
	out.Reset()
	if err = run(append(args, "-fail-on-regression"), &out, &diag); !errors.Is(err, errRegression) || out.Len() == 0 {
		t.Fatalf("gate: %v %s", err, out.String())
	}
	out.Reset()
	if err = run(append(args, "-format", "html", "-fail-on-regression"), &out, &diag); !errors.Is(err, errRegression) || !strings.Contains(out.String(), "Outcome regressions detected") {
		t.Fatal("HTML gate failed")
	}
	out.Reset()
	if err = run([]string{"-baseline", paths[0], "-candidate", paths[0]}, &out, &diag); err == nil || out.Len() != 0 {
		t.Fatal("invalid comparison emitted output")
	}
}
