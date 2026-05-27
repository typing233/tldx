package cli

import (
	"testing"
	"time"
)

func TestParseCheckArgsBasic(t *testing.T) {
	args, err := ParseCheckArgs([]string{"-k", "hello,world", "-t", "com,ai"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args.Keywords) != 2 || args.Keywords[0] != "hello" || args.Keywords[1] != "world" {
		t.Errorf("keywords = %v, want [hello world]", args.Keywords)
	}
	if len(args.TLDs) != 2 || args.TLDs[0] != "com" || args.TLDs[1] != "ai" {
		t.Errorf("TLDs = %v, want [com ai]", args.TLDs)
	}
}

func TestParseCheckArgsPositional(t *testing.T) {
	args, err := ParseCheckArgs([]string{"hello", "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args.Keywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(args.Keywords))
	}
}

func TestParseCheckArgsDefaults(t *testing.T) {
	args, err := ParseCheckArgs([]string{"-k", "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.Concurrency != 5 {
		t.Errorf("concurrency = %d, want 5", args.Concurrency)
	}
	if args.Timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", args.Timeout)
	}
	if args.Format != "text" {
		t.Errorf("format = %q, want text", args.Format)
	}
	if args.Retries != 1 {
		t.Errorf("retries = %d, want 1", args.Retries)
	}
}

func TestParseCheckArgsAllFlags(t *testing.T) {
	args, err := ParseCheckArgs([]string{
		"-k", "test",
		"-p", "get,my",
		"-s", "app,hq",
		"-t", "com,io",
		"--max-length", "15",
		"--limit", "3",
		"--concurrency", "10",
		"--timeout", "5s",
		"--retries", "2",
		"--format", "json-array",
		"--no-color",
		"--available-only",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args.Prefixes) != 2 {
		t.Errorf("prefixes = %v, want 2 items", args.Prefixes)
	}
	if len(args.Suffixes) != 2 {
		t.Errorf("suffixes = %v, want 2 items", args.Suffixes)
	}
	if args.MaxLength != 15 {
		t.Errorf("max-length = %d, want 15", args.MaxLength)
	}
	if args.Limit != 3 {
		t.Errorf("limit = %d, want 3", args.Limit)
	}
	if args.Concurrency != 10 {
		t.Errorf("concurrency = %d, want 10", args.Concurrency)
	}
	if args.Timeout != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", args.Timeout)
	}
	if args.Retries != 2 {
		t.Errorf("retries = %d, want 2", args.Retries)
	}
	if args.Format != "json-array" {
		t.Errorf("format = %q, want json-array", args.Format)
	}
	if !args.NoColor {
		t.Error("expected no-color to be true")
	}
	if !args.AvailableOnly {
		t.Error("expected available-only to be true")
	}
}

func TestParseCheckArgsNoKeywords(t *testing.T) {
	_, err := ParseCheckArgs([]string{"-t", "com"})
	if err == nil {
		t.Error("expected error for missing keywords")
	}
}

func TestParseCheckArgsBadFormat(t *testing.T) {
	_, err := ParseCheckArgs([]string{"-k", "test", "--format", "xml"})
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"a,,b", []string{"a", "b"}},
	}
	for _, tt := range tests {
		result := splitAndTrim(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitAndTrim(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i, r := range result {
			if r != tt.expected[i] {
				t.Errorf("splitAndTrim(%q)[%d] = %q, want %q", tt.input, i, r, tt.expected[i])
			}
		}
	}
}
