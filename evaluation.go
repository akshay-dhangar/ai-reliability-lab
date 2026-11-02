package core

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type Task struct {
	ID       string  `json:"id"`
	Prompt   string  `json:"prompt"`
	Expected *string `json:"expected"`
	Match    string  `json:"match"`
}
type Dataset struct {
	Version int    `json:"version"`
	ID      string `json:"id"`
	Tasks   []Task `json:"tasks"`
}

func decodeBounded(r io.Reader, target any) error {
	data, err := io.ReadAll(io.LimitReader(r, MaxResultsBytes+1))
	if err != nil {
		return err
	}
	if len(data) > MaxResultsBytes {
		return errors.New("input exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil {
		return errors.New("invalid JSON document")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return errors.New("expected one JSON document")
	}
	return nil
}
func validID(s string) bool {
	return strings.TrimSpace(s) != "" && len(s) <= 256 && !strings.ContainsAny(s, "\x00\r\n")
}
func (d Dataset) Validate() error {
	if d.Version != 1 || !validID(d.ID) || len(d.Tasks) == 0 || len(d.Tasks) > MaxRecords {
		return errors.New("dataset requires version 1, an ID and 1-10000 tasks")
	}
	seen := map[string]bool{}
	for i, t := range d.Tasks {
		if !validID(t.ID) || seen[t.ID] || strings.TrimSpace(t.Prompt) == "" || len(t.Prompt) > 8192 || t.Expected == nil || len(*t.Expected) > 8192 || (t.Match != "exact" && t.Match != "trim_space") {
			return fmt.Errorf("task %d: invalid ID, prompt, expected output or match rule", i+1)
		}
		seen[t.ID] = true
	}
	return nil
}
func LoadDataset(r io.Reader) (Dataset, error) {
	var d Dataset
	if err := decodeBounded(r, &d); err != nil {
		return d, err
	}
	return d, d.Validate()
}

// Providers receive prompts and case IDs, never the evaluator's expected answer.
// Complete must honor context cancellation. Go cannot forcibly stop an arbitrary implementation.
type Request struct{ CaseID, Prompt string }
type Completion struct {
	Text                      string
	InputTokens, OutputTokens int
	UsageKnown                bool
}
type Provider interface {
	Complete(context.Context, Request) (Completion, error)
}
type EvaluationOptions struct {
	RunID, Model, ProviderMode string
	Timeout                    time.Duration
}

func Evaluate(ctx context.Context, d Dataset, p Provider, o EvaluationOptions) ([]Record, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	if p == nil || !validID(o.RunID) || !validID(o.Model) || !validID(o.ProviderMode) || o.Timeout <= 0 || o.Timeout > 5*time.Minute {
		return nil, errors.New("provider, identifiers and timeout (0-5 minutes) required")
	}
	canonical, _ := json.Marshal(d)
	sum := sha256.Sum256(canonical)
	digest := hex.EncodeToString(sum[:])
	records := make([]Record, 0, len(d.Tasks))
	for _, task := range d.Tasks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		callCtx, cancel := context.WithTimeout(ctx, o.Timeout)
		start := time.Now()
		result, err := p.Complete(callCtx, Request{task.ID, task.Prompt})
		elapsed := time.Since(start).Milliseconds()
		callErr := callCtx.Err()
		cancel()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		success := false
		kind := ""
		known := result.UsageKnown
		if errors.Is(callErr, context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			kind = "timeout"
		} else if err != nil {
			kind = "provider_error"
		} else if len(result.Text) > 65536 || result.InputTokens < 0 || result.OutputTokens < 0 {
			kind = "invalid_response"
		} else {
			actual, expected := result.Text, *task.Expected
			if task.Match == "trim_space" {
				actual = strings.TrimSpace(actual)
				expected = strings.TrimSpace(expected)
			}
			success = actual == expected
			if !success {
				kind = "mismatch"
			}
		}
		if kind != "" && kind != "mismatch" {
			known = false
			result.InputTokens = 0
			result.OutputTokens = 0
		}
		if !known {
			result.InputTokens = 0
			result.OutputTokens = 0
		}
		r := Record{RunID: o.RunID, Dataset: d.ID, CaseID: task.ID, Model: o.Model, Success: &success, LatencyMS: &elapsed, InputTokens: &result.InputTokens, OutputTokens: &result.OutputTokens, ErrorKind: kind, UsageKnown: &known, DatasetSHA256: digest, ProviderMode: o.ProviderMode}
		if err := validateRecord(r); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

type ReplayEntry struct {
	Text         *string `json:"text"`
	InputTokens  *int    `json:"input_tokens"`
	OutputTokens *int    `json:"output_tokens"`
}
type Replay struct {
	Responses map[string]ReplayEntry `json:"responses"`
}

func LoadReplay(r io.Reader, d Dataset) (Replay, error) {
	var replay Replay
	if err := decodeBounded(r, &replay); err != nil {
		return replay, err
	}
	if len(replay.Responses) != len(d.Tasks) {
		return replay, errors.New("replay must cover exactly the dataset tasks")
	}
	for _, task := range d.Tasks {
		e, ok := replay.Responses[task.ID]
		if !ok || e.Text == nil || len(*e.Text) > 65536 || e.InputTokens == nil || e.OutputTokens == nil || *e.InputTokens < 0 || *e.OutputTokens < 0 {
			return replay, errors.New("invalid or missing replay response")
		}
	}
	return replay, nil
}
func (r Replay) Complete(ctx context.Context, request Request) (Completion, error) {
	if err := ctx.Err(); err != nil {
		return Completion{}, err
	}
	e, ok := r.Responses[request.CaseID]
	if !ok || e.Text == nil || e.InputTokens == nil || e.OutputTokens == nil {
		return Completion{}, errors.New("missing replay response")
	}
	return Completion{*e.Text, *e.InputTokens, *e.OutputTokens, true}, nil
}
