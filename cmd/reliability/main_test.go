package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCLIExample(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{"-input", "../../examples/scripted-results.jsonl"}, nil, &output, &output); err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile("../../examples/expected-report.json")
	if err != nil {
		t.Fatal(err)
	}
	if output.String() != string(expected) {
		t.Fatalf("unexpected report:\n%s", output.String())
	}
}
func TestCLIInvalidDoesNotEmitPartialReport(t *testing.T) {
	var output, diagnostics bytes.Buffer
	if err := run(nil, strings.NewReader("{}"), &output, &diagnostics); err == nil || output.Len() != 0 {
		t.Fatalf("err=%v output=%s", err, output.String())
	}
}
