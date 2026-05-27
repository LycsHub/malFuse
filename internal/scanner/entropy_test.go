package scanner

import (
	"testing"
)

func TestShannonEntropyEmpty(t *testing.T) {
	e := shannonEntropy([]byte{})
	if e != 0.0 {
		t.Errorf("expected 0.0 for empty input, got %f", e)
	}
}

func TestShannonEntropySingleByte(t *testing.T) {
	e := shannonEntropy([]byte("aaaaaaaaaa"))
	if e != 0.0 {
		t.Errorf("expected 0.0 for all same byte, got %f", e)
	}
}

func TestShannonEntropyUniform(t *testing.T) {
	// All 256 byte values once = maximum entropy
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}
	e := shannonEntropy(data)
	if e < 7.9 || e > 8.1 {
		t.Errorf("expected entropy near 8.0 for uniform 256 values, got %f", e)
	}
}

func TestShannonEntropyNormalText(t *testing.T) {
	data := []byte("This is normal English text for testing purposes.")
	e := shannonEntropy(data)
	if e > 5.0 {
		t.Errorf("expected entropy < 5.0 for normal English text, got %f", e)
	}
}

func TestEntropyCheckBlocksHighEntropy(t *testing.T) {
	// Random-like data should have high entropy
	data := make([]byte, 1000)
	for i := range data {
		data[i] = byte(i * 37 % 256)
	}
	result := entropyCheck(data, 4.5)
	if !result.Block {
		t.Error("expected Block true for high entropy data")
	}
	if result.Reason != "entropy" {
		t.Errorf("expected Reason entropy, got %s", result.Reason)
	}
}

func TestEntropyCheckPassesLowEntropy(t *testing.T) {
	data := []byte("plain text with normal distribution of characters")
	result := entropyCheck(data, 4.5)
	if result.Block {
		t.Error("expected Block false for normal text")
	}
}
