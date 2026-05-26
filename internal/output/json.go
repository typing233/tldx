package output

import (
	"encoding/json"
	"io"
	"tldx/internal/checker"
)

type JSONOutput struct {
	Available []string  `json:"available"`
	Stats     JSONStats `json:"stats"`
}

type JSONStats struct {
	Total       int   `json:"total"`
	Available   int   `json:"available"`
	Unavailable int   `json:"unavailable"`
	Errors      int   `json:"errors"`
	DurationMs  int64 `json:"duration_ms"`
}

func PrintJSON(w io.Writer, available []string, stats checker.Stats) error {
	out := JSONOutput{
		Available: available,
		Stats: JSONStats{
			Total:       stats.Total,
			Available:   stats.Available,
			Unavailable: stats.Unavailable,
			Errors:      stats.Errors,
			DurationMs:  stats.Duration.Milliseconds(),
		},
	}
	if out.Available == nil {
		out.Available = []string{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
