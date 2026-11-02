# ai-reliability-lab: seven-day MVP

Goal: Evaluate model task success, latency and token usage with reproducible datasets.

Status: queued. The initial core component is implemented; the full MVP is not.

- [ ] Day 1: Load JSONL evaluation results and aggregate metrics.
- [ ] Day 2: Define reproducible task cases and expected outcomes.
- [ ] Day 3: Add provider-independent evaluation runner.
- [ ] Day 4: Compare two runs and flag regressions.
- [ ] Day 5: Export readable HTML and JSON reports.
- [ ] Day 6: Test failures, timeouts and malformed cases.
- [ ] Day 7: Document measured results and record an MVP demo.

## Definition of done

- A documented local command runs the scoped end-to-end example.
- Meaningful positive and failure-path tests pass locally and in CI.
- README distinguishes implemented features, mocks, measured results, and limitations.
- Example inputs and outputs are reproducible.
- A short demo script and evidence-based LinkedIn draft are saved locally.
- Real model quality is not claimed without a live evaluation.

Seven days is a target. Reduce scope or record unfinished work explicitly; never mark a project complete merely because its week ended.
