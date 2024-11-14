package service

import (
	"fmt"
	"hash"
	"io"
	"os"
)

// function to calculate the hash of a file with a specific hash function (MD5, SHA1, SHA256)
func calculateHash(filePath string, hashFunc func() hash.Hash) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := hashFunc()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	hashBytes := h.Sum(nil)
	return fmt.Sprintf("%x", hashBytes), nil
}
