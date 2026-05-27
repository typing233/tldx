package output

import (
	"encoding/json"
	"io"
)

func PrintJSONArray(w io.Writer, available []string) error {
	if available == nil {
		available = []string{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(available)
}
