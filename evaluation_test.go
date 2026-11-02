package core

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type providerFunc func(context.Context, Request) (Completion, error)

func (f providerFunc) Complete(c context.Context, r Request) (Completion, error) { return f(c, r) }
func dataset() Dataset {
	expected := "201"
	return Dataset{Version: 1, ID: "d", Tasks: []Task{{ID: "c", Prompt: "HTTP Created code", Expected: &expected, Match: "exact"}}}
}
func options() EvaluationOptions {
	return EvaluationOptions{RunID: "r", Model: "fake", ProviderMode: "test", Timeout: time.Second}
}
func TestEvaluateGradingAndProvenance(t *testing.T) {
	d := dataset()
	p := providerFunc(func(_ context.Context, r Request) (Completion, error) {
		if r.CaseID != "c" || r.Prompt != d.Tasks[0].Prompt {
			t.Fatal(r)
		}
		return Completion{Text: " 201\n", InputTokens: 10, OutputTokens: 2, UsageKnown: true}, nil
	})
	first, e := Evaluate(context.Background(), d, p, options())
	if e != nil || *first[0].Success || first[0].ErrorKind != "mismatch" || len(first[0].DatasetSHA256) != 64 {
		t.Fatalf("%+v %v", first, e)
	}
	d.Tasks[0].Match = "trim_space"
	second, e := Evaluate(context.Background(), d, p, options())
	if e != nil || !*second[0].Success || second[0].DatasetSHA256 == first[0].DatasetSHA256 {
		t.Fatal("grading rule not reflected in digest")
	}
}
func TestProviderFailureAndTimeout(t *testing.T) {
	for _, kind := range []string{"provider_error", "timeout", "invalid_response"} {
		t.Run(kind, func(t *testing.T) {
			p := providerFunc(func(ctx context.Context, _ Request) (Completion, error) {
				switch kind {
				case "provider_error":
					return Completion{}, errors.New("secret upstream body")
				case "timeout":
					<-ctx.Done()
					return Completion{}, ctx.Err()
				default:
					return Completion{InputTokens: -1}, nil
				}
			})
			o := options()
			o.Timeout = time.Millisecond
			r, e := Evaluate(context.Background(), dataset(), p, o)
			if e != nil || r[0].ErrorKind != kind || *r[0].Success || *r[0].UsageKnown {
				t.Fatalf("%+v %v", r, e)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e := Evaluate(ctx, dataset(), providerFunc(func(context.Context, Request) (Completion, error) {
		t.Fatal("called after cancellation")
		return Completion{}, nil
	}), options())
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}
func TestDatasetAndReplayValidation(t *testing.T) {
	for _, raw := range []string{`{}`, `null`, `{"version":1,"id":"d","tasks":[]}`, `{"version":1,"id":"d","tasks":[{"id":"c","prompt":"p","match":"exact"}]}`} {
		if _, e := LoadDataset(strings.NewReader(raw)); e == nil {
			t.Fatal("invalid dataset accepted")
		}
	}
	d := dataset()
	d.Tasks = append(d.Tasks, d.Tasks[0])
	if d.Validate() == nil {
		t.Fatal("duplicate accepted")
	}
	if _, e := LoadReplay(strings.NewReader(`{"responses":{}}`), dataset()); e == nil {
		t.Fatal("missing replay accepted")
	}
	if _, e := LoadReplay(strings.NewReader(`{"responses":{"c":{"text":"201","input_tokens":0,"output_tokens":0}}}`), dataset()); e != nil {
		t.Fatal(e)
	}
}
