# AI Reliability Lab

A Go toolkit for inspecting model evaluation results. **Milestone 1 is implemented: JSONL ingestion and per-model metrics. The full evaluation MVP is still in progress.**

## Run

Requires Go 1.25+. No provider credentials or external dependencies are needed.

```sh
go run ./cmd/reliability -input examples/scripted-results.jsonl
# Or stream measurements:
cat examples/scripted-results.jsonl | go run ./cmd/reliability
```

The output matches [expected-report.json](examples/expected-report.json). Example measurements are synthetic: scripted-a passes 1/2 cases with a 200 ms mean; scripted-b passes 2/2 with a 300 ms mean. These numbers demonstrate aggregation, not actual model quality or performance.

## Input contract

Each nonblank JSONL line is one result with these required fields:

```json
{"run_id":"run-1","dataset":"tasks-v1","case_id":"case-1","model":"model-a","success":true,"latency_ms":120,"input_tokens":40,"output_tokens":12}
```

- `success` is the evaluator's supplied boolean judgment; this command does not independently grade outputs.
- Latency is an integer in milliseconds; token counts are nonnegative integers. Zero is valid; missing and null measurements are rejected. Failed cases remain in latency and token totals.
- A file represents one run and dataset. Each model/case pair must be unique. Identifiers must be nonblank, single-line strings of at most 256 bytes.
- Unknown fields, malformed/trailing JSON, mixed runs/datasets, oversized data and overflowing totals are rejected. Errors identify the line without echoing raw records.
- Input is bounded to 16 MiB, 10,000 records and 64 KiB per line. Blank lines are ignored.

## Report semantics

JSON output has schema version 1 and models sorted by name. Success rate is a fraction from 0 to 1. Mean latency includes all cases; p95 uses the nearest-rank method (`ceil(0.95 × count)`) and is not interpolated. Token totals include failures. No cost is estimated without pricing data.

Models may have different case coverage; inspect counts and input cases before comparing their rates. This milestone does not claim paired comparisons, statistical significance, or representative model rankings. A later milestone will add matched-run comparisons.

Output goes to stdout only after the entire input validates. Diagnostics go to stderr and invalid input exits nonzero. To save a report, redirect stdout to a new file.

## Verify

```sh
go vet ./...
go test -race ./...
```

Tests cover the documented example, malformed/missing measurements, mixed runs, duplicate cases, bounded input, numeric overflow, deterministic ordering and CLI failure output.

Next: reproducible task definitions and evaluator expectations, followed by a provider-independent runner, regression comparisons and HTML reporting. See [ROADMAP.md](ROADMAP.md) and [PROGRESS.md](PROGRESS.md). No live provider calls, deployments, or model-quality claims have been made.
