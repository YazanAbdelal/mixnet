package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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
		return nil, errors.New("genAESKey:Error generating AESm key: " + err.Error())
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
	io.ReadFull(rand.Reader, nonce)

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

	// unencrypt
	decryptedBytes, err := gcm.Open(nil, nonce, encrypteBytes, nil)
	if err != nil {
		return nil, errors.New("decryptAES: Error decrypting bytes: " + err.Error())
	}

	return decryptedBytes, nil
}

// EncryptMessage encrypts a slice of bytes with hybrid encryption using the provided keys.
func EncryptMessage(unencryptedBytes []byte, rsaKey *rsa.PublicKey, aesKeySize int) ([]byte, error) {
	// generate AES key
	aesKey, err := genAESKey(aesKeySize)
	if err != nil {
		return nil, errors.New("EncryptMessage: Error generating AES key: " + err.Error())
	}

	// first we encrypt with AES
	sealedMessage, err := encryptWithAES(unencryptedBytes, aesKey)
	if err != nil {
		return nil, errors.New("EncryptMessage: Error encrypting with AES: " + err.Error())
	}

	// then we encrypt the key with OAEP
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, aesKey, nil)
	if err != nil {
		return nil, errors.New("EncryptMessage: Error encrypting AES key: " + err.Error())
	}

	return append(encryptedKey, sealedMessage...), nil
}

// DecryptMessage decrypts a slice of bytes that was encrypted with a hybrid encryption using the provided keys.
func DecryptMessage(encryptedPacket []byte, rsaKey *rsa.PrivateKey) ([]byte, error) {
	blockSize := rsaKey.Size()
	encryptedKey, encryptedMessage := encryptedPacket[:blockSize], encryptedPacket[blockSize:]

	// decrypt AES key using RSA key
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, encryptedKey, nil)
	if err != nil {
		return nil, errors.New("DecryptMessage: Error decrypting AES key: " + err.Error())
	}

	// decrypt message bytes using the AES key
	decryptedBytes, err := decryptAES(encryptedMessage, aesKey)
	if err != nil {
		return nil, errors.New("DecryptMessage: Error decrypting message using AES key: " + err.Error())
	}

	return decryptedBytes, nil
}
