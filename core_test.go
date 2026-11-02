package core

import (
	"testing"
	"time"
)

func TestSummary(t *testing.T) {
	r, e := Summarize([]Sample{{true, time.Second, 10, 3}, {false, 3 * time.Second, 20, 4}})
	if e != nil || r.SuccessRate != 0.5 || r.MeanLatency != 2*time.Second || r.OutputTokens != 7 {
		t.Fatalf("%+v %v", r, e)
	}
	if _, e = Summarize(nil); e == nil {
		t.Fatal("empty accepted")
	}
	if _, e = Summarize([]Sample{{Latency: -1}}); e == nil {
		t.Fatal("negative accepted")
	}
}
