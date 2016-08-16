package util

import (
	"testing"
)

func TestLookupGroup(t *testing.T) {
	testData := map[string]int{
		"dingjun": 1000,
		"root":    0,
	}

	for name, gid := range testData {
		g := LookupGroup(name)
		if g == nil {
			t.Errorf("lookup group failed for: %s\n", name)
			continue
		}
		if g.Gid != gid {
			t.Errorf("expected gid: %d, got: %d\n", gid, g.Gid)
		}
	}
}

func TestLookupGroupId(t *testing.T) {
	testData := map[int]string{
		0:    "root",
		1000: "dingjun",
	}
	for gid, name := range testData {
		g := LookupGroupID(gid)
		if g == nil {
			t.Errorf("lookup group id failed for: %d\n", gid)
			continue
		}
		if g.Name != name {
			t.Errorf("expected name: %s, got: %s\n", name, g.Name)
		}
	}
}
