package keygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

const (
	MessageSize int = 8192
)

// GenerateKeys generates a pair of private and public RSA keys of the given size.
// Minimum size is 1024 bits. Use 4096 for good measure.
func GenerateKeys(numBits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// rand.Reader is a CSPRNG (Cryptographically secure pseudo-random number generator)
	privateKey, err := rsa.GenerateKey(rand.Reader, numBits)
	if err != nil {
		return nil, nil, errors.New("GenerateKeys: Error generating private key: " + err.Error())
	}

	return privateKey, &privateKey.PublicKey, nil
}

// ExportPrivateKey writes the private key to the provided file path.
func ExportPrivateKey(key *rsa.PrivateKey, dir string, filename string) error {
	// write PEM file content
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key), // converts PrivateKey struct into []byte using PKCS#1 structure
	}

	// create file to write the key to
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.New("ExportPrivateKey: Error creating directory: " + err.Error())
	}
	privateFile, err := os.Create(dir + "/" + filename)
	if err != nil {
		return errors.New("ExportPrivateKey: Error creating file: " + err.Error())
	}
	defer privateFile.Close()

	// write the contents to the new file using base64 encoding
	err = pem.Encode(privateFile, privateBlock)
	if err != nil {
		return errors.New("ExportPrivateKey: Error writing key to the file: " + err.Error())
	}

	return nil
}

// ExportPrivateKey writes the public key to the provided file path.
func ExportPublicKey(key *rsa.PublicKey, dir string, filename string) error {
	// write PEM file content
	privateBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key), // converts PrivateKey struct into []byte using PKCS#1 structure
	}

	// create file to write the key to
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.New("ExportPublicKey: Error creating directory: " + err.Error())
	}
	publicFile, err := os.Create(dir + "/" + filename)
	if err != nil {
		return errors.New("ExportPublicKey: Error creating file: " + err.Error())
	}
	defer publicFile.Close()

	// write the contents to the new file using base64 encoding
	err = pem.Encode(publicFile, privateBlock)
	if err != nil {
		return errors.New("ExportPublicKey: Error writing key to the file: " + err.Error())
	}

	return nil
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	// load bytes from file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("LoadPrivateKey: Error loading key from file: " + err.Error())
	}

	// extract block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("LoadPrivateKey: Error decoding PEM block.")
	}

	// extract private key from block
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("LoadPrivateKey: Error extracting private key from PEM block: " + err.Error())
	}

	return privateKey, nil
}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	// load bytes from file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("LoadPublicKey: Error loading key from file: " + err.Error())
	}

	// extract block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("LoadPublicKey: Error decoding PEM block.")
	}

	// extract public key from block
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("LoadPublicKey: Error extracting public key from PEM block: " + err.Error())
	}

	return publicKey, nil
}
