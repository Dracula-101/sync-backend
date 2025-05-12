package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func LoadPEMFileInto(path string) ([]byte, error) {
	parts := strings.Split(path, "/")
	filename := strings.Split(parts[len(parts)-1], ".")[0]

	base64Value := os.Getenv(fmt.Sprintf("%s_BASE64_PEM", strings.ToUpper(filename)))
	if base64Value != "" {
		data, err := base64.StdEncoding.DecodeString(base64Value)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func CopyWithProgress(dst io.Writer, src io.Reader, totalSize int64) (int64, error) {
	return io.Copy(dst, src)
}

func GetFileExtension(filePath string) string {
	parts := strings.Split(filePath, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}
