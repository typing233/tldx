package cli

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type CheckArgs struct {
	Keywords      []string
	Prefixes      []string
	Suffixes      []string
	TLDs          []string
	Preset        string
	MaxLength     int
	Limit         int
	Concurrency   int
	Timeout       time.Duration
	Retries       int
	Format        string
	NoColor       bool
	AvailableOnly bool
}

func ParseCheckArgs(args []string) (*CheckArgs, error) {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)

	var keywords, prefixes, suffixes, tlds, file string

	fs.StringVar(&keywords, "k", "", "Comma-separated keywords")
	fs.StringVar(&keywords, "keywords", "", "Comma-separated keywords")
	fs.StringVar(&file, "f", "", "File with keywords (one per line, use - for stdin)")
	fs.StringVar(&file, "file", "", "File with keywords (one per line, use - for stdin)")
	fs.StringVar(&prefixes, "p", "", "Comma-separated prefixes")
	fs.StringVar(&prefixes, "prefix", "", "Comma-separated prefixes")
	fs.StringVar(&suffixes, "s", "", "Comma-separated suffixes")
	fs.StringVar(&suffixes, "suffix", "", "Comma-separated suffixes")
	fs.StringVar(&tlds, "t", "", "Comma-separated TLDs (without dots)")
	fs.StringVar(&tlds, "tld", "", "Comma-separated TLDs (without dots)")

	result := &CheckArgs{
		Concurrency: 5,
		Timeout:     10 * time.Second,
		Retries:     1,
		Format:      "text",
	}

	fs.StringVar(&result.Preset, "preset", "", "Use a preset TLD set (e.g. popular, tech, startup)")
	fs.IntVar(&result.MaxLength, "max-length", 0, "Maximum domain name length (0 = no limit)")
	fs.IntVar(&result.Limit, "limit", 0, "Stop after finding N available domains (0 = no limit)")
	fs.IntVar(&result.Limit, "n", 0, "Stop after finding N available domains (0 = no limit)")
	fs.IntVar(&result.Concurrency, "concurrency", 5, "Number of concurrent workers")
	fs.IntVar(&result.Concurrency, "c", 5, "Number of concurrent workers")
	fs.DurationVar(&result.Timeout, "timeout", 10*time.Second, "HTTP request timeout")
	fs.IntVar(&result.Retries, "retries", 1, "Number of retries on transient errors")
	fs.StringVar(&result.Format, "format", "text", "Output format: text or json-array")
	fs.BoolVar(&result.NoColor, "no-color", false, "Disable colored output")
	fs.BoolVar(&result.AvailableOnly, "available-only", false, "Only show available domains")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tldx check [options] [keywords...]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		return nil, err
	}

	var kws []string

	if file != "" {
		fileKws, err := readKeywordsFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading keywords file: %w", err)
		}
		kws = append(kws, fileKws...)
	}

	if keywords != "" {
		kws = append(kws, splitAndTrim(keywords)...)
	}

	kws = append(kws, fs.Args()...)

	if len(kws) == 0 {
		if isStdinPipe() {
			stdinKws, err := readKeywordsFromReader(os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("reading stdin: %w", err)
			}
			kws = append(kws, stdinKws...)
		}
	}

	if len(kws) == 0 {
		return nil, fmt.Errorf("no keywords provided (use -k, -f, positional args, or pipe to stdin)")
	}

	result.Keywords = kws

	if prefixes != "" {
		result.Prefixes = splitAndTrim(prefixes)
	}
	if suffixes != "" {
		result.Suffixes = splitAndTrim(suffixes)
	}
	if tlds != "" {
		result.TLDs = splitAndTrim(tlds)
	}

	if result.Format != "text" && result.Format != "json-array" {
		return nil, fmt.Errorf("unsupported format %q (use text or json-array)", result.Format)
	}
	if result.Concurrency < 1 {
		result.Concurrency = 1
	}

	return result, nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func readKeywordsFromFile(path string) ([]string, error) {
	if path == "-" {
		return readKeywordsFromReader(os.Stdin)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readKeywordsFromReader(f)
}

func readKeywordsFromReader(r io.Reader) ([]string, error) {
	var keywords []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		keywords = append(keywords, line)
	}
	return keywords, scanner.Err()
}

func isStdinPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}
