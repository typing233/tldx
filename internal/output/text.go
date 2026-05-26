package output

import (
	"fmt"
	"io"
	"os"
	"time"
	"tldx/internal/checker"
)

var ColorEnabled = true

func init() {
	if os.Getenv("NO_COLOR") != "" {
		ColorEnabled = false
	}
	if fi, err := os.Stdout.Stat(); err == nil {
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			ColorEnabled = false
		}
	}
}

func green(s string) string {
	if !ColorEnabled {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}

func red(s string) string {
	if !ColorEnabled {
		return s
	}
	return "\033[31m" + s + "\033[0m"
}

func yellow(s string) string {
	if !ColorEnabled {
		return s
	}
	return "\033[33m" + s + "\033[0m"
}

func bold(s string) string {
	if !ColorEnabled {
		return s
	}
	return "\033[1m" + s + "\033[0m"
}

func dim(s string) string {
	if !ColorEnabled {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

func PrintResult(w io.Writer, r checker.Result, availableOnly bool) {
	if r.Error != nil {
		if !availableOnly {
			fmt.Fprintf(w, "  %s %s\n", yellow("ERR"), dim(r.Domain+" — "+r.Error.Error()))
		}
		return
	}
	if r.Available {
		fmt.Fprintf(w, "  %s %s\n", green("✓"), bold(r.Domain))
	} else if !availableOnly {
		fmt.Fprintf(w, "  %s %s\n", red("✗"), dim(r.Domain))
	}
}

func PrintStats(w io.Writer, stats checker.Stats) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", bold("── Statistics ──"))
	fmt.Fprintf(w, "  Total checked:  %d\n", stats.Total)
	fmt.Fprintf(w, "  Available:      %s\n", green(fmt.Sprintf("%d", stats.Available)))
	fmt.Fprintf(w, "  Unavailable:    %s\n", red(fmt.Sprintf("%d", stats.Unavailable)))
	fmt.Fprintf(w, "  Errors:         %s\n", yellow(fmt.Sprintf("%d", stats.Errors)))
	fmt.Fprintf(w, "  Duration:       %s\n", stats.Duration.Round(time.Millisecond).String())
}

func PrintHeader(w io.Writer, total int) {
	fmt.Fprintf(w, "\n%s checking %d candidate domains...\n\n", bold("tldx"), total)
}
