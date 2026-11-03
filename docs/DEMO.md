# Reproduce the local MVP

Requirements: Go 1.25+ and Bash. No API keys, database, network services or paid model calls are required. Run from the repository root:

```sh
./scripts/demo.sh
```

The script creates a fresh ignored `build/demo.*` directory and prints its path. Open `comparison.html` in a browser; `baseline.html` provides the single-run summary. JSONL inputs, JSON reports and the regression-gate result are saved alongside them. The script never overwrites a previous demo directory.

## What to show

1. The three-task dataset uses narrow backend questions with exact or whitespace-trimmed expected answers.
2. Baseline and candidate both pass 2/3 tasks. The candidate breaks the HTTP Created answer while fixing the FIFO answer.
3. The HTML comparison shows one regression and one improvement despite a zero change in aggregate accuracy.
4. The built regression gate returns exit 2. The demo script expects and verifies that code, then exits successfully.
5. Dataset digests, provider mode and run IDs are carried into reports. Changing dataset contents or case coverage causes comparisons to fail.

The adapter replays fixed responses. Token counts come from fixtures, and latency measures local replay execution. This demo verifies evaluation plumbing, not model quality. Outcomes and token counts are reproducible; run timings vary.

## Report commands

```sh
go run ./cmd/reliability -input /path/to/results.jsonl -format html > summary.html
go run ./cmd/compare -baseline /path/to/baseline.jsonl -candidate /path/to/candidate.jsonl -format html > comparison.html
```

JSON remains the default. HTML is standalone, uses no JavaScript or remote assets, and escapes untrusted labels. The comparison gate works with either format. An invalid format or incompatible input emits no report. Shell redirection may still create/truncate the target file; choose a new output path.

## Scope and limitations

The completed MVP is a local Go evaluation toolkit: validated JSONL ingestion, versioned tasks, grading, a provider interface with scripted replay, paired comparisons, optional CI gate and JSON/HTML output. No live provider adapter, cost model, statistical confidence intervals, web service, dashboard persistence or open-ended semantic grading is implemented.

Provider deadlines are cooperative; adapters must honor context cancellation. Failure text is redacted. Failed cases contribute to success-rate and latency totals. Unknown usage is excluded from token totals, and token comparisons require explicit known usage on every pair. Missing legacy usage provenance is not sufficient for token deltas.

Do not claim general model reliability, benchmark superiority or production scalability from these three fixtures. Live evaluation is a separate future milestone.

## Verification

`go vet ./...`, `go test -race ./...` and the demo run in CI. Tests cover malformed/missing records, duplicate IDs, mixed data, oversized inputs, overflow, grading, cancellation, timeouts, provider errors, paired regressions, report escaping and writer errors. HTML has no interactive controls; inspect it directly in a browser.
