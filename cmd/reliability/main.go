package main

import (
	"encoding/json"
	"flag"
	"fmt"
	core "github.com/akshay-dhangar/ai-reliability-lab"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "reliability:", err)
		os.Exit(1)
	}
}
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("reliability", flag.ContinueOnError)
	flags.SetOutput(stderr)
	input := flags.String("input", "-", "JSONL results file, or - for stdin")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	reader := stdin
	if *input != "-" {
		file, err := os.Open(*input)
		if err != nil {
			return fmt.Errorf("open input: %w", err)
		}
		defer file.Close()
		reader = file
	}
	report, err := core.Analyze(reader)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
