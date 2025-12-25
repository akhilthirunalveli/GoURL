package service

import (
	"errors"
	"strings"
)

const (
	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base     = uint64(len(alphabet))
)

// Encode converts a uint64 ID to a Base62 string
func Encode(id uint64) string {
	if id == 0 {
		return string(alphabet[0])
	}

	var sb strings.Builder
	for id > 0 {
		sb.WriteByte(alphabet[id%base])
		id /= base
	}

	// Reverse the string
	runes := []rune(sb.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// Decode converts a Base62 string back to a uint64 ID
func Decode(encoded string) (uint64, error) {
	var id uint64

	for _, r := range encoded {
		idx := strings.IndexRune(alphabet, r)
		if idx == -1 {
			return 0, errors.New("invalid character in base62 string")
		}
		id = id*base + uint64(idx)
	}

	return id, nil
}
