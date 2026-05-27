package scanner

import "testing"

func TestBase64DetectLong(t *testing.T) {
	// 100+ base64 chars
	data := []byte("aaaa" + "dGVzdGluZ2Jhc2U2NAp0aGlzaXNhYmFzZTY0c3RyaW5ndGhhdGlzMTAwY2hhcmFjdGVyc2xvbmdhbmRzaG91bGRiZWRldGVjdGVkYXM=" + "zzzz")
	result := obfuscationCheck(data, 100)
	if !result.Block {
		t.Error("expected Block true for long base64 string")
	}
}

func TestBase64DetectShort(t *testing.T) {
	data := []byte(`some data with short base64: dGVzdA== that is it`)
	result := obfuscationCheck(data, 100)
	if result.Block {
		t.Error("expected Block false for short base64 strings")
	}
}

func TestHexDetectManyEscapes(t *testing.T) {
	data := []byte(`\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64\x21\x54\x65\x73\x74\x69\x6e\x67\x21\x21`)
	result := obfuscationCheck(data, 20)
	if !result.Block {
		t.Error("expected Block true for many hex escapes")
	}
}

func TestHexDetectFewEscapes(t *testing.T) {
	data := []byte(`\x68\x65\x6c\x6c\x6f`)
	result := obfuscationCheck(data, 20)
	if result.Block {
		t.Error("expected Block false for few hex escapes")
	}
}

func TestEvalDetect(t *testing.T) {
	tests := []struct {
		name  string
		input string
		block bool
	}{
		{"eval atob", `eval(atob("dGVzdA=="))`, true},
		{"eval base64 decode", `eval(base64.b64decode("dGVzdA=="))`, true},
		{"Function constructor", `new Function("return 'evil'")()`, true},
		{"exec code", `exec("import os; os.system('rm -rf /')")`, true},
		{"normal function call", `calculateValue(42)`, false},
		{"no eval", `print("hello world")`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := obfuscationCheck([]byte(tt.input), 100)
			if result.Block != tt.block {
				t.Errorf("expected Block=%v, got %v for %q", tt.block, result.Block, tt.input)
			}
		})
	}
}
