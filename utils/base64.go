package utils

import (
	"encoding/base64"
	"fmt"
)

func EncodeBase64(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded
}

func DecodeBase64(data string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return decoded, nil
}
