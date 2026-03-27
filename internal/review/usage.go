package review

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// CallUsage captures token usage and cost for a single Claude CLI invocation.
type CallUsage struct {
	Phase       string  `json:"phase"`
	Model       string  `json:"model"`
	InputTokens int     `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	CacheReadTokens   int `json:"cache_read_tokens,omitempty"`
	CacheCreateTokens int `json:"cache_creation_tokens,omitempty"`
	CostUSD     float64 `json:"cost_usd"`
	DurationMS  int     `json:"duration_ms"`
}

// UsageReport is the top-level structure written to the usage JSON file.
type UsageReport struct {
	PR                string      `json:"pr"`
	Timestamp         string      `json:"timestamp"`
	Calls             []CallUsage `json:"calls"`
	TotalInputTokens  int         `json:"total_input_tokens"`
	TotalOutputTokens int         `json:"total_output_tokens"`
	TotalCostUSD      float64     `json:"total_cost_usd"`
}

// UsageTracker accumulates usage across multiple Claude calls (thread-safe).
type UsageTracker struct {
	mu    sync.Mutex
	calls []CallUsage
}

// Add records a single Claude call's usage.
func (u *UsageTracker) Add(call CallUsage) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.calls = append(u.calls, call)
}

// Report builds a UsageReport from accumulated calls.
func (u *UsageTracker) Report(repo string, prNumber int) *UsageReport {
	u.mu.Lock()
	defer u.mu.Unlock()

	r := &UsageReport{
		PR:        fmt.Sprintf("%s#%d", repo, prNumber),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Calls:     u.calls,
	}
	for _, c := range u.calls {
		r.TotalInputTokens += c.InputTokens
		r.TotalOutputTokens += c.OutputTokens
		r.TotalCostUSD += c.CostUSD
	}
	return r
}

// WriteUsageFile writes the usage report to disk as JSON.
// Path defaults to "codecanary-usage.json" but can be overridden via
// the CODECANARY_USAGE_FILE environment variable.
func WriteUsageFile(report *UsageReport) error {
	path := os.Getenv("CODECANARY_USAGE_FILE")
	if path == "" {
		path = "codecanary-usage.json"
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling usage report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing usage file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Usage report written to %s\n", path)
	return nil
}

// claudeJSONResponse represents the JSON output from `claude --print --output-format json`.
type claudeJSONResponse struct {
	Result     string  `json:"result"`
	IsError    bool    `json:"is_error"`
	CostUSD    float64 `json:"total_cost_usd"`
	DurationMS int     `json:"duration_ms"`
	Usage      struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
	ModelUsage map[string]struct {
		CostUSD float64 `json:"costUSD"`
	} `json:"modelUsage"`
}

// firstModel extracts the model name from the modelUsage map.
func (r *claudeJSONResponse) firstModel() string {
	for k := range r.ModelUsage {
		return k
	}
	return ""
}
