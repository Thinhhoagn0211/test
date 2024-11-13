package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// loadPublicKey loads a public key from a PEM file
func loadPublicKey(pubKeyPath string) (crypto.PublicKey, error) {
	keyData, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid public key format")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	return publicKey, nil
}

// verifySignature verifies the digital signature of a file
func verifySignature(publicKey crypto.PublicKey, filePath, sigPath string) error {
	// Read the file
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Read the signature
	sigData, err := ioutil.ReadFile(sigPath)
	if err != nil {
		return fmt.Errorf("failed to read signature: %w", err)
	}

	// Hash the file data
	hasher := crypto.SHA256.New()
	hasher.Write(fileData)
	hashed := hasher.Sum(nil)

	// Verify the signature based on the key type
	switch key := publicKey.(type) {
	case *rsa.PublicKey:
		err = rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed, sigData)
	case *ecdsa.PublicKey:
		ok := ecdsa.VerifyASN1(key, hashed, sigData)
		if !ok {
			err = fmt.Errorf("ECDSA signature verification failed")
		}
	case ed25519.PublicKey:
		if !ed25519.Verify(key, hashed, sigData) {
			err = fmt.Errorf("Ed25519 signature verification failed")
		}
	default:
		err = fmt.Errorf("unsupported public key type")
	}

	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s <public-key.pem> <file-to-check> <signature.sig>", os.Args[0])
	}

	pubKeyPath := os.Args[1]
	filePath := os.Args[2]
	sigPath := os.Args[3]

	publicKey, err := loadPublicKey(pubKeyPath)
	if err != nil {
		log.Fatalf("Error loading public key: %v", err)
	}

	err = verifySignature(publicKey, filePath, sigPath)
	if err != nil {
		log.Fatalf("Signature verification failed: %v", err)
	}

	fmt.Println("Signature is valid!")
}
