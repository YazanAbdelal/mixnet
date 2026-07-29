package keygen

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
)

// GenerateECCKeys generates random 32-byte (256 bits) ECC keys.
func GenerateECCKeys() (*ecdh.PrivateKey, *ecdh.PublicKey, error) {
	privateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, errors.New("GenerateECCKeys: Error generating private key using ecdh.X25519().GenerateKey: " + err.Error())
	}

	return privateKey, privateKey.PublicKey(), nil
}

// ExportPrivateECCKey writes the private ECC key to the provided file path.
func ExportPrivateECCKey(privateKey *ecdh.PrivateKey, dir string, filename string) error {
	// convert ecdh.PrivateKey to bytes
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return errors.New("ExportPrivateECCKey: Error converting private key to bytes using x509.MarshalPKCS8PrivateKey: " + err.Error())
	}

	// write PEM file content
	privateBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// create file to write the key to
	// create file to write the key to
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.New("ExportPrivateECCKey: Error creating directory: " + err.Error())
	}
	privateFile, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return errors.New("ExportPrivateECCKey: Error creating file: " + err.Error())
	}
	defer privateFile.Close()

	// write the contents to the new file using base64 encoding
	err = pem.Encode(privateFile, privateBlock)
	if err != nil {
		return errors.New("ExportPrivateECCKey: Error writing key to the file: " + err.Error())
	}

	return nil
}

// ExportPublicECCKey writes the public ECC key to the provided file path.
func ExportPublicECCKey(publicKey *ecdh.PublicKey, dir string, filename string) error {
	// convert ecdh.PublicKey to bytes
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return errors.New("ExportPublicECCKey: Error converting public key to bytes using x509.MarshalPKIXPublicKey: " + err.Error())
	}

	// write PEM file content
	publicBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// create file to write the key to
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.New("ExportPublicECCKey: Error creating directory: " + err.Error())
	}
	publicFile, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return errors.New("ExportPublicECCKey: Error creating file: " + err.Error())
	}
	defer publicFile.Close()

	// write the contents to the new file using base64 encoding
	err = pem.Encode(publicFile, publicBlock)
	if err != nil {
		return errors.New("ExportPublicECCKey: Error writing key to the file: " + err.Error())
	}

	return nil
}

// LoadPrivateECCKey loads a ecdh.PrivateKey from a given .pem file
func LoadPrivateECCKey(path string) (*ecdh.PrivateKey, error) {
	// load bytes from file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("LoadPrivateECCKey: Error loading ECC key from file: " + err.Error())
	}

	// extract block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("LoadPrivateECCKey: Error decoding PEM block.")
	}

	// extract private key from block
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("LoadPrivateECCKey: Error extracting private ECC key from PEM block: " + err.Error())
	}

	// convert to ecdh.PrivateKey struct using type assertion
	privateKey, ok := parsedKey.(*ecdh.PrivateKey)
	if !ok {
		return nil, errors.New("LoadPrivateECCKey: Decoded key is not a valid ECDH private key.")
	}

	return privateKey, nil
}

// LoadPublicECCKey loads a ecdh.PublicKey from a given .pem file
func LoadPublicECCKey(path string) (*ecdh.PublicKey, error) {
	// load bytes from file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("LoadPublicECCKey: Error loading key from file: " + err.Error())
	}

	// extract block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("LoadPublicECCKey: Error decoding PEM block.")
	}

	// extract public key from block
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("LoadPublicECCKey: Error extracting public ECC key from PEM block: " + err.Error())
	}

	// convert to ecdh.PublicKey struct using type assertion
	publicKey, ok := parsedKey.(*ecdh.PublicKey)
	if !ok {
		return nil, errors.New("LoadPublicECCKey: Decoded key is not a valid ECDH public key.")
	}

	return publicKey, nil
}
