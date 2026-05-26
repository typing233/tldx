package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"tldx/internal/preset"
)

func runPresets(_ []string) int {
	fmt.Fprintf(os.Stdout, "\nAvailable TLD Presets:\n\n")

	names := preset.Names()
	sort.Strings(names)

	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	all := preset.List()
	for _, name := range names {
		tlds := all[name]
		padding := strings.Repeat(" ", maxLen-len(name))
		fmt.Fprintf(os.Stdout, "  %s%s  %s\n", name, padding, strings.Join(tlds, ", "))
	}

	fmt.Fprintf(os.Stdout, "\nUsage: tldx --preset <name> -k <keywords>\n\n")
	return 0
}
