#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
# Create a new directory every time: never overwrite prior evidence.
mkdir -p build
output=$(mktemp -d "$PWD/build/demo.XXXXXX")
go run ./cmd/evaluate -dataset examples/tasks.json -replay examples/replay.json -run-id baseline > "$output/baseline.jsonl"
go run ./cmd/evaluate -dataset examples/tasks.json -replay examples/replay-candidate.json -run-id candidate > "$output/candidate.jsonl"
go run ./cmd/reliability -input "$output/baseline.jsonl" > "$output/baseline.json"
go run ./cmd/reliability -input "$output/baseline.jsonl" -format html > "$output/baseline.html"
go run ./cmd/compare -baseline "$output/baseline.jsonl" -candidate "$output/candidate.jsonl" > "$output/comparison.json"
go run ./cmd/compare -baseline "$output/baseline.jsonl" -candidate "$output/candidate.jsonl" -format html > "$output/comparison.html"
go build -o "$output/compare" ./cmd/compare
set +e
"$output/compare" -baseline "$output/baseline.jsonl" -candidate "$output/candidate.jsonl" -fail-on-regression > "$output/gated-comparison.json"
status=$?
set -e
if [ "$status" -ne 2 ]; then
  echo "Expected regression exit 2, received $status" >&2
  exit 1
fi
printf 'Scripted demo passed: paired regression detected despite equal overall accuracy.\nReports: %s\n' "$output"
