package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	core "github.com/akshay-dhangar/ai-reliability-lab"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "evaluate:", err)
		os.Exit(1)
	}
}
func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	f := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	f.SetOutput(stderr)
	dataset := f.String("dataset", "", "task dataset JSON")
	replay := f.String("replay", "", "scripted responses JSON (no live calls)")
	id := f.String("run-id", "", "unique evaluation run ID")
	model := f.String("model", "scripted-replay", "label for recorded responses")
	timeout := f.Duration("timeout", 30*time.Second, "per-case cooperative deadline")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *dataset == "" || *replay == "" || *id == "" {
		return fmt.Errorf("-dataset, -replay and -run-id are required; no positional arguments")
	}
	df, err := os.Open(*dataset)
	if err != nil {
		return err
	}
	defer df.Close()
	d, err := core.LoadDataset(df)
	if err != nil {
		return err
	}
	rf, err := os.Open(*replay)
	if err != nil {
		return err
	}
	defer rf.Close()
	provider, err := core.LoadReplay(rf, d)
	if err != nil {
		return err
	}
	records, err := core.Evaluate(ctx, d, provider, core.EvaluationOptions{RunID: *id, Model: *model, ProviderMode: "scripted_replay", Timeout: *timeout})
	if err != nil {
		return err
	}
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	for _, record := range records {
		if err = encoder.Encode(record); err != nil {
			return err
		}
	}
	// Check compatibility and aggregate overflow before emitting any result.
	if _, err = core.Analyze(bytes.NewReader(output.Bytes())); err != nil {
		return err
	}
	_, err = io.Copy(stdout, &output)
	return err
}
