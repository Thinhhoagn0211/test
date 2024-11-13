package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	pubBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pubBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PublicKey), nil
}
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	privBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privBytes)
	if block == nil || (block.Type != "PRIVATE KEY" && block.Type != "RSA PRIVATE KEY") {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	// Kiểm tra và parse PKCS#1
	if block.Type == "RSA PRIVATE KEY" {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	// Nếu là PKCS#8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// Xác minh kiểu của khóa
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaKey, nil
}

// Load RSA public and private keys as before (same as your original loadPublicKey and loadPrivateKey functions)

func encryptFileChunked(inputPath, outputPath, pubKeyPath string, chunkSize int) error {
	pubKey, err := loadPublicKey(pubKeyPath)
	if err != nil {
		return err
	}

	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	buffer := make([]byte, chunkSize)
	for {
		n, err := inFile.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// Encrypt the chunk
		encryptedChunk, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, buffer[:n])
		if err != nil {
			return err
		}

		// Write encrypted chunk size followed by the chunk itself
		chunkSizeBytes := int32(len(encryptedChunk))
		if err := binary.Write(outFile, binary.LittleEndian, chunkSizeBytes); err != nil {
			return err
		}
		if _, err := outFile.Write(encryptedChunk); err != nil {
			return err
		}
	}
	return nil
}

func decryptFileChunked(inputPath, outputPath, privKeyPath string) error {
	privKey, err := loadPrivateKey(privKeyPath)
	if err != nil {
		return err
	}

	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	for {
		var chunkSize int32
		if err := binary.Read(inFile, binary.LittleEndian, &chunkSize); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		encryptedChunk := make([]byte, chunkSize)
		if _, err := io.ReadFull(inFile, encryptedChunk); err != nil {
			return err
		}

		// Decrypt the chunk
		decryptedChunk, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, encryptedChunk)
		if err != nil {
			return err
		}

		// Write the decrypted chunk to the output file
		if _, err := outFile.Write(decryptedChunk); err != nil {
			return err
		}
	}

	return nil
}

func measureEncryptionCPU(inputPath, outputPath, pubKeyPath string, chunkSize int) error {
	// Lấy % CPU trước khi bắt đầu mã hóa
	initialCPU, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}

	startTime := time.Now()

	// Thực hiện mã hóa với chế độ chunked
	err = encryptFileChunked(inputPath, outputPath, pubKeyPath, chunkSize)
	if err != nil {
		return err
	}

	// Đo % CPU sau khi mã hóa
	finalCPU, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("Encryption completed in: %v\n", elapsedTime)
	fmt.Printf("Initial CPU usage: %.2f%%\n", initialCPU[0])
	fmt.Printf("Final CPU usage: %.2f%%\n", finalCPU[0])
	fmt.Printf("CPU usage increase during encryption: %.2f%%\n", finalCPU[0]-initialCPU[0])

	return nil
}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run main.go <mode> <input> <output> <keyPath>")
		fmt.Println("mode: encrypt or decrypt")
		os.Exit(1)
	}

	mode := os.Args[1]
	input := os.Args[2]
	output := os.Args[3]
	keyPath := os.Args[4]

	chunkSize := 190 // Adjust chunk size depending on the RSA key length and encryption padding

	switch mode {
	case "encrypt":
		err := measureEncryptionCPU(input, output, keyPath, chunkSize)
		if err != nil {
			fmt.Println("Encryption error:", err)
			os.Exit(1)
		}
		fmt.Println("File encrypted successfully in chunks.")
	case "decrypt":
		err := decryptFileChunked(input, output, keyPath)
		if err != nil {
			fmt.Println("Decryption error:", err)
			os.Exit(1)
		}
		fmt.Println("File decrypted successfully in chunks.")
	default:
		fmt.Println("Invalid mode. Use 'encrypt' or 'decrypt'")
		os.Exit(1)
	}
}
