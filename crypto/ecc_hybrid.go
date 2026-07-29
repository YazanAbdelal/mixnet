package crypto

import (
	"crypto/ecdh"
	"errors"
)

// EncryptWithECC encrypts a slice of bytes using ECC with the provided ecdh.PrivateKey and ecdh.PublicKey.
// Both keys must use the same curve.
func EncryptWithECC(unencryptedBytes []byte, privateKey *ecdh.PrivateKey, publicKey *ecdh.PublicKey) ([]byte, error) {
	// get shared AES key
	aesKey, err := DeriveAESKey(privateKey, publicKey)
	if err != nil {
		return nil, errors.New("EncryptWithECC: " + err.Error())
	}

	// encrypt with the AES key
	encryptedBytes, err := encryptWithAES(unencryptedBytes, aesKey)
	if err != nil {
		return nil, errors.New("EncryptWithECC: " + err.Error())
	}

	return encryptedBytes, nil
}

// DecryptWithECC decrypts a slice of bytes that was encrypted using ECC with the provided ecdh.PrivateKey and
// ecdh.PublicKey.
// Both keys must use the same curve.
func DecryptWithECC(encryptedPacket []byte, privateKey *ecdh.PrivateKey, publicKey *ecdh.PublicKey) ([]byte, error) {
	// get shared AES key
	aesKey, err := DeriveAESKey(privateKey, publicKey)
	if err != nil {
		return nil, errors.New("DecryptWithECC: " + err.Error())
	}

	// decrypt packet using the AES key
	decryptedBytes, err := decryptAES(encryptedPacket, aesKey)
	if err != nil {
		return nil, errors.New("DecryptWithECC: " + err.Error())
	}

	return decryptedBytes, nil
}
