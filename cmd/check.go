package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"tldx/internal/checker"
	"tldx/internal/cli"
	"tldx/internal/domain"
	"tldx/internal/output"
	"tldx/internal/preset"
	"tldx/internal/rdap"
)

func runCheck(args []string) int {
	checkArgs, err := cli.ParseCheckArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'tldx --help' for usage.\n")
		return 1
	}

	if checkArgs.NoColor {
		output.ColorEnabled = false
	}

	tlds := checkArgs.TLDs
	if checkArgs.Preset != "" {
		presetTLDs, ok := preset.Get(checkArgs.Preset)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: unknown preset %q. Run 'tldx presets' to see available presets.\n", checkArgs.Preset)
			return 1
		}
		tlds = append(tlds, presetTLDs...)
	}
	if len(tlds) == 0 {
		tlds = []string{"com"}
	}

	tlds = dedup(tlds)

	candidates := domain.Generate(domain.GenerateConfig{
		Keywords:  checkArgs.Keywords,
		Prefixes:  checkArgs.Prefixes,
		Suffixes:  checkArgs.Suffixes,
		TLDs:      tlds,
		MaxLength: checkArgs.MaxLength,
	})

	if len(candidates) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no candidate domains generated (check max-length filter)\n")
		return 1
	}

	httpClient := &http.Client{Timeout: checkArgs.Timeout}

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 15*time.Second)
	bootstrap := rdap.NewBootstrap(bootstrapCtx, httpClient)
	bootstrapCancel()

	rdapClient := rdap.NewClient(httpClient, bootstrap, checkArgs.Retries)
	pool := checker.NewPool(rdapClient, checkArgs.Concurrency)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startTime := time.Now()

	if checkArgs.Format == "text" {
		output.PrintHeader(os.Stdout, len(candidates))
	}

	results := pool.Run(ctx, candidates)

	var stats checker.Stats
	var available []string
	foundCount := 0

	for r := range results {
		stats.Total++
		if r.Error != nil {
			stats.Errors++
		} else if r.Available {
			stats.Available++
			available = append(available, r.Domain)
			foundCount++
		} else {
			stats.Unavailable++
		}

		if checkArgs.Format == "text" {
			output.PrintResult(os.Stdout, r, checkArgs.AvailableOnly)
		}

		if checkArgs.Limit > 0 && foundCount >= checkArgs.Limit {
			cancel()
			for r := range results {
				stats.Total++
				if r.Error != nil {
					stats.Errors++
				} else if r.Available {
					stats.Available++
					available = append(available, r.Domain)
				} else {
					stats.Unavailable++
				}
			}
			break
		}
	}

	stats.Duration = time.Since(startTime)

	if checkArgs.Format == "json-array" {
		output.PrintJSONArray(os.Stdout, available)
		output.PrintStats(os.Stderr, stats)
	} else {
		output.PrintStats(os.Stdout, stats)
	}

	return 0
}

func dedup(items []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
