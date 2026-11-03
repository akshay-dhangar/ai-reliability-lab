# Progress

## 2 November 2026 — result ingestion

Implemented a dependency-free Go CLI reading a JSONL file or stdin and emitting a versioned JSON report. Per-model metrics include counts, success fraction, mean/p95 latency and token totals. Input validation rejects mixed runs/datasets, duplicate cases, missing measurements, out-of-range numbers and oversized input. The seed aggregator now rejects integer/duration overflow.

The checked-in synthetic example and expected report provide a reproducible CLI check. Tests also cover validation failures and ensure invalid input emits no partial report.

Next: define task cases and expected outputs so a runner can produce these records reproducibly. Provider execution, actual grading, comparisons and HTML reporting remain unimplemented. No real model was evaluated.

## 2 November 2026 — task definitions and runner

Added bounded versioned dataset loading, exact/trim-space grading, canonical dataset hashes, and a provider interface that excludes expected answers. The sequential runner records mismatches, redacted provider failures, cooperative timeouts and unknown usage. Parent cancellation aborts output. A scripted replay adapter validates complete coverage and produces JSONL compatible with the metrics CLI.

The three-case backend fixture has two passes and one intentional mismatch. Token counts are synthetic and latency measures replay execution. Tests cover grading, digest changes, malformed datasets, replay coverage, provider failure, deadline expiry, cancellation and the complete CLI-to-report pipeline. No live provider call was made.

Next: compare matched runs and flag regressions without comparing different dataset digests or incompatible case coverage.

## 2 November 2026 — paired comparisons

Added a comparison API and CLI requiring distinct runs, matching nonempty dataset digests/provider modes and identical model/case coverage. Reports preserve baseline/candidate summaries, metric deltas and sorted case-level changes. Token deltas are omitted without explicit known usage on every pair. An optional exit-2 gate detects any pass-to-fail transition, regardless of the net success-rate change.

The candidate replay fixture demonstrates one regression and one improvement with an unchanged 2/3 success rate. Tests cover this masking scenario, reordered input, incompatible coverage/provenance, unknown usage and CLI output/exit behavior. Latency/token deltas remain descriptive; no statistical significance or live-model performance is claimed.

Next: HTML report export, final failure-case review and a reproducible demo with a factual LinkedIn draft.

## 3 November 2026 — local MVP completion

Added standalone, script-free HTML summaries and comparisons using Go html/template with escaped labels and responsive tables. Both CLIs accept -format json|html; the regression gate preserves its exit-2 contract for HTML. Fixed unknown-usage records contributing external placeholder values to token totals.

The repeatable demo creates a new output directory, exports JSON/HTML and verifies the built gate. Tests cover HTML escaping, writer errors, format rejection, HTML gate behavior and unknown token totals. A factual LinkedIn draft is saved but not posted.

Completed scope: local evaluation and comparison toolkit with scripted replay. Live providers, semantic grading, statistical significance and production hosting remain unimplemented. See docs/DEMO.md.
