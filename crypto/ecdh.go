package crypto

import (
	"crypto/ecdh"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
)

// deriveSharedKey takes a private ECDH key and a public ECDH key and returns the shared secret using the Diffie-Hellman algorithm.
// Private and public keys must use the same curve.
func deriveSharedSecret(privateKey *ecdh.PrivateKey, publicKey *ecdh.PublicKey) ([]byte, error) {
	sharedSecret, err := privateKey.ECDH(publicKey)
	if err != nil {
		return nil, errors.New("deriveSharedKey: Error deriving shared key: " + err.Error())
	}

	return sharedSecret, nil
}

// DeriveAESKey takes a private ECDH key and a public ECDH key and returns a 32-Byte AES key that is derived from
// the ECDH shared secret.
// Private and public keys must use the same curve.
func DeriveAESKey(privateKey *ecdh.PrivateKey, publicKey *ecdh.PublicKey) ([]byte, error) {
	// get shared secret using DH
	sharedSecret, err := deriveSharedSecret(privateKey, publicKey)
	if err != nil {
		return nil, errors.New("DeriveAESKey: " + err.Error())
	}

	// derive uniformly random AES key from shared secret
	key := make([]byte, 32)
	h := hkdf.New(sha256.New, sharedSecret, nil, []byte("mixnet-layer"))
	io.ReadFull(h, key)

	return key, nil
}
