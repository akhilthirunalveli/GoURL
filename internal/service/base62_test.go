package service

import (
	"testing"
)

func TestBase62(t *testing.T) {
	tests := []struct {
		id       uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "A"},
		{61, "z"},
		{62, "10"},
		{12345, "3D7"},
	}

	for _, tt := range tests {
		encoded := Encode(tt.id)
		if encoded != tt.expected {
			t.Errorf("Encode(%d): expected %s, got %s", tt.id, tt.expected, encoded)
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Decode(%s) returned error: %v", encoded, err)
		}
		if decoded != tt.id {
			t.Errorf("Decode(%s): expected %d, got %d", encoded, tt.id, decoded)
		}
	}
}

func TestDecodeError(t *testing.T) {
	_, err := Decode("invalid_char$")
	if err == nil {
		t.Error("Expected error for invalid character, got nil")
	}
}
