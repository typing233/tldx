package preset

import "testing"

func TestGet(t *testing.T) {
	tlds, ok := Get("popular")
	if !ok {
		t.Fatal("expected 'popular' preset to exist")
	}
	if len(tlds) == 0 {
		t.Fatal("expected non-empty TLD list")
	}
	found := false
	for _, tld := range tlds {
		if tld == "com" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'com' in popular preset")
	}
}

func TestGetUnknown(t *testing.T) {
	_, ok := Get("nonexistent")
	if ok {
		t.Error("expected false for unknown preset")
	}
}

func TestList(t *testing.T) {
	all := List()
	if len(all) == 0 {
		t.Fatal("expected non-empty preset list")
	}
	if _, ok := all["tech"]; !ok {
		t.Error("expected 'tech' in preset list")
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) != len(Registry) {
		t.Errorf("expected %d names, got %d", len(Registry), len(names))
	}
}
