package hvutil

import (
	"crypto/rand"
	"encoding/hex"
)

func UniqueID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	const alphanum = "abcdefghijklmnopqrstuvwxyz0123456789"
	for i, byt := range b {
		b[i] = alphanum[int(byt)%len(alphanum)]
	}

	return string(b), nil
}

func GenerateMAC() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// unicast flag
	b[0] = (b[0] | 2) & 0xfe

	return hex.EncodeToString(b), nil
}
