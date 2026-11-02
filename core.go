package core

import (
	"errors"
	"time"
)

type Sample struct {
	Success                   bool
	Latency                   time.Duration
	InputTokens, OutputTokens int
}
type Summary struct {
	Count, Passed, InputTokens, OutputTokens int
	SuccessRate                              float64
	MeanLatency                              time.Duration
}

func Summarize(samples []Sample) (Summary, error) {
	if len(samples) == 0 {
		return Summary{}, errors.New("samples cannot be empty")
	}
	r := Summary{Count: len(samples)}
	var total time.Duration
	for _, s := range samples {
		if s.Latency < 0 || s.InputTokens < 0 || s.OutputTokens < 0 {
			return Summary{}, errors.New("negative measurement")
		}
		if s.Success {
			r.Passed++
		}
		total += s.Latency
		r.InputTokens += s.InputTokens
		r.OutputTokens += s.OutputTokens
	}
	r.SuccessRate = float64(r.Passed) / float64(r.Count)
	r.MeanLatency = total / time.Duration(r.Count)
	return r, nil
}
