package core

import (
	"fmt"
	"io"
	"sort"
)

type CaseChange struct {
	CaseID         string `json:"case_id"`
	Change         string `json:"change"`
	BaselineError  string `json:"baseline_error,omitempty"`
	CandidateError string `json:"candidate_error,omitempty"`
}
type ModelComparison struct {
	Model                    string       `json:"model"`
	Baseline                 ModelReport  `json:"baseline"`
	Candidate                ModelReport  `json:"candidate"`
	SuccessRateDelta         float64      `json:"success_rate_delta"`
	MeanLatencyDeltaMS       float64      `json:"mean_latency_delta_ms"`
	P95LatencyDeltaMS        int64        `json:"p95_latency_delta_ms"`
	InputTokensDelta         *int         `json:"input_tokens_delta"`
	OutputTokensDelta        *int         `json:"output_tokens_delta"`
	TokenComparisonAvailable bool         `json:"token_comparison_available"`
	Regressions              int          `json:"regressions"`
	Improvements             int          `json:"improvements"`
	Changes                  []CaseChange `json:"changes"`
}
type Comparison struct {
	SchemaVersion  int               `json:"schema_version"`
	Dataset        string            `json:"dataset"`
	DatasetSHA256  string            `json:"dataset_sha256"`
	ProviderMode   string            `json:"provider_mode"`
	BaselineRun    string            `json:"baseline_run"`
	CandidateRun   string            `json:"candidate_run"`
	HasRegressions bool              `json:"has_regressions"`
	Models         []ModelComparison `json:"models"`
}

// Compare requires identical model/case keys and explicit matching provenance.
// Outcome regressions are paired pass-to-fail transitions, not net pass counts.
func Compare(baseline, candidate io.Reader) (Comparison, error) {
	before, err := LoadResults(baseline)
	if err != nil {
		return Comparison{}, fmt.Errorf("baseline: %w", err)
	}
	after, err := LoadResults(candidate)
	if err != nil {
		return Comparison{}, fmt.Errorf("candidate: %w", err)
	}
	b, a := before[0], after[0]
	if b.RunID == a.RunID {
		return Comparison{}, fmt.Errorf("baseline and candidate must have distinct run IDs")
	}
	if b.DatasetSHA256 == "" || b.ProviderMode == "" || b.Dataset != a.Dataset || b.DatasetSHA256 != a.DatasetSHA256 || b.ProviderMode != a.ProviderMode {
		return Comparison{}, fmt.Errorf("matching dataset ID, nonempty digest and provider mode required")
	}
	if len(before) != len(after) {
		return Comparison{}, fmt.Errorf("model/case coverage differs")
	}
	key := func(r Record) string { return r.Model + "\x00" + r.CaseID }
	lookup := map[string]Record{}
	for _, r := range after {
		lookup[key(r)] = r
	}
	for _, r := range before {
		if _, ok := lookup[key(r)]; !ok {
			return Comparison{}, fmt.Errorf("model/case coverage differs")
		}
	}
	bs, err := summarizeRecords(before)
	if err != nil {
		return Comparison{}, fmt.Errorf("baseline: %w", err)
	}
	as, err := summarizeRecords(after)
	if err != nil {
		return Comparison{}, fmt.Errorf("candidate: %w", err)
	}
	result := Comparison{SchemaVersion: 1, Dataset: b.Dataset, DatasetSHA256: b.DatasetSHA256, ProviderMode: b.ProviderMode, BaselineRun: b.RunID, CandidateRun: a.RunID, Models: make([]ModelComparison, 0, len(bs.Models))}
	for i, bm := range bs.Models {
		am := as.Models[i]
		m := ModelComparison{Model: bm.Model, Baseline: bm, Candidate: am, SuccessRateDelta: am.SuccessRate - bm.SuccessRate, MeanLatencyDeltaMS: am.MeanLatencyMS - bm.MeanLatencyMS, P95LatencyDeltaMS: am.P95LatencyMS - bm.P95LatencyMS, Changes: []CaseChange{}, TokenComparisonAvailable: true}
		for _, old := range before {
			if old.Model != bm.Model {
				continue
			}
			next := lookup[key(old)]
			// Legacy records with no usage provenance are deliberately not comparable.
			if old.UsageKnown == nil || next.UsageKnown == nil || !*old.UsageKnown || !*next.UsageKnown {
				m.TokenComparisonAvailable = false
			}
			change := ""
			if *old.Success && !*next.Success {
				change = "regression"
				m.Regressions++
			} else if !*old.Success && *next.Success {
				change = "improvement"
				m.Improvements++
			} else if old.ErrorKind != next.ErrorKind {
				change = "failure_kind_changed"
			}
			if change != "" {
				m.Changes = append(m.Changes, CaseChange{old.CaseID, change, old.ErrorKind, next.ErrorKind})
			}
		}
		if m.TokenComparisonAvailable {
			input, output := am.InputTokens-bm.InputTokens, am.OutputTokens-bm.OutputTokens
			m.InputTokensDelta = &input
			m.OutputTokensDelta = &output
		}
		sort.Slice(m.Changes, func(i, j int) bool { return m.Changes[i].CaseID < m.Changes[j].CaseID })
		result.HasRegressions = result.HasRegressions || m.Regressions > 0
		result.Models = append(result.Models, m)
	}
	return result, nil
}
