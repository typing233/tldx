package domain

import (
	"testing"
)

func TestGenerateBasic(t *testing.T) {
	cfg := GenerateConfig{
		Keywords: []string{"hello", "world"},
		TLDs:    []string{"com", "ai"},
	}
	results := Generate(cfg)
	expected := []string{"hello.com", "hello.ai", "world.com", "world.ai"}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d: %v", len(expected), len(results), results)
	}
	for i, r := range results {
		if r != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, r, expected[i])
		}
	}
}

func TestGenerateWithPrefixSuffix(t *testing.T) {
	cfg := GenerateConfig{
		Keywords: []string{"app"},
		Prefixes: []string{"get", "my"},
		Suffixes: []string{"hq", "io"},
		TLDs:    []string{"com"},
	}
	results := Generate(cfg)
	expected := []string{
		"getapphq.com", "getappio.com",
		"myapphq.com", "myappio.com",
	}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d: %v", len(expected), len(results), results)
	}
	for i, r := range results {
		if r != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, r, expected[i])
		}
	}
}

func TestGenerateMaxLength(t *testing.T) {
	cfg := GenerateConfig{
		Keywords:  []string{"longkeyword", "hi"},
		TLDs:     []string{"com"},
		MaxLength: 10,
	}
	results := Generate(cfg)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %v", len(results), results)
	}
	if results[0] != "hi.com" {
		t.Errorf("expected hi.com, got %s", results[0])
	}
}

func TestGenerateEmptyKeywords(t *testing.T) {
	cfg := GenerateConfig{
		Keywords: []string{},
		TLDs:    []string{"com"},
	}
	results := Generate(cfg)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestGenerateWithEmptyPrefix(t *testing.T) {
	cfg := GenerateConfig{
		Keywords: []string{"test"},
		Prefixes: []string{"", "get"},
		TLDs:    []string{"com"},
	}
	results := Generate(cfg)
	expected := []string{"test.com", "gettest.com"}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d: %v", len(expected), len(results), results)
	}
	for i, r := range results {
		if r != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, r, expected[i])
		}
	}
}

func TestGenerateDefaultTLD(t *testing.T) {
	cfg := GenerateConfig{
		Keywords: []string{"test"},
	}
	results := Generate(cfg)
	if len(results) != 1 || results[0] != "test.com" {
		t.Errorf("expected [test.com], got %v", results)
	}
}
