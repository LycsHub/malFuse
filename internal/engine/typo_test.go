package engine

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"kitten", "sitting", 3},
		{"", "", 0},
		{"abc", "abc", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"requets", "requests", 1},
		{"lodaash", "lodash", 1},
		{"requests", "requests", 0},
		{"flask", "flask", 0},
		{"a", "b", 1},
		{"ab", "ba", 2},
	}
	for _, tt := range tests {
		result := levenshteinDistance(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestTypoCheckBlocksCloseMatch(t *testing.T) {
	popular := []string{"requests", "flask", "django", "numpy", "lodash"}
	check := TypoCheck(popular, 2)

	result := check(nil, Request{Name: "requets"})
	if !result.Block {
		t.Error("expected Block true for typo-squatting 'requets' (close to 'requests')")
	}
	if result.Reason != "typo-squatting" {
		t.Errorf("expected Reason typo-squatting, got %s", result.Reason)
	}
}

func TestTypoCheckPassesExactMatch(t *testing.T) {
	popular := []string{"requests", "flask"}
	check := TypoCheck(popular, 2)

	result := check(nil, Request{Name: "requests"})
	if result.Block {
		t.Error("expected Block false for exact match 'requests'")
	}
}

func TestTypoCheckPassesFarMatch(t *testing.T) {
	popular := []string{"requests", "flask"}
	check := TypoCheck(popular, 2)

	result := check(nil, Request{Name: "my-obscure-lib"})
	if result.Block {
		t.Error("expected Block false for far match")
	}
}

func TestTypoCheckSkipsShortNames(t *testing.T) {
	popular := []string{"ab", "requests"}
	check := TypoCheck(popular, 2)

	result := check(nil, Request{Name: "xy"})
	if result.Block {
		t.Error("expected Block false for short name (skipped)")
	}
}

func TestTypoCheckSkipsEmptyList(t *testing.T) {
	check := TypoCheck(nil, 2)

	result := check(nil, Request{Name: "anything"})
	if result.Block {
		t.Error("expected Block false with empty popular list")
	}
}

func TestTypoCheckHigherThreshold(t *testing.T) {
	popular := []string{"requests"}
	check := TypoCheck(popular, 3)

	result := check(nil, Request{Name: "requeets"})
	if !result.Block {
		t.Error("expected Block true for distance 2 with threshold 3")
	}
}
