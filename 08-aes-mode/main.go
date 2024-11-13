package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

const cipherKey = "eaba2d73e2474eed9c3352f5aefe50e3"

func main() {
	var mode string
	flag.Func("mode", "choose mode to encrypt or decrypt", func(s string) error {
		if s == "encrypt" || s == "decrypt" {
			mode = s
			return nil
		} else {
			return fmt.Errorf("unsupported mode")
		}
	})
	inputPath := flag.String("input", "", "")
	outputPath := flag.String("output", "", "")
	flag.Parse()

	switch mode {
	case "encrypt":
		err := measureEncryption(*inputPath, *outputPath, cipherKey)
		if err != nil {
			log.Fatal(err)
		}
	case "decrypt":
		err := decryptCBC(cipherKey, *inputPath, *outputPath)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func measureEncryption(inputPath, outputPath, cipherKey string) error {
	// Measure initial CPU usage
	initialCPU, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}

	// Measure encryption start time
	startTime := time.Now()

	// Perform encryption
	err = encryptCBC(cipherKey, inputPath, outputPath)
	if err != nil {
		return err
	}

	// Measure final CPU usage and encryption duration
	finalCPU, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}
	elapsedTime := time.Since(startTime)

	// Output the results
	fmt.Printf("Encryption completed in: %v\n", elapsedTime)
	fmt.Printf("Initial CPU usage: %.2f%%\n", initialCPU[0])
	fmt.Printf("Final CPU usage: %.2f%%\n", finalCPU[0])
	fmt.Printf("CPU usage increase during encryption: %.2f%%\n", finalCPU[0]-initialCPU[0])

	// Calculate encryption speed (bytes per second)
	fileSize, err := getFileSize(inputPath)
	if err != nil {
		return err
	}
	speed := float64(fileSize) / elapsedTime.Seconds()
	fmt.Printf("Encryption speed: %.2f bytes per second\n", speed)

	return nil
}

func encryptCBC(cipherKey string, inputPath string, outputPath string) error {
	key := []byte(cipherKey)

	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read file")
	}

	plainText := content
	plainText, err = pkcs7pad(content, aes.BlockSize)
	if err != nil {
		return fmt.Errorf(`plainText: "%s" has error`, plainText)
	}

	if len(plainText)%aes.BlockSize != 0 {
		err := fmt.Errorf(`plainText: "%s" has the wrong block size`, plainText)
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

	if err := os.WriteFile(outputPath, cipherText, 0644); err != nil {
		return fmt.Errorf("cannot write encrypt file into %s", outputPath)
	}
	return nil
}

func decryptCBC(cipherKey string, inputPath string, outputPath string) error {
	key := []byte(cipherKey)

	cipherText, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read file: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	if len(cipherText) < aes.BlockSize {
		return errors.New("ciphertext too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	plainText, err := pkcs7strip(cipherText, aes.BlockSize)
	if err != nil {
		return fmt.Errorf("error during unpadding: %v", err)
	}

	if err := os.WriteFile(outputPath, plainText, 0644); err != nil {
		return fmt.Errorf("cannot write decrypted file into %s: %v", outputPath, err)
	}
	return nil
}

// pkcs7strip remove pkcs7 padding
func pkcs7strip(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: Data is not block-aligned")
	}
	padLen := int(data[length-1])
	ref := bytes.Repeat([]byte{byte(padLen)}, padLen)
	if padLen > blockSize || padLen == 0 || !bytes.HasSuffix(data, ref) {
		return nil, errors.New("pkcs7: Invalid padding")
	}
	return data[:length-padLen], nil
}

// pkcs7pad add pkcs7 padding
func pkcs7pad(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 1 || blockSize >= 256 {
		return nil, fmt.Errorf("pkcs7: Invalid block size %d", blockSize)
	} else {
		padLen := blockSize - len(data)%blockSize
		padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
		return append(data, padding...), nil
	}
}

func getFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}
