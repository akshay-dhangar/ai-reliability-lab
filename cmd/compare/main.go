package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	core "github.com/akshay-dhangar/ai-reliability-lab"
	"io"
	"os"
)

var errRegression = errors.New("paired outcome regressions detected")

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "compare:", err)
		if errors.Is(err, errRegression) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
func run(args []string, out, diagnostics io.Writer) error {
	f := flag.NewFlagSet("compare", flag.ContinueOnError)
	f.SetOutput(diagnostics)
	format := f.String("format", "json", "output format: json or html")
	baseline := f.String("baseline", "", "baseline JSONL file")
	candidate := f.String("candidate", "", "candidate JSONL file")
	gate := f.Bool("fail-on-regression", false, "exit 2 after writing report if any case changes from pass to fail")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *baseline == "" || *candidate == "" {
		return fmt.Errorf("-baseline and -candidate are required; no positional arguments")
	}
	if *format != "json" && *format != "html" {
		return fmt.Errorf("format must be json or html")
	}
	b, err := os.Open(*baseline)
	if err != nil {
		return err
	}
	defer b.Close()
	c, err := os.Open(*candidate)
	if err != nil {
		return err
	}
	defer c.Close()
	report, err := core.Compare(b, c)
	if err != nil {
		return err
	}
	if *format == "html" {
		err = core.WriteComparisonHTML(out, report)
	} else {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(report)
	}
	if err != nil {
		return err
	}
	if *gate && report.HasRegressions {
		return errRegression
	}
	return nil
}
