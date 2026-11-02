# AI Reliability Lab

A Go toolkit for inspecting model evaluation results. **Implemented: JSONL metrics, task definitions, a provider-independent runner with a scripted replay adapter, and paired-run comparisons. The full evaluation MVP is still in progress.**

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

Next: readable HTML reports and final MVP documentation. See [ROADMAP.md](ROADMAP.md) and [PROGRESS.md](PROGRESS.md). No live provider calls, deployments, or model-quality claims have been made.

## Run an evaluation (scripted replay)

```sh
go run ./cmd/evaluate -dataset examples/tasks.json -replay examples/replay.json -run-id demo-1 > results.jsonl
go run ./cmd/reliability -input results.jsonl
```

The example grades three backend questions: two pass and one intentionally fails. The 39 input/5 output tokens are fixture values, not measured provider usage. Latency measures local replay execution and varies between runs; it is not model inference latency. Reports preserve `provider_mode: scripted_replay` to make this distinction explicit.

Dataset version 1 contains an ID and unique tasks with `id`, `prompt`, `expected` and `match`. Supported match rules are `exact` (byte-for-byte string equality) and `trim_space` (trim leading/trailing Unicode whitespace on both sides). Empty expected strings are supported; missing expected values are rejected. Case, punctuation and interior whitespace remain significant. These graders suit narrow tasks, not open-ended answer quality.

Each result includes a SHA-256 digest of the decoded, canonically encoded dataset, including prompts, expectations, task order and grading rules. Mixed dataset digests or provider modes are rejected by ingestion. Use a new run ID for each evaluation; the CLI does not maintain a global ID registry.

The Go `Provider` interface receives only a task ID and prompt, never the expected answer. `Evaluate` runs sequentially with a per-case context deadline (`-timeout`, default 30s, maximum 5m). Adapters must honor cancellation: this is a cooperative timeout, not a process sandbox. Provider errors become redacted `provider_error`, `timeout` or `invalid_response` results; mismatched answers use `mismatch`. Parent cancellation aborts the CLI without emitting partial output. It performs no automatic retries.

`usage_known: false` means token counts are unknown and stored as zero placeholders. Reports count these records in `unknown_usage_count`; token totals then represent known usage only. Legacy input without the flag retains its supplied counts. Replay token counts are explicitly supplied fixture data. Raw responses and prompts are not copied into results, and upstream error text is not emitted.

The replay adapter validates complete task coverage before execution. A future real provider adapter can implement the same interface. Live providers and HTML reports remain unimplemented.

## Compare matched runs

```sh
go run ./cmd/evaluate -dataset examples/tasks.json -replay examples/replay.json -run-id baseline > baseline.jsonl
go run ./cmd/evaluate -dataset examples/tasks.json -replay examples/replay-candidate.json -run-id candidate > candidate.jsonl
go run ./cmd/compare -baseline baseline.jsonl -candidate candidate.jsonl
```

Both fixtures pass 2/3 tasks, but the candidate regresses on `http-created` and improves on `queue-order`. Comparing only aggregate success rates would hide this regression. Changes are paired by model and case ID, and output order is deterministic.

Comparison requires distinct run IDs, matching dataset IDs and nonempty dataset digests, matching nonempty provider modes, and identical model/case coverage. It rejects mismatched data instead of silently intersecting cases. Model labels must match: use a stable experiment label for successive configurations, and keep configuration details with your run artifacts. This is not a cross-model ranking command. The supplied digest is provenance metadata, not proof that an external producer evaluated the dataset honestly.

Reports include baseline/candidate summaries, candidate-minus-baseline success and latency deltas, per-case regressions/improvements and failure-kind changes. Token deltas are available only when every paired record explicitly has `usage_known: true`; otherwise they are JSON null with `token_comparison_available: false`. This does not make fixture token values into live measurements.

`has_regressions` means at least one case changed from pass to fail, even if other cases improved. Timing and token deltas are descriptive; they do not trigger this gate and are not evidence of statistical significance. Replay timing is local execution time, not model latency.

For CI, build the command and request an outcome gate:

```sh
go build -o /tmp/reliability-compare ./cmd/compare
/tmp/reliability-compare -baseline baseline.jsonl -candidate candidate.jsonl -fail-on-regression
```

Exit codes: 0 for a valid report (or no regressions with the gate), 1 for invalid input or an operational error, 2 for detected regressions when gated. A valid regression report is written before exit 2; invalid comparisons emit no report. `go run` wraps executable exit codes, so use the built binary when CI needs to distinguish 1 from 2.
