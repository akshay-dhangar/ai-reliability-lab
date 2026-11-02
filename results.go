package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"
)

const MaxResultsBytes = 16 << 20
const MaxRecords = 10000

// Record contains observed measurements, not model outputs or evaluator logic.
type Record struct {
	RunID        string `json:"run_id"`
	Dataset      string `json:"dataset"`
	CaseID       string `json:"case_id"`
	Model        string `json:"model"`
	Success      *bool  `json:"success"`
	LatencyMS    *int64 `json:"latency_ms"`
	InputTokens  *int   `json:"input_tokens"`
	OutputTokens *int   `json:"output_tokens"`
}
type ModelReport struct {
	Model         string  `json:"model"`
	Count         int     `json:"count"`
	Passed        int     `json:"passed"`
	Failed        int     `json:"failed"`
	SuccessRate   float64 `json:"success_rate"`
	MeanLatencyMS float64 `json:"mean_latency_ms"`
	P95LatencyMS  int64   `json:"p95_latency_ms"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
}
type Report struct {
	SchemaVersion int           `json:"schema_version"`
	RunID         string        `json:"run_id"`
	Dataset       string        `json:"dataset"`
	Models        []ModelReport `json:"models"`
}

// LoadResults rejects mixed runs, missing measurements and duplicate model/case pairs.
// Input is bounded before decoding; errors identify lines without echoing input data.
func LoadResults(reader io.Reader) ([]Record, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxResultsBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read results: %w", err)
	}
	if len(data) > MaxResultsBytes {
		return nil, fmt.Errorf("results exceed %d bytes", MaxResultsBytes)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), 64<<10)
	var records []Record
	seen := map[string]bool{}
	line := 0
	for scanner.Scan() {
		line++
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}
		if len(records) == MaxRecords {
			return nil, fmt.Errorf("line %d: too many records", line)
		}
		var record Record
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&record); err != nil {
			return nil, fmt.Errorf("line %d: invalid result object", line)
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			return nil, fmt.Errorf("line %d: expected one JSON object", line)
		}
		if err = validateRecord(record); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if len(records) > 0 && (record.RunID != records[0].RunID || record.Dataset != records[0].Dataset) {
			return nil, fmt.Errorf("line %d: mixed run or dataset", line)
		}
		key := record.Model + "\x00" + record.CaseID
		if seen[key] {
			return nil, fmt.Errorf("line %d: duplicate model/case pair", line)
		}
		seen[key] = true
		records = append(records, record)
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("line %d: record exceeds line limit", line+1)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("results contain no records")
	}
	return records, nil
}
func validateRecord(r Record) error {
	for _, v := range []string{r.RunID, r.Dataset, r.CaseID, r.Model} {
		if strings.TrimSpace(v) == "" || len(v) > 256 || strings.ContainsAny(v, "\x00\r\n") {
			return fmt.Errorf("identifiers must be nonblank, single-line strings of at most 256 bytes")
		}
	}
	if r.Success == nil || r.LatencyMS == nil || r.InputTokens == nil || r.OutputTokens == nil {
		return fmt.Errorf("success, latency_ms, input_tokens and output_tokens are required")
	}
	if *r.LatencyMS < 0 || *r.LatencyMS > math.MaxInt64/int64(time.Millisecond) || *r.InputTokens < 0 || *r.OutputTokens < 0 {
		return fmt.Errorf("measurements are negative or out of range")
	}
	return nil
}
func Analyze(reader io.Reader) (Report, error) {
	records, err := LoadResults(reader)
	if err != nil {
		return Report{}, err
	}
	grouped := map[string][]Sample{}
	latencies := map[string][]int64{}
	for _, r := range records {
		grouped[r.Model] = append(grouped[r.Model], Sample{*r.Success, time.Duration(*r.LatencyMS) * time.Millisecond, *r.InputTokens, *r.OutputTokens})
		latencies[r.Model] = append(latencies[r.Model], *r.LatencyMS)
	}
	names := make([]string, 0, len(grouped))
	for name := range grouped {
		names = append(names, name)
	}
	sort.Strings(names)
	report := Report{SchemaVersion: 1, RunID: records[0].RunID, Dataset: records[0].Dataset, Models: make([]ModelReport, 0, len(names))}
	for _, name := range names {
		summary, e := Summarize(grouped[name])
		if e != nil {
			return Report{}, fmt.Errorf("aggregate measurements: %w", e)
		}
		values := latencies[name]
		sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		report.Models = append(report.Models, ModelReport{name, summary.Count, summary.Passed, summary.Count - summary.Passed, summary.SuccessRate, float64(summary.MeanLatency) / float64(time.Millisecond), values[(95*len(values)+99)/100-1], summary.InputTokens, summary.OutputTokens})
	}
	return report, nil
}
