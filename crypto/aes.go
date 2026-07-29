package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// genAESKey generates an random key of the given size (in bytes).
func genAESKey(size int) ([]byte, error) {
	// make an empty slice of byte of the given size
	aesKey := make([]byte, size)

	// fill the slice with random bytes
	_, err := io.ReadFull(rand.Reader, aesKey)
	if err != nil {
		return nil, errors.New("genAESKey:Error generating AES key: " + err.Error())
	}

	return aesKey, nil
}

// encryptWithAES encrypts a slice of bytes using the provided AES key.
func encryptWithAES(unencryptedBytes []byte, aesKey []byte) ([]byte, error) {
	// generate block cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, errors.New("encryptWithAES: Error generating AES block: " + err.Error())
	}

	// generate GCM (for encryption + authentication)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("encryptWithAES: Error generating GCM: " + err.Error())
	}

	// generate random nonce to protect against repeat attacks
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, errors.New("encryptWithAES: Error generating random nonce: " + err.Error())
	}

	// encrypt and validate message
	return gcm.Seal(nonce, nonce, unencryptedBytes, nil), nil
}

// decryptAES decrypts a slice of bytes that was encrypted with AES using the provided key.
func decryptAES(sealedMsg []byte, aesKey []byte) ([]byte, error) {
	// generate block cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, errors.New("decryptAES: Error generating AES block: " + err.Error())
	}

	// generate GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("decryptAES: Error generating GCM: " + err.Error())
	}

	nonceSize := gcm.NonceSize()
	nonce, encrypteBytes := sealedMsg[:nonceSize], sealedMsg[nonceSize:]

	// decrypt
	decryptedBytes, err := gcm.Open(nil, nonce, encrypteBytes, nil)
	if err != nil {
		return nil, errors.New("decryptAES: Error decrypting bytes: " + err.Error())
	}

	return decryptedBytes, nil
}
