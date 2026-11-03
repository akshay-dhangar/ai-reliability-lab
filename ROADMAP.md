# ai-reliability-lab: seven-day MVP

Goal: Evaluate model task success, latency and token usage with reproducible datasets.

Status: scoped local MVP completed 3 November 2026. Started 2 November; original target window 2–8 November. Live provider evaluation remains outside the completed scope.

- [x] Day 1: Load JSONL evaluation results and aggregate metrics.
- [x] Day 2: Define reproducible task cases and expected outcomes.
- [x] Day 3: Add provider-independent evaluation runner.
- [x] Day 4: Compare two runs and flag regressions.
- [x] Day 5: Export readable HTML and JSON reports.
- [x] Day 6: Test failures, timeouts and malformed cases.
- [x] Day 7: Document measured results and record an MVP demo.

## Definition of done

- A documented local command runs the scoped end-to-end example.
- Meaningful positive and failure-path tests pass locally and in CI.
- README distinguishes implemented features, mocks, measured results, and limitations.
- Example inputs and outputs are reproducible.
- A short demo script and evidence-based LinkedIn draft are saved locally.
- Real model quality is not claimed without a live evaluation.

Seven days is a target. Reduce scope or record unfinished work explicitly; never mark a project complete merely because its week ended.

Completion evidence: JSON and standalone HTML reports, a one-command scripted demo, CI regression gate, failure-path tests and a local LinkedIn draft. See docs/DEMO.md for exact scope and reproduction.
