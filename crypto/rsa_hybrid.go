package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
)

// EncryptWithRSA encrypts a slice of bytes with hybrid RSA encryption using the provided keys.
func EncryptWithRSA(unencryptedBytes []byte, rsaKey *rsa.PublicKey, aesKeySize int) ([]byte, error) {
	// generate AES key
	aesKey, err := genAESKey(aesKeySize)
	if err != nil {
		return nil, errors.New("EncryptWithRSA: Error generating AES key: " + err.Error())
	}

	// first we encrypt with AES
	sealedMessage, err := encryptWithAES(unencryptedBytes, aesKey)
	if err != nil {
		return nil, errors.New("EncryptWithRSA: Error encrypting with AES: " + err.Error())
	}

	// then we encrypt the key with OAEP
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, aesKey, nil)
	if err != nil {
		return nil, errors.New("EncryptWithRSA: Error encrypting AES key: " + err.Error())
	}

	return append(encryptedKey, sealedMessage...), nil
}

// DecryptWithRSA decrypts a slice of bytes that was encrypted with a hybrid RSA encryption using the provided keys.
func DecryptWithRSA(encryptedPacket []byte, rsaKey *rsa.PrivateKey) ([]byte, error) {
	blockSize := rsaKey.Size()
	encryptedKey, encryptedMessage := encryptedPacket[:blockSize], encryptedPacket[blockSize:]

	// decrypt AES key using RSA key
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, encryptedKey, nil)
	if err != nil {
		return nil, errors.New("DecryptWithRSA: Error decrypting AES key: " + err.Error())
	}

	// decrypt message bytes using the AES key
	decryptedBytes, err := decryptAES(encryptedMessage, aesKey)
	if err != nil {
		return nil, errors.New("DecryptWithRSA: Error decrypting message using AES key: " + err.Error())
	}

	return decryptedBytes, nil
}
