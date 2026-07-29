package keygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
)

const (
	MessageSize int = 8192
)

// GenerateRSAKeys generates a pair of private and public RSA keys of the given size.
// Minimum size is 1024 bits. Use 4096 for good measure.
func GenerateRSAKeys(numBits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// rand.Reader is a CSPRNG (Cryptographically secure pseudo-random number generator)
	privateKey, err := rsa.GenerateKey(rand.Reader, numBits)
	if err != nil {
		return nil, nil, errors.New("GenerateRSAKeys: Error generating private key: " + err.Error())
	}

	return privateKey, &privateKey.PublicKey, nil
}

// ExportPrivateRSAKey writes the private RSA key to the provided file path.
func ExportPrivateRSAKey(key *rsa.PrivateKey, dir string, filename string) error {
	// write PEM file content
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key), // converts PrivateKey struct into []byte using PKCS#1 structure
	}

	// create file to write the key to
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.New("ExportPrivateRSAKey: Error creating directory: " + err.Error())
	}
	privateFile, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return errors.New("ExportPrivateRSAKey: Error creating file: " + err.Error())
	}
	defer privateFile.Close()

	// write the contents to the new file using base64 encoding
	err = pem.Encode(privateFile, privateBlock)
	if err != nil {
		return errors.New("ExportPrivateRSAKey: Error writing key to the file: " + err.Error())
	}

	return nil
}

// ExportPrivateKey writes the public RSA key to the provided file path.
func ExportPublicRSAKey(key *rsa.PublicKey, dir string, filename string) error {
	// write PEM file content
	publicBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key), // converts PrivateKey struct into []byte using PKCS#1 structure
	}

	// create file to write the key to
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.New("ExportPublicRSAKey: Error creating directory: " + err.Error())
	}
	publicFile, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return errors.New("ExportPublicRSAKey: Error creating file: " + err.Error())
	}
	defer publicFile.Close()

	// write the contents to the new file using base64 encoding
	err = pem.Encode(publicFile, publicBlock)
	if err != nil {
		return errors.New("ExportPublicRSAKey: Error writing key to the file: " + err.Error())
	}

	return nil
}

// LoadPrivateRSAKey loads a rsa.PrivateKey from a given .pem file
func LoadPrivateRSAKey(path string) (*rsa.PrivateKey, error) {
	// load bytes from file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("LoadPrivateRSAKey: Error loading key from file: " + err.Error())
	}

	// extract block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("LoadPrivateRSAKey: Error decoding PEM block.")
	}

	// extract private key from block
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("LoadPrivateRSAKey: Error extracting private key from PEM block: " + err.Error())
	}

	return privateKey, nil
}

// LoadPublicRSAKey loads a rsa.PublicKey from a given .pem file
func LoadPublicRSAKey(path string) (*rsa.PublicKey, error) {
	// load bytes from file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("LoadPrivateRSAKey: Error loading key from file: " + err.Error())
	}

	// extract block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("LoadPrivateRSAKey: Error decoding PEM block.")
	}

	// extract public key from block
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("LoadPrivateRSAKey: Error extracting public key from PEM block: " + err.Error())
	}

	return publicKey, nil
}
