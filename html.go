package core

import (
	"bytes"
	"html/template"
	"io"
	"strconv"
)

const pageStart = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'"><title>AI Reliability Lab</title><style>
body{margin:0;background:#101b26;color:#e7eef4;font:16px system-ui,sans-serif}main{max-width:1100px;margin:auto;padding:40px 24px}header{border-bottom:1px solid #3c5264;padding-bottom:20px;color:#89cbd2;letter-spacing:2px;font-size:13px}h1{font-size:clamp(30px,5vw,48px);letter-spacing:-1px}h2{font-size:23px}p{line-height:1.6}section{background:#192b39;padding:24px;border:1px solid #3c5264;border-radius:12px;margin:24px 0}.meta,code{overflow-wrap:anywhere}.muted{color:#b3c3d1}.alert{border-left:4px solid #efab75;padding:12px 18px;background:#362c25}.table{overflow-x:auto}table{border-collapse:collapse;width:100%;font-size:14px}th,td{text-align:left;padding:13px 10px;border-bottom:1px solid #3c5264}th{color:#9ed7db}td{overflow-wrap:anywhere}footer{margin-top:30px;color:#b3c3d1;font-size:13px}@media print{body{background:white;color:black}section{background:white}th,.muted,footer{color:#333}}
</style></head><body><main><header>AI RELIABILITY LAB / EVALUATION REPORT</header>`
const pageEnd = `<footer>Local report · No external assets or scripts · Metrics describe supplied observations, not statistical significance.</footer></main></body></html>`
const summaryHTML = pageStart + `<h1>Run summary</h1><p class="meta">Run <strong>{{.RunID}}</strong> · Dataset <strong>{{.Dataset}}</strong></p><p class="meta">Provider mode: {{if .ProviderMode}}{{.ProviderMode}}{{else}}unspecified{{end}}<br>Dataset SHA-256: {{if .DatasetSHA256}}{{.DatasetSHA256}}{{else}}not supplied{{end}}</p>{{if eq .ProviderMode "scripted_replay"}}<p class="alert">Scripted replay: token counts are fixture values; latency measures local execution, not model inference.</p>{{end}}<section><h2>Model results</h2><div class="table"><table><thead><tr><th>Model</th><th>Passed / cases</th><th>Success</th><th>Mean ms</th><th>P95 ms</th><th>Input tokens</th><th>Output tokens</th><th>Unknown usage</th></tr></thead><tbody>{{range .Models}}<tr><td>{{.Model}}</td><td>{{.Passed}} / {{.Count}}</td><td>{{pct .SuccessRate}}%</td><td>{{num .MeanLatencyMS}}</td><td>{{.P95LatencyMS}}</td><td>{{.InputTokens}}</td><td>{{.OutputTokens}}</td><td>{{.UnknownUsageCount}}</td></tr>{{end}}</tbody></table></div></section><p class="muted">Latency and token totals include failures. With unknown usage, totals cover known usage only. Different model case coverage can make rates incomparable; this summary is not a ranking.</p>` + pageEnd
const comparisonHTML = pageStart + `<h1>{{if .HasRegressions}}Outcome regressions detected{{else}}No outcome regressions{{end}}</h1><p class="meta"><strong>{{.BaselineRun}}</strong> → <strong>{{.CandidateRun}}</strong><br>Dataset: {{.Dataset}}<br>Provider mode: {{.ProviderMode}}<br>Dataset SHA-256: {{.DatasetSHA256}}</p>{{if eq .ProviderMode "scripted_replay"}}<p class="alert">Scripted replay: supplied token counts and local execution latency are not live model measurements.</p>{{end}}{{range .Models}}<section><h2>{{.Model}}</h2><p>{{.Regressions}} regressions · {{.Improvements}} improvements · {{.Baseline.Count}} matched cases</p><div class="table"><table><thead><tr><th>Metric</th><th>Baseline</th><th>Candidate</th><th>Change</th></tr></thead><tbody><tr><td>Success rate</td><td>{{pct .Baseline.SuccessRate}}%</td><td>{{pct .Candidate.SuccessRate}}%</td><td>{{pct .SuccessRateDelta}} percentage points</td></tr><tr><td>Mean latency (ms)</td><td>{{num .Baseline.MeanLatencyMS}}</td><td>{{num .Candidate.MeanLatencyMS}}</td><td>{{num .MeanLatencyDeltaMS}}</td></tr><tr><td>P95 latency (ms)</td><td>{{.Baseline.P95LatencyMS}}</td><td>{{.Candidate.P95LatencyMS}}</td><td>{{.P95LatencyDeltaMS}}</td></tr><tr><td>Input tokens</td><td>{{.Baseline.InputTokens}}</td><td>{{.Candidate.InputTokens}}</td><td>{{if .TokenComparisonAvailable}}{{.InputTokensDelta}}{{else}}Unavailable{{end}}</td></tr><tr><td>Output tokens</td><td>{{.Baseline.OutputTokens}}</td><td>{{.Candidate.OutputTokens}}</td><td>{{if .TokenComparisonAvailable}}{{.OutputTokensDelta}}{{else}}Unavailable{{end}}</td></tr></tbody></table></div>{{if not .TokenComparisonAvailable}}<p class="muted">Token deltas unavailable: explicit known usage is required for every paired case.</p>{{end}}<h3>Changed cases</h3>{{if .Changes}}<div class="table"><table><thead><tr><th>Case</th><th>Change</th><th>Previous error</th><th>Current error</th></tr></thead><tbody>{{range .Changes}}<tr><td>{{.CaseID}}</td><td>{{.Change}}</td><td>{{.BaselineError}}</td><td>{{.CandidateError}}</td></tr>{{end}}</tbody></table></div>{{else}}<p>No outcome or failure-kind changes.</p>{{end}}</section>{{end}}<p class="muted">The regression gate considers pass-to-fail changes, even if other cases improve. Latency and token deltas are descriptive and do not trigger the gate.</p>` + pageEnd

func renderHTML(w io.Writer, source string, data any) error {
	t, err := template.New("report").Funcs(template.FuncMap{"pct": func(v float64) string { return format(v * 100) }, "num": format}).Parse(source)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	if err = t.Execute(&buffer, data); err != nil {
		return err
	}
	_, err = buffer.WriteTo(w)
	return err
}
func WriteReportHTML(w io.Writer, r Report) error         { return renderHTML(w, summaryHTML, r) }
func WriteComparisonHTML(w io.Writer, r Comparison) error { return renderHTML(w, comparisonHTML, r) }

func format(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }
